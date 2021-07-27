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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"log"
	"net/http"
)

/////////////////////////
//		api
/////////////////////////

const FilterDevicesOfGroupByAccess = true

func (this *Controller) ReadDeviceGroup(id string, token string) (result model.DeviceGroup, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetDeviceGroup(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err := this.security.CheckBool(token, this.config.DeviceGroupTopic, id, model.READ)
	if err != nil {
		result = model.DeviceGroup{}
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		result = model.DeviceGroup{}
		return result, errors.New("access denied"), http.StatusForbidden
	}
	if FilterDevicesOfGroupByAccess {
		return this.FilterDevicesOfGroupByAccess(token, result)
	} else {
		return result, nil, http.StatusOK
	}
}

func (this *Controller) FilterDevicesOfGroupByAccess(token string, group model.DeviceGroup) (result model.DeviceGroup, err error, code int) {
	if len(group.DeviceIds) == 0 {
		return group, nil, http.StatusOK
	}
	access, err := this.security.CheckMultiple(token, this.config.DeviceTopic, group.DeviceIds, model.EXECUTE)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	result = group
	result.DeviceIds = []string{}
	for _, id := range group.DeviceIds {
		if access[id] {
			result.DeviceIds = append(result.DeviceIds, id)
		} else if this.config.Debug {
			log.Println("DEBUG: filtered " + id + " from result, because user lost execution access to the device")
		}
	}
	return result, nil, http.StatusOK
}

//only the first element of group.Devices is checked.
//this should be enough because every used device should be referenced in each element of group.Devices
//use ValidateDeviceGroup() to ensure that this constraint is adhered to
func (this *Controller) CheckAccessToDevicesOfGroup(token string, group model.DeviceGroup) (err error, code int) {
	if len(group.DeviceIds) == 0 {
		return nil, http.StatusOK
	}
	access, err := this.security.CheckMultiple(token, this.config.DeviceTopic, group.DeviceIds, model.EXECUTE)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	//looping one element of group.Devices is enough because ValidateDeviceGroup() ensures that every used device is referenced in each group.Devices element
	for _, id := range group.DeviceIds {
		if !access[id] {
			return errors.New("no execution access to device " + id), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) ValidateDeviceGroup(group model.DeviceGroup) (err error, code int) {
	if group.Id == "" {
		return errors.New("missing device-group id"), http.StatusBadRequest
	}
	if group.Name == "" {
		return errors.New("missing device-group name"), http.StatusBadRequest
	}
	return this.ValidateDeviceGroupSelection(group.Criteria, group.DeviceIds)
}

func (this *Controller) ValidateDeviceGroupSelection(criteria []model.DeviceGroupFilterCriteria, devices []string) (error, int) {
	deviceCache := map[string]model.Device{}
	deviceTypeCache := map[string]model.DeviceType{}
	deviceUsageCount := map[string]int{}
	for _, c := range criteria {
		deviceUsedInMapping := map[string]bool{}
		for _, deviceId := range devices {
			if deviceUsedInMapping[deviceId] {
				return errors.New("multiple uses of device-id " + deviceId + " for the same filter-criteria"), http.StatusBadRequest
			}
			deviceUsedInMapping[deviceId] = true
			deviceUsageCount[deviceId] = deviceUsageCount[deviceId] + 1
			err, code := this.selectionMatchesCriteria(&deviceCache, &deviceTypeCache, c, deviceId)
			if err != nil {
				return err, code
			}
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) selectionMatchesCriteria(
	dcache *map[string]model.Device,
	dtcache *map[string]model.DeviceType,
	criteria model.DeviceGroupFilterCriteria,
	deviceId string) (err error, code int) {

	ctx, _ := getTimeoutContext()
	var exists bool

	device, ok := (*dcache)[deviceId]
	if !ok {
		device, exists, err = this.db.GetDevice(ctx, deviceId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown device-id: " + deviceId), http.StatusBadRequest
		}
		(*dcache)[deviceId] = device
	}

	deviceType, ok := (*dtcache)[device.DeviceTypeId]
	if !ok {
		deviceType, exists, err = this.db.GetDeviceType(ctx, device.DeviceTypeId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown device-type-id: " + device.DeviceTypeId), http.StatusBadRequest
		}
		(*dtcache)[device.DeviceTypeId] = deviceType
	}

	deviceClassMatches := criteria.DeviceClassId == "" || criteria.DeviceClassId == deviceType.DeviceClassId
	if !deviceClassMatches {
		return errors.New("device " + deviceId + " does not match device-class of filter-criteria"), http.StatusBadRequest
	}

	serviceMatches := false
	for _, service := range deviceType.Services {
		interactionMatches := service.Interaction == criteria.Interaction
		if service.Interaction == model.EVENT_AND_REQUEST {
			interactionMatches = true
		}
		aspectMatches := criteria.AspectId == ""
		for _, aspectId := range service.AspectIds {
			if criteria.AspectId == "" || criteria.AspectId == aspectId {
				aspectMatches = true
				break
			}
		}
		functionMatches := false
		for _, functionId := range service.FunctionIds {
			if criteria.FunctionId == functionId {
				functionMatches = true
				break
			}
		}
		if interactionMatches && functionMatches && aspectMatches {
			serviceMatches = true
			break
		}
	}
	if !serviceMatches {
		return errors.New("no service of the device " + deviceId + " matches filter-criteria"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDeviceGroup(deviceGroup model.DeviceGroup, owner string) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.SetDeviceGroup(ctx, deviceGroup)
}

func (this *Controller) DeleteDeviceGroup(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveDeviceGroup(ctx, id)
}
