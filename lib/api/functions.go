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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &FunctionsEndpoints{})
}

type FunctionsEndpoints struct{}

// ListFunctions godoc
// @Summary      list functions
// @Description  list functions
// @Tags         list, functions
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        rdf_type query string false "filter; https://senergy.infai.org/ontology/ControllingFunction || https://senergy.infai.org/ontology/MeasuringFunction"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Success      200 {array}  models.Function
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /functions [GET]
func (this *FunctionsEndpoints) ListFunctions(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /functions", func(writer http.ResponseWriter, request *http.Request) {
		listoptions := model.FunctionListOptions{
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

		listoptions.RdfType = request.URL.Query().Get("rdf_type")

		listoptions.Search = request.URL.Query().Get("search")
		listoptions.SortBy = request.URL.Query().Get("sort")
		if listoptions.SortBy == "" {
			listoptions.SortBy = "name.asc"
		}
		result, total, err, errCode := control.ListFunctions(listoptions)
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

// ListControllingFunctions godoc
// @Summary      list controlling-functions
// @Description  list controlling-functions
// @Tags         list, controlling-functions, functions
// @Produce      json
// @Security Bearer
// @Success      200 {array}  models.Function
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /controlling-functions [GET]
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

// ListMeasuringFunctions godoc
// @Summary      list measuring-functions
// @Description  list measuring-functions
// @Tags         list, measuring-functions, functions
// @Produce      json
// @Security Bearer
// @Success      200 {array}  models.Function
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /measuring-functions [GET]
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

// Get godoc
// @Summary      get function
// @Description  get function
// @Tags         get, functions
// @Produce      json
// @Security Bearer
// @Param        id path string true "Function Id"
// @Success      200 {object}  models.DeviceClass
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /functions/{id} [GET]
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

// Validate godoc
// @Summary      validate function
// @Description  validate function
// @Tags         validate, functions
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.Function true "Function to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /functions [PUT]
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

// Delete godoc
// @Summary      delete function
// @Description  delete function; may only be called by admins; can also be used to only validate deletes
// @Tags         validate, functions
// @Security Bearer
// @Param        dry-run query bool false "only validate deletion"
// @Param        id path string true "Functions Id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /functions/{id} [DELETE]
func (this *FunctionsEndpoints) Delete(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /functions/{id}", func(writer http.ResponseWriter, request *http.Request) {
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
			err, code := control.ValidateFunctionDelete(id)
			if err != nil {
				http.Error(writer, err.Error(), code)
				return
			}
			writer.WriteHeader(http.StatusOK)
			return
		}
		token := util.GetAuthToken(request)
		err, code := control.DeleteFunction(token, id)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// Create godoc
// @Summary      create function
// @Description  create function
// @Tags         create, functions
// @Produce      json
// @Security Bearer
// @Param        message body models.Function true "element"
// @Success      200 {object}  models.Function
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /functions [POST]
func (this *FunctionsEndpoints) Create(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /functions", func(writer http.ResponseWriter, request *http.Request) {
		function := models.Function{}
		err := json.NewDecoder(request.Body).Decode(&function)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)

		result, err, errCode := control.SetFunction(token, function)
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
// @Summary      set function
// @Description  set function
// @Tags         set, functions
// @Produce      json
// @Security Bearer
// @Param        id path string true "Function Id"
// @Param        message body models.Function true "element"
// @Success      200 {object}  models.Function
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /functions/{id} [PUT]
func (this *FunctionsEndpoints) Set(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /functions/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		function := models.Function{}
		err := json.NewDecoder(request.Body).Decode(&function)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)

		if function.Id != id {
			http.Error(writer, "id in body unequal to id in request endpoint", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetFunction(token, function)
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
