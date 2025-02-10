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
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &ExtendedHubEndpoints{})
}

type ExtendedHubEndpoints struct{}

// Get godoc
// @Summary      get extended-hubs
// @Description  get extended-hubs
// @Tags         get, hubs, extended-hubs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Hub Id"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Success      200 {object}  models.ExtendedHub
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /extended-hubs/{id} [GET]
func (this *ExtendedHubEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /extended-hubs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
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
}

// List godoc
// @Summary      list extended-hubs
// @Description  list extended-hubs
// @Tags         list, hubs, extended-hubs
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Param        connection-state query integer false "filter; valid values are 'online', 'offline' and an empty string for unknown states"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Success      200 {array}  models.ExtendedHub
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /extended-hubs [GET]
func (this *ExtendedHubEndpoints) List(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /extended-hubs", func(writer http.ResponseWriter, request *http.Request) {
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

		hubListOptions.LocalDeviceId = request.URL.Query().Get("local-device-id")
		hubListOptions.OwnerId = request.URL.Query().Get("owner")

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

		result, total, err, errCode := control.ListExtendedHubs(util.GetAuthToken(request), hubListOptions)
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
