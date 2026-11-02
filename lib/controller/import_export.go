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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func (this *Controller) Export(ctx context.Context, token string, options model.ImportExportOptions) (result model.ImportExport, err error, code int) {
	result = model.ImportExport{}

	result.Protocols, err, code = this.ExportProtocols(ctx, options)
	if err != nil {
		return result, err, code
	}

	result.Functions, err, code = this.ExportFunctions(ctx, options)
	if err != nil {
		return result, err, code
	}

	result.Aspects, err, code = this.ExportAspects(ctx, options)
	if err != nil {
		return result, err, code
	}

	result.AspectClasses, err, code = this.ExportAspectClasses(ctx, options)
	if err != nil {
		return result, err, code
	}

	result.Concepts, err, code = this.ExportConcepts(ctx, options)
	if err != nil {
		return result, err, code
	}

	result.Characteristics, err, code = this.ExportCharacteristics(ctx, options)
	if err != nil {
		return result, err, code
	}

	result.DeviceClasses, err, code = this.ExportDeviceClasses(ctx, options)
	if err != nil {
		return result, err, code
	}

	result.DeviceTypes, err, code = this.ExportDeviceTypes(ctx, options)
	if err != nil {
		return result, err, code
	}

	if options.IncludeOwnedInformation {
		jwtToken, err := jwt.Parse(token)
		if err != nil {
			return result, err, http.StatusBadRequest
		}
		if !jwtToken.IsAdmin() {
			return result, errors.New("only admins may export owned information"), http.StatusForbidden
		}

		var tempPerm []client.Resource

		result.Devices, tempPerm, err, code = this.ExportDevices(ctx, token, options)
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)

		result.DeviceGroups, tempPerm, err, code = this.ExportDeviceGroups(ctx, token, options)
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)

		result.Hubs, tempPerm, err, code = this.ExportHubs(ctx, token, options)
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)

		result.Locations, tempPerm, err, code = this.ExportLocations(ctx, token, options)
		if err != nil {
			return result, err, code
		}
		result.Permissions = append(result.Permissions, tempPerm...)
	}
	result.Sort()
	return result, nil, http.StatusOK
}

