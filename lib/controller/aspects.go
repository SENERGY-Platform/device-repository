/*
 * Copyright 2022 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"net/http"
	"strings"
)

func (this *Controller) SetAspect(aspect model.Aspect, owner string) error {
	ctx, _ := getTimeoutContext()
	err := this.db.SetAspect(ctx, aspect)
	if err != nil {
		return err
	}
	descendentNodeIds := getDescendentNodeIds(aspect)
	err = this.handleMovedSubAspects(aspect, descendentNodeIds, owner)
	if err != nil {
		return err
	}
	return this.setAspectNodes(aspect)
}

func getDescendentNodeIds(aspect model.Aspect) (result []string) {
	result = []string{aspect.Id}
	for _, sub := range aspect.SubAspects {
		result = append(result, getDescendentNodeIds(sub)...)
	}
	return result
}

func (this *Controller) handleMovedSubAspects(aspect model.Aspect, descendentNodesIds []string, owner string) error {
	ctx, _ := getTimeoutContext()
	nodes, err := this.db.ListAspectNodesByIdList(ctx, descendentNodesIds)
	if err != nil {
		return err
	}
	movedFrom := map[string][]string{}
	deletedAspect := map[string]bool{}
	for _, node := range nodes {
		//is moved aspect
		if node.RootId != aspect.Id {
			movedFrom[node.RootId] = append(movedFrom[node.RootId], node.Id)
		}
		//sub aspect is moved root aspect
		if node.Id == node.RootId && node.Id != aspect.Id {
			deletedAspect[node.Id] = true
			err = this.producer.PublishAspectDelete(node.Id, owner)
			if err != nil {
				return err
			}
		}
	}
	for rootId, movedIds := range movedFrom {
		if !deletedAspect[rootId] {
			sourceAspect, exists, err := this.db.GetAspect(ctx, rootId)
			if err != nil {
				return err
			}
			if exists {
				changedAspect := filterSubAspects(sourceAspect, movedIds)
				err = this.producer.PublishAspectUpdate(changedAspect, owner)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func filterSubAspects(aspect model.Aspect, ids []string) model.Aspect {
	filterSubAspect := map[string]bool{}
	for _, id := range ids {
		filterSubAspect[id] = true
	}
	subAspects := []model.Aspect{}
	for _, sub := range aspect.SubAspects {
		if !filterSubAspect[sub.Id] {
			subAspects = append(subAspects, filterSubAspects(sub, ids))
		}
	}
	aspect.SubAspects = subAspects
	return aspect
}

func (this *Controller) DeleteAspect(id string) error {
	ctx, _ := getTimeoutContext()
	err := this.db.RemoveAspectNodesByRootId(ctx, id)
	if err != nil {
		return err
	}
	return this.db.RemoveAspect(ctx, id)
}

func (this *Controller) GetAspects() (result []model.Aspect, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllAspects(ctx)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetAspect(id string) (result model.Aspect, err error, code int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetAspect(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) GetAspectsWithMeasuringFunction(ancestors bool, descendants bool) (result []model.Aspect, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAspectsWithMeasuringFunction(ctx, ancestors, descendants)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) ValidateAspect(aspect model.Aspect) (err error, code int) {
	if aspect.Id == "" {
		return errors.New("missing aspect id"), http.StatusBadRequest
	}
	if !strings.HasPrefix(aspect.Id, model.URN_PREFIX) {
		return errors.New("invalid aspect id"), http.StatusBadRequest
	}
	if aspect.Name == "" {
		return errors.New("missing aspect name"), http.StatusBadRequest
	}
	for _, sub := range aspect.SubAspects {
		err, code = this.ValidateAspect(sub)
		if err != nil {
			return err, code
		}
	}

	//check for deleted sub aspects
	ctx, _ := getTimeoutContext()
	old, exists, err := this.db.GetAspectNode(ctx, aspect.Id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return nil, http.StatusOK
	}
	newDescendentNodesIds := getDescendentNodeIds(aspect)
	newDescendentsSet := map[string]bool{}
	for _, id := range newDescendentNodesIds {
		newDescendentsSet[id] = true
	}
	for _, id := range old.DescendentIds {
		if !newDescendentsSet[id] {
			isUsed, where, err := this.db.AspectIsUsed(ctx, id)
			if err != nil {
				return err, http.StatusInternalServerError
			}
			if isUsed {
				return errors.New("sub aspect " + id + " is still in use: " + strings.Join(where, ",")), http.StatusBadRequest
			}
		}
	}

	return nil, http.StatusOK
}

func (this *Controller) ValidateAspectDelete(id string) (err error, code int) {
	ctx, _ := getTimeoutContext()
	aspect, exists, err := this.db.GetAspectNode(ctx, id)
	if !exists {
		//deleting nothing is ok
		return nil, http.StatusOK
	}
	if err != nil {
		return err, http.StatusInternalServerError
	}
	isUsed, where, err := this.db.AspectIsUsed(ctx, id)
	if isUsed {
		return errors.New("still in use: " + strings.Join(where, ",")), http.StatusBadRequest
	}
	for _, sub := range aspect.DescendentIds {
		isUsed, where, err = this.db.AspectIsUsed(ctx, sub)
		if isUsed {
			return errors.New("sub aspect " + sub + " is still in use: " + strings.Join(where, ",")), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}
