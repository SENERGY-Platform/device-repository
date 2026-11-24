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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func init() {
	endpoints = append(endpoints, &GraphEndpoints{})
}

type GraphEndpoints struct{}

// parseGraphTimeQuery reads an optional RFC3339 instant from the query.
func parseGraphTimeQuery(request *http.Request, name string) (result *time.Time, err error) {
	if !request.URL.Query().Has(name) {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, request.URL.Query().Get(name))
	if err != nil {
		return nil, fmt.Errorf("unable to parse %v as RFC3339 (a '+' in the offset must be url-encoded as %%2B): %w", name, err)
	}
	return &t, nil
}

// Get godoc
// @Summary      get graph
// @Description  get graph; elements of a graph have a validity [valid_from, valid_until) and the graph is answered as it is at one instant: now, or the instant given by 'at'; an instant at which no node is valid is answered with the graph without nodes and edges; with versions=all the stored document is answered, with all versions of nodes, edges and attributes
// @Tags         graphs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Graph Id"
// @Param        at query string false "RFC3339 instant; default now; the graph as it is valid at this instant; may not be combined with versions"
// @Param        versions query string false "'all' answers the stored document with all versions; may not be combined with at"
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
		at, err := parseGraphTimeQuery(request, "at")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		versions := request.URL.Query().Get("versions")
		if versions != "" && versions != "all" {
			http.Error(writer, "unknown value for versions, use 'all'", http.StatusBadRequest)
			return
		}
		if versions != "" && at != nil {
			http.Error(writer, "at and versions may not be combined", http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)
		var result models.Graph
		var errCode int
		switch {
		case versions == "all":
			result, err, errCode = control.ReadGraphVersions(request.Context(), token, id)
		case at != nil:
			result, err, errCode = control.ReadGraphAt(request.Context(), token, id, *at)
		default:
			result, err, errCode = control.ReadGraph(request.Context(), token, id)
		}
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

// GetSegments godoc
// @Summary      get graph segments
// @Description  get the graph cut at every validity bound (valid_from or valid_until of a node, an edge or an attribute) inside the range [start, end); each piece [from, until) comes with the graph as it is valid in that piece; pieces are not merged; without start the first piece has from null, without end the last has until null, without both the answer is the whole history
// @Tags         graphs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Graph Id"
// @Param        start query string false "RFC3339 instant, inclusive; default unbounded"
// @Param        end query string false "RFC3339 instant, exclusive, must be after start; default unbounded"
// @Success      200 {array}  models.GraphSegment
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /graphs/{id}/segments [GET]
func (this *GraphEndpoints) GetSegments(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /graphs/{id}/segments", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		start, err := parseGraphTimeQuery(request, "start")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		end, err := parseGraphTimeQuery(request, "end")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err, errCode := control.ReadGraphSegments(request.Context(), util.GetAuthToken(request), id, start, end)
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
// @Description  list graph; every graph is answered as it is now, see GET /graphs/{id}
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

		result, total, err, errCode := control.ListGraphs(request.Context(), util.GetAuthToken(request), graphListOptions)
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

		result, err, errCode := control.SetGraph(request.Context(), token, graph)
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
// @Description  set graph; the body is the edited graph as it is valid at the instant 'at' (default now), the server compares it with the stored graph at that instant by identity (a node by its id, an edge by from_node_id and to_node_id, an attribute by its key within its owner) and sets the validities; versions that are not valid at 'at' are never touched or dropped; valid_from and valid_until of the body are ignored for elements that exist and refused for elements that are added; the resource of a node may not change; the stored graph is validated for every segment between its validity bounds before it is written, a 400 names the first invalid instant; mode 'correction' (default) modifies the versions valid at 'at' in place over their whole span, adds elements without bounds (except that they end where a later version of the same identity starts) and deletes removed ones; mode 'change' splits a changed version at 'at' (the old one ends, the new one runs to the old one's end), adds elements valid from 'at' until the next version of their identity and ends removed ones at 'at'; with 'until', only for mode 'change', the edit holds exactly for [at, until) for every identity the edit touches: versions of it that go on beyond 'until' are resumed there with their old content, whatever lies inside the span is replaced, and a node that is valid beyond 'until' can not be removed; the write is compare-and-set against concurrent writes of the same graph and is applied again a few times before the answer is 409; the answer is the stored graph as it is now
// @Tags         graphs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Graph Id"
// @Param        at query string false "RFC3339 instant the body is the graph at; default now"
// @Param        until query string false "RFC3339 instant, must be after at, only with mode=change; the edit holds for [at, until)"
// @Param        mode query string false "'correction' (default) or 'change'"
// @Param        message body models.Graph true "element"
// @Success      200 {object}  models.Graph
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      409
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

		options := model.GraphEditOptions{Mode: request.URL.Query().Get("mode")}
		options.At, err = parseGraphTimeQuery(request, "at")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		options.Until, err = parseGraphTimeQuery(request, "until")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		result, err, errCode := control.EditGraph(request.Context(), token, graph, options)
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

// MoveNode godoc
// @Summary      move graph node
// @Description  moves a node below another node from the instant 'at' (default now): the outgoing edges of the node valid at 'at' end there and one edge to to_node_id with weight 100 starts there; with 'until' the move holds for [at, until) only and the old edges are resumed (a month booked to the wrong parent is one request); the end node can not be moved; history before 'at' is not touched; the answer is the graph as it is now; 400 names the instant at which the result is invalid; 409 if the graph was changed concurrently and stays so after a few attempts; needs write on the graph
// @Tags         graphs
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        id path string true "Graph Id"
// @Param        node_id path string true "Node Id"
// @Param        message body model.GraphNodeMove true "move"
// @Success      200 {object}  models.Graph
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      409
// @Failure      500
// @Router       /graphs/{id}/nodes/{node_id}/move [POST]
func (this *GraphEndpoints) MoveNode(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /graphs/{id}/nodes/{node_id}/move", func(writer http.ResponseWriter, request *http.Request) {
		move := model.GraphNodeMove{}
		err := json.NewDecoder(request.Body).Decode(&move)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err, errCode := control.MoveGraphNode(request.Context(), util.GetAuthToken(request), request.PathValue("id"), request.PathValue("node_id"), move)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writeGraphJson(config, writer, result)
	})
}

