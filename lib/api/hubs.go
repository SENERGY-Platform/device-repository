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
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &HubEndpoints{})
}

type HubEndpoints struct{}

func (this *HubEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /hubs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		permission, err := model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if permission == models.UnsetPermissionFlag {
			permission = model.READ
		}
		result, err, errCode := control.ReadHub(id, util.GetAuthToken(request), permission)
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

func (this *HubEndpoints) List(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /hubs", func(writer http.ResponseWriter, request *http.Request) {
		hubListOptions := model.HubListOptions{
			Limit:  100,
			Offset: 0,
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			hubListOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			hubListOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				hubListOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				hubListOptions.Ids = []string{}
			}
		}

		hubListOptions.Search = request.URL.Query().Get("search")
		hubListOptions.SortBy = request.URL.Query().Get("sort")
		if hubListOptions.SortBy == "" {
			hubListOptions.SortBy = "name.asc"
		}

		if request.URL.Query().Has("connection-state") {
			searchedState := request.URL.Query().Get("connection-state")
			if !slices.Contains([]models.ConnectionState{models.ConnectionStateOnline, models.ConnectionStateOffline, models.ConnectionStateUnknown}, searchedState) {
				http.Error(writer, "invalid connection state:"+searchedState, http.StatusBadRequest)
				return
			}
			hubListOptions.ConnectionState = &searchedState
		}

		hubListOptions.Permission, err = model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if hubListOptions.Permission == models.UnsetPermissionFlag {
			hubListOptions.Permission = model.READ
		}

		result, err, errCode := control.ListHubs(util.GetAuthToken(request), hubListOptions)
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

func (this *HubEndpoints) GetDevices(config config.Config, router *http.ServeMux, control Controller) {
	//use 'p' query parameter to limit selection to a permission;
	//		used internally to guarantee that user has needed permission for the resource
	//		example: 'p=x' guaranties the user has execution rights
	//use 'as' query parameter to decide if a list of device.Id or device.LocalId should be returned
	//		default is LocalId
	//		allowed values are 'id', 'local_id' and 'localId'
	router.HandleFunc("GET /hubs/{id}/devices", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		permission, err := model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if permission == models.UnsetPermissionFlag {
			permission = model.READ
		}

		var asLocalId bool
		asParam := request.URL.Query().Get("as")
		switch asParam {
		case "":
			asLocalId = true
		case "id":
			asLocalId = false
		case "localId":
			asLocalId = true
		case "local_id":
			asLocalId = true
		default:
			http.Error(writer, "expect 'id', 'localId' or 'local_id' as value for 'as' query-parameter if it is used", http.StatusBadRequest)
			return
		}
		result, err, errCode := control.ListHubDeviceIds(id, util.GetAuthToken(request), permission, asLocalId)
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

func (this *HubEndpoints) Head(config config.Config, router *http.ServeMux, control Controller) {
	//use 'p' query parameter to limit selection to a permission;
	//		used internally to guarantee that user has needed permission for the resource
	//		example: 'p=x' guaranties the user has execution rights
	router.HandleFunc("HEAD /hubs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		permission, err := model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if permission == models.UnsetPermissionFlag {
			permission = model.READ
		}
		id := request.PathValue("id")
		_, _, errCode := control.ReadHub(id, util.GetAuthToken(request), permission)
		writer.WriteHeader(errCode)
		return
	})
}

func (this *HubEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /hubs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		hub := models.Hub{}
		err = json.NewDecoder(request.Body).Decode(&hub)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateHub(util.GetAuthToken(request), hub)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
