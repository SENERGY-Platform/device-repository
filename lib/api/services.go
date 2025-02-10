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
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"log"
	"net/http"
)

func init() {
	endpoints = append(endpoints, &ServiceEndpoints{})
}

type ServiceEndpoints struct{}

// Get godoc
// @Summary      get service
// @Description  get service
// @Tags         get, services
// @Produce      json
// @Security Bearer
// @Param        id path string true "Service Id"
// @Success      200 {object}  models.Service
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /services/{id} [GET]
func (this *ServiceEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /services/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.GetService(id)
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
}
