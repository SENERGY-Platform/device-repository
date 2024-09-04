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
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, &ConceptEndpoints{})
}

type ConceptEndpoints struct{}

func (this *ConceptEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /concepts/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		subClass, err := strconv.ParseBool(request.URL.Query().Get("sub-class"))
		resultConceptWithCharacteristics := models.ConceptWithCharacteristics{}
		resultConcept := models.Concept{}
		errCode := 0
		if subClass {
			resultConceptWithCharacteristics, err, errCode = control.GetConceptWithCharacteristics(id)
		} else {
			resultConcept, err, errCode = control.GetConceptWithoutCharacteristics(id)
		}
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		if subClass {
			err = json.NewEncoder(writer).Encode(resultConceptWithCharacteristics)
		} else {
			err = json.NewEncoder(writer).Encode(resultConcept)
		}
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})
}

func (this *ConceptEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /concepts", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		concept := models.Concept{}
		err = json.NewDecoder(request.Body).Decode(&concept)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateConcept(concept)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

func (this *ConceptEndpoints) ValidateDelete(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /concepts/{id}", func(writer http.ResponseWriter, request *http.Request) {
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
		err, code := control.ValidateConceptDelete(id)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
