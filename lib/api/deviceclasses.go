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
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	endpoints = append(endpoints, &DeviceClassEndpoints{})
}

type DeviceClassEndpoints struct{}

// ListDeviceClasses godoc
// @Summary      list device-classes
// @Description  list device-classes
// @Tags         device-classes
// @Produce      json
// @Security Bearer
// @Param        limit query integer false "default 100, will be ignored if 'ids' is set"
// @Param        offset query integer false "default 0, will be ignored if 'ids' is set"
// @Param        search query string false "filter"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Param        used_with_controlling_function query bool false "filter; only 'true' is a valid value; if set, returns device-classes used in combination with controlling-function"
// @Success      200 {array}  models.DeviceClass
// @Header       200 {integer}  X-Total-Count  "count of all matching elements; used for pagination"
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /v2/device-classes [GET]
func (this *DeviceClassEndpoints) ListDeviceClasses(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /v2/device-classes", func(writer http.ResponseWriter, request *http.Request) {
		listoptions := model.DeviceClassListOptions{
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

		if request.URL.Query().Has("used_with_controlling_function") {
			listoptions.UsedWithControllingFunction, err = strconv.ParseBool(request.URL.Query().Get("used_with_controlling_function"))
			if err != nil {
				http.Error(writer, "unable to parse used_with_controlling_function as bool: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		listoptions.Search = request.URL.Query().Get("search")
		listoptions.SortBy = request.URL.Query().Get("sort")
		if listoptions.SortBy == "" {
			listoptions.SortBy = "name.asc"
		}
		result, total, err, errCode := control.ListDeviceClasses(listoptions)
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
// @Deprecated
// @Summary      deprecated list device-classes; use GET /v2/device-classes
// @Description  deprecated list device-classes; use GET /v2/device-classes
// @Tags         device-classes
// @Produce      json
// @Security Bearer
// @Param        function query string false "filter; only 'controlling-function' is a valid value; if set, returns device-classes used in combination with controlling-function"
// @Success      200 {array}  models.DeviceClass
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-classes [GET]
func (this *DeviceClassEndpoints) List(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /device-classes", func(writer http.ResponseWriter, request *http.Request) {
		var result []models.DeviceClass
		var err error
		var errCode int

		function := request.URL.Query().Get("function")

		if function == "" {
			result, err, errCode = control.GetDeviceClasses()
			if err != nil {
				http.Error(writer, err.Error(), errCode)
				return
			}
		} else {
			if function == "controlling-function" {
				result, err, errCode = control.GetDeviceClassesWithControllingFunctions()
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
// @Summary      get device-class
// @Description  get device-class
// @Tags         device-classes
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Class Id"
// @Success      200 {object}  models.DeviceClass
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-classes/{id} [GET]
func (this *DeviceClassEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /device-classes/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.GetDeviceClass(id)
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

// GetFunctions godoc
// @Summary      list device-class functions
// @Description  list functions used in combination with this device-class
// @Tags         device-classes
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Class Id"
// @Success      200 {array}  models.Function
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-classes/{id}/functions [GET]
func (this *DeviceClassEndpoints) GetFunctions(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /device-classes/{id}/functions", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.GetDeviceClassesFunctions(id)
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

// GetControllingFunctions godoc
// @Summary      list device-class functions
// @Description  list controlling-functions used in combination with this device-class
// @Tags         device-classes
// @Produce      json
// @Security Bearer
// @Param        id path string true "Device Class Id"
// @Success      200 {array}  models.Function
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-classes/{id}/controlling-functions [GET]
func (this *DeviceClassEndpoints) GetControllingFunctions(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /device-classes/{id}/controlling-functions", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		result, err, errCode := control.GetDeviceClassesControllingFunctions(id)
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
// @Summary      validate device-class
// @Description  validate device-class
// @Tags         device-classes
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.DeviceClass true "Device-Class to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /device-classes [PUT]
func (this *DeviceClassEndpoints) Validate(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /device-classes", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		deviceclass := models.DeviceClass{}
		err = json.NewDecoder(request.Body).Decode(&deviceclass)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err, code := control.ValidateDeviceClass(deviceclass)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// Delete godoc
// @Summary      delete device-class
// @Description  delete device-class; may only be called by admins; can also be used to only validate deletes
// @Tags         device-classes
// @Security Bearer
// @Param        dry-run query bool false "only validate deletion"
// @Param        id path string true "DeviceClasses Id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /device-classes/{id} [DELETE]
func (this *DeviceClassEndpoints) Delete(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /device-classes/{id}", func(writer http.ResponseWriter, request *http.Request) {
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
			err, code := control.ValidateDeviceClassDelete(id)
			if err != nil {
				http.Error(writer, err.Error(), code)
				return
			}
			writer.WriteHeader(http.StatusOK)
			return
		}
		token := util.GetAuthToken(request)
		err, code := control.DeleteDeviceClass(token, id)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}

// Create godoc
// @Summary      create device-class
// @Description  create device-class
// @Tags         device-classes
// @Produce      json
// @Security Bearer
// @Param        message body models.DeviceClass true "element"
// @Success      200 {object}  models.DeviceClass
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-classes [POST]
func (this *DeviceClassEndpoints) Create(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /device-classes", func(writer http.ResponseWriter, request *http.Request) {
		deviceClass := models.DeviceClass{}
		err := json.NewDecoder(request.Body).Decode(&deviceClass)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		token := util.GetAuthToken(request)

		result, err, errCode := control.SetDeviceClass(token, deviceClass)
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
// @Summary      set device-class
// @Description  set device-class
// @Tags         device-classes
// @Produce      json
// @Security Bearer
// @Param        id path string true "DeviceClass Id"
// @Param        message body models.DeviceClass true "element"
// @Success      200 {object}  models.DeviceClass
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Router       /device-classes/{id} [PUT]
func (this *DeviceClassEndpoints) Set(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("PUT /device-classes/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		deviceClass := models.DeviceClass{}
		err := json.NewDecoder(request.Body).Decode(&deviceClass)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		token := util.GetAuthToken(request)

		if deviceClass.Id != id {
			http.Error(writer, "id in body unequal to id in request endpoint", http.StatusBadRequest)
			return
		}

		result, err, errCode := control.SetDeviceClass(token, deviceClass)
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
