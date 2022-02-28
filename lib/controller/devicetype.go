/*
 * Copyright 2019 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"log"
	"net/http"
	"sort"
	"time"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadDeviceType(id string, token string) (result model.DeviceType, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	deviceType, exists, err := this.db.GetDeviceType(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return deviceType, nil, http.StatusOK
}

func (this *Controller) ListDeviceTypes(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string) (result []model.DeviceType, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListDeviceTypes(ctx, limit, offset, sort, filter, interactionsFilter)
	return
}

func (this *Controller) ValidateDeviceType(dt model.DeviceType) (err error, code int) {
	if dt.Id == "" {
		return errors.New("missing device-type id"), http.StatusBadRequest
	}
	if dt.Name == "" {
		return errors.New("missing device-type name"), http.StatusBadRequest
	}
	if len(dt.Services) == 0 {
		return errors.New("expect at least one service"), http.StatusBadRequest
	}
	for _, service := range dt.Services {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		deviceTypes, err := this.db.GetDeviceTypesByServiceId(ctx, service.Id)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if len(deviceTypes) > 1 {
			return errors.New("reused service id"), http.StatusBadRequest
		}
		if len(deviceTypes) == 1 && deviceTypes[0].Id != dt.Id {
			return errors.New("reused service id"), http.StatusBadRequest
		}
		err, code = this.ValidateService(service)
		if err != nil {
			return err, code
		}
	}
	err = ValidateServiceGroups(dt.ServiceGroups, dt.Services)
	if err != nil {
		return err, http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func ValidateServiceGroups(groups []model.ServiceGroup, services []model.Service) error {
	groupIndex := map[string]bool{}
	for _, g := range groups {
		if _, ok := groupIndex[g.Key]; ok {
			return errors.New("duplicate service-group key: " + g.Key)
		}
		groupIndex[g.Key] = true
	}
	for _, s := range services {
		if s.ServiceGroupKey != "" {
			_, ok := groupIndex[s.ServiceGroupKey]
			if !ok {
				return errors.New("unknown service-group key: " + s.ServiceGroupKey)
			}
		}
	}
	return nil
}

func (this *Controller) GetDeviceTypeSelectables(query []model.FilterCriteria, pathPrefix string, interactionsFilter []string) (result []model.DeviceTypeSelectable, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.getDeviceTypeSelectables(ctx, query, pathPrefix, interactionsFilter)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) getDeviceTypeSelectables(ctx context.Context, query []model.FilterCriteria, pathPrefix string, interactionsFilter []string) (result []model.DeviceTypeSelectable, err error) {
	result = []model.DeviceTypeSelectable{}
	deviceTypes, err := this.db.GetDeviceTypeIdsByFilterCriteria(ctx, query, interactionsFilter)
	if err != nil {
		return result, err
	}
	groupByDeviceType := map[string][]model.DeviceTypeCriteria{}
	for _, criteria := range query {
		dtCriteria, err := this.db.GetDeviceTypeCriteriaForDeviceTypeIdsAndFilterCriteria(ctx, deviceTypes, criteria)
		if err != nil {
			return result, err
		}
		for _, element := range dtCriteria {
			groupByDeviceType[element.DeviceTypeId] = append(groupByDeviceType[element.DeviceTypeId], element)
		}
	}
	aspectCache := &map[string]model.AspectNode{}
	for dtId, dtCriteria := range groupByDeviceType {
		dt, _, err := this.db.GetDeviceType(ctx, dtId)
		if err != nil {
			return result, err
		}
		element := model.DeviceTypeSelectable{
			DeviceTypeId:       dtId,
			Services:           []model.Service{},
			ServicePathOptions: map[string][]model.ServicePathOption{},
		}
		for _, criteria := range dtCriteria {
			aspectNode, err := this.getAspectNodeForDeviceTypeSelectables(aspectCache, criteria.AspectId)
			if err != nil {
				return result, err
			}
			element.ServicePathOptions[criteria.ServiceId] = append(element.ServicePathOptions[criteria.ServiceId], model.ServicePathOption{
				ServiceId:        criteria.ServiceId,
				Path:             pathPrefix + criteria.ContentVariablePath,
				CharacteristicId: criteria.CharacteristicId,
				AspectNode:       aspectNode,
				FunctionId:       criteria.FunctionId,
				IsVoid:           criteria.IsVoid,
			})
		}
		for sid, options := range element.ServicePathOptions {
			for _, service := range dt.Services {
				if service.Id == sid {
					element.Services = append(element.Services, service)
				}
			}
			sort.Slice(options, func(i, j int) bool {
				return options[i].Path < options[j].Path
			})
			element.ServicePathOptions[sid] = options
		}
		result = append(result, element)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].DeviceTypeId < result[j].DeviceTypeId
	})
	return result, nil
}

func (this *Controller) getAspectNodeForDeviceTypeSelectables(aspectCache *map[string]model.AspectNode, aspectId string) (aspectNode model.AspectNode, err error) {
	if aspectId == "" {
		return aspectNode, nil
	}
	var ok bool
	aspectNode, ok = (*aspectCache)[aspectId]
	if !ok {
		ctx, _ := getTimeoutContext()
		aspectNode, ok, err = this.db.GetAspectNode(ctx, aspectId)
		if err != nil {
			log.Println("WARNING: unable to load aspect node", aspectId, err)
			return aspectNode, err
		}
		if !ok {
			log.Println("WARNING: unknown aspect node", aspectId)
			return aspectNode, err
		}
		(*aspectCache)[aspectId] = aspectNode
	}
	return aspectNode, nil
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDeviceType(deviceType model.DeviceType, owner string) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.SetDeviceType(ctx, deviceType)
}

func (this *Controller) DeleteDeviceType(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveDeviceType(ctx, id)
}
