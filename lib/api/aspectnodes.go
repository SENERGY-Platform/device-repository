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
	"strings"
)

func init() {
	endpoints = append(endpoints, &AspectNodeEndpoints{})
}

type AspectNodeEndpoints struct{}

type AspectNodeQuery struct {
	Ids *[]string `json:"ids,omitempty"`
}

// Query godoc
// @Summary      query aspect-nodes
// @Description  query aspect-nodes
// @Tags         query, aspect-nodes
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        message body AspectNodeQuery true "AspectNodeQuery"
// @Success      200 {array}  models.AspectNode
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /query/aspect-nodes [POST]
func (this *AspectNodeEndpoints) Query(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /query/aspect-nodes", func(writer http.ResponseWriter, request *http.Request) {
		query := AspectNodeQuery{}
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if query.Ids != nil {
			result, err, errCode := control.GetAspectNodesByIdList(*query.Ids)
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
		}
		http.Error(writer, "no known query content found", http.StatusBadRequest)
		return
	})
}

// ListAspectNodes godoc
// @Summary      list aspect-nodes
// @Description  list aspect-nodes
// @Tags         list, aspect-nodes
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Success      200 {array}  models.AspectNode
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/aspect-nodes [GET]
func (this *AspectEndpoints) ListAspectNodes(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /v2/aspect-nodes", func(writer http.ResponseWriter, request *http.Request) {
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
		result, total, err, errCode := control.ListAspectNodes(listoptions)
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
// @Summary      deprecated list aspect-nodes
// @Description  deprecated list aspect-nodes
// @Tags         list, aspect-nodes, aspects, deprecated
// @Produce      json
// @Security Bearer
// @Param        function query string false "filter; only 'measuring-function' is a valid value; if set, returns aspect-nodes used in combination with measuring-functions"
// @Param        ancestors query bool false "filter; in combination with 'function'; if true, returns also ancestor nodes of matching nodes"
// @Param        descendants query bool false "filter; in combination with 'function'; if true, returns also descendant nodes of matching nodes"
// @Success      200 {array}  models.AspectNode
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /aspect-nodes [GET]
func (this *AspectNodeEndpoints) List(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /aspect-nodes", func(writer http.ResponseWriter, request *http.Request) {
		var result []models.AspectNode
		var err error
		var errCode int

		function := request.URL.Query().Get("function")

		if function == "" {
			result, err, errCode = control.GetAspectNodes()
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
				result, err, errCode = control.GetAspectNodesWithMeasuringFunction(ancestors, descendants)
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
// @Summary      get aspect-node
// @Description  get aspect-node
// @Tags         get, aspect-nodes, aspects
// @Produce      json
// @Security Bearer
// @Param        id path string true "Aspect-Node Id"
// @Success      200 {object}  models.AspectNode
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /aspect-nodes/{id} [GET]
func (this *AspectNodeEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /aspect-nodes/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.GetAspectNode(id)
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
// @Summary      list aspect-node measuring-functions
// @Description  list measuring-functions used in combination with this aspect-node
// @Tags         list, aspect-nodes, aspects, functions
// @Produce      json
// @Security Bearer
// @Param        id path string true "Aspect-Node Id"
// @Success      200 {array}  models.Function
// @Param        ancestors query bool false "filter; if true, returns also functions used in combination with ancestors of the input aspect-node"
// @Param        descendants query bool false "filter; if true, returns also functions used in combination with descendants of the input aspect-node"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /aspect-nodes/{id}/measuring-functions [GET]
func (this *AspectNodeEndpoints) ListMeasuringFunctions(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /aspect-nodes/{id}/measuring-functions", func(writer http.ResponseWriter, request *http.Request) {
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
