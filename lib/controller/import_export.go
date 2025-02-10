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

	result.Protocols, err = this.db.ListProtocols(context.Background(), 0, 0, "name.asc")
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	result.Functions, _, err = this.db.ListFunctions(context.Background(), model.FunctionListOptions{})
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	result.Aspects, _, err = this.db.ListAspects(context.Background(), model.AspectListOptions{})
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	result.Concepts, _, err = this.db.ListConcepts(context.Background(), model.ConceptListOptions{})
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	result.Characteristics, _, err = this.db.ListCharacteristics(context.Background(), model.CharacteristicListOptions{})
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	result.DeviceClasses, _, err = this.db.ListDeviceClasses(context.Background(), model.DeviceClassListOptions{})
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	result.DeviceTypes, _, err = this.db.ListDeviceTypesV3(context.Background(), model.DeviceTypeListOptions{})
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	if options.IncludeOwnedInformation {
		tempDevices, _, err := this.db.ListDevices(context.Background(), model.DeviceListOptions{}, false)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		for _, d := range tempDevices {
			result.Devices = append(result.Devices, d.Device)
		}
		tempPerm, err, code := this.permissionsV2Client.ListResourcesWithAdminPermission(token, this.config.DeviceTopic, client.ListOptions{})
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)

		result.DeviceGroups, _, err = this.db.ListDeviceGroups(context.Background(), model.DeviceGroupListOptions{IgnoreGenerated: true})
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		tempPerm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermission(token, this.config.DeviceGroupTopic, client.ListOptions{})
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)

		tempHubs, _, err := this.db.ListHubs(context.Background(), model.HubListOptions{}, false)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		for _, h := range tempHubs {
			result.Hubs = append(result.Hubs, h.Hub)
		}
		tempPerm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermission(token, this.config.HubTopic, client.ListOptions{})
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)

		result.Locations, _, err = this.db.ListLocations(context.Background(), model.LocationListOptions{})
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		tempPerm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermission(token, this.config.LocationTopic, client.ListOptions{})
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)
	}
	return result, nil, http.StatusOK
}

func (this *Controller) Import(token string, importModel model.ImportExport, options model.ImportExportOptions) (err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		return errors.New("only admins may export"), http.StatusForbidden
	}

	for _, p := range importModel.Protocols {
		err, code = this.ValidateProtocol(p)
		if err != nil {
			return err, code
		}
		err = this.setProtocol(p)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}

	for _, characteristic := range importModel.Characteristics {
		err, code = this.ValidateCharacteristics(characteristic)
		if err != nil {
			return err, code
		}
		err = this.setCharacteristic(characteristic)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}

	for _, concept := range importModel.Concepts {
		err, code = this.ValidateConcept(concept)
		if err != nil {
			return err, code
		}
		err = this.setConcept(concept)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}

	for _, f := range importModel.Functions {
		err, code = this.ValidateFunction(f)
		if err != nil {
			return err, code
		}
		err = this.setFunction(f)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}

	for _, a := range importModel.Aspects {
		err, code = this.ValidateAspect(a)
		if err != nil {
			return err, code
		}
		err = this.setAspect(a)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}

	for _, dc := range importModel.DeviceClasses {
		err, code = this.ValidateDeviceClass(dc)
		if err != nil {
			return err, code
		}
		err = this.setDeviceClass(dc)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}

	for _, dt := range importModel.DeviceTypes {
		err, code = this.ValidateDeviceType(dt, model.ValidationOptions{})
		if err != nil {
			return err, code
		}
		err = this.setDeviceType(dt)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}

	if options.IncludeOwnedInformation {
		//set permissions before and after local db, to ensure validation and no permission changes
		for _, p := range importModel.Permissions {
			_, err, code = this.permissionsV2Client.SetPermission(token, p.TopicId, p.Id, p.ResourcePermissions)
			if err != nil {
				return err, code
			}
		}
		for _, d := range importModel.Devices {
			err, code = this.ValidateDevice(token, d)
			if err != nil {
				return err, code
			}
			_, err, _ = this.setDevice(d)
			if err != nil {
				return err, http.StatusInternalServerError
			}
		}
		for _, h := range importModel.Hubs {
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
		for _, dg := range importModel.DeviceGroups {
			err, code = this.ValidateDeviceGroup(token, dg)
			if err != nil {
				return err, code
			}
			err = this.setDeviceGroup(dg, jwtToken.GetUserId())
			if err != nil {
				return err, http.StatusInternalServerError
			}
		}
		for _, l := range importModel.Locations {
			err, code = this.ValidateLocation(l)
			if err != nil {
				return err, code
			}
			err = this.setLocation(l, jwtToken.GetUserId())
			if err != nil {
				return err, http.StatusInternalServerError
			}
		}
		//set permissions before and after local db, to ensure validation and no permission changes
		for _, p := range importModel.Permissions {
			_, err, code = this.permissionsV2Client.SetPermission(token, p.TopicId, p.Id, p.ResourcePermissions)
			if err != nil {
				return err, code
			}
		}
	}
	return nil, http.StatusOK
}
