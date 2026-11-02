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
	"context"
	"errors"
	"net/http"

	"github.com/SENERGY-Platform/device-repository/v3/lib/database"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func (this *Controller) setAspectNodes(ctx context.Context, aspect models.Aspect) (err error) {
	ctx, _ = getTimeoutContext(ctx)
	err = this.db.RemoveAspectNodesByRootId(ctx, aspect.Id)
	if err != nil {
		return err
	}
	_, err = CreateAspectNodes(ctx, this.db, aspect, aspect.Id, "", []string{})
	return err
}

func CreateAspectNodes(ctx context.Context, db database.Database, aspect models.Aspect, rootId string, parentId string, ancestors []string) (descendents []string, err error) {
	descendents = []string{}
	children := []string{}
	for _, sub := range aspect.SubAspects {
		children = append(children, sub.Id)
		temp, err := CreateAspectNodes(ctx, db, sub, rootId, aspect.Id, append(ancestors, aspect.Id))
		if err != nil {
			return descendents, err
		}
		descendents = append(descendents, temp...)
	}
	ctx, _ = getTimeoutContext(ctx)
	err = db.SetAspectNode(ctx, models.AspectNode{
		Id:            aspect.Id,
		Name:          aspect.Name,
		AspectClassId: aspect.AspectClassId,
		RootId:        rootId,
		ParentId:      parentId,
		ChildIds:      children,
		AncestorIds:   ancestors,
		DescendentIds: descendents,
	})
	return append(descendents, aspect.Id), err
}

func (this *Controller) ListAspectNodes(ctx context.Context, listOptions model.AspectListOptions) (result []models.AspectNode, total int64, err error, errCode int) {
	ctx, _ = getTimeoutContext(ctx)
	result, total, err = this.db.ListAspectNodes(ctx, listOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) GetAspectNode(ctx context.Context, id string) (result models.AspectNode, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	result, exists, err := this.db.GetAspectNode(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) GetAspectNodes(ctx context.Context) (result []models.AspectNode, err error, code int) {
	code = http.StatusOK
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.ListAllAspectNodes(ctx)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetAspectNodesWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) (result []models.AspectNode, err error, code int) {
	code = http.StatusOK
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.ListAspectNodesWithMeasuringFunction(ctx, ancestors, descendants)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetAspectNodesWithControllingFunction(ctx context.Context, ancestors bool, descendants bool) (result []models.AspectNode, err error, code int) {
	code = http.StatusOK
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.ListAspectNodesWithControllingFunction(ctx, ancestors, descendants)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetAspectNodesByIdList(ctx context.Context, ids []string) (result []models.AspectNode, err error, code int) {
	code = http.StatusOK
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.ListAspectNodesByIdList(ctx, ids)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}