// EndNode godoc
// @Summary      end graph node
// @Description  takes a node out of the graph from the instant 'at' (default now): the node and its edges end there and the edges that led to it are rerouted to where it led, as when a device is deleted; everything before 'at' stays; a thing or device the node stands for is not touched, in particular the open device link of a meter thing is NOT ended, ending or swapping it is a separate call (POST /things/{id}/device-links/{link_id}/end or /swap) the caller decides on; the answer is the graph as it is now; 400 names the instant at which the result is invalid; 409 if the graph was changed concurrently and stays so after a few attempts; needs write on the graph
// @Tags         graphs
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        id path string true "Graph Id"
// @Param        node_id path string true "Node Id"
// @Param        message body model.GraphNodeEnd false "end"
// @Success      200 {object}  models.Graph
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      409
// @Failure      500
// @Router       /graphs/{id}/nodes/{node_id}/end [POST]
func (this *GraphEndpoints) EndNode(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /graphs/{id}/nodes/{node_id}/end", func(writer http.ResponseWriter, request *http.Request) {
		end := model.GraphNodeEnd{}
		err := json.NewDecoder(request.Body).Decode(&end)
		if err != nil && !errors.Is(err, io.EOF) { //io.EOF: empty body
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err, errCode := control.EndGraphNode(request.Context(), util.GetAuthToken(request), request.PathValue("id"), request.PathValue("node_id"), end)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writeGraphJson(config, writer, result)
	})
}

func writeGraphJson(config configuration.Config, writer http.ResponseWriter, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(writer).Encode(value)
	if err != nil {
		config.GetLogger().Info("unable to encode response", "error", err.Error())
	}
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

		err, errCode := control.DeleteGraph(request.Context(), token, id)
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

// DeleteMany godoc
// @Summary      delete multiple graphs
// @Description  delete multiple graphs; rights and preconditions of all ids are checked before the first one is deleted, unknown ids are ignored; at most 10000 ids
// @Tags         graphs
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        message body []string true "ids to be deleted"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      500
// @Router       /graphs [DELETE]
func (this *GraphEndpoints) DeleteMany(config configuration.Config, router *http.ServeMux, control Controller) {
	handleDeleteMany(config, router, "/graphs", control, Controller.DeleteGraphs)
}
