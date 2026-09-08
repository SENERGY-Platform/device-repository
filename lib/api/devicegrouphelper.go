/*
 * Copyright 2026 InfAI (CC SES)
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
	"net/http"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/v2/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/v2/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v2/lib/model"
)

func init() {
	endpoints = append(endpoints, &DeviceGroupHelperEndpoints{})
}

type DeviceGroupHelperEndpoints struct{}

// Helper godoc
// @Summary      device-group helper
// @Description  helps to build a valid device-group: returns the criteria a device-group of the given devices would have, and the devices that may be added to it
// @Tags         device-groups, helper
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100"
// @Param        offset query integer false "default 0"
// @Param        search query string false "filter on the listed options"
// @Param        maintains_group_usability query bool false "filter; list only options that leave the group with at least one counting criteria"
// @Param        function_block_list query string false "comma-separated list of function ids that do not count towards keeping the group usable"
// @Param        message body []string true "device id list"
// @Success      200 {object}  model.DeviceGroupHelperResult
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-group-helper [POST]
func (this *DeviceGroupHelperEndpoints) Helper(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /device-group-helper", func(writer http.ResponseWriter, request *http.Request) {
		deviceIds := []string{}
		err := json.NewDecoder(request.Body).Decode(&deviceIds)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		options := model.DeviceGroupHelperOptions{
			Limit:  100,
			Offset: 0,
			Search: request.URL.Query().Get("search"),
		}

		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			options.Limit, err = strconv.ParseInt(limitParam, 10, 64)
			if err != nil {
				http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
				return
			}
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			options.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
			if err != nil {
				http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
				return
			}
		}

		maintainsParam := request.URL.Query().Get("maintains_group_usability")
		if maintainsParam != "" {
			options.MaintainsGroupUsability, err = strconv.ParseBool(maintainsParam)
			if err != nil {
				http.Error(writer, "unable to parse maintains_group_usability:"+err.Error(), http.StatusBadRequest)
				return
			}
		}

		functionBlockList := request.URL.Query().Get("function_block_list")
		if functionBlockList != "" {
			options.FunctionBlockList = strings.Split(functionBlockList, ",")
		}

		result, err, errCode := control.DeviceGroupHelper(util.GetAuthToken(request), deviceIds, options)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}
