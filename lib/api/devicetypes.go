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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, DeviceTypeEndpoints)
}

func DeviceTypeEndpoints(config config.Config, control Controller, router *jwt_http_router.Router) {
	resource := "/device-types"

	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		result, err, errCode := control.ReadDeviceType(id, jwt)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
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
	*/
	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
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

		result, err, errCode := control.ListDeviceTypes(jwt, limit, offset, sort)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})
}
