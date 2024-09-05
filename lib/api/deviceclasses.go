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
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, &DeviceClassEndpoints{})
}

type DeviceClassEndpoints struct{}

// List godoc
// @Summary      list device-classes
// @Description  list device-classes
// @Tags         list, device-classes
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
func (this *DeviceClassEndpoints) List(config config.Config, router *http.ServeMux, control Controller) {
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
// @Tags         get, device-classes
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
func (this *DeviceClassEndpoints) Get(config config.Config, router *http.ServeMux, control Controller) {
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
// @Tags         list, device-classes, functions
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
func (this *DeviceClassEndpoints) GetFunctions(config config.Config, router *http.ServeMux, control Controller) {
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
// @Tags         list, device-classes, functions
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
func (this *DeviceClassEndpoints) GetControllingFunctions(config config.Config, router *http.ServeMux, control Controller) {
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
// @Tags         validate, device-classes
// @Accept       json
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not an update but a validation"
// @Param        message body models.DeviceClass true "Device-Class to be validated"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /device-classes [PUT]
func (this *DeviceClassEndpoints) Validate(config config.Config, router *http.ServeMux, control Controller) {
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

// ValidateDelete godoc
// @Summary      validate device-class delete
// @Description  validate if device-class may be deleted
// @Tags         validate, device-classes
// @Security Bearer
// @Param        dry-run query bool true "must be true; reminder, that this is not a delete but a validation"
// @Param        id path string true "DeviceClass Id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /device-classes/{id} [DELETE]
func (this *DeviceClassEndpoints) ValidateDelete(config config.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /device-classes/{id}", func(writer http.ResponseWriter, request *http.Request) {
		dryRun, err := strconv.ParseBool(request.URL.Query().Get("dry-run"))
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if !dryRun {
			http.Error(writer, "only with query-parameter 'dry-run=true' allowed", http.StatusNotImplemented)
			return
		}
		id := request.PathValue("id")
		err, code := control.ValidateDeviceClassDelete(id)
		if err != nil {
			http.Error(writer, err.Error(), code)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
