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
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/v3/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func init() {
	endpoints = append(endpoints, &ThingEndpoints{})
}

type ThingEndpoints struct{}

// Get godoc
// @Summary      get thing
// @Description  get thing
// @Tags         things
// @Produce      json
// @Security Bearer
// @Param        id path string true "Thing Id"
// @Success      200 {object}  models.Thing
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /things/{id} [GET]
func (this *ThingEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /things/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.ReadThing(request.Context(), util.GetAuthToken(request), id)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writeThingJson(config, writer, result)
	})
}

// List godoc
// @Summary      list things
// @Description  list things
// @Tags         things
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter; matches id and name"
// @Param        sort query string false "default id.asc; supported: id, name"
// @Param        ids query string false "filter; ignores limit/offset; comma-separated list"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Param        device_ids query string false "filter; comma-separated list of device ids; lists things with a device link, open or ended, to one of the devices"
// @Success      200 {array}  models.Thing
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /things [GET]
func (this *ThingEndpoints) List(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /things", func(writer http.ResponseWriter, request *http.Request) {
		options := model.ThingListOptions{
			Limit:  100,
			Offset: 0,
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			options.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			options.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		if request.URL.Query().Has("ids") {
			idsParam := request.URL.Query().Get("ids")
			if idsParam != "" {
				options.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				options.Ids = []string{}
			}
		}

		if request.URL.Query().Has("device_ids") {
			deviceIdsParam := request.URL.Query().Get("device_ids")
			if deviceIdsParam != "" {
				options.DeviceIds = strings.Split(strings.TrimSpace(deviceIdsParam), ",")
			} else {
				options.DeviceIds = []string{}
			}
		}

		options.Search = request.URL.Query().Get("search")
		options.SortBy = request.URL.Query().Get("sort")
		if options.SortBy == "" {
			options.SortBy = "id.asc"
		}

		options.Permission, err = model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if options.Permission == models.UnsetPermissionFlag {
			options.Permission = model.READ
		}

		result, total, err, errCode := control.ListThings(request.Context(), util.GetAuthToken(request), options)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
		writeThingJson(config, writer, result)
	})
}

// Create godoc
// @Summary      create thing
// @Description  create thing; the id is assigned by the server, device_links in the body are ignored, use the device-links endpoints
// @Tags         things
// @Produce      json
// @Security Bearer
// @Param        message body models.Thing true "thing"
// @Success      200 {object}  models.Thing
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /things [POST]
func (this *ThingEndpoints) Create(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /things", func(writer http.ResponseWriter, request *http.Request) {
		thing := models.Thing{}
		err := json.NewDecoder(request.Body).Decode(&thing)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if thing.Id != "" {
			http.Error(writer, "thing may not contain a preset id. please use PUT to update a thing", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetThing(request.Context(), util.GetAuthToken(request), thing)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writeThingJson(config, writer, result)
	})
}

// Set godoc
// @Summary      set thing
// @Description  set thing; device_links in the body are ignored, the stored links are kept
// @Tags         things
// @Produce      json
// @Security Bearer
// @Param        id path string true "Thing Id"
// @Param        message body models.Thing true "element"
// @Success      200 {object}  models.Thing
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /things/{id} [PUT]
func (this *ThingEndpoints) Set(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /things/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		thing := models.Thing{}
		err := json.NewDecoder(request.Body).Decode(&thing)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if id == "" {
			http.Error(writer, "missing id in path", http.StatusBadRequest)
			return
		}
		if thing.Id == "" {
			thing.Id = id
		}
		if thing.Id != id {
			http.Error(writer, "id in body unequal to id in request endpoint", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetThing(request.Context(), util.GetAuthToken(request), thing)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writeThingJson(config, writer, result)
	})
}

// Delete godoc
// @Summary      delete thing
// @Description  delete thing; the thing is also removed from all graphs
// @Tags         things
// @Produce      json
// @Security Bearer
// @Param        id path string true "Thing Id"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /things/{id} [DELETE]
func (this *ThingEndpoints) Delete(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /things/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		err, errCode := control.DeleteThing(request.Context(), util.GetAuthToken(request), id)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writeThingJson(config, writer, true)
	})
}

// AddDeviceLink godoc
// @Summary      link device to thing
// @Description  link a device to a thing for the half-open period [from, until); from defaults to now, until to open; periods of one device may not overlap, not even across things
// @Tags         things
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        id path string true "Thing Id"
// @Param        message body model.ThingDeviceLinkCreate true "device link"
// @Success      200 {object}  models.Thing
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /things/{id}/device-links [POST]
func (this *ThingEndpoints) AddDeviceLink(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /things/{id}/device-links", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		create := model.ThingDeviceLinkCreate{}
		err := json.NewDecoder(request.Body).Decode(&create)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err, errCode := control.AddThingDeviceLink(request.Context(), util.GetAuthToken(request), id, create)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writeThingJson(config, writer, result)
	})
}

// EndDeviceLink godoc
// @Summary      end device link of thing
// @Description  sets until of the device link; an empty body or a missing until ends the link now
// @Tags         things
// @Accept       json
// @Produce      json
// @Security Bearer
// @Param        id path string true "Thing Id"
// @Param        link_id path string true "Device Link Id"
// @Param        message body model.ThingDeviceLinkEnd false "end of the link"
// @Success      200 {object}  models.Thing
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /things/{id}/device-links/{link_id}/end [POST]
func (this *ThingEndpoints) EndDeviceLink(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /things/{id}/device-links/{link_id}/end", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		linkId := request.PathValue("link_id")
		end := model.ThingDeviceLinkEnd{}
		err := json.NewDecoder(request.Body).Decode(&end)
		if err != nil && !errors.Is(err, io.EOF) { //io.EOF: empty body
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		result, err, errCode := control.EndThingDeviceLink(request.Context(), util.GetAuthToken(request), id, linkId, end)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writeThingJson(config, writer, result)
	})
}

func writeThingJson(config configuration.Config, writer http.ResponseWriter, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(writer).Encode(value)
	if err != nil {
		config.GetLogger().Info("unable to encode response", "error", err.Error())
	}
}
