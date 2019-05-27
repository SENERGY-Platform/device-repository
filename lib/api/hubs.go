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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
)

func init() {
	endpoints = append(endpoints, HubEndpoints)
}

func HubEndpoints(config config.Config, control Controller, router *jwt_http_router.Router) {

	resource := "/hubs"

	router.POST(resource, func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.HEAD(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.PUT(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.DELETE(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.GET(resource+"/:id/name", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.PUT(resource+"/:id/name", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.GET(resource+"/:id/hash", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.PUT(resource+"/:id/hash", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.GET(resource+"/:id/devices", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})

	router.PUT(resource+"/:id/devices", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		//TODO
		http.Error(writer, "not implemented", http.StatusNotImplemented)
	})
}
