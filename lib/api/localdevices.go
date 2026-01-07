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

package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func init() {
	endpoints = append(endpoints, &LocalDevicesEndpoints{})
}

type LocalDevicesEndpoints struct{}

// List godoc
// @Summary      list devices (local-id variant)
// @Description  list devices (local-id variant)
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        ids query string false "comma separated list of local ids"
// @Param        owner_id query string false "defaults to requesting user; used in combination with local_id to find devices"
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        device-type-ids query string false "filter; comma-seperated list"
// @Param        attr-keys query string false "filter; comma-seperated list; lists elements only if they have an attribute key that is in the given list"
// @Param        attr-values query string false "filter; comma-seperated list; lists elements only if they have an attribute value that is in the given list"
// @Param        connection-state query integer false "filter; valid values are 'online', 'offline' and an empty string for unknown states"
// @Param        device-attribute-blacklist query string false "JSON encoded []models.Attribute, attribute value and origin will only be checked if set, otherwise all values or origins will be blacklisted"
// @Success      200 {array}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /local-devices/{id} [GET]
func (this *LocalDevicesEndpoints) List(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /local-devices", func(writer http.ResponseWriter, request *http.Request) {
		deviceListOptions := model.DeviceListOptions{
			Limit:  100,
			Offset: 0,
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			deviceListOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			deviceListOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				deviceListOptions.LocalIds = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				deviceListOptions.LocalIds = []string{}
			}
		}

		deviceListOptions.Owner = request.URL.Query().Get("owner")

		deviceTypeIdsParam := request.URL.Query().Get("device-type-ids")
		if request.URL.Query().Has("device-type-ids") {
			if deviceTypeIdsParam != "" {
				deviceListOptions.DeviceTypeIds = strings.Split(strings.TrimSpace(deviceTypeIdsParam), ",")
			} else {
				deviceListOptions.DeviceTypeIds = []string{}
			}
		}

		attrKeysParam := request.URL.Query().Get("attr-keys")
		if request.URL.Query().Has("attr-keys") {
			if attrKeysParam != "" {
				deviceListOptions.AttributeKeys = strings.Split(strings.TrimSpace(attrKeysParam), ",")
			} else {
				deviceListOptions.AttributeKeys = []string{}
			}
		}
		attrValuesParam := request.URL.Query().Get("attr-values")
		if request.URL.Query().Has("attr-values") {
			if attrValuesParam != "" {
				deviceListOptions.AttributeValues = strings.Split(strings.TrimSpace(attrValuesParam), ",")
			} else {
				deviceListOptions.AttributeValues = []string{}
			}
		}

		deviceListOptions.Search = request.URL.Query().Get("search")
		deviceListOptions.SortBy = request.URL.Query().Get("sort")
		if deviceListOptions.SortBy == "" {
			deviceListOptions.SortBy = "name.asc"
		}

		if request.URL.Query().Has("connection-state") {
			searchedState := request.URL.Query().Get("connection-state")
			if !slices.Contains([]models.ConnectionState{models.ConnectionStateOnline, models.ConnectionStateOffline, models.ConnectionStateUnknown}, searchedState) {
				http.Error(writer, "invalid connection state:"+searchedState, http.StatusBadRequest)
				return
			}
			deviceListOptions.ConnectionState = &searchedState
		}

		deviceListOptions.Permission, err = model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if deviceListOptions.Permission == models.UnsetPermissionFlag {
			deviceListOptions.Permission = model.READ
		}

		deviceAttributeBlacklistParam := request.URL.Query().Get("device-attribute-blacklist")
		if deviceAttributeBlacklistParam != "" {
			deviceAttributeBlacklistParam, err = url.QueryUnescape(deviceAttributeBlacklistParam)
			if err != nil {
				http.Error(writer, "unable to decode device-attribute-blacklist: "+err.Error(), http.StatusBadRequest)
				return
			}
			var blacklist []models.Attribute
			err = json.Unmarshal([]byte(deviceAttributeBlacklistParam), &blacklist)
			if err != nil {
				http.Error(writer, "unable to parse device-attribute-blacklist: "+err.Error(), http.StatusBadRequest)
				return
			}
			deviceListOptions.DeviceAttributeBlacklist = blacklist
		}

		result, err, errCode := control.ListDevices(util.GetAuthToken(request), deviceListOptions)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}

// Get godoc
// @Summary      get device by local id
// @Description  get device by local id
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Local Id"
// @Param        owner_id query string false "defaults to requesting user; used in combination with id to find device"
// @Success      200 {object}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /local-devices/{id} [GET]
func (this *LocalDevicesEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /local-devices/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token, err := jwt.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		ownerId := request.URL.Query().Get("owner_id")
		if ownerId == "" {
			ownerId = token.GetUserId()
		}
		result, err, errCode := control.ReadDeviceByLocalId(ownerId, id, token.Jwt(), model.READ)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}

// Create godoc
// @Summary      create device (local-id variant)
// @Description  create device (local-id variant)
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        message body models.Device true "element"
// @Success      200 {object}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /local-devices/{id} [POST]
func (this *LocalDevicesEndpoints) Create(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /local-devices", func(writer http.ResponseWriter, request *http.Request) {
		device := models.Device{}
		err := json.NewDecoder(request.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)

		if device.Id != "" {
			http.Error(writer, "body may not contain a preset id. please use the PUT method for updates", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.CreateDevice(token, device)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}

// Set godoc
// @Summary      set device (local-id variant)
// @Description  set device (local-id variant)
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Local Id"
// @Param        update-only-same-origin-attributes query string false "comma separated list; ensure that no attribute from another origin is overwritten"
// @Param        message body models.Device true "element"
// @Success      200 {object}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /local-devices/{id} [PUT]
func (this *LocalDevicesEndpoints) Set(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /local-devices/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token, err := jwt.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		ownerId := token.GetUserId()
		old, err, errCode := control.ReadDeviceByLocalId(ownerId, id, token.Jwt(), model.WRITE)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		id = old.Id

		device := models.Device{}
		err = json.NewDecoder(request.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if device.Id != "" && device.Id != id {
			http.Error(writer, "device contains a different id then the id from the url", http.StatusBadRequest)
			return
		}
		device.Id = id

		options := model.DeviceUpdateOptions{}
		if request.URL.Query().Has(UpdateOnlySameOriginAttributesKey) {
			temp := request.URL.Query().Get(UpdateOnlySameOriginAttributesKey)
			options.UpdateOnlySameOriginAttributes = strings.Split(temp, ",")
		}

		result, err, errCode := control.SetDevice(token.Jwt(), device, options)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}

// Delete godoc
// @Summary      delete device (local-id variant)
// @Description  delete device (local-id variant)
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Local Id"
// @Param        owner_id query string false "defaults to requesting user; used in combination with id to find device"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /local-devices/{id} [DELETE]
func (this *LocalDevicesEndpoints) Delete(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /local-devices/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token, err := jwt.GetParsedToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		ownerId := request.URL.Query().Get("owner_id")
		if ownerId == "" {
			ownerId = token.GetUserId()
		}
		old, err, errCode := control.ReadDeviceByLocalId(ownerId, id, token.Jwt(), model.ADMINISTRATE)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		id = old.Id

		err, errCode = control.DeleteDevice(token.Jwt(), id)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(true)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}
