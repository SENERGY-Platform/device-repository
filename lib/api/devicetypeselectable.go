/*
 * Copyright 2022 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, DeviceTypeSelectableEndpoints)
}

func DeviceTypeSelectableEndpoints(config config.Config, control Controller, router *httprouter.Router) {
	resource := "/device-type-selectables"

	router.POST("/query"+resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		query := []model.FilterCriteria{}
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		pathPrefix := request.URL.Query().Get("path-prefix")
		interactionsFilter := []string{}
		interactionsFilterStr := request.URL.Query().Get("interactions-filter")
		if interactionsFilterStr != "" {
			for _, interaction := range strings.Split(interactionsFilterStr, ",") {
				interactionsFilter = append(interactionsFilter, strings.TrimSpace(interaction))
			}
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
		result, err, errCode := control.GetDeviceTypeSelectables(query, pathPrefix, interactionsFilter, includeModified)
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

	router.POST("/v2/query"+resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		query := []model.FilterCriteria{}
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		pathPrefix := request.URL.Query().Get("path-prefix")
		includeModifiedStr := request.URL.Query().Get("include_id_modified")
		includeModified := false
		if includeModifiedStr != "" {
			includeModified, err = strconv.ParseBool(includeModifiedStr)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		result, err, errCode := control.GetDeviceTypeSelectablesV2(query, pathPrefix, includeModified)
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
