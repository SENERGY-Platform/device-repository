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
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &DeviceTypeEndpoints{})
}

type DeviceTypeEndpoints struct{}

// Get godoc
// @Summary      get device-type
// @Description  get device-type
// @Tags         get, device-types
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device-Type Id"
// @Success      200 {object}  models.DeviceType
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-types/{id} [GET]
func (this *DeviceTypeEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /device-types/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.ReadDeviceType(id, util.GetAuthToken(request))
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

// ListV3 godoc
// @Summary      list device-types
// @Description  list device-types
// @Tags         list, device-types
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Param        protocol-ids query string false "filter; comma-seperated list; lists elements only if they use a protocol that is in the given list"
// @Param        attr-keys query string false "filter; comma-seperated list; lists elements only if they have an attribute key that is in the given list"
// @Param        attr-values query string false "filter; comma-seperated list; lists elements only if they have an attribute value that is in the given list"
// @Param        include-modified query bool false "include id-modified device-types"
// @Param        ignore-unmodified query bool false "no unmodified device-types"
// @Param        criteria query string false "filter; json encoded []model.FilterCriteria"
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; does not count modified elements; used for pagination"
// @Success      200 {array}  models.DeviceType
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v3/device-types [GET]
func (this *DeviceTypeEndpoints) ListV3(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /v3/device-types", func(writer http.ResponseWriter, request *http.Request) {
		options := model.DeviceTypeListOptions{
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

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				options.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				options.Ids = []string{}
			}
		}

		attrKeysParam := request.URL.Query().Get("attr-keys")
		if request.URL.Query().Has("attr-keys") {
			if attrKeysParam != "" {
				options.AttributeKeys = strings.Split(strings.TrimSpace(attrKeysParam), ",")
			} else {
				options.AttributeKeys = []string{}
			}
		}
		attrValuesParam := request.URL.Query().Get("attr-values")
		if request.URL.Query().Has("attr-values") {
			if attrValuesParam != "" {
				options.AttributeValues = strings.Split(strings.TrimSpace(attrValuesParam), ",")
			} else {
				options.AttributeValues = []string{}
			}
		}

		protocolIdsParam := request.URL.Query().Get("protocol-ids")
		if request.URL.Query().Has("protocol-ids") {
			if protocolIdsParam != "" {
				options.ProtocolIds = strings.Split(strings.TrimSpace(protocolIdsParam), ",")
			} else {
				options.ProtocolIds = []string{}
			}
		}

		options.Search = request.URL.Query().Get("search")
		options.SortBy = request.URL.Query().Get("sort")
		if options.SortBy == "" {
			options.SortBy = "name.asc"
		}

		includeModifiedStr := request.URL.Query().Get("include-modified")
		if includeModifiedStr != "" {
			options.IncludeModified, err = strconv.ParseBool(includeModifiedStr)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		ignoreUnmodified := request.URL.Query().Get("ignore-unmodified")
		if ignoreUnmodified != "" {
			options.IgnoreUnmodified, err = strconv.ParseBool(ignoreUnmodified)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		criteria := request.URL.Query().Get("criteria")
		if criteria != "" {
			criteriaList := []model.FilterCriteria{}
			err = json.Unmarshal([]byte(criteria), &criteriaList)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
			options.Criteria = criteriaList
		}

		result, total, err, code := control.ListDeviceTypesV3(util.GetAuthToken(request), options)
		if err != nil {
			http.Error(writer, err.Error(), code)
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

// deprecated
func (this *DeviceTypeEndpoints) List(config config.Config, router *http.ServeMux, control Controller) {
	/*
			query params:
			- limit: number; default 100
		    - offset: number; default 0
			- sort: <field>[.<direction>]; optional;
				- field: 'name', 'id'; defined at github.com/SENERGY-Platform/device-repository/lib/database/mongo/devicetype.go ListDeviceTypes()
				- direction: 'asc' || 'desc'; optional
				- examples:
					?sort=name.asc
					?sort=name
			- filter: json encoded []model.FilterCriteria; optional
					all criteria must be satisfied
			- interactions-filter: comma seperated list of interactions
					deprecated: use interactions field in filter (model.FilterCriteria.Interaction)
					if set: returns only device-types with at least one matching interaction on criteria matching services
					ignored if empty
			- include_id_modified: bool; add service-group modified device-types to result
	*/
	router.HandleFunc("GET /device-types", func(writer http.ResponseWriter, request *http.Request) {
		var err error
		limitParam := request.URL.Query().Get("limit")
		var limit int64 = 100
		if limitParam != "" {
			limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		var offset int64 = 0
		if offsetParam != "" {
			offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		sort := request.URL.Query().Get("sort")
		if sort == "" {
			sort = "name.asc"
		}

		includeModifiedStr := request.URL.Query().Get("include_id_modified")
		includeModified := false
		if includeModifiedStr != "" {
			includeModified, err = strconv.ParseBool(includeModifiedStr)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		includeUnmodifiedStr := request.URL.Query().Get("include_id_unmodified")
		includeUnmodified := true
		if includeUnmodifiedStr != "" {
			includeUnmodified, err = strconv.ParseBool(includeUnmodifiedStr)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		filter := request.URL.Query().Get("filter")
		deviceTypesFilter := []model.FilterCriteria{}
		if filter != "" {
			err = json.Unmarshal([]byte(filter), &deviceTypesFilter)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		var result []models.DeviceType
		var errCode int
		interactionsFilterStr := request.URL.Query().Get("interactions-filter")
		if interactionsFilterStr != "" {
			interactionsFilter := []string{}
			for _, interaction := range strings.Split(interactionsFilterStr, ",") {
				interactionsFilter = append(interactionsFilter, strings.TrimSpace(interaction))
			}
			result, err, errCode = control.ListDeviceTypes(util.GetAuthToken(request), limit, offset, sort, deviceTypesFilter, interactionsFilter, includeModified, includeUnmodified)
		} else {
			result, err, errCode = control.ListDeviceTypesV2(util.GetAuthToken(request), limit, offset, sort, deviceTypesFilter, includeModified, includeUnmodified)
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
// @Summary      validate device-type
// @Description  validate device-type
// @Tags         validate, device-types
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.DeviceType true "Device-Type to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /device-types [PUT]
func (this *DeviceTypeEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /device-types", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		options, err := model.LoadDeviceTypeValidationOptions(request.URL.Query())
		if err != nil {
			http.Error(writer, "invalid validation options: "+err.Error(), http.StatusBadRequest)
			return
		}
		dt := models.DeviceType{}
		err = json.NewDecoder(request.Body).Decode(&dt)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateDeviceType(dt, options)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// Create godoc
// @Summary      create device-type
// @Description  create device-type
// @Tags         create, device-types
// @Produce      json
// @Security Bearer
// @Param        distinct_attributes query string false "comma separated list of attribute keys; no other device-type with the same attribute key/value may exist"
// @Param        message body models.DeviceType true "element"
// @Success      200 {object}  models.DeviceType
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-types [POST]
func (this *DeviceTypeEndpoints) Create(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /device-types", func(writer http.ResponseWriter, request *http.Request) {
		devicetype := models.DeviceType{}
		err := json.NewDecoder(request.Body).Decode(&devicetype)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)
		if devicetype.Id != "" {
			http.Error(writer, "body may not contain a preset id. please use the PUT method for updates", http.StatusBadRequest)
			return
		}

		distinctAttr := request.URL.Query().Get("distinct_attributes")
		if distinctAttr != "" {
			err = control.ValidateDistinctDeviceTypeAttributes(devicetype, strings.Split(distinctAttr, ","))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		result, err, errCode := control.SetDeviceType(token, devicetype)
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
// @Summary      set device-type
// @Description  set device-type
// @Tags         set, device-types
// @Produce      json
// @Security Bearer
// @Param        id path string true "DeviceType Id"
// @Param        distinct_attributes query string false "comma separated list of attribute keys; no other device-type with the same attribute key/value may exist"
// @Param        message body models.DeviceType true "element"
// @Success      200 {object}  models.DeviceType
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-types/{id} [PUT]
func (this *DeviceTypeEndpoints) Set(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /device-types/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		devicetype := models.DeviceType{}
		err := json.NewDecoder(request.Body).Decode(&devicetype)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if id != devicetype.Id || devicetype.Id == "" {
			http.Error(writer, "expect body and path to contain the same device-type id", http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)
		distinctAttr := request.URL.Query().Get("distinct_attributes")
		if distinctAttr != "" {
			err = control.ValidateDistinctDeviceTypeAttributes(devicetype, strings.Split(distinctAttr, ","))
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		result, err, errCode := control.SetDeviceType(token, devicetype)
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
// @Summary      delete device-type
// @Description  delete device-type
// @Tags         delete, device-types
// @Produce      json
// @Security Bearer
// @Param        id path string true "DeviceType Id"
// @Success      200
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-types/{id} [DELETE]
func (this *DeviceTypeEndpoints) Delete(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /device-types/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token := util.GetAuthToken(request)

		err, errCode := control.DeleteDeviceType(token, id)
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
