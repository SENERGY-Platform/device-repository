/*
 * Copyright 2025 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
)

func init() {
	endpoints = append(endpoints, &DefaultsEndpoints{})
}

type DefaultsEndpoints struct{}

// GetDefaultDeviceAttributes godoc
// @Summary      get default device attributes
// @Description  get default attributes for devices where the owner is the requesting user
// @Tags         defaults
// @Produce      json
// @Security Bearer
// @Success      200 {array}  models.Attribute
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /defaults/devices/attributes [GET]
func (this *DefaultsEndpoints) GetDefaultDeviceAttributes(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /defaults/devices/attributes", func(writer http.ResponseWriter, request *http.Request) {
		result, err, errCode := control.GetDefaultDeviceAttributes(util.GetAuthToken(request))
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

// SetDefaultDeviceAttributes godoc
// @Summary      set default device attributes
// @Description  set default attributes for devices where the owner is the requesting user
// @Tags         defaults
// @Accept       json
// @Security Bearer
// @Param        attributes body []models.Attribute true "attributes"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /defaults/devices/attributes [PUT]
func (this *DefaultsEndpoints) SetDefaultDeviceAttributes(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /defaults/devices/attributes", func(writer http.ResponseWriter, request *http.Request) {
		var attributes []models.Attribute
		err := json.NewDecoder(request.Body).Decode(&attributes)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, errCode := control.SetDefaultDeviceAttributes(util.GetAuthToken(request), attributes)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.WriteHeader(http.StatusOK)
		return
	})
}
