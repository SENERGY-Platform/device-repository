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

	router.GET(resource+"/:uri", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		uri := params.ByName("uri")
		result, err, errCode := control.ReadDeviceByUri(uri, jwt)
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

	router.POST(resource, func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	/*
		update device instance by uri
		id, uri, gateway, user-tags and image in body will be ignored
	*/
	router.PUT(resource+"/:uri", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.DELETE(resource+"/:uri", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.HEAD(resource+"/:uri", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

}
