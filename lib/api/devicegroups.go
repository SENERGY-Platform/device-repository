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
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &DeviceGroupEndpoints{})
}

type DeviceGroupEndpoints struct{}

// List godoc
// @Summary      list device-group
// @Description  list device-group
// @Tags         device-groups
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Param        device-ids query string false "filter; comma-seperated list"
// @Param        ignore-generated query bool false "filter; remove generated groups from result"
// @Param        attr-keys query string false "filter; comma-seperated list; lists elements only if they have an attribute key that is in the given list"
// @Param        attr-values query string false "filter; comma-seperated list; lists elements only if they have an attribute value that is in the given list"
// @Param        criteria query string false "filter; json encoded []model.FilterCriteria"
// @Param        p query string false "default 'r'; used to check permissions on request; valid values are 'r', 'w', 'x', 'a' for read, write, execute, administrate"
// @Param        filter_generic_duplicate_criteria query bool false "remove criteria that are more generalized variations of already listed criteria (ref SNRGY-3027)"
// @Success      200 {array}  models.DeviceGroup
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-groups [GET]
func (this *DeviceGroupEndpoints) List(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /device-groups", func(writer http.ResponseWriter, request *http.Request) {
		deviceGroupListOptions := model.DeviceGroupListOptions{
			Limit:                          100,
			Offset:                         0,
			FilterGenericDuplicateCriteria: request.URL.Query().Get("filter_generic_duplicate_criteria") == "true",
		}
		var err error
		limitParam := request.URL.Query().Get("limit")
		if limitParam != "" {
			deviceGroupListOptions.Limit, err = strconv.ParseInt(limitParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse limit:"+err.Error(), http.StatusBadRequest)
			return
		}

		offsetParam := request.URL.Query().Get("offset")
		if offsetParam != "" {
			deviceGroupListOptions.Offset, err = strconv.ParseInt(offsetParam, 10, 64)
		}
		if err != nil {
			http.Error(writer, "unable to parse offset:"+err.Error(), http.StatusBadRequest)
			return
		}

		idsParam := request.URL.Query().Get("ids")
		if request.URL.Query().Has("ids") {
			if idsParam != "" {
				deviceGroupListOptions.Ids = strings.Split(strings.TrimSpace(idsParam), ",")
			} else {
				deviceGroupListOptions.Ids = []string{}
			}
		}

		deviceGroupListOptions.Search = request.URL.Query().Get("search")
		deviceGroupListOptions.SortBy = request.URL.Query().Get("sort")
		if deviceGroupListOptions.SortBy == "" {
			deviceGroupListOptions.SortBy = "name.asc"
		}

		if request.URL.Query().Has("ignore-generated") {
			deviceGroupListOptions.IgnoreGenerated, err = strconv.ParseBool(request.URL.Query().Get("ignore-generated"))
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
			deviceGroupListOptions.Criteria = criteriaList
		}

		attrKeysParam := request.URL.Query().Get("attr-keys")
		if request.URL.Query().Has("attr-keys") {
			if attrKeysParam != "" {
				deviceGroupListOptions.AttributeKeys = strings.Split(strings.TrimSpace(attrKeysParam), ",")
			} else {
				deviceGroupListOptions.AttributeKeys = []string{}
			}
		}
		attrValuesParam := request.URL.Query().Get("attr-values")
		if request.URL.Query().Has("attr-values") {
			if attrValuesParam != "" {
				deviceGroupListOptions.AttributeValues = strings.Split(strings.TrimSpace(attrValuesParam), ",")
			} else {
				deviceGroupListOptions.AttributeValues = []string{}
			}
		}

		deviceIdsParam := request.URL.Query().Get("device-ids")
		if request.URL.Query().Has("device-ids") && deviceIdsParam != "" {
			deviceGroupListOptions.DeviceIds = strings.Split(strings.TrimSpace(deviceIdsParam), ",")
		}

		deviceGroupListOptions.Permission, err = model.GetPermissionFlagFromQuery(request.URL.Query())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if deviceGroupListOptions.Permission == models.UnsetPermissionFlag {
			deviceGroupListOptions.Permission = model.READ
		}

		result, total, err, errCode := control.ListDeviceGroups(util.GetAuthToken(request), deviceGroupListOptions)
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

// Get godoc
// @Summary      get device-group
// @Description  get device-group
// @Tags         device-groups
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Group Id"
// @Param        filter_generic_duplicate_criteria query bool false "remove criteria that are more generalized variations of already listed criteria (ref SNRGY-3027)"
// @Success      200 {object}  models.DeviceGroup
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-groups/{id} [GET]
func (this *DeviceGroupEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /device-groups/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")

		//ref https://bitnify.atlassian.net/browse/SNRGY-3027
		filterGenericDuplicateCriteria := request.URL.Query().Get("filter_generic_duplicate_criteria") == "true"

		result, err, errCode := control.ReadDeviceGroup(id, util.GetAuthToken(request), filterGenericDuplicateCriteria)
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

// Validate godoc
// @Summary      validate device-group
// @Description  validate device-group
// @Tags         device-groups
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.DeviceGroup true "DeviceGroup to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /device-groups [PUT]
func (this *DeviceGroupEndpoints) Validate(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /device-groups", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		group := models.DeviceGroup{}
		err = json.NewDecoder(request.Body).Decode(&group)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateDeviceGroup(util.GetAuthToken(request), group)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// Delete godoc
// @Summary      delete device-group
// @Description  delete device-group; may only be called by admins; can also be used to only validate deletes
// @Tags         device-groups
// @Security Bearer
// @Param        dry-run query bool false "only validate deletion"
// @Param        id path string true "DeviceGroup Id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /device-groups/{id} [DELETE]
func (this *DeviceGroupEndpoints) Delete(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /device-groups/{id}", func(writer http.ResponseWriter, request *http.Request) {
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
		token := util.GetAuthToken(request)
		if dryRun {
			err, code := control.ValidateDeviceGroupDelete(token, id)
			if err != nil {
				http.Error(writer, err.Error(), code)
				return
			}
			writer.WriteHeader(http.StatusOK)
			return
		}
		err, code := control.DeleteDeviceGroup(token, id)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// Create godoc
// @Summary      create device-group
// @Description  create device-group
// @Tags         device-groups
// @Produce      json
// @Security Bearer
// @Param        message body models.DeviceGroup true "element"
// @Success      200 {object}  models.DeviceGroup
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-groups [POST]
func (this *DeviceGroupEndpoints) Create(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /device-groups", func(writer http.ResponseWriter, request *http.Request) {
		deviceGroup := models.DeviceGroup{}
		err := json.NewDecoder(request.Body).Decode(&deviceGroup)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)

		if deviceGroup.Id != "" {
			http.Error(writer, "body may not contain a preset id. please use the PUT method for updates", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetDeviceGroup(token, deviceGroup)
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
// @Summary      set device-group
// @Description  set device-group
// @Tags         device-groups
// @Produce      json
// @Security Bearer
// @Param        id path string true "DeviceGroup Id"
// @Param        message body models.DeviceGroup true "element"
// @Success      200 {object}  models.DeviceGroup
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-groups/{id} [PUT]
func (this *DeviceGroupEndpoints) Set(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /device-groups/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		deviceGroup := models.DeviceGroup{}
		err := json.NewDecoder(request.Body).Decode(&deviceGroup)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)

		if deviceGroup.Id != id {
			http.Error(writer, "id in body unequal to id in request endpoint", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetDeviceGroup(token, deviceGroup)
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
