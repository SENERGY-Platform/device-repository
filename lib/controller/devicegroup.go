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
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
)

/////////////////////////
//		api
/////////////////////////

const FilterDevicesOfGroupByAccess = true

func (this *Controller) ReadDeviceGroup(id string, jwt jwt_http_router.Jwt) (result model.DeviceGroup, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	deviceGroup, exists, err := this.db.GetDeviceGroup(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err := this.security.CheckBool(jwt, this.config.DeviceGroupTopic, id, model.READ)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	if FilterDevicesOfGroupByAccess {
		return this.FilterDevicesOfGroupByAccess(jwt, result)
	} else {
		return deviceGroup, nil, http.StatusOK
	}
}

func (this *Controller) FilterDevicesOfGroupByAccess(jwt jwt_http_router.Jwt, group model.DeviceGroup) (result model.DeviceGroup, err error, code int) {
	deviceIds := []string{}
	if len(group.Devices) == 0 {
		return group, nil, http.StatusOK
	}
	//looping one element of group.Devices is enough because ValidateDeviceGroup() ensures that every used device is referenced in each group.Devices element
	for _, selection := range group.Devices[0].Selection {
		deviceIds = append(deviceIds, selection.DeviceId)
	}
	access, err := this.security.CheckMultiple(jwt, this.config.DeviceGroupTopic, deviceIds, model.EXECUTE)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	result = group
	for i, element := range group.Devices {
		result.Devices[i].Selection = []model.Selection{}
		for _, selection := range element.Selection {
			if access[selection.DeviceId] {
				result.Devices[i].Selection = append(result.Devices[i].Selection, selection)
			} else if this.config.Debug {
				log.Println("DEBUG: filtered " + selection.DeviceId + " from result, because user lost execution access to the device")
			}
		}
	}
	return result, nil, http.StatusOK
}

//only the first element of group.Devices is checked.
//this should be enough because every used device should be referenced in each element of group.Devices
//use ValidateDeviceGroup() to ensure that this constraint is adhered to
func (this *Controller) CheckAccessToDevicesOfGroup(jwt jwt_http_router.Jwt, group model.DeviceGroup) (err error, code int) {
	if len(group.Devices) == 0 {
		return nil, http.StatusOK
	}
	deviceIds := []string{}
	//looping one element of group.Devices is enough because ValidateDeviceGroup() ensures that every used device is referenced in each group.Devices element
	for _, selection := range group.Devices[0].Selection {
		deviceIds = append(deviceIds, selection.DeviceId)
	}
	access, err := this.security.CheckMultiple(jwt, this.config.DeviceGroupTopic, deviceIds, model.EXECUTE)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	//looping one element of group.Devices is enough because ValidateDeviceGroup() ensures that every used device is referenced in each group.Devices element
	for _, selection := range group.Devices[0].Selection {
		if !access[selection.DeviceId] {
			return errors.New("no execution access to device " + selection.DeviceId), http.StatusBadRequest
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
	switch group.BlockedInteraction {
	case "":
	case model.EVENT:
	case model.REQUEST:
	case model.EVENT_AND_REQUEST:
	default:
		return errors.New("unknown interaction in blocked_interaction: " + string(group.BlockedInteraction)), http.StatusBadRequest
	}
	return this.ValidateDeviceGroupMapping(group.BlockedInteraction, group.Devices)
}

func (this *Controller) ValidateDeviceGroupMapping(blockedInteraction model.Interaction, mapping []model.DeviceGroupMapping) (error, int) {
	deviceCache := map[string]model.Device{}
	deviceTypeCache := map[string]model.DeviceType{}
	deviceUsageCount := map[string]int{}
	for _, m := range mapping {
		deviceUsedInMapping := map[string]bool{}
		for _, s := range m.Selection {
			deviceId := s.DeviceId
			if deviceUsedInMapping[deviceId] {
				return errors.New("multiple uses of device-id " + deviceId + " for the same filter-criteria"), http.StatusBadRequest
			}
			deviceUsedInMapping[deviceId] = true
			deviceUsageCount[deviceId] = deviceUsageCount[deviceId] + 1
			err, code := this.selectionMatchesCriteria(&deviceCache, &deviceTypeCache, blockedInteraction, m.Criteria, s)
			if err != nil {
				return err, code
			}
		}
	}
	expectedCount := len(mapping)
	for deviceId, count := range deviceUsageCount {
		if count != expectedCount {
			return errors.New("expect " + deviceId + " to be referenced for every filter-criteria"), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) selectionMatchesCriteria(
	dcache *map[string]model.Device,
	dtcache *map[string]model.DeviceType,
	blockedInteraction model.Interaction,
	criteria model.FilterCriteria,
	selection model.Selection) (err error, code int) {

	ctx, _ := getTimeoutContext()
	var exists bool

	device, ok := (*dcache)[selection.DeviceId]
	if !ok {
		device, exists, err = this.db.GetDevice(ctx, selection.DeviceId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown device-id: " + selection.DeviceId), http.StatusBadRequest
		}
		(*dcache)[selection.DeviceId] = device
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
		return errors.New("device does not match device-class of filter-criteria"), http.StatusBadRequest
	}

	for _, serviceId := range selection.ServiceIds {
		var service *model.Service
		for _, s := range deviceType.Services {
			if s.Id == serviceId {
				service = &s
				break
			}
		}
		if service == nil {
			return errors.New("service (" + serviceId + ") not part of device (" + device.Id + ")"), http.StatusBadRequest
		}

		if service.Interaction == blockedInteraction {
			return errors.New("device/service uses blocked interaction: " + string(blockedInteraction)), http.StatusBadRequest
		}

		aspectMatches := false
		for _, aspectId := range service.AspectIds {
			if criteria.AspectId == "" || criteria.AspectId == aspectId {
				aspectMatches = true
				break
			}
		}
		if !aspectMatches {
			return errors.New("device/service does not match aspect of filter-criteria"), http.StatusBadRequest
		}
		functionMatches := false
		for _, functionId := range service.FunctionIds {
			if criteria.FunctionId == functionId {
				functionMatches = true
				break
			}
		}
		if !functionMatches {
			return errors.New("device/service does not match function of filter-criteria"), http.StatusBadRequest
		}
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
