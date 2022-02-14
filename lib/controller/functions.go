/*
 * Copyright 2022 InfAI (CC SES)
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

package controller

import (
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"net/http"
	"strings"
)

func (this *Controller) SetFunction(function model.Function, owner string) error {
	model.SetFunctionRdfType(&function)
	ctx, _ := getTimeoutContext()
	return this.db.SetFunction(ctx, function)
}

func (this *Controller) DeleteFunction(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveFunction(ctx, id)
}

func (this *Controller) GetFunctionsByType(rdfType string) (result []model.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllFunctionsByType(ctx, rdfType)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetAspectsMeasuringFunctions(aspect string) (result []model.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllMeasuringFunctionsByAspect(ctx, aspect)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetDeviceClassesFunctions(deviceClass string) (result []model.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllFunctionsByDeviceClass(ctx, deviceClass)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetDeviceClassesControllingFunctions(deviceClass string) (result []model.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllControllingFunctionsByDeviceClass(ctx, deviceClass)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetFunction(id string) (result model.Function, err error, code int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetFunction(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ValidateFunction(function model.Function) (err error, code int) {
	if function.Id == "" {
		return errors.New("missing function id"), http.StatusBadRequest
	}
	if !strings.HasPrefix(function.Id, model.URN_PREFIX) {
		return errors.New("invalid function id"), http.StatusBadRequest
	}
	if function.Name == "" {
		return errors.New("missing function name"), http.StatusBadRequest
	}
	if !(function.RdfType == model.SES_ONTOLOGY_CONTROLLING_FUNCTION || function.RdfType == model.SES_ONTOLOGY_MEASURING_FUNCTION) {
		return errors.New("wrong function type"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}
