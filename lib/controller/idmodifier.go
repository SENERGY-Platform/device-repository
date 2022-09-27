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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"net/http"
	"strings"
)

func (this *Controller) modifyDevice(device model.Device, modifier map[string][]string) (result model.Device, err error, code int) {
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

func (this *Controller) modifyDeviceType(dt model.DeviceType, modifier map[string][]string) (result model.DeviceType, err error, code int) {
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

const ServiceGroupSelectionIdModifier = "service_group_selection"

func (this *Controller) modifyDeviceServiceGroupSelection(device model.Device, params []string) (result model.Device, err error, ode int) {
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
	serviceGroupList := []model.ServiceGroup{}
	if this.config.DeviceServiceGroupSelectionAllowNotFound {
		serviceGroupList = append(dt.ServiceGroups, model.ServiceGroup{
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

func (this *Controller) modifyDeviceTypeServiceGroupSelection(dt model.DeviceType, params []string) (result model.DeviceType, err error, ode int) {
	if len(params) == 0 {
		return result, errors.New("missing service-group-key in " + ServiceGroupSelectionIdModifier + " id parameter"), http.StatusBadRequest
	}
	result = dt
	sgKey := params[0]

	newServiceList := []model.Service{}

	for _, service := range dt.Services {
		if service.ServiceGroupKey == sgKey || service.ServiceGroupKey == "" {
			newServiceList = append(newServiceList, service)
		}
	}

	result.Services = newServiceList
	return result, nil, http.StatusOK
}
