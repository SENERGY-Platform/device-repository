/*
 *
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *
 */

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &LocationEndpoints{})
}

type LocationEndpoints struct{}

// Get godoc
// @Summary      get location
// @Description  get location
// @Tags         get, locations
// @Produce      json
// @Security Bearer
// @Param        id path string true "Location Id"
// @Success      200 {object}  models.Location
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /locations/{id} [GET]
func (this *LocationEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /locations/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.GetLocation(id, util.GetAuthToken(request))
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
// @Summary      validate location
// @Description  validate location
// @Tags         validate, locations
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.Location true "Location to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /functions [PUT]
func (this *LocationEndpoints) Validate(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /locations", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		location := models.Location{}
		err = json.NewDecoder(request.Body).Decode(&location)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateLocation(location)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// List godoc
// @Summary      list location
// @Description  list location
// @Tags         list, locations
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Success      200 {array}  models.Location
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /locations [GET]
func (this *LocationEndpoints) List(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /locations", func(writer http.ResponseWriter, request *http.Request) {
		locationListOptions := model.LocationListOptions{
			Limit:  100,
			Offset: 0,
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			locationListOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			locationListOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				locationListOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				locationListOptions.Ids = []string{}
			}
		}

		locationListOptions.Search = request.URL.Query().Get("search")
		locationListOptions.SortBy = request.URL.Query().Get("sort")
		if locationListOptions.SortBy == "" {
			locationListOptions.SortBy = "name.asc"
		}

		locationListOptions.Permission, err = model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if locationListOptions.Permission == models.UnsetPermissionFlag {
			locationListOptions.Permission = model.READ
		}

		result, total, err, errCode := control.ListLocations(util.GetAuthToken(request), locationListOptions)
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

// Create godoc
// @Summary      create location
// @Description  create location
// @Tags         create, locations
// @Produce      json
// @Security Bearer
// @Param        id path string true "Location Id"
// @Param        message body models.Location true "element"
// @Success      200 {object}  models.Location
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /locations [POST]
func (this *LocationEndpoints) Create(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /locations", func(writer http.ResponseWriter, request *http.Request) {
		location := models.Location{}
		err := json.NewDecoder(request.Body).Decode(&location)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)
		if location.Id != "" {
			http.Error(writer, "location may not contain a preset id. please use PUT to update a location", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetLocation(token, location)
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

// Set godoc
// @Summary      set location
// @Description  set location
// @Tags         set, locations
// @Produce      json
// @Security Bearer
// @Param        id path string true "Location Id"
// @Param        message body models.Location true "element"
// @Success      200 {object}  models.Location
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /locations/{id} [PUT]
func (this *LocationEndpoints) Set(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /locations/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		location := models.Location{}
		err := json.NewDecoder(request.Body).Decode(&location)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if location.Id != id {
			http.Error(writer, "id in body unequal to id in request endpoint", http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)

		result, err, errCode := control.SetLocation(token, location)
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

// Delete godoc
// @Summary      delete location
// @Description  delete location
// @Tags         delete, locations
// @Produce      json
// @Security Bearer
// @Param        id path string true "Location Id"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /locations/{id} [DELETE]
func (this *LocationEndpoints) Delete(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /locations/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token := util.GetAuthToken(request)

		err, errCode := control.DeleteLocation(token, id)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(true)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})
}
