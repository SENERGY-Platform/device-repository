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
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, DeviceTypeEndpoints)
}

func DeviceTypeEndpoints(config config.Config, control Controller, router *httprouter.Router) {
	resource := "/device-types"

	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		result, err, errCode := control.ReadDeviceType(id, util.GetAuthToken(request))
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

	/*
			query params:
			- limit: number; default 100
		    - offset: number; default 0
			- sort: <field>[.<direction>]; optional;
				- field: 'name', 'id'; defined at github.com/SENERGY-Platform/device-repository/lib/database/mongo/devicetype.go ListDeviceTypes()
				- direction: 'asc' || 'desc'; optional
				- examples:
					?sort=name.asc
					?sort=name
			- filter: json encoded []model.FilterCriteria; optional
					all criteria must be satisfied
			- interactions-filter: comma seperated list of interactions
					deprecated: use interactions field in filter (model.FilterCriteria.Interaction)
					if set: returns only device-types with at least one matching interaction on criteria matching services
					ignored if empty
			- include_id_modified: bool; add service-group modified device-types to result
	*/
	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var err error
		limitParam := request.URL.Query().Get("limit")
		var limit int64 = 100
		if limitParam != "" {
			limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		var offset int64 = 0
		if offsetParam != "" {
			offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		sort := request.URL.Query().Get("sort")
		if sort == "" {
			sort = "name.asc"
		}

		includeModifiedStr := request.URL.Query().Get("include_id_modified")
		includeModified := false
		if includeModifiedStr != "" {
			includeModified, err = strconv.ParseBool(includeModifiedStr)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		includeUnmodifiedStr := request.URL.Query().Get("include_id_unmodified")
		includeUnmodified := true
		if includeUnmodifiedStr != "" {
			includeUnmodified, err = strconv.ParseBool(includeUnmodifiedStr)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		filter := request.URL.Query().Get("filter")
		deviceTypesFilter := []model.FilterCriteria{}
		if filter != "" {
			err = json.Unmarshal([]byte(filter), &deviceTypesFilter)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		var result []models.DeviceType
		var errCode int
		interactionsFilterStr := request.URL.Query().Get("interactions-filter")
		if interactionsFilterStr != "" {
			interactionsFilter := []string{}
			for _, interaction := range strings.Split(interactionsFilterStr, ",") {
				interactionsFilter = append(interactionsFilter, strings.TrimSpace(interaction))
			}
			result, err, errCode = control.ListDeviceTypes(util.GetAuthToken(request), limit, offset, sort, deviceTypesFilter, interactionsFilter, includeModified, includeUnmodified)
		} else {
			result, err, errCode = control.ListDeviceTypesV2(util.GetAuthToken(request), limit, offset, sort, deviceTypesFilter, includeModified, includeUnmodified)
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

	router.PUT(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		options, err := model.LoadDeviceTypeValidationOptions(request.URL.Query())
		if err != nil {
			http.Error(writer, "invalid validation options: "+err.Error(), http.StatusBadRequest)
			return
		}
		dt := models.DeviceType{}
		err = json.NewDecoder(request.Body).Decode(&dt)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateDeviceType(dt, options)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

}
