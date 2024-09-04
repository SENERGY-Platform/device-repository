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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, &FunctionsEndpoints{})
}

type FunctionsEndpoints struct{}

func (this *FunctionsEndpoints) ListControllingFunctions(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /controlling-functions", func(writer http.ResponseWriter, request *http.Request) {
		result, err, errCode := control.GetFunctionsByType(model.SES_ONTOLOGY_CONTROLLING_FUNCTION)
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

func (this *FunctionsEndpoints) ListMeasuringFunctions(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /measuring-functions", func(writer http.ResponseWriter, request *http.Request) {
		result, err, errCode := control.GetFunctionsByType(model.SES_ONTOLOGY_MEASURING_FUNCTION)
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

func (this *FunctionsEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /functions/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.GetFunction(id)
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

func (this *FunctionsEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /functions", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		function := models.Function{}
		err = json.NewDecoder(request.Body).Decode(&function)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		model.SetFunctionRdfType(&function)
		err, code := control.ValidateFunction(function)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

func (this *FunctionsEndpoints) ValidateDelete(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /functions/{id}", func(writer http.ResponseWriter, request *http.Request) {
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
		err, code := control.ValidateFunctionDelete(id)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
