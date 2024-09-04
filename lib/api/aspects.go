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
