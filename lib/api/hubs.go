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
	"github.com/SENERGY-Platform/device-repository/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &HubEndpoints{})
}

type HubEndpoints struct{}

// Get godoc
// @Summary      get hub
// @Description  get hub
// @Tags         get, hubs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Hub Id"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Success      200 {object}  models.Hub
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /hubs/{id} [GET]
func (this *HubEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /hubs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		permission, err := model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if permission == models.UnsetPermissionFlag {
			permission = model.READ
		}
		result, err, errCode := control.ReadHub(id, util.GetAuthToken(request), permission)
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

// List godoc
// @Summary      list hubs
// @Description  list hubs
// @Tags         list, hubs
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Param        connection-state query integer false "filter; valid values are 'online', 'offline' and an empty string for unknown states"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Success      200 {array}  models.Hub
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /hubs [GET]
func (this *HubEndpoints) List(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /hubs", func(writer http.ResponseWriter, request *http.Request) {
		hubListOptions := model.HubListOptions{
			Limit:  100,
			Offset: 0,
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			hubListOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			hubListOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				hubListOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				hubListOptions.Ids = []string{}
			}
		}

		hubListOptions.LocalDeviceId = request.URL.Query().Get("local-device-id")
		hubListOptions.OwnerId = request.URL.Query().Get("owner")

		hubListOptions.Search = request.URL.Query().Get("search")
		hubListOptions.SortBy = request.URL.Query().Get("sort")
		if hubListOptions.SortBy == "" {
			hubListOptions.SortBy = "name.asc"
		}

		if request.URL.Query().Has("connection-state") {
			searchedState := request.URL.Query().Get("connection-state")
			if !slices.Contains([]models.ConnectionState{models.ConnectionStateOnline, models.ConnectionStateOffline, models.ConnectionStateUnknown}, searchedState) {
				http.Error(writer, "invalid connection state:"+searchedState, http.StatusBadRequest)
				return
			}
			hubListOptions.ConnectionState = &searchedState
		}

		hubListOptions.Permission, err = model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if hubListOptions.Permission == models.UnsetPermissionFlag {
			hubListOptions.Permission = model.READ
		}

		result, err, errCode := control.ListHubs(util.GetAuthToken(request), hubListOptions)
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

// GetDevices godoc
// @Summary      get device ids of hub
// @Description  get device ids of hub
// @Tags         get, hubs, devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Hub Id"
// @Param        as query string false "default 'local_id'; valid values 'local_id', 'localId', 'id'; selects if device ids or device local-ids should be returned"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Success      200 {array}  string
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /hubs/{id}/devices [GET]
func (this *HubEndpoints) GetDevices(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /hubs/{id}/devices", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		permission, err := model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if permission == models.UnsetPermissionFlag {
			permission = model.READ
		}

		var asLocalId bool
		asParam := request.URL.Query().Get("as")
		switch asParam {
		case "":
			asLocalId = true
		case "id":
			asLocalId = false
		case "localId":
			asLocalId = true
		case "local_id":
			asLocalId = true
		default:
			http.Error(writer, "expect 'id', 'localId' or 'local_id' as value for 'as' query-parameter if it is used", http.StatusBadRequest)
			return
		}
		result, err, errCode := control.ListHubDeviceIds(id, util.GetAuthToken(request), permission, asLocalId)
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

// Head godoc
// @Summary      head hub
// @Description  head hub
// @Tags         head, hubs
// @Security Bearer
// @Param        id path string true "Hub Id"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /hubs/{id} [HEAD]
func (this *HubEndpoints) Head(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("HEAD /hubs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		permission, err := model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if permission == models.UnsetPermissionFlag {
			permission = model.READ
		}
		id := request.PathValue("id")
		_, _, errCode := control.ReadHub(id, util.GetAuthToken(request), permission)
		writer.WriteHeader(errCode)
		return
	})
}

// Validate godoc
// @Summary      validate hub
// @Description  validate hub
// @Tags         validate, hubs
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.Hub true "Hub to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /hubs [PUT]
func (this *HubEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /hubs", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		hub := models.Hub{}
		err = json.NewDecoder(request.Body).Decode(&hub)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateHub(util.GetAuthToken(request), hub)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})

	//legacy
	router.HandleFunc("PUT /hubs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		hub := models.Hub{}
		err = json.NewDecoder(request.Body).Decode(&hub)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateHub(util.GetAuthToken(request), hub)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// Create godoc
// @Summary      create hub
// @Description  create hub
// @Tags         create, hubs
// @Produce      json
// @Security Bearer
// @Param        message body models.Hub true "element"
// @Success      200 {object}  models.Hub
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /hubs [POST]
func (this *HubEndpoints) Create(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /hubs", func(writer http.ResponseWriter, request *http.Request) {
		hub := models.Hub{}
		err := json.NewDecoder(request.Body).Decode(&hub)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)

		if hub.Id != "" {
			http.Error(writer, "body may not contain a preset id. please use the PUT method for updates", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetHub(token, hub)
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
// @Summary      set hub
// @Description  set hub
// @Tags         set, hubs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Hub Id"
// @Param        user_id query string false "only admins may set user_id; overwrites hub.OwnerId; defaults to existing hub.OwnerId and falls back to user-id of requesting user if hub does not exist"
// @Param        message body models.Hub true "element"
// @Success      200 {object}  models.Hub
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /hubs/{id} [PUT]
func (this *HubEndpoints) Set(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /hubs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		userId := request.URL.Query().Get("user_id")
		hub := models.Hub{}
		err := json.NewDecoder(request.Body).Decode(&hub)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if hub.Id != id || hub.Id == "" {
			http.Error(writer, "hub id in body unequal to hub id in request endpoint", http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)
		jwtToken, err := jwt.Parse(token)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if userId != "" && !jwtToken.IsAdmin() {
			http.Error(writer, "only admins may set user_id", http.StatusForbidden)
			return
		}
		if userId != "" {
			hub.OwnerId = userId
		}

		result, err, errCode := control.SetHub(token, hub)
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

// SetName godoc
// @Summary      set hub name
// @Description  set hub name
// @Tags         set, hubs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Hub Id"
// @Param        message body string true "name"
// @Success      200 {object}  models.Hub
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /hubs/{id}/name [PUT]
func (this *HubEndpoints) SetName(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /hubs/{id}/name", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		name := ""
		err := json.NewDecoder(request.Body).Decode(&name)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)
		hub, err, code := control.ReadHub(token, id, model.WRITE)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		hub.Name = name

		result, err, errCode := control.SetHub(token, hub)
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

// Delete godoc
// @Summary      delete hub
// @Description  delete hub
// @Tags         delete, hubs
// @Produce      json
// @Security Bearer
// @Param        id path string true "Hub Id"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /hubs/{id} [DELETE]
func (this *HubEndpoints) Delete(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /hubs/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token := util.GetAuthToken(request)
		err, errCode := control.DeleteHub(token, id)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(true)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})
}
