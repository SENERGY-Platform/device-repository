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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, &AspectEndpoints{})
}

type AspectEndpoints struct{}

// List godoc
// @Summary      list aspects
// @Description  list aspects
// @Tags         list, aspects
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
func (this *AspectEndpoints) List(config config.Config, router *http.ServeMux, control Controller) {
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
// @Tags         get, aspects
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
func (this *AspectEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
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
// @Tags         validate, aspects
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.Aspect true "Aspect to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /aspects [PUT]
func (this *AspectEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
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

// ValidateDelete godoc
// @Summary      validate aspect delete
// @Description  validate if aspect may be deleted
// @Tags         validate, aspects
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not a delete but a validation"
// @Param        id path string true "Aspect Id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /aspects/{id} [DELETE]
func (this *AspectEndpoints) ValidateDelete(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /aspects/{id}", func(writer http.ResponseWriter, request *http.Request) {
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
		err, code := control.ValidateAspectDelete(id)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// GetMeasuringFunctions godoc
// @Summary      list aspect measuring-functions
// @Description  list measuring-functions used in combination with this aspect
// @Tags         list, aspects, functions
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
func (this *AspectEndpoints) GetMeasuringFunctions(config config.Config, router *http.ServeMux, control Controller) {
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
