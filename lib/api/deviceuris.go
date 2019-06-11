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
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
)

func init() {
	endpoints = append(endpoints, DeviceUrisEndpoints)
}

//view of device instances where uri is used as id
func DeviceUrisEndpoints(config config.Config, control Controller, router *jwt_http_router.Router) {

	resource := "/device-uris"

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
	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, ps jwt_http_router.Params, jwt jwt_http_router.Jwt) {
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

	/*
		query params:
		- permission: 'r' || 'w' || 'x' || 'x'; default 'r'
	*/
	router.GET(resource+"/:uri", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		uri := params.ByName("uri")
		p := request.URL.Query().Get("permission")
		result, err, errCode := control.ReadDeviceByUri(uri, p, jwt)
		if err != nil {
			log.Println("DEBUG: unknown uri", uri)
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
		- permission: 'r' || 'w' || 'x' || 'x'; default 'r'
	*/
	router.HEAD(resource+"/:uri", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		uri := params.ByName("uri")
		p := request.URL.Query().Get("permission")
		result, err, errCode := control.ReadDeviceByUri(uri, p, jwt)
		if err != nil {
			log.Println("DEBUG: unknown uri", uri)
			http.Error(writer, err.Error(), errCode)
			return
		}
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

	if config.Commands {
		router.POST(resource, func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
			device := model.DeviceInstance{}
			err := json.NewDecoder(request.Body).Decode(&device)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
			result, err, errCode := control.PublishDeviceCreate(jwt, device)
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
			update device instance by uri
			id, uri, gateway, user-tags and image in body will be ignored
		*/
		router.PUT(resource+"/:uri", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
			uri := params.ByName("uri")
			device := model.DeviceInstance{}
			err := json.NewDecoder(request.Body).Decode(&device)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
			result, err, errCode := control.PublishDeviceUriUpdate(jwt, uri, device)
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

		router.DELETE(resource+"/:uri", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
			uri := params.ByName("uri")
			err, errCode := control.PublishDeviceUriDelete(jwt, uri)
			if err != nil {
				http.Error(writer, err.Error(), errCode)
				return
			}
			err = json.NewEncoder(writer).Encode(true)
			if err != nil {
				log.Println("ERROR: unable to encode response", err)
			}
			return
		})
	}

}
