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
	endpoints = append(endpoints, &ConceptEndpoints{})
}

type ConceptEndpoints struct{}

// ListConcepts godoc
// @Summary      list concepts
// @Description  list concepts
// @Tags         list, concepts
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Success      200 {array}  models.Concept
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/concepts [GET]
func (this *ConceptEndpoints) ListConcepts(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /v2/concepts", func(writer http.ResponseWriter, request *http.Request) {
		listoptions := model.ConceptListOptions{
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
		result, total, err, errCode := control.ListConcepts(listoptions)
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

// ListConceptsWithCharacteristics godoc
// @Summary      list concepts with characteristics
// @Description  list concepts with characteristics
// @Tags         list, concepts
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Success      200 {array}  models.ConceptWithCharacteristics
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/concepts-with-characteristics [GET]
func (this *ConceptEndpoints) ListConceptsWithCharacteristics(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /v2/concepts-with-characteristics", func(writer http.ResponseWriter, request *http.Request) {
		listoptions := model.ConceptListOptions{
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
		result, total, err, errCode := control.ListConceptsWithCharacteristics(listoptions)
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

// Get godoc
// @Summary      get concept
// @Description  get concept
// @Tags         get, concepts
// @Produce      json
// @Security Bearer
// @Param        id path string true "Concepts Id"
// @Param        sub-class query bool false "default=false; true -> returns models.ConceptWithCharacteristics; false -> returns models.Concept"
// @Success      200 {object}  models.Concept
// @Success      200 {object}  models.ConceptWithCharacteristics
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /concepts/{id} [GET]
func (this *ConceptEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /concepts/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		subClassStr := request.URL.Query().Get("sub-class")
		subClass := false
		var err error
		if subClassStr != "" {
			subClass, err = strconv.ParseBool(subClassStr)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
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

// Validate godoc
// @Summary      validate concept
// @Description  validate concept
// @Tags         validate, concepts
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.Concept true "Concept to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /concepts [PUT]
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

// ValidateDelete godoc
// @Summary      validate concepts delete
// @Description  validate if concept may be deleted
// @Tags         validate, concepts
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not a delete but a validation"
// @Param        id path string true "Concepts Id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /concepts/{id} [DELETE]
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
