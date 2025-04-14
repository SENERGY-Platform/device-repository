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
	endpoints = append(endpoints, &AspectEndpoints{})
}

type AspectEndpoints struct{}

// ListAspects godoc
// @Summary      list aspects
// @Description  list aspects
// @Tags         aspects
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Success      200 {array}  models.Aspect
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/aspects [GET]
func (this *AspectEndpoints) ListAspects(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /v2/aspects", func(writer http.ResponseWriter, request *http.Request) {
		listoptions := model.AspectListOptions{
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
		result, total, err, errCode := control.ListAspects(listoptions)
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
// @Deprecated
// @Summary      deprecated list aspects; use GET /v2/aspects
// @Description  deprecated list aspects; use GET /v2/aspects
// @Tags         aspects
// @Produce      json
// @Security Bearer
// @Param        function query string false "filter; only 'measuring-function' is a valid value; if set, returns aspects used in combination with measuring-functions"
// @Param        ancestors query bool false "filter; in combination with 'function'; if true, returns also ancestor nodes of matching aspects"
// @Param        descendants query bool false "filter; in combination with 'function'; if true, returns also descendant nodes of matching aspects"
// @Success      200 {array}  models.Aspect
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /aspects [GET]
func (this *AspectEndpoints) List(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /aspects", func(writer http.ResponseWriter, request *http.Request) {
		var result []models.Aspect
		var err error
		var errCode int

		function := request.URL.Query().Get("function")

		if function == "" {
			result, err, errCode = control.GetAspects()
			if err != nil {
				http.Error(writer, err.Error(), errCode)
				return
			}
		} else {
			ancestors := false
			descendants := true
			ancestorsQuery := request.URL.Query().Get("ancestors")
			if ancestorsQuery != "" {
				ancestors, err = strconv.ParseBool(ancestorsQuery)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusBadRequest)
					return
				}
			}
			descendantsQuery := request.URL.Query().Get("descendants")
			if descendantsQuery != "" {
				descendants, err = strconv.ParseBool(descendantsQuery)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusBadRequest)
					return
				}
			}
			if function == "measuring-function" {
				result, err, errCode = control.GetAspectsWithMeasuringFunction(ancestors, descendants)
				if err != nil {
					http.Error(writer, err.Error(), errCode)
					return
				}
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

// Get godoc
// @Summary      get aspect
// @Description  get aspect
// @Tags         aspects
// @Produce      json
// @Security Bearer
// @Param        id path string true "Aspect Id"
// @Success      200 {object}  models.Aspect
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /aspects/{id} [GET]
func (this *AspectEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /aspects/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.GetAspect(id)
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
// @Summary      validate aspect
// @Description  validate aspect
// @Tags         aspects
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.Aspect true "Aspect to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /aspects [PUT]
func (this *AspectEndpoints) Validate(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /aspects", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		aspect := models.Aspect{}
		err = json.NewDecoder(request.Body).Decode(&aspect)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateAspect(aspect)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// Set godoc
// @Summary      set aspect
// @Description  set aspect
// @Tags         aspects
// @Produce      json
// @Security Bearer
// @Param        id path string true "Aspect Id"
// @Param        message body models.Aspect true "element"
// @Success      200 {object}  models.Aspect
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /aspects/{id} [PUT]
func (this *AspectEndpoints) Set(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /aspects/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		aspect := models.Aspect{}
		err := json.NewDecoder(request.Body).Decode(&aspect)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)

		if aspect.Id != id {
			http.Error(writer, "id in body unequal to id in request endpoint", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetAspect(token, aspect)
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

// Create godoc
// @Summary      create aspect
// @Description  create aspect with generated id
// @Tags         aspects
// @Produce      json
// @Security Bearer
// @Param        wait query bool false "wait for done message in kafka before responding"
// @Param        message body models.Aspect true "element"
// @Success      200 {object}  models.Aspect
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /aspects [POST]
func (this *AspectEndpoints) Create(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /aspects", func(writer http.ResponseWriter, request *http.Request) {
		aspect := models.Aspect{}
		err := json.NewDecoder(request.Body).Decode(&aspect)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)

		result, err, errCode := control.SetAspect(token, aspect)
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

// DeleteAspect godoc
// @Summary      delete aspect
// @Description  delete aspect; may only be called by admins; can also be used to only validate deletes
// @Tags         aspects
// @Security Bearer
// @Param        dry-run query bool false "only validate deletion"
// @Param        id path string true "Aspect Id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /aspects/{id} [DELETE]
func (this *AspectEndpoints) DeleteAspect(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /aspects/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		dryRun := false
		if request.URL.Query().Has("dry-run") {
			var err error
			dryRun, err = strconv.ParseBool(request.URL.Query().Get("dry-run"))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if dryRun {
			err, code := control.ValidateAspectDelete(id)
			if err != nil {
				http.Error(writer, err.Error(), code)
				return
			}
			writer.WriteHeader(http.StatusOK)
			return
		}
		token := util.GetAuthToken(request)
		err, errCode := control.DeleteAspect(token, id)
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

// GetMeasuringFunctions godoc
// @Summary      list aspect measuring-functions
// @Description  list measuring-functions used in combination with this aspect
// @Tags         aspects
// @Produce      json
// @Security Bearer
// @Param        id path string true "Aspect Id"
// @Success      200 {array}  models.Function
// @Param        ancestors query bool false "filter; if true, returns also functions used in combination with ancestors of the input aspect"
// @Param        descendants query bool false "filter; if true, returns also functions used in combination with descendants of the input aspect"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /aspects/{id}/measuring-functions [GET]
func (this *AspectEndpoints) GetMeasuringFunctions(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /aspects/{id}/measuring-functions", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		ancestors := false
		descendants := true
		var err error
		ancestorsQuery := request.URL.Query().Get("ancestors")
		if ancestorsQuery != "" {
			ancestors, err = strconv.ParseBool(ancestorsQuery)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		descendantsQuery := request.URL.Query().Get("descendants")
		if descendantsQuery != "" {
			descendants, err = strconv.ParseBool(descendantsQuery)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		result, err, errCode := control.GetAspectNodesMeasuringFunctions(id, ancestors, descendants)
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
