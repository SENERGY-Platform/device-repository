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
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, &DeviceGroupEndpoints{})
}

type DeviceGroupEndpoints struct{}

func (this *DeviceGroupEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /device-groups/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")

		//ref https://bitnify.atlassian.net/browse/SNRGY-3027
		filterGenericDuplicateCriteria := request.URL.Query().Get("filter_generic_duplicate_criteria") == "true"

		result, err, errCode := control.ReadDeviceGroup(id, util.GetAuthToken(request), filterGenericDuplicateCriteria)
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

func (this *DeviceGroupEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /device-groups", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		group := models.DeviceGroup{}
		err = json.NewDecoder(request.Body).Decode(&group)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateDeviceGroup(group)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		err, code = control.CheckAccessToDevicesOfGroup(util.GetAuthToken(request), group)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
