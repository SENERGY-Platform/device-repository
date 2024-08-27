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
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, ExtendedHubEndpoints)
}

func ExtendedHubEndpoints(config config.Config, control Controller, router *httprouter.Router) {
	resource := "/extended-hubs"

	//use 'p' query parameter to limit selection to a permission;
	//		used internally to guarantee that user has needed permission for the resource
	//		example: 'p=x' guaranties the user has execution rights
	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		permission, err := model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if permission == models.UnsetPermissionFlag {
			permission = model.READ
		}
		result, err, errCode := control.ReadExtendedHub(id, util.GetAuthToken(request), permission)
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

	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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

		result, err, errCode := control.ListExtendedHubs(util.GetAuthToken(request), hubListOptions)
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
