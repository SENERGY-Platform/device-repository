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
	endpoints = append(endpoints, &DeviceEndpoints{})
}

type DeviceEndpoints struct{}

func (this *DeviceEndpoints) List(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /devices", func(writer http.ResponseWriter, request *http.Request) {
		//query-parameter:
		//		- limit: number; default 100, will be ignored if 'ids' is set
		//		- offset: number; default 0, will be ignored if 'ids' is set
		//		- ids: filter by comma seperated id list

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
				deviceListOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				deviceListOptions.Ids = []string{}
			}
		}

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

		result, err, errCode := control.ListDevices(util.GetAuthToken(request), deviceListOptions)
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

func (this *DeviceEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /devices/{id}", func(writer http.ResponseWriter, request *http.Request) {
		//use 'as=local_id' query parameter to search device by local_id
		//		may use the 'owner_id' query parameter, which will default to the user/subject of the Auth-Token
		//use 'p' query parameter to limit selection to a permission;
		//		used internally to guarantee that user has needed permission for the resource
		//		example: 'p=x' guaranties the user has execution rights
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
		var result models.Device
		var errCode int
		if as == "local_id" {
			result, err, errCode = control.ReadDeviceByLocalId(ownerId, id, util.GetAuthToken(request), permission)
		} else {
			result, err, errCode = control.ReadDevice(id, util.GetAuthToken(request), permission)
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

func (this *DeviceEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /devices", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		device := models.Device{}
		err = json.NewDecoder(request.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateDevice(util.GetAuthToken(request), device)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
