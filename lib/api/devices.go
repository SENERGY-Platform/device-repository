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
	"log"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func init() {
	endpoints = append(endpoints, &DeviceEndpoints{})
}

type DeviceEndpoints struct{}

const UpdateOnlySameOriginAttributesKey = "update-only-same-origin-attributes"
const DisplayNameAttributeKey = "shared/nickname"
const DisplayNameAttributeOrigin = "shared"

// List godoc
// @Summary      list devices
// @Description  list devices
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Param        local_ids query string false "in combination with owner; fills ids filter; comma-seperated list"
// @Param        owner query string false "used in combination with local_ids to fill ids filter; defaults to requesting user"
// @Param        device-type-ids query string false "filter; comma-seperated list"
// @Param        attr-keys query string false "filter; comma-seperated list; lists elements only if they have an attribute key that is in the given list"
// @Param        attr-values query string false "filter; comma-seperated list; lists elements only if they have an attribute value that is in the given list"
// @Param        connection-state query integer false "filter; valid values are 'online', 'offline' and an empty string for unknown states"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Param        device-attribute-blacklist query string false "JSON encoded []models.Attribute, attribute value and origin will only be checked if set, otherwise all values or origins will be blacklisted"
// @Success      200 {array}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /devices [GET]
func (this *DeviceEndpoints) List(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /devices", func(writer http.ResponseWriter, request *http.Request) {
		deviceListOptions := model.DeviceListOptions{
			Limit:  100,
			Offset: 0,
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			deviceListOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			deviceListOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				deviceListOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				deviceListOptions.Ids = []string{}
			}
		}

		localIdsParam := request.URL.Query().Get("local_ids")
		if request.URL.Query().Has("local_ids") {
			if localIdsParam != "" {
				deviceListOptions.LocalIds = strings.Split(strings.TrimSpace(localIdsParam), ",")
			} else {
				deviceListOptions.LocalIds = []string{}
			}
		}

		deviceListOptions.Owner = request.URL.Query().Get("owner")

		deviceTypeIdsParam := request.URL.Query().Get("device-type-ids")
		if request.URL.Query().Has("device-type-ids") {
			if deviceTypeIdsParam != "" {
				deviceListOptions.DeviceTypeIds = strings.Split(strings.TrimSpace(deviceTypeIdsParam), ",")
			} else {
				deviceListOptions.DeviceTypeIds = []string{}
			}
		}

		attrKeysParam := request.URL.Query().Get("attr-keys")
		if request.URL.Query().Has("attr-keys") {
			if attrKeysParam != "" {
				deviceListOptions.AttributeKeys = strings.Split(strings.TrimSpace(attrKeysParam), ",")
			} else {
				deviceListOptions.AttributeKeys = []string{}
			}
		}
		attrValuesParam := request.URL.Query().Get("attr-values")
		if request.URL.Query().Has("attr-values") {
			if attrValuesParam != "" {
				deviceListOptions.AttributeValues = strings.Split(strings.TrimSpace(attrValuesParam), ",")
			} else {
				deviceListOptions.AttributeValues = []string{}
			}
		}

		deviceListOptions.Search = request.URL.Query().Get("search")
		deviceListOptions.SortBy = request.URL.Query().Get("sort")
		if deviceListOptions.SortBy == "" {
			deviceListOptions.SortBy = "name.asc"
		}

		if request.URL.Query().Has("connection-state") {
			searchedState := request.URL.Query().Get("connection-state")
			if !slices.Contains([]models.ConnectionState{models.ConnectionStateOnline, models.ConnectionStateOffline, models.ConnectionStateUnknown}, searchedState) {
				http.Error(writer, "invalid connection state:"+searchedState, http.StatusBadRequest)
				return
			}
			deviceListOptions.ConnectionState = &searchedState
		}

		deviceListOptions.Permission, err = model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if deviceListOptions.Permission == models.UnsetPermissionFlag {
			deviceListOptions.Permission = model.READ
		}

		deviceAttributeBlacklistParam := request.URL.Query().Get("device-attribute-blacklist")
		if deviceAttributeBlacklistParam != "" {
			deviceAttributeBlacklistParam, err = url.QueryUnescape(deviceAttributeBlacklistParam)
			if err != nil {
				http.Error(writer, "unable to decode device-attribute-blacklist: "+err.Error(), http.StatusBadRequest)
				return
			}
			var blacklist []models.Attribute
			err = json.Unmarshal([]byte(deviceAttributeBlacklistParam), &blacklist)
			if err != nil {
				http.Error(writer, "unable to parse device-attribute-blacklist: "+err.Error(), http.StatusBadRequest)
				return
			}
			deviceListOptions.DeviceAttributeBlacklist = blacklist
		}

		result, err, errCode := control.ListDevices(util.GetAuthToken(request), deviceListOptions)
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
// @Summary      get device
// @Description  get device
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Id"
// @Param        as query string false "interprets the id as local_id if as=='local_id'"
// @Param        owner_id query string false "default requesting user; used in combination with local_id (as=='local_id') to identify the device"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Success      200 {object}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /devices/{id} [GET]
func (this *DeviceEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /devices/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		as := request.URL.Query().Get("as")
		ownerId := request.URL.Query().Get("owner_id")
		if ownerId == "" {
			token, err := jwt.GetParsedToken(request)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusUnauthorized)
				return
			}
			ownerId = token.GetUserId()
		}
		permission, err := model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if permission == models.UnsetPermissionFlag {
			permission = model.READ
		}
		var result models.Device
		var errCode int
		if as == "local_id" {
			result, err, errCode = control.ReadDeviceByLocalId(ownerId, id, util.GetAuthToken(request), permission)
		} else {
			result, err, errCode = control.ReadDevice(id, util.GetAuthToken(request), permission)
		}
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
// @Summary      validate device
// @Description  validate device
// @Tags         devices
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.Device true "Device to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /devices [PUT]
func (this *DeviceEndpoints) Validate(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /devices", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		device := models.Device{}
		err = json.NewDecoder(request.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateDevice(util.GetAuthToken(request), device)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// Create godoc
// @Summary      create device
// @Description  create device
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        message body models.Device true "element"
// @Success      200 {object}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /devices [POST]
func (this *DeviceEndpoints) Create(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /devices", func(writer http.ResponseWriter, request *http.Request) {
		device := models.Device{}
		err := json.NewDecoder(request.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)

		if device.Id != "" {
			http.Error(writer, "body may not contain a preset id. please use the PUT method for updates", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.CreateDevice(token, device)
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
// @Summary      set device
// @Description  set device; admins may create new devices but only without using the UpdateOnlySameOriginAttributesKey query parameter
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Id"
// @Param        update-only-same-origin-attributes query string false "comma separated list; ensure that no attribute from another origin is overwritten"
// @Param        message body models.Device true "element"
// @Success      200 {object}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /devices/{id} [PUT]
func (this *DeviceEndpoints) Set(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /devices/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		device := models.Device{}
		err := json.NewDecoder(request.Body).Decode(&device)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)

		if device.Id != id {
			http.Error(writer, "id in body unequal to id in request endpoint", http.StatusBadRequest)
			return
		}

		options := model.DeviceUpdateOptions{}
		if request.URL.Query().Has(UpdateOnlySameOriginAttributesKey) {
			temp := request.URL.Query().Get(UpdateOnlySameOriginAttributesKey)
			options.UpdateOnlySameOriginAttributes = strings.Split(temp, ",")
		}

		result, err, errCode := control.SetDevice(token, device, options)
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

// SetAttributes godoc
// @Summary      set device attributes
// @Description  set device attributes
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Id"
// @Param        update-only-same-origin-attributes query string false "comma separated list; ensure that no attribute from another origin is overwritten"
// @Param        message body []models.Attribute true "attributes"
// @Success      200 {object}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /devices/{id}/attributes [PUT]
func (this *DeviceEndpoints) SetAttributes(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /devices/{id}/attributes", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		attributes := []models.Attribute{}
		err := json.NewDecoder(request.Body).Decode(&attributes)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		options := model.DeviceUpdateOptions{}
		if request.URL.Query().Has(UpdateOnlySameOriginAttributesKey) {
			temp := request.URL.Query().Get(UpdateOnlySameOriginAttributesKey)
			options.UpdateOnlySameOriginAttributes = strings.Split(temp, ",")
		}

		device, err, errCode := control.ReadDevice(id, token, model.WRITE)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		device.Attributes = attributes

		result, err, errCode := control.SetDevice(token, device, options)
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

// SetDisplayName godoc
// @Summary      set device display name
// @Description  set device display name
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Id"
// @Param        message body string true "display name"
// @Success      200 {object}  models.Device
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /devices/{id}/display_name [PUT]
func (this *DeviceEndpoints) SetDisplayName(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /devices/{id}/display_name", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		displayName := ""

		err := json.NewDecoder(request.Body).Decode(&displayName)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)

		device, err, errCode := control.ReadDevice(id, token, model.WRITE)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}

		displayNameAttrFound := false
		for i, attr := range device.Attributes {
			if attr.Key == DisplayNameAttributeKey {
				attr.Value = displayName
				device.Attributes[i] = attr
				displayNameAttrFound = true
			}
		}
		if !displayNameAttrFound {
			device.Attributes = append(device.Attributes, models.Attribute{Key: DisplayNameAttributeKey, Value: displayName, Origin: DisplayNameAttributeOrigin})
		}

		result, err, errCode := control.SetDevice(token, device, model.DeviceUpdateOptions{})
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
// @Summary      delete device
// @Description  delete device
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Id"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /devices/{id} [DELETE]
func (this *DeviceEndpoints) Delete(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /devices/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token := util.GetAuthToken(request)

		err, errCode := control.DeleteDevice(token, id)
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

// DeleteMany godoc
// @Summary      delete multiple devices
// @Description  delete multiple devices
// @Tags         devices
// @Produce      json
// @Security Bearer
// @Param        message body []string true "ids to be deleted"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /devices [DELETE]
func (this *DeviceEndpoints) DeleteMany(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /devices", func(writer http.ResponseWriter, request *http.Request) {
		ids := []string{}
		err := json.NewDecoder(request.Body).Decode(&ids)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)

		for _, id := range ids {
			err, errCode := control.DeleteDevice(token, id)
			if err != nil {
				http.Error(writer, err.Error(), errCode)
				return
			}
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(true)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})
}

// SetConnectionState godoc
// @Summary      set device connection-state
// @Description  set device connection-state
// @Tags         devices
// @Security Bearer
// @Param        message body bool true "connected true/false"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /devices/{id}/connection-state [PUT]
func (this *DeviceEndpoints) SetConnectionState(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /devices/{id}/connection-state", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		if id == "" {
			http.Error(writer, "missing id", http.StatusBadRequest)
			return
		}
		connected := false
		err := json.NewDecoder(request.Body).Decode(&connected)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)

		err, errCode := control.SetDeviceConnectionState(token, id, connected)
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