func (this *Controller) ExportProtocols(ctx context.Context, options model.ImportExportOptions) (result []models.Protocol, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "protocols") {
		result, err = this.db.ListProtocols(context.WithoutCancel(ctx), 0, 0, "name.asc")
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

func (this *Controller) ExportAspectClasses(ctx context.Context, options model.ImportExportOptions) (result []models.AspectClass, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "aspect-classes") {
		result, _, err = this.db.ListAspectClasses(context.WithoutCancel(ctx), model.AspectClassListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportConcepts(ctx context.Context, options model.ImportExportOptions) (result []models.Concept, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "concepts") {
		result, _, err = this.db.ListConcepts(context.WithoutCancel(ctx), model.ConceptListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportFunctions(ctx context.Context, options model.ImportExportOptions) (result []models.Function, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "functions") {
		result, _, err = this.db.ListFunctions(context.WithoutCancel(ctx), model.FunctionListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportAspects(ctx context.Context, options model.ImportExportOptions) (result []models.Aspect, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "aspects") {
		result, _, err = this.db.ListAspects(context.WithoutCancel(ctx), model.AspectListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportCharacteristics(ctx context.Context, options model.ImportExportOptions) (result []models.Characteristic, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "characteristics") {
		result, _, err = this.db.ListCharacteristics(context.WithoutCancel(ctx), model.CharacteristicListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportDeviceClasses(ctx context.Context, options model.ImportExportOptions) (result []models.DeviceClass, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "device-classes") {
		result, _, err = this.db.ListDeviceClasses(context.WithoutCancel(ctx), model.DeviceClassListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return result, err, http.StatusOK
}

func (this *Controller) ExportDeviceTypes(ctx context.Context, options model.ImportExportOptions) (result []models.DeviceType, err error, code int) {
	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "device-types") {
		result, _, err = this.db.ListDeviceTypesV3(context.WithoutCancel(ctx), model.DeviceTypeListOptions{Ids: options.FilterIds})
	}
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	setContentVariableAspectIdOnReadList(result)
	return result, err, http.StatusOK
}

func (this *Controller) ExportDevices(ctx context.Context, token string, options model.ImportExportOptions) (result []models.Device, perm []client.Resource, err error, code int) {
	if options.FilterResourceTypes != nil && !slices.Contains(options.FilterResourceTypes, "devices") {
		return nil, nil, nil, http.StatusOK
	}
	tempDevices, _, err := this.db.ListDevices(context.WithoutCancel(ctx), model.DeviceListOptions{Ids: options.FilterIds}, false)
	if err != nil {
		return result, perm, err, http.StatusInternalServerError
	}
	for _, d := range tempDevices {
		result = append(result, d.Device)
	}
	perm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermissionContext(ctx, token, this.config.DeviceTopic, client.ListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, code
	}
	return result, perm, nil, http.StatusOK
}

func (this *Controller) ExportDeviceGroups(ctx context.Context, token string, options model.ImportExportOptions) (result []models.DeviceGroup, perm []client.Resource, err error, code int) {
	if options.FilterResourceTypes != nil && !slices.Contains(options.FilterResourceTypes, "device-groups") {
		return nil, nil, nil, http.StatusOK
	}
	result, _, err = this.db.ListDeviceGroups(context.WithoutCancel(ctx), model.DeviceGroupListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, http.StatusInternalServerError
	}
	//read straight from the database: device-groups written before the migration carry only
	//the deprecated DeviceGroupFilterCriteria.AspectId
	setDeviceGroupCriteriaAspectIdsOnReadList(result)
	perm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermissionContext(ctx, token, this.config.DeviceGroupTopic, client.ListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, code
	}
	return result, perm, nil, http.StatusOK
}

func (this *Controller) ExportHubs(ctx context.Context, token string, options model.ImportExportOptions) (result []models.Hub, perm []client.Resource, err error, code int) {
	if options.FilterResourceTypes != nil && !slices.Contains(options.FilterResourceTypes, "hubs") {
		return nil, nil, nil, http.StatusOK
	}
	tempHubs, _, err := this.db.ListHubs(context.WithoutCancel(ctx), model.HubListOptions{Ids: options.FilterIds}, false)
	if err != nil {
		return result, perm, err, http.StatusInternalServerError
	}
	for _, d := range tempHubs {
		result = append(result, d.Hub)
	}
	perm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermissionContext(ctx, token, this.config.HubTopic, client.ListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, code
	}
	return result, perm, nil, http.StatusOK
}

func (this *Controller) ExportLocations(ctx context.Context, token string, options model.ImportExportOptions) (result []models.Location, perm []client.Resource, err error, code int) {
	if options.FilterResourceTypes != nil && !slices.Contains(options.FilterResourceTypes, "locations") {
		return nil, nil, nil, http.StatusOK
	}
	result, _, err = this.db.ListLocations(context.WithoutCancel(ctx), model.LocationListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, http.StatusInternalServerError
	}
	perm, err, code = this.permissionsV2Client.ListResourcesWithAdminPermissionContext(ctx, token, this.config.LocationTopic, client.ListOptions{Ids: options.FilterIds})
	if err != nil {
		return result, perm, err, code
	}
	return result, perm, nil, http.StatusOK
}

func (this *Controller) Import(ctx context.Context, token string, importModel model.ImportExport, options model.ImportExportOptions) (err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		return errors.New("only admins may import"), http.StatusForbidden
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "protocols") {
		for _, p := range importModel.Protocols {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, p.Id) {
				err, code = this.ValidateProtocol(ctx, p)
				if err != nil {
					return err, code
				}
				err = this.setProtocol(ctx, p)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "characteristics") {
		for _, characteristic := range importModel.Characteristics {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, characteristic.Id) {
				err, code = this.ValidateCharacteristics(ctx, characteristic)
				if err != nil {
					return err, code
				}
				err = this.setCharacteristic(ctx, characteristic)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "concepts") {
		for _, concept := range importModel.Concepts {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, concept.Id) {
				err, code = this.ValidateConcept(ctx, concept)
				if err != nil {
					return err, code
				}
				err = this.setConcept(ctx, concept)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "functions") {
		for _, f := range importModel.Functions {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, f.Id) {
				err, code = this.ValidateFunction(ctx, f)
				if err != nil {
					return err, code
				}
				err = this.setFunction(ctx, f)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "aspects") {
		for _, a := range importModel.Aspects {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, a.Id) {
				err, code = ResolveAspectClassIds(&a)
				if err != nil {
					return err, code
				}
				err, code = this.ValidateAspect(ctx, a)
				if err != nil {
					return err, code
				}
				err = this.setAspect(ctx, a)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "aspect-classes") {
		for _, ac := range importModel.AspectClasses {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, ac.Id) {
				err, code = this.ValidateAspectClass(ctx, ac)
				if err != nil {
					return err, code
				}
				err = this.setAspectClass(ctx, ac)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "device-classes") {
		for _, dc := range importModel.DeviceClasses {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, dc.Id) {
				err, code = this.ValidateDeviceClass(ctx, dc)
				if err != nil {
					return err, code
				}
				err = this.setDeviceClass(ctx, dc)
				if err != nil {
					return err, http.StatusInternalServerError
				}
			}
		}
	}

	if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "device-types") {
		for _, dt := range importModel.DeviceTypes {
			if options.FilterIds == nil || slices.Contains(options.FilterIds, dt.Id) {
				SetContentVariableAspectIdsOnWrite(&dt)
				err, code = this.ValidateDeviceType(ctx, dt, model.ValidationOptions{})
				if err != nil {
					return err, code
				}
				err = this.setDeviceType(ctx, dt)
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
				_, err, code = this.permissionsV2Client.SetPermissionContext(ctx, token, p.TopicId, p.Id, p.ResourcePermissions)
				if err != nil {
					return err, code
				}
			}
		}

		if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "devices") {
			for _, d := range importModel.Devices {
				if options.FilterIds == nil || slices.Contains(options.FilterIds, d.Id) {
					err, code = this.ValidateDevice(ctx, token, d)
					if err != nil {
						this.logger.Warn("device validation failed with error: " + err.Error())
						return err, code
					}
					_, err, _ = this.setDevice(ctx, d)
					if err != nil {
						return err, http.StatusInternalServerError
					}
				}
			}
		}

		if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "hubs") {
			for _, h := range importModel.Hubs {
				if options.FilterIds == nil || slices.Contains(options.FilterIds, h.Id) {
					err, code = this.ValidateHub(ctx, token, h)
					if err != nil {
						return err, code
					}
					err = this.setHub(ctx, model.HubWithConnectionState{
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
					//written with db.setDeviceGroup below, so the write normalization of
					//Controller.SetDeviceGroup never runs on an imported group
					SetDeviceGroupCriteriaAspectIdsOnWrite(&dg)
					err, code = this.ValidateDeviceGroup(ctx, token, dg)
					if err != nil {
						return err, code
					}
					err = this.setDeviceGroup(ctx, dg, jwtToken.GetUserId())
					if err != nil {
						return err, http.StatusInternalServerError
					}
				}
			}
		}

		if options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, "locations") {
			for _, l := range importModel.Locations {
				if options.FilterIds == nil || slices.Contains(options.FilterIds, l.Id) {
					err, code = this.ValidateLocation(ctx, l)
					if err != nil {
						return err, code
					}
					err = this.setLocation(ctx, l, jwtToken.GetUserId())
					if err != nil {
						return err, http.StatusInternalServerError
					}
				}
			}
		}

		//set permissions before and after local db, to ensure validation and no permission changes
		for _, p := range importModel.Permissions {
			if (options.FilterResourceTypes == nil || slices.Contains(options.FilterResourceTypes, p.TopicId)) && (options.FilterIds == nil || slices.Contains(options.FilterIds, p.Id)) {
				_, err, code = this.permissionsV2Client.SetPermissionContext(ctx, token, p.TopicId, p.Id, p.ResourcePermissions)
				if err != nil {
					return err, code
				}
			}
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) ImportFrom(ctx context.Context, token string, includeOwnedInformation bool, options model.ImportFromOptions) (err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		return errors.New("only admins may import"), http.StatusForbidden
	}

	if !hasBearerPrefix(options.RemoteAuthToken) {
		options.RemoteAuthToken = "bearer " + options.RemoteAuthToken
	}

	ieoptions := model.ImportExportOptions{
		IncludeOwnedInformation: includeOwnedInformation,
		FilterResourceTypes:     options.FilterResourceTypes,
		FilterIds:               options.FilterIds,
	}

	req, err := GetExportHttpRequest(options.RemoteDeviceRepository, options.RemoteAuthToken, ieoptions)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp)), resp.StatusCode
	}
	export := model.ImportExport{}
	err = json.NewDecoder(resp.Body).Decode(&export)
	if err != nil {
		_, _ = io.ReadAll(resp.Body) //ensure resp.Body is read to EOF
		return err, http.StatusInternalServerError
	}

	return this.Import(ctx, token, export, ieoptions)
}

func hasBearerPrefix(token string) bool {
	return len(token) > 7 && strings.ToLower(token[:7]) == "bearer "
}

func GetExportHttpRequest(deviceRepoUrl string, token string, options model.ImportExportOptions) (*http.Request, error) {
	queryString := ""
	query := url.Values{}
	if options.IncludeOwnedInformation {
		query.Set("include_owned_information", "true")
	}
	if options.FilterIds != nil {
		query.Set("filter_ids", strings.Join(options.FilterIds, ","))
	}
	if options.FilterResourceTypes != nil {
		query.Set("filter_resource_types", strings.Join(options.FilterResourceTypes, ","))
	}
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, deviceRepoUrl+"/export"+queryString, nil)
	if err != nil {
		return req, err
	}
	req.Header.Set("Authorization", token)
	return req, nil
}
