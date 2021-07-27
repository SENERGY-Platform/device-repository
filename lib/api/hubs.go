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
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, HubEndpoints)
}

func HubEndpoints(config config.Config, control Controller, router *httprouter.Router) {
	resource := "/hubs"

	//use 'p' query parameter to limit selection to a permission;
	//		used internally to guarantee that user has needed permission for the resource
	//		example: 'p=x' guaranties the user has execution rights
	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		permission := model.AuthAction(request.URL.Query().Get("p"))
		if permission == "" {
			permission = model.READ
		}
		result, err, errCode := control.ReadHub(id, util.GetAuthToken(request), permission)
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
	//use 'as' query parameter to decide if a list of device.Id or device.LocalId should be returned
	//		default is LocalId
	//		allowed values are 'id', 'local_id' and 'localId'
	router.GET(resource+"/:id/devices", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		permission := model.AuthAction(request.URL.Query().Get("p"))
		if permission == "" {
			permission = model.READ
		}

		var asLocalId bool
		asParam := request.URL.Query().Get("as")
		switch asParam {
		case "":
			asLocalId = true
		case "id":
			asLocalId = false
		case "localId":
			asLocalId = true
		case "local_id":
			asLocalId = true
		default:
			http.Error(writer, "expect 'id', 'localId' or 'local_id' as value for 'as' query-parameter if it is used", http.StatusBadRequest)
			return
		}
		result, err, errCode := control.ListHubDeviceIds(id, util.GetAuthToken(request), permission, asLocalId)
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
	router.HEAD(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		permission := model.AuthAction(request.URL.Query().Get("p"))
		if permission == "" {
			permission = model.READ
		}
		id := params.ByName("id")
		_, _, errCode := control.ReadHub(id, util.GetAuthToken(request), permission)
		writer.WriteHeader(errCode)
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
