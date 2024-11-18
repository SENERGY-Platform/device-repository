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

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &ExtendedDeviceEndpoints{})
}

type ExtendedDeviceEndpoints struct{}

// List godoc
// @Summary      list extended-device
// @Description  list extended-device
// @Tags         list, devices, extended-devices
// @Produce      json
// @Security Bearer
// @Param        fulldt query bool false "if true, result contains full device-type"
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Param        local_ids query string false "in combination with owner; fills ids filter; comma-seperated list"
// @Param        owner query string false "used in combination with local_ids to fill ids filter; defaults to requesting user"
// @Param        device-type-ids query string false "filter; comma-seperated list"
// @Param        attr-keys query string false "filter; comma-seperated list; lists elements only if they have an attribute key that is in the given list"
// @Param        attr-values query string false "filter; comma-seperated list; lists elements only if they have an attribute value that is in the given list"
// @Param        connection-state query integer false "filter; valid values are 'online', 'offline' and an empty string for unknown states"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Success      200 {array}  models.ExtendedDevice
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /extended-devices [GET]
func (this *ExtendedDeviceEndpoints) List(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /extended-devices", func(writer http.ResponseWriter, request *http.Request) {
		deviceListOptions := model.ExtendedDeviceListOptions{
			Limit:  100,
			Offset: 0,
			FullDt: request.URL.Query().Get("fulldt") == "true",
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
				deviceListOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				deviceListOptions.Ids = []string{}
			}
		}

		localIdsParam := request.URL.Query().Get("local_ids")
		if request.URL.Query().Has("local_ids") {
			if localIdsParam != "" {
				deviceListOptions.LocalIds = strings.Split(strings.TrimSpace(localIdsParam), ",")
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

		result, total, err, errCode := control.ListExtendedDevices(util.GetAuthToken(request), deviceListOptions)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})
}

// Get godoc
// @Summary      get extended-device
// @Description  get extended-device
// @Tags         get, devices, extended-devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Id"
// @Param        as query string false "interprets the id as local_id if as=='local_id'"
// @Param        owner_id query string false "default requesting user; used in combination with local_id (as=='local_id') to identify the device"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Param        fulldt query bool false "if true, result contains full device-type"
// @Success      200 {object}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /extended-devices/{id} [GET]
func (this *ExtendedDeviceEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /extended-devices/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		as := request.URL.Query().Get("as")
		ownerId := request.URL.Query().Get("owner_id")
		if ownerId == "" {
			token, err := jwt.GetParsedToken(request)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusUnauthorized)
				return
			}
			ownerId = token.GetUserId()
		}
		permission, err := model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if permission == models.UnsetPermissionFlag {
			permission = model.READ
		}
		fulldt := request.URL.Query().Get("fulldt") == "true"
		var result models.ExtendedDevice
		var errCode int
		if as == "local_id" {
			result, err, errCode = control.ReadExtendedDeviceByLocalId(ownerId, id, util.GetAuthToken(request), permission, fulldt)
		} else {
			result, err, errCode = control.ReadExtendedDevice(id, util.GetAuthToken(request), permission, fulldt)
		}
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})
}
