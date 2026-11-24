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
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/v3/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"strings"
)

// MaxDeleteManyIds is the maximum number of ids in the body of the DELETE /<resource> endpoints
const MaxDeleteManyIds = 10000

// deleteManyMaxBodyBytes bounds the body that is read before the number of ids can be checked
const deleteManyMaxBodyBytes = 8 << 20

// handleDeleteMany registers 'DELETE <path>' with a json list of ids as body.
// the controller method checks rights and preconditions of all ids before it deletes the first one.
func handleDeleteMany(config configuration.Config, router *http.ServeMux, path string, control Controller, deleteMany func(control Controller, ctx context.Context, token string, ids []string) (error, int)) {
	router.HandleFunc("DELETE "+path, func(writer http.ResponseWriter, request *http.Request) {
		ids := []string{}
		err := json.NewDecoder(http.MaxBytesReader(writer, request.Body, deleteManyMaxBodyBytes)).Decode(&ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if len(ids) > MaxDeleteManyIds {
			http.Error(writer, "too many ids: at most 10000 ids may be deleted with one request", http.StatusBadRequest)
			return
		}
		err, errCode := deleteMany(control, request.Context(), util.GetAuthToken(request), ids)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(true)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
	})
}

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
func (this *HelperEndpoints) Id(config configuration.Config, router *http.ServeMux, control Controller) {
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
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}
