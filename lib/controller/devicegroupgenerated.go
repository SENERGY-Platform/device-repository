/*
 * Copyright 2024 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"slices"
	"strings"
)

func (this *Controller) CreateGeneratedDeviceGroup(device models.Device) (err error) {
	virtualDgId := deviceIdToGeneratedDeviceGroupId(device.Id)
	dg := models.DeviceGroup{
		Id:                    virtualDgId,
		Name:                  getDeviceDisplayName(device) + "_group",
		DeviceIds:             []string{device.Id},
		AutoGeneratedByDevice: device.Id,
	}
	dg.Criteria, err, _ = this.getDeviceGroupCriteria(device)
	if err != nil {
		return err
	}
	dg.SetShortCriteria()
	return this.PublishDeviceGroup(dg, device.OwnerId)
}

func getDeviceDisplayName(device models.Device) string {
	displayName := device.Name
	for _, attr := range device.Attributes {
		if attr.Key == "shared/nickname" && attr.Value != "" {
			displayName = attr.Value
		}
	}
	return displayName
}

func (this *Controller) RemoveGeneratedDeviceGroup(deviceid string, owner string) error {
	virtualDgId := deviceIdToGeneratedDeviceGroupId(deviceid)
	ctx, _ := getTimeoutContext()
	dg, exists, err := this.db.GetDeviceGroup(ctx, virtualDgId)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	dg.DeviceIds = slices.DeleteFunc(dg.DeviceIds, func(s string) bool {
		return s == deviceid
	})
	if len(dg.DeviceIds) > 0 {
		dg.AutoGeneratedByDevice = ""
		return this.PublishDeviceGroup(dg, owner)
	} else {
		return this.PublishDeviceGroupDelete(virtualDgId, owner)
	}
}

func deviceIdToGeneratedDeviceGroupId(deviceId string) string {
	return models.URN_PREFIX + "device-group:" + strings.TrimPrefix(deviceId, models.URN_PREFIX+"device:")
}

func (this *Controller) getDeviceGroupCriteria(device models.Device) (result []models.DeviceGroupFilterCriteria, err error, code int) {
	ctx, _ := getTimeoutContext()
	deviceType, _, err := this.db.GetDeviceType(ctx, device.DeviceTypeId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	resultSet := map[string]models.DeviceGroupFilterCriteria{}
	for _, service := range deviceType.Services {
		interactions := []models.Interaction{service.Interaction}
		if service.Interaction == models.EVENT_AND_REQUEST {
			interactions = []models.Interaction{models.EVENT, models.REQUEST}
		}
		work := []models.ContentVariable{}
		for _, content := range service.Inputs {
			work = append(work, content.ContentVariable)
		}
		for _, content := range service.Outputs {
			work = append(work, content.ContentVariable)
		}
		for i := 0; i < len(work); i++ {
			current := work[i]
			if current.FunctionId != "" {
				for _, interaction := range interactions {
					if isMeasuringFunctionId(current.FunctionId) {
						criteria := models.DeviceGroupFilterCriteria{
							FunctionId:  current.FunctionId,
							AspectId:    current.AspectId,
							Interaction: interaction,
						}
						resultSet[criteriaHash(criteria)] = criteria
						if current.AspectId != "" {
							aspectNode, _, err := this.db.GetAspectNode(ctx, current.AspectId)
							if err != nil {
								return result, err, http.StatusInternalServerError
							}
							for _, aspect := range aspectNode.AncestorIds {
								criteria := models.DeviceGroupFilterCriteria{
									FunctionId:  current.FunctionId,
									AspectId:    aspect,
									Interaction: interaction,
								}
								resultSet[criteriaHash(criteria)] = criteria
							}
						}
					} else {
						criteria := models.DeviceGroupFilterCriteria{
							FunctionId:    current.FunctionId,
							DeviceClassId: deviceType.DeviceClassId,
							Interaction:   interaction,
						}
						resultSet[criteriaHash(criteria)] = criteria
					}
				}
			}
			if len(current.SubContentVariables) > 0 {
				work = append(work, current.SubContentVariables...)
			}
		}
	}
	for _, element := range resultSet {
		result = append(result, element)
	}
	return result, nil, http.StatusOK
}

func criteriaHash(criteria models.DeviceGroupFilterCriteria) string {
	return criteria.FunctionId + "_" + criteria.AspectId + "_" + criteria.DeviceClassId + "_" + string(criteria.Interaction)
}

func isMeasuringFunctionId(id string) bool {
	if strings.HasPrefix(id, model.MEASURING_FUNCTION_PREFIX) {
		return true
	}
	return false
}
