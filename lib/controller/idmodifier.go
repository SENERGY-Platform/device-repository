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
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"slices"
	"strings"
)

func (this *Controller) modifyDevice(device models.Device, modifier map[string][]string) (result models.Device, err error, code int) {
	code = http.StatusOK
	result = device
	for key, params := range modifier {
		switch key {
		case ServiceGroupSelectionIdModifier:
			result, err, code = this.modifyDeviceServiceGroupSelection(result, params)
			if err != nil {
				return result, err, code
			}
		}
	}
	return result, err, code
}

func (this *Controller) modifyDeviceType(dt models.DeviceType, modifier map[string][]string) (result models.DeviceType, err error, code int) {
	code = http.StatusOK
	result = dt
	for key, params := range modifier {
		switch key {
		case ServiceGroupSelectionIdModifier:
			result, err, code = this.modifyDeviceTypeServiceGroupSelection(result, params)
			if err != nil {
				return result, err, code
			}
		}
	}
	return result, err, code
}

func (this *Controller) modifyDeviceTypeList(list []models.DeviceType, sort string, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, code int) {
	result = []models.DeviceType{}
	for _, dt := range list {
		pureId, modifier := idmodifier.SplitModifier(dt.Id)
		isModified := pureId != dt.Id && len(modifier) > 0
		if !isModified && includeUnmodified {
			result = append(result, dt)
		}
		if isModified && includeModified {
			dt, err, code = this.modifyDeviceType(dt, modifier)
			if err != nil {
				return result, err, code
			}
			result = append(result, dt)
		}
	}
	if includeModified {
		result = sortDeviceTypes(result, sort)
	}
	return result, nil, http.StatusOK
}

func sortDeviceTypes(list []models.DeviceType, sort string) []models.DeviceType {
	parts := strings.Split(sort, ".")
	direction := 1
	if len(parts) > 1 && parts[1] == "desc" {
		direction = -1
	}
	switch parts[0] {
	case "id":
		slices.SortFunc(list, func(a, b models.DeviceType) int {
			return strings.Compare(a.Id, b.Id) * direction
		})
	case "name":
		slices.SortFunc(list, func(a, b models.DeviceType) int {
			return strings.Compare(a.Name, b.Name) * direction
		})
	default:
		return list
	}
	return list
}

const ServiceGroupSelectionIdModifier = "service_group_selection"

func (this *Controller) modifyDeviceServiceGroupSelection(device models.Device, params []string) (result models.Device, err error, ode int) {
	if len(params) == 0 {
		return result, errors.New("missing service-group-key in " + ServiceGroupSelectionIdModifier + " id parameter"), http.StatusBadRequest
	}
	result = device
	sgKey := params[0]
	ctx, _ := getTimeoutContext()
	dt, exists, err := this.db.GetDeviceType(ctx, device.DeviceTypeId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("unable to use " + ServiceGroupSelectionIdModifier + " modifier: device-type not found"), http.StatusInternalServerError
	}
	if !strings.Contains(result.DeviceTypeId, idmodifier.Seperator) {
		result.DeviceTypeId = result.DeviceTypeId + idmodifier.Seperator
	} else {
		result.DeviceTypeId = result.DeviceTypeId + "&"
	}
	result.DeviceTypeId = result.DeviceTypeId + idmodifier.EncodeModifierParameter(map[string][]string{ServiceGroupSelectionIdModifier: params})
	serviceGroupList := []models.ServiceGroup{}
	if this.config.DeviceServiceGroupSelectionAllowNotFound {
		serviceGroupList = append(dt.ServiceGroups, models.ServiceGroup{
			Key:  sgKey,
			Name: sgKey,
		})
	} else {
		serviceGroupList = dt.ServiceGroups
	}
	for _, sg := range serviceGroupList {
		if sg.Key == sgKey {
			if result.Name != "" {
				result.Name = result.Name + " " + sg.Name
			}
			for i, attr := range result.Attributes {
				if attr.Key == DisplayNameAttributeName {
					attr.Value = attr.Value + " " + sg.Name
				}
				result.Attributes[i] = attr
			}
			return result, nil, http.StatusOK
		}
	}

	return result, errors.New("no matching service-group-key found for " + ServiceGroupSelectionIdModifier + " id parameter"), http.StatusOK
}

func (this *Controller) modifyDeviceTypeServiceGroupSelection(dt models.DeviceType, params []string) (result models.DeviceType, err error, ode int) {
	if len(params) == 0 {
		return result, errors.New("missing service-group-key in " + ServiceGroupSelectionIdModifier + " id parameter"), http.StatusBadRequest
	}
	result = dt
	sgKey := params[0]

	newServiceList := []models.Service{}

	for _, service := range dt.Services {
		if service.ServiceGroupKey == sgKey || service.ServiceGroupKey == "" {
			newServiceList = append(newServiceList, service)
		}
	}

	for _, sg := range dt.ServiceGroups {
		if sg.Key == sgKey {
			result.Name = result.Name + " " + sg.Name
			break
		}
	}

	result.Services = newServiceList
	return result, nil, http.StatusOK
}
