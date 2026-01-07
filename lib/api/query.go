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
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"net/http"
)

func init() {
	endpoints = append(endpoints, &QueryEndpoint{})
}

type QueryEndpoint struct{}

// Query godoc
// @Summary      query used-in-device-type
// @Description  query used-in-device-type
// @Tags         device-types
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        message body model.UsedInDeviceTypeQuery true "filter"
// @Success      200 {object}  model.UsedInDeviceTypeResponse
// @Failure      400
// @Failure      404
// @Failure      500
// @Router       /query/used-in-device-type [POST]
func (this *QueryEndpoint) Query(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /query/used-in-device-type", func(writer http.ResponseWriter, request *http.Request) {
		query := model.UsedInDeviceTypeQuery{}
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err, errCode := control.GetUsedInDeviceType(query)
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
