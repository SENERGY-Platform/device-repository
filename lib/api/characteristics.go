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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &CharacteristicsEndpoints{})
}

type CharacteristicsEndpoints struct{}

// ListCharacteristics godoc
// @Summary      list characteristics
// @Description  list characteristics
// @Tags         list, characteristics
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Success      200 {array}  models.Characteristic
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/characteristics [GET]
func (this *CharacteristicsEndpoints) ListCharacteristics(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /v2/characteristics", func(writer http.ResponseWriter, request *http.Request) {
		listoptions := model.CharacteristicListOptions{
			Limit:  100,
			Offset: 0,
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			listoptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			listoptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				listoptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				listoptions.Ids = []string{}
			}
		}

		listoptions.Search = request.URL.Query().Get("search")
		listoptions.SortBy = request.URL.Query().Get("sort")
		if listoptions.SortBy == "" {
			listoptions.SortBy = "name.asc"
		}
		result, total, err, errCode := control.ListCharacteristics(listoptions)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}

		writer.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})
}

// List godoc
// @Summary      list characteristics
// @Description  list characteristics
// @Tags         list, characteristics
// @Produce      json
// @Security Bearer
// @Param        leafsOnly query bool false "default=true; filter; return only characteristics that have no sub-characteristics"
// @Success      200 {array}  models.Characteristic
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /characteristics [GET]
func (this *CharacteristicsEndpoints) List(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /characteristics", func(writer http.ResponseWriter, request *http.Request) {
		leafsOnlyStr := request.URL.Query().Get("leafsOnly")
		leafsOnly := true
		var err error
		if leafsOnlyStr != "" {
			leafsOnly, err = strconv.ParseBool(leafsOnlyStr)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		result, err, errCode := control.GetCharacteristics(leafsOnly)
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

// Get godoc
// @Summary      get characteristic
// @Description  get characteristic
// @Tags         get, characteristics
// @Produce      json
// @Security Bearer
// @Param        id path string true "Characteristic Id"
// @Success      200 {object}  models.Characteristic
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /characteristics/{id} [GET]
func (this *CharacteristicsEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /characteristics/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.GetCharacteristic(id)
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

// Validate godoc
// @Summary      validate characteristic
// @Description  validate characteristic
// @Tags         validate, characteristics
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.Characteristic true "Characteristic to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /characteristics [PUT]
func (this *CharacteristicsEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /characteristics", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		characteristic := models.Characteristic{}
		err = json.NewDecoder(request.Body).Decode(&characteristic)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateCharacteristics(characteristic)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// ValidateDelete godoc
// @Summary      validate characteristic delete
// @Description  validate if characteristic may be deleted
// @Tags         validate, characteristics
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not a delete but a validation"
// @Param        id path string true "Characteristics Id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /characteristics/{id} [DELETE]
func (this *CharacteristicsEndpoints) ValidateDelete(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /characteristics/{id}", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		id := request.PathValue("id")
		err, code := control.ValidateCharacteristicDelete(id)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
