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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, HubEndpoints)
}

func HubEndpoints(config config.Config, control Controller, router *jwt_http_router.Router) {
	resource := "/hubs"

	//use 'p' query parameter to limit selection to a permission;
	//		used internally to guarantee that user has needed permission for the resource
	//		example: 'p=x' guaranties the user has execution rights
	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		id := params.ByName("id")
		permission := model.AuthAction(request.URL.Query().Get("p"))
		if permission == "" {
			permission = model.READ
		}
		result, err, errCode := control.ReadHub(id, jwt, permission)
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

	//use 'p' query parameter to limit selection to a permission;
	//		used internally to guarantee that user has needed permission for the resource
	//		example: 'p=x' guaranties the user has execution rights
	router.HEAD(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		permission := model.AuthAction(request.URL.Query().Get("p"))
		if permission == "" {
			permission = model.READ
		}
		id := params.ByName("id")
		_, _, errCode := control.ReadHub(id, jwt, permission)
		writer.WriteHeader(errCode)
		return
	})

	router.PUT(resource, func(writer http.ResponseWriter, request *http.Request, params jwt_http_router.Params, jwt jwt_http_router.Jwt) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		hub := model.Hub{}
		err = json.NewDecoder(request.Body).Decode(&hub)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateHub(hub)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

}
