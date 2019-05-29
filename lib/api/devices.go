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
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
)

func init() {
	endpoints = append(endpoints, DeviceEndpoints)
}

func DeviceEndpoints(config config.Config, control Controller, router *jwt_http_router.Router) {
	resource := "/devices"

	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		result, err, errCode := control.ReadDevice(id, jwt)
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
			- permission: 'r' || 'w' || 'x' || 'x'; default 'r'
			- sort: <field>[.<direction>]; optional;
				- field: declared by https://github.com/SENERGY-Platform/permission-search/config.json -> Resources.deviceinstance.Features[*].Name
				- direction: 'asc' || 'desc'; optional
				- examples:
					?sort=name.asc
					?sort=name
	*/
	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		options, err := listoptions.FromQueryParameter(request, 100, 0)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		options.Strict()
		result, err, errCode := control.ListDevices(jwt, options)
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

	router.PUT(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.POST(resource, func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})
}
