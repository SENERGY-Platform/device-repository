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
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"net/http"
	"strings"
)

func (this *Controller) setAspectSyncHandler(aspect models.Aspect) (err error) {
	descendentNodeIds := getDescendentNodeIds(aspect)
	err = this.handleMovedSubAspects(aspect, descendentNodeIds)
	if err != nil {
		return err
	}
	err = this.setAspectNodes(aspect)
	if err != nil {
		return err
	}
	err = this.publisher.PublishAspect(aspect)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) setAspect(aspect models.Aspect) error {
	ctx, _ := getTimeoutContext()
	err := this.db.SetAspect(ctx, aspect, this.setAspectSyncHandler)
	if err != nil {
		return err
	}
	return nil
}

func getDescendentNodeIds(aspect models.Aspect) (result []string) {
	result = []string{aspect.Id}
	for _, sub := range aspect.SubAspects {
		result = append(result, getDescendentNodeIds(sub)...)
	}
	return result
}

func (this *Controller) handleMovedSubAspects(aspect models.Aspect, descendentNodesIds []string) error {
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
			err = this.deleteAspect(node.Id)
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
				err = this.setAspect(changedAspect)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func filterSubAspects(aspect models.Aspect, ids []string) models.Aspect {
	filterSubAspect := map[string]bool{}
	for _, id := range ids {
		filterSubAspect[id] = true
	}
	subAspects := []models.Aspect{}
	for _, sub := range aspect.SubAspects {
		if !filterSubAspect[sub.Id] {
			subAspects = append(subAspects, filterSubAspects(sub, ids))
		}
	}
	aspect.SubAspects = subAspects
	return aspect
}

func (this *Controller) SetAspect(token string, aspect models.Aspect) (result models.Aspect, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return result, errors.New("token is not an admin"), http.StatusUnauthorized
	}

	//ensure ids
	aspect.GenerateId()

	err, code = this.ValidateAspect(aspect)
	if err != nil {
		return aspect, err, code
	}
	err = this.setAspect(aspect)
	if err != nil {
		return aspect, err, http.StatusInternalServerError
	}
	return aspect, nil, http.StatusOK
}

func (this *Controller) DeleteAspect(token string, id string) (err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return errors.New("token is not an admin"), http.StatusUnauthorized
	}
	err, code = this.ValidateAspectDelete(id)
	if err != nil {
		return err, code
	}
	err = this.deleteAspect(id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteAspectSyncHandler(aspect models.Aspect) (err error) {
	ctx, _ := getTimeoutContext()
	err = this.db.RemoveAspectNodesByRootId(ctx, aspect.Id)
	if err != nil {
		return err
	}
	err = this.publisher.PublishAspectDelete(aspect.Id)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) deleteAspect(id string) (err error) {
	ctx, _ := getTimeoutContext()
	err = this.db.RemoveAspect(ctx, id, this.deleteAspectSyncHandler)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) GetAspects() (result []models.Aspect, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllAspects(ctx)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) ListAspects(listOptions model.AspectListOptions) (result []models.Aspect, total int64, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListAspects(ctx, listOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) GetAspect(id string) (result models.Aspect, err error, code int) {
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

func (this *Controller) GetAspectsWithMeasuringFunction(ancestors bool, descendants bool) (result []models.Aspect, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAspectsWithMeasuringFunction(ctx, ancestors, descendants)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) ValidateAspect(aspect models.Aspect) (err error, code int) {
	return this.validateAspect(aspect, true)
}

func (this *Controller) validateAspect(aspect models.Aspect, checkDelete bool) (err error, code int) {
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
		err, code = this.validateAspect(sub, false)
		if err != nil {
			return err, code
		}
	}

	//check for deleted sub aspects; but only for the root aspect, to prevent errors when moving sub aspect
	if checkDelete {
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
