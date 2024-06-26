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
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, DeviceEndpoints)
}

func DeviceEndpoints(config config.Config, control Controller, router *httprouter.Router) {
	resource := "/devices"

	//query-parameter:
	//		- limit: number; default 100, will be ignored if 'ids' is set
	//		- offset: number; default 0, will be ignored if 'ids' is set
	//		- ids: filter by comma seperated id list
	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		deviceListOptions := model.DeviceListOptions{
			Limit:  100,
			Offset: 0,
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			deviceListOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			deviceListOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				deviceListOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				deviceListOptions.Ids = []string{}
			}
		}

		result, err, errCode := control.ListDevices(util.GetAuthToken(request), deviceListOptions)
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

	//use 'as=local_id' query parameter to search device by local_id
	//		may use the 'owner_id' query parameter, which will default to the user/subject of the Auth-Token
	//use 'p' query parameter to limit selection to a permission;
	//		used internally to guarantee that user has needed permission for the resource
	//		example: 'p=x' guaranties the user has execution rights
	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		as := request.URL.Query().Get("as")
		ownerId := request.URL.Query().Get("owner_id")
		if ownerId == "" {
			token, err := jwt.GetParsedToken(request)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusUnauthorized)
				return
			}
			ownerId = token.GetUserId()
		}
		permission := model.AuthAction(request.URL.Query().Get("p"))
		if permission == "" {
			permission = model.READ
		}
		var result models.Device
		var err error
		var errCode int
		if as == "local_id" {
			result, err, errCode = control.ReadDeviceByLocalId(ownerId, id, util.GetAuthToken(request), permission)
		} else {
			result, err, errCode = control.ReadDevice(id, util.GetAuthToken(request), permission)
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
		device := models.Device{}
		err = json.NewDecoder(request.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateDevice(util.GetAuthToken(request), device)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

}
