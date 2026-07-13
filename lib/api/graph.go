/*
 * Copyright 2026 InfAI (CC SES)
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
	"net/http"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func init() {
	endpoints = append(endpoints, &GraphEndpoints{})
}

type GraphEndpoints struct{}

// Get godoc
// @Summary      get graph
// @Description  get graph
// @Tags         graphs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Graph Id"
// @Success      200 {object}  models.Graph
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /graphs/{id} [GET]
func (this *GraphEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /graphs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.ReadGraph(util.GetAuthToken(request), id)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}

// List godoc
// @Summary      list graph
// @Description  list graph
// @Tags         graphs
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-separated list"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Param        device_ids query string false "filter; comma-separated list of device ids that must be found in a node resource_id with resource_type 'device'"
// @Param        attributes query string false "filter; comma-separated list of key value pairs, value is optional; lists elements only if they have an attribute that is in the given list, example: ?attributes=attrKey1,attrKey2:attrVal2,attrKey3"
// @Param        attributes_json query string false "filter; json encoded []model.Attribute, alternative for attributes query parameter, may be useful if if the attribute origin is important or the key/value contains ',' or ':'"
// @Success      200 {array}  models.Graph
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /graphs [GET]
func (this *GraphEndpoints) List(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /graphs", func(writer http.ResponseWriter, request *http.Request) {
		graphListOptions := model.GraphListOptions{
			Limit:  100,
			Offset: 0,
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			graphListOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			graphListOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				graphListOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				graphListOptions.Ids = []string{}
			}
		}

		if request.URL.Query().Has("device_ids") {
			deviceIdsParam := request.URL.Query().Get("device_ids")
			if deviceIdsParam != "" {
				graphListOptions.DeviceIds = strings.Split(strings.TrimSpace(deviceIdsParam), ",")
			} else {
				graphListOptions.DeviceIds = []string{}
			}
		}

		if request.URL.Query().Has("attributes") {
			attributesParam := request.URL.Query().Get("attributes")
			if attributesParam != "" {
				for _, attr := range strings.Split(attributesParam, ",") {
					parts := strings.Split(attr, ":")
					key := strings.TrimSpace(parts[0])
					value := ""
					if len(parts) >= 2 {
						value = strings.TrimSpace(parts[1])
					}
					graphListOptions.Attributes = append(graphListOptions.Attributes, models.Attribute{
						Key:   key,
						Value: value,
					})
				}
			}
		}
		if request.URL.Query().Has("attributes_json") {
			attributesParam := request.URL.Query().Get("attributes_json")
			temp := []models.Attribute{}
			err = json.Unmarshal([]byte(attributesParam), &temp)
			if err != nil {
				http.Error(writer, "unable to parse attributes_json:"+err.Error(), http.StatusBadRequest)
				return
			}
			graphListOptions.Attributes = append(graphListOptions.Attributes, temp...)
		}

		graphListOptions.Search = request.URL.Query().Get("search")
		graphListOptions.SortBy = request.URL.Query().Get("sort")
		if graphListOptions.SortBy == "" {
			graphListOptions.SortBy = "id.asc"
		}

		graphListOptions.Permission, err = model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if graphListOptions.Permission == models.UnsetPermissionFlag {
			graphListOptions.Permission = model.READ
		}

		result, total, err, errCode := control.ListGraphs(util.GetAuthToken(request), graphListOptions)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}

// Create godoc
// @Summary      create graph
// @Description  create graph
// @Tags         graphs
// @Produce      json
// @Security Bearer
// @Param        message body models.Graph true "graph"
// @Success      200 {object}  models.Graph
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /graphs [POST]
func (this *GraphEndpoints) Create(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /graphs", func(writer http.ResponseWriter, request *http.Request) {
		graph := models.Graph{}
		err := json.NewDecoder(request.Body).Decode(&graph)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)
		if graph.Id != "" {
			http.Error(writer, "graph may not contain a preset id. please use PUT to update a graph", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetGraph(token, graph)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}

// Set godoc
// @Summary      set graph
// @Description  set graph
// @Tags         graphs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Graph Id"
// @Param        message body models.Graph true "element"
// @Success      200 {object}  models.Graph
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /graphs/{id} [PUT]
func (this *GraphEndpoints) Set(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /graphs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		graph := models.Graph{}
		err := json.NewDecoder(request.Body).Decode(&graph)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if id == "" {
			http.Error(writer, "missing id in path", http.StatusBadRequest)
		}
		if graph.Id == "" {
			graph.Id = id
		}
		if graph.Id != id {
			http.Error(writer, "id in body unequal to id in request endpoint", http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)

		result, err, errCode := control.SetGraph(token, graph)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}

// Delete godoc
// @Summary      delete graph
// @Description  delete graph
// @Tags         graphs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Graph Id"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /graphs/{id} [DELETE]
func (this *GraphEndpoints) Delete(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /graphs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token := util.GetAuthToken(request)

		err, errCode := control.DeleteGraph(token, id)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(true)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}
