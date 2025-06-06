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
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, &InvalidElements{})
}

type InvalidElements struct{}

type ValidationError struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Error string `json:"error"`
}

// DeviceTypes godoc
// @Summary
// @Description  validate existing device-types
// @Tags         device-types
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100; limit of checked device-types not of returned errors"
// @Param        offset query integer false "default 0"
// @Param        sort query string false "default name.asc"
// @Param        allow_none_leaf_aspect_nodes_in_device_types query string false "allow none leaf aspect nodes in device-types"
// @Success      200 {array}  ValidationError
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /invalid/device-type [GET]
func (this *InvalidElements) DeviceTypes(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /invalid/device-types", func(writer http.ResponseWriter, request *http.Request) {
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
