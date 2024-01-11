/*
 * Copyright 2024 InfAI (CC SES)
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
	endpoints = append(endpoints, InvalidElements)
}

type ValidationError struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Error string `json:"error"`
}

func InvalidElements(config config.Config, control Controller, router *httprouter.Router) {
	resource := "/invalid"

	router.GET(resource+"/device-types", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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

		options, err := model.LoadDeviceTypeValidationOptions(request.URL.Query())
		if err != nil {
			http.Error(writer, "invalid validation options: "+err.Error(), http.StatusBadRequest)
			return
		}

		list, err, code := control.ListDeviceTypesV2(util.GetAuthToken(request), limit, offset, sort, nil, false, true)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		result := []ValidationError{}
		for _, e := range list {
			err, code = control.ValidateDeviceType(e, options)
			if err != nil {
				if code != http.StatusBadRequest {
					http.Error(writer, err.Error(), code)
					return
				}
				result = append(result, ValidationError{
					Id:    e.Id,
					Name:  e.Name,
					Error: err.Error(),
				})
			}
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

}
