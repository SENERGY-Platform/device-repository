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
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadDeviceType(id string, token string) (result models.DeviceType, err error, errCode int) {
	return this.readDeviceType(id)
}

func (this *Controller) readDeviceType(id string) (result models.DeviceType, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	pureId, modifier := idmodifier.SplitModifier(id)
	deviceType, exists, err := this.db.GetDeviceType(ctx, pureId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	deviceType.Id = id
	if modifier != nil && len(modifier) > 0 {
		deviceType, err, errCode = this.modifyDeviceType(deviceType, modifier)
		if err != nil {
			return result, err, errCode
		}
	}
	return deviceType, nil, http.StatusOK
}

func (this *Controller) ListDeviceTypes(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	temp, err := this.db.ListDeviceTypes(ctx, limit, offset, sort, filter, interactionsFilter, includeModified)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return this.modifyDeviceTypeList(temp, sort, includeModified, includeUnmodified)
}

func (this *Controller) ListDeviceTypesV2(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	temp, err := this.db.ListDeviceTypesV2(ctx, limit, offset, sort, filter, includeModified)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return this.modifyDeviceTypeList(temp, sort, includeModified, includeUnmodified)
}

func (this *Controller) ListDeviceTypesV3(token string, listOptions model.DeviceTypeListOptions) (result []models.DeviceType, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	temp, err := this.db.ListDeviceTypesV3(ctx, listOptions)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return this.modifyDeviceTypeList(temp, listOptions.SortBy, listOptions.IncludeModified, !listOptions.IgnoreUnmodified)
}

func (this *Controller) ValidateDeviceType(dt models.DeviceType, options model.ValidationOptions) (err error, code int) {
	if dt.Id == "" {
		return errors.New("missing device-type id"), http.StatusBadRequest
	}
	if dt.Name == "" {
		return errors.New("missing device-type name"), http.StatusBadRequest
	}
	if len(dt.Services) == 0 {
		return errors.New("expect at least one service"), http.StatusBadRequest
	}
	protocolCache := &map[string]models.Protocol{}
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
		err, code = this.ValidateService(service, protocolCache, options)
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

func ValidateServiceGroups(groups []models.ServiceGroup, services []models.Service) error {
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

func (this *Controller) GetDeviceTypeSelectables(query []model.FilterCriteria, pathPrefix string, interactionsFilter []string, includeModified bool) (result []model.DeviceTypeSelectable, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.getDeviceTypeSelectables(ctx, query, pathPrefix, interactionsFilter, includeModified)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetDeviceTypeSelectablesV2(query []model.FilterCriteria, pathPrefix string, includeModified bool, servicesMustMatchAllCriteria bool) (result []model.DeviceTypeSelectable, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.getDeviceTypeSelectablesV2(ctx, query, pathPrefix, includeModified, servicesMustMatchAllCriteria)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) getDeviceTypeSelectables(ctx context.Context, query []model.FilterCriteria, pathPrefix string, interactionsFilter []string, includeModified bool) (result []model.DeviceTypeSelectable, err error) {
	if len(query) == 0 {
		query = append(query, model.FilterCriteria{})
	}

	result = []model.DeviceTypeSelectable{}

	//EVENT|REQUEST should also find EVENT_AND_REQUEST
	if (contains(interactionsFilter, string(models.EVENT)) || contains(interactionsFilter, string(models.REQUEST))) && !contains(interactionsFilter, string(models.EVENT_AND_REQUEST)) {
		interactionsFilter = append(interactionsFilter, string(models.EVENT_AND_REQUEST))
	}

	deviceTypes, err := this.db.GetDeviceTypeIdsByFilterCriteria(ctx, query, interactionsFilter, includeModified)
	if err != nil {
		return result, err
	}
	groupByDeviceType := map[string][]model.DeviceTypeCriteria{}
	for _, criteria := range query {
		dtCriteria, err := this.db.GetDeviceTypeCriteriaForDeviceTypeIdsAndFilterCriteria(ctx, deviceTypes, criteria, includeModified)
		if err != nil {
			return result, err
		}
		for _, element := range dtCriteria {
			groupByDeviceType[element.DeviceTypeId] = append(groupByDeviceType[element.DeviceTypeId], element)
		}
	}
	aspectCache := &map[string]models.AspectNode{}
	for dtId, dtCriteria := range groupByDeviceType {
		dt, err, _ := this.readDeviceType(dtId)
		if err != nil {
			return result, err
		}
		element := model.DeviceTypeSelectable{
			DeviceTypeId:       dtId,
			Services:           []models.Service{},
			ServicePathOptions: map[string][]model.ServicePathOption{},
		}
		for _, criteria := range dtCriteria {
			aspectNode, err := this.getAspectNodeForDeviceTypeSelectables(aspectCache, criteria.AspectId)
			if err != nil {
				return result, err
			}
			element.ServicePathOptions[criteria.ServiceId] = append(element.ServicePathOptions[criteria.ServiceId], model.ServicePathOption{
				ServiceId:             criteria.ServiceId,
				Path:                  pathPrefix + criteria.ContentVariablePath,
				CharacteristicId:      criteria.CharacteristicId,
				AspectNode:            aspectNode,
				FunctionId:            criteria.FunctionId,
				IsVoid:                criteria.IsVoid,
				Value:                 criteria.Value,
				Type:                  criteria.Type,
				IsControllingFunction: criteria.IsControllingFunction,
			})
		}
		for sid, options := range element.ServicePathOptions {
			configurablesCandidates, err := this.db.GetConfigurableCandidates(ctx, sid)
			if err != nil {
				return result, err
			}
			for i, option := range options {
				options[i].Configurables, err = this.getConfigurables(configurablesCandidates, option)
				if err != nil {
					return result, err
				}
			}
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

func (this *Controller) getDeviceTypeSelectablesV2(ctx context.Context, query []model.FilterCriteria, pathPrefix string, includeModified bool, servicesMustMatchAllCriteria bool) (result []model.DeviceTypeSelectable, err error) {
	if len(query) == 0 {
		query = append(query, model.FilterCriteria{})
	}

	result = []model.DeviceTypeSelectable{}

	deviceTypes, err := this.db.GetDeviceTypeIdsByFilterCriteriaV2(ctx, query, includeModified)
	if err != nil {
		return result, err
	}
	groupByDeviceType := map[string][]model.DeviceTypeCriteria{}
	serviceCriteriaCount := map[string]int{}
	for _, criteria := range query {
		dtCriteria, err := this.db.GetDeviceTypeCriteriaForDeviceTypeIdsAndFilterCriteria(ctx, deviceTypes, criteria, includeModified)
		if err != nil {
			return result, err
		}
		serviceIndex := map[string]bool{}
		for _, element := range dtCriteria {
			groupByDeviceType[element.DeviceTypeId] = append(groupByDeviceType[element.DeviceTypeId], element)
			serviceIndex[element.ServiceId] = true
		}
		for sid, _ := range serviceIndex {
			serviceCriteriaCount[sid] = serviceCriteriaCount[sid] + 1
		}
	}
	validService := map[string]bool{}
	for sid, count := range serviceCriteriaCount {
		if count == len(query) {
			validService[sid] = true
		}
	}

	aspectCache := &map[string]models.AspectNode{}
	for dtId, dtCriteria := range groupByDeviceType {
		dt, err, _ := this.readDeviceType(dtId)
		if err != nil {
			return result, err
		}
		element := model.DeviceTypeSelectable{
			DeviceTypeId:       dtId,
			Services:           []models.Service{},
			ServicePathOptions: map[string][]model.ServicePathOption{},
		}
		usedPaths := map[string]map[string]bool{}
		for _, criteria := range dtCriteria {
			if !servicesMustMatchAllCriteria || validService[criteria.ServiceId] {
				aspectNode, err := this.getAspectNodeForDeviceTypeSelectables(aspectCache, criteria.AspectId)
				if err != nil {
					return result, err
				}
				if _, ok := usedPaths[criteria.ServiceId]; !ok {
					usedPaths[criteria.ServiceId] = map[string]bool{}
				}
				if !usedPaths[criteria.ServiceId][pathPrefix+criteria.ContentVariablePath] {
					usedPaths[criteria.ServiceId][pathPrefix+criteria.ContentVariablePath] = true
					element.ServicePathOptions[criteria.ServiceId] = append(element.ServicePathOptions[criteria.ServiceId], model.ServicePathOption{
						ServiceId:             criteria.ServiceId,
						Path:                  pathPrefix + criteria.ContentVariablePath,
						CharacteristicId:      criteria.CharacteristicId,
						AspectNode:            aspectNode,
						FunctionId:            criteria.FunctionId,
						IsVoid:                criteria.IsVoid,
						Value:                 criteria.Value,
						Type:                  criteria.Type,
						IsControllingFunction: criteria.IsControllingFunction,
						Interaction:           models.Interaction(criteria.Interaction),
					})
				}
			}
		}
		for sid, options := range element.ServicePathOptions {
			configurablesCandidates, err := this.db.GetConfigurableCandidates(ctx, sid)
			if err != nil {
				return result, err
			}
			for i, option := range options {
				options[i].Configurables, err = this.getConfigurables(configurablesCandidates, option)
				if err != nil {
					return result, err
				}
			}
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (this *Controller) getAspectNodeForDeviceTypeSelectables(aspectCache *map[string]models.AspectNode, aspectId string) (aspectNode models.AspectNode, err error) {
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

func (this *Controller) getConfigurables(candidates []model.DeviceTypeCriteria, pathOption model.ServicePathOption) (result []model.Configurable, err error) {
	for _, candidate := range candidates {
		aspectNode := models.AspectNode{}
		if candidate.AspectId != "" {
			aspectNode, err, _ = this.GetAspectNode(candidate.AspectId)
			if err != nil {
				return result, err
			}
		}
		if !pathOption.IsControllingFunction || !pathOptionIsAncestorOfConfigurableCandidate(pathOption, candidate) {
			result = append(result, model.Configurable{
				Path:             candidate.ContentVariablePath,
				CharacteristicId: candidate.CharacteristicId,
				AspectNode:       aspectNode,
				FunctionId:       candidate.FunctionId,
				Type:             candidate.Type,
				Value:            candidate.Value,
			})
		}
	}
	return result, nil
}

func pathOptionIsAncestorOfConfigurableCandidate(option model.ServicePathOption, candidate model.DeviceTypeCriteria) bool {
	if candidate.ContentVariablePath == option.Path {
		return true
	}
	if strings.HasPrefix(candidate.ContentVariablePath, option.Path+".") {
		return true
	}
	return false
}

func (this *Controller) GetUsedInDeviceType(query model.UsedInDeviceTypeQuery) (result model.UsedInDeviceTypeResponse, err error, errCode int) {
	result = model.UsedInDeviceTypeResponse{}
	if query.Resource == "" {
		return result, errors.New("unknown resource"), http.StatusBadRequest
	}
	if query.With == "" {
		query.With = "device-type"
	}
	if query.CountBy == "" {
		query.CountBy = "device-type"
	}

	ctx, _ := getTimeoutContext()

	criteria := []model.DeviceTypeCriteria{}

	switch query.Resource {
	case "aspects":
		criteria, err = this.db.GetDeviceTypeCriteriaByAspectIds(ctx, query.Ids, false)
	case "device-classes":
		criteria, err = this.db.GetDeviceTypeCriteriaByDeviceClassIds(ctx, query.Ids, false)
	case "functions":
		criteria, err = this.db.GetDeviceTypeCriteriaByFunctionIds(ctx, query.Ids, false)
	case "characteristics":
		criteria, err = this.db.GetDeviceTypeCriteriaByCharacteristicIds(ctx, query.Ids, false)
	default:
		return result, fmt.Errorf("unknown resource=\"%v\"", query.Resource), http.StatusBadRequest
	}

	for _, id := range query.Ids {
		result[id] = model.RefInDeviceTypeResponseElement{UsedIn: []model.DeviceTypeReference{}}
	}

	//collect ids
	for _, c := range criteria {
		id := ""
		switch query.Resource {
		case "aspects":
			id = c.AspectId
		case "device-classes":
			id = c.DeviceClassId
		case "functions":
			id = c.FunctionId
		case "characteristics":
			id = c.CharacteristicId
		default:
			return result, fmt.Errorf("unknown resource=\"%v\"", query.Resource), http.StatusBadRequest
		}
		temp := result[id]
		temp.UsedIn = addDeviceTypeCriteriaToDeviceTypeRefs(temp.UsedIn, c)
		result[id] = temp
	}

	//collect count
	for key, value := range result {
		switch query.CountBy {
		case "device-type":
			value.Count = len(value.UsedIn)
		case "service":
			for _, dt := range value.UsedIn {
				value.Count = value.Count + len(dt.UsedIn)
			}
		case "variable":
			for _, dt := range value.UsedIn {
				for _, s := range dt.UsedIn {
					value.Count = value.Count + len(s.UsedIn)
				}
			}
		default:
			return result, fmt.Errorf("unknown count_by=\"%v\"", query.CountBy), http.StatusBadRequest
		}
		result[key] = value
	}

	//remove unwanted used-in lists
	for key, value := range result {
		switch query.With {
		case "device-type":
			for i, dtRef := range value.UsedIn {
				dtRef.UsedIn = []model.ServiceReference{}
				value.UsedIn[i] = dtRef
			}
		case "service":
			for i, dtRef := range value.UsedIn {
				for j, sRef := range dtRef.UsedIn {
					sRef.UsedIn = []model.VariableReference{}
					dtRef.UsedIn[j] = sRef
				}
				value.UsedIn[i] = dtRef
			}
		case "variable":
			//nothing to do here
		default:
			return result, fmt.Errorf("unknown with=\"%v\"", query.With), http.StatusBadRequest
		}
		result[key] = value
	}

	//add names
	for key, value := range result {
		for i, dtRef := range value.UsedIn {
			dt, _, err := this.db.GetDeviceType(ctx, dtRef.Id)
			if err != nil {
				return result, err, http.StatusInternalServerError
			}
			dtRef.Name = dt.Name
			for j, sRef := range dtRef.UsedIn {
				for _, s := range dt.Services {
					if s.Id == sRef.Id {
						sRef.Name = s.Name
						break
					}
				}
				dtRef.UsedIn[j] = sRef
			}
			value.UsedIn[i] = dtRef
		}
		result[key] = value
	}

	return result, nil, 200
}

func addDeviceTypeCriteriaToDeviceTypeRefs(list []model.DeviceTypeReference, criteria model.DeviceTypeCriteria) []model.DeviceTypeReference {
	found := false
	for i, e := range list {
		if e.Id == criteria.DeviceTypeId {
			e.UsedIn = addDeviceTypeCriteriaToServiceRefs(e.UsedIn, criteria)
			list[i] = e
			found = true
		}
	}
	if !found {
		list = append(list, model.DeviceTypeReference{
			Id:     criteria.DeviceTypeId,
			UsedIn: addDeviceTypeCriteriaToServiceRefs([]model.ServiceReference{}, criteria),
		})
	}
	return list
}

func addDeviceTypeCriteriaToServiceRefs(list []model.ServiceReference, criteria model.DeviceTypeCriteria) []model.ServiceReference {
	found := false
	for i, e := range list {
		if e.Id == criteria.ServiceId {
			e.UsedIn = addDeviceTypeCriteriaToVariableRefs(e.UsedIn, criteria)
			list[i] = e
			found = true
		}
	}
	if !found {
		list = append(list, model.ServiceReference{
			Id:     criteria.ServiceId,
			UsedIn: addDeviceTypeCriteriaToVariableRefs([]model.VariableReference{}, criteria),
		})
	}
	return list
}

func addDeviceTypeCriteriaToVariableRefs(list []model.VariableReference, criteria model.DeviceTypeCriteria) []model.VariableReference {
	list = append(list, model.VariableReference{
		Id:   criteria.ContentVariableId,
		Path: criteria.ContentVariablePath,
	})
	return list
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDeviceType(deviceType models.DeviceType, owner string) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.SetDeviceType(ctx, deviceType)
}

func (this *Controller) DeleteDeviceType(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveDeviceType(ctx, id)
}
