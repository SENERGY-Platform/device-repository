/*
 * Copyright 2025 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"net/http"
	"slices"
)

func (this *Controller) Export(token string, options model.ImportExportOptions) (result model.ImportExport, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		return result, errors.New("only admins may export"), http.StatusForbidden
	}
	result = model.ImportExport{}

	result.Protocols, err, code = this.ExportProtocols(token, options)
	if err != nil {
		return result, err, code
	}

	result.Functions, err, code = this.ExportFunctions(token, options)
	if err != nil {
		return result, err, code
	}

	result.Aspects, err, code = this.ExportAspects(token, options)
	if err != nil {
		return result, err, code
	}

	result.Concepts, err, code = this.ExportConcepts(token, options)
	if err != nil {
		return result, err, code
	}

	result.Characteristics, err, code = this.ExportCharacteristics(token, options)
	if err != nil {
		return result, err, code
	}

	result.DeviceClasses, err, code = this.ExportDeviceClasses(token, options)
	if err != nil {
		return result, err, code
	}

	result.DeviceTypes, err, code = this.ExportDeviceTypes(token, options)
	if err != nil {
		return result, err, code
	}

	if options.IncludeOwnedInformation {
		var tempPerm []client.Resource

		result.Devices, tempPerm, err, code = this.ExportDevices(token, options)
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)

		result.DeviceGroups, tempPerm, err, code = this.ExportDeviceGroups(token, options)
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)

		result.Hubs, tempPerm, err, code = this.ExportHubs(token, options)
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)

		result.Locations, tempPerm, err, code = this.ExportLocations(token, options)
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)
	}
	result.Sort()
	return result, nil, http.StatusOK
}

func (this *Controller) ExportProtocols(token string, options model.ImportExportOptions) (result []models.Protocol, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "protocols") {
		result, err = this.db.ListProtocols(context.Background(), 0, 0, "name.asc")
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	if options.FilterIds != nil {
		temp := []models.Protocol{}
		for _, e := range result {
			if slices.Contains(options.FilterIds, e.Id) {
				temp = append(temp, e)
			}
		}
		result = temp
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportConcepts(token string, options model.ImportExportOptions) (result []models.Concept, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "concepts") {
		result, _, err = this.db.ListConcepts(context.Background(), model.ConceptListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportFunctions(token string, options model.ImportExportOptions) (result []models.Function, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "functions") {
		result, _, err = this.db.ListFunctions(context.Background(), model.FunctionListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportAspects(token string, options model.ImportExportOptions) (result []models.Aspect, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "aspects") {
		result, _, err = this.db.ListAspects(context.Background(), model.AspectListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportCharacteristics(token string, options model.ImportExportOptions) (result []models.Characteristic, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "characteristics") {
		result, _, err = this.db.ListCharacteristics(context.Background(), model.CharacteristicListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportDeviceClasses(token string, options model.ImportExportOptions) (result []models.DeviceClass, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "device-classes") {
		result, _, err = this.db.ListDeviceClasses(context.Background(), model.DeviceClassListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportDeviceTypes(token string, options model.ImportExportOptions) (result []models.DeviceType, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "device-types") {
		result, _, err = this.db.ListDeviceTypesV3(context.Background(), model.DeviceTypeListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportDevices(token string, options model.ImportExportOptions) (result []models.Device, perm []client.Resource, err error, code int) {
	if options.FilterResourceTypes != nil && !slices.Contains(options.FilterResourceTypes, "devices") {
		return nil, nil, nil, http.StatusOK
	}
	tempDevices, _, err := this.db.ListDevices(context.Background(), model.DeviceListOptions{Ids: options.FilterIds}, false)
	if err != nil {
		return result, perm, err, http.StatusInternalServerError
	}
	for _, d := range tempDevices {
		result = append(result, d.Device)
	}
	perm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermission(token, this.config.DeviceTopic, client.ListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, code
	}
	return result, perm, nil, http.StatusOK
}

func (this *Controller) ExportDeviceGroups(token string, options model.ImportExportOptions) (result []models.DeviceGroup, perm []client.Resource, err error, code int) {
	if options.FilterResourceTypes != nil && !slices.Contains(options.FilterResourceTypes, "device-groups") {
		return nil, nil, nil, http.StatusOK
	}
	result, _, err = this.db.ListDeviceGroups(context.Background(), model.DeviceGroupListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, http.StatusInternalServerError
	}
	perm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermission(token, this.config.DeviceGroupTopic, client.ListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, code
	}
	return result, perm, nil, http.StatusOK
}

func (this *Controller) ExportHubs(token string, options model.ImportExportOptions) (result []models.Hub, perm []client.Resource, err error, code int) {
	if options.FilterResourceTypes != nil && !slices.Contains(options.FilterResourceTypes, "hubs") {
		return nil, nil, nil, http.StatusOK
	}
	tempHubs, _, err := this.db.ListHubs(context.Background(), model.HubListOptions{Ids: options.FilterIds}, false)
	if err != nil {
		return result, perm, err, http.StatusInternalServerError
	}
	for _, d := range tempHubs {
		result = append(result, d.Hub)
	}
	perm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermission(token, this.config.HubTopic, client.ListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, code
	}
	return result, perm, nil, http.StatusOK
}

func (this *Controller) ExportLocations(token string, options model.ImportExportOptions) (result []models.Location, perm []client.Resource, err error, code int) {
	if options.FilterResourceTypes != nil && !slices.Contains(options.FilterResourceTypes, "locations") {
		return nil, nil, nil, http.StatusOK
	}
	result, _, err = this.db.ListLocations(context.Background(), model.LocationListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, http.StatusInternalServerError
	}
	perm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermission(token, this.config.LocationTopic, client.ListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, code
	}
	return result, perm, nil, http.StatusOK
}

func (this *Controller) Import(token string, importModel model.ImportExport, options model.ImportExportOptions) (err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		return errors.New("only admins may export"), http.StatusForbidden
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "protocols") {
		for _, p := range importModel.Protocols {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, p.Id) {
				err, code = this.ValidateProtocol(p)
				if err != nil {
					return err, code
				}
				err = this.setProtocol(p)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "characteristics") {
		for _, characteristic := range importModel.Characteristics {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, characteristic.Id) {
				err, code = this.ValidateCharacteristics(characteristic)
				if err != nil {
					return err, code
				}
				err = this.setCharacteristic(characteristic)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "concepts") {
		for _, concept := range importModel.Concepts {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, concept.Id) {
				err, code = this.ValidateConcept(concept)
				if err != nil {
					return err, code
				}
				err = this.setConcept(concept)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "functions") {
		for _, f := range importModel.Functions {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, f.Id) {
				err, code = this.ValidateFunction(f)
				if err != nil {
					return err, code
				}
				err = this.setFunction(f)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "aspects") {
		for _, a := range importModel.Aspects {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, a.Id) {
				err, code = this.ValidateAspect(a)
				if err != nil {
					return err, code
				}
				err = this.setAspect(a)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "device-classes") {
		for _, dc := range importModel.DeviceClasses {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, dc.Id) {
				err, code = this.ValidateDeviceClass(dc)
				if err != nil {
					return err, code
				}
				err = this.setDeviceClass(dc)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "device-types") {
		for _, dt := range importModel.DeviceTypes {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, dt.Id) {
				err, code = this.ValidateDeviceType(dt, model.ValidationOptions{})
				if err != nil {
					return err, code
				}
				err = this.setDeviceType(dt)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.IncludeOwnedInformation {
		//set permissions before and after local db, to ensure validation and no permission changes
		for _, p := range importModel.Permissions {
			if (options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, p.TopicId)) && (options.FilterIds == nil || slices.Contains(options.FilterIds, p.Id)) {
				_, err, code = this.permissionsV2Client.SetPermission(token, p.TopicId, p.Id, p.ResourcePermissions)
				if err != nil {
					return err, code
				}
			}
		}

		if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "devices") {
			for _, d := range importModel.Devices {
				if options.FilterIds == nil || slices.Contains(options.FilterIds, d.Id) {
					err, code = this.ValidateDevice(token, d)
					if err != nil {
						return err, code
					}
					_, err, _ = this.setDevice(d)
					if err != nil {
						return err, http.StatusInternalServerError
					}
				}
			}
		}

		if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "hubs") {
			for _, h := range importModel.Hubs {
				if options.FilterIds == nil || slices.Contains(options.FilterIds, h.Id) {
					err, code = this.ValidateHub(token, h)
					if err != nil {
						return err, code
					}
					err = this.setHub(model.HubWithConnectionState{
						Hub:             h,
						ConnectionState: models.ConnectionStateUnknown,
					})
					if err != nil {
						return err, http.StatusInternalServerError
					}
				}
			}
		}

		if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "device-groups") {
			for _, dg := range importModel.DeviceGroups {
				if options.FilterIds == nil || slices.Contains(options.FilterIds, dg.Id) {
					err, code = this.ValidateDeviceGroup(token, dg)
					if err != nil {
						return err, code
					}
					err = this.setDeviceGroup(dg, jwtToken.GetUserId())
					if err != nil {
						return err, http.StatusInternalServerError
					}
				}
			}
		}

		if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "locations") {
			for _, l := range importModel.Locations {
				if options.FilterIds == nil || slices.Contains(options.FilterIds, l.Id) {
					err, code = this.ValidateLocation(l)
					if err != nil {
						return err, code
					}
					err = this.setLocation(l, jwtToken.GetUserId())
					if err != nil {
						return err, http.StatusInternalServerError
					}
				}
			}
		}

		//set permissions before and after local db, to ensure validation and no permission changes
		for _, p := range importModel.Permissions {
			if (options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, p.TopicId)) && (options.FilterIds == nil || slices.Contains(options.FilterIds, p.Id)) {
				_, err, code = this.permissionsV2Client.SetPermission(token, p.TopicId, p.Id, p.ResourcePermissions)
				if err != nil {
					return err, code
				}
			}
		}
	}
	return nil, http.StatusOK
}
