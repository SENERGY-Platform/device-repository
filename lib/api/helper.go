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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strings"
)

func init() {
	endpoints = append(endpoints, &HelperEndpoints{})
}

type HelperEndpoints struct{}

// Id godoc
// @Summary      transforms short id to long id
// @Description  transforms short id to long id
// @Tags         helper
// @Produce      json
// @Security Bearer
// @Param        short_id query string true "short id"
// @Param        prefix query string true "prefix added to generated long id"
// @Success      200 {object} string
// @Failure      400
// @Router       /helper/id [GET]
func (this *HelperEndpoints) Id(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /helper/id", func(writer http.ResponseWriter, request *http.Request) {
		shortId := strings.TrimSpace(request.URL.Query().Get("short_id"))
		prefix := strings.TrimSpace(request.URL.Query().Get("prefix"))
		uuidPart, err := models.LongId(shortId)
		result := prefix + uuidPart
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
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
