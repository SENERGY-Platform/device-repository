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
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"net/http"
	"strings"
)

func (this *Controller) SetFunction(token string, function models.Function) (result models.Function, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return result, errors.New("token is not an admin"), http.StatusUnauthorized
	}

	//ensure ids
	if !DisableFeaturesForTestEnv {
		function.GenerateId()
	}
	model.SetFunctionRdfType(&function)
	err, code = this.ValidateFunction(function)
	if err != nil {
		return result, err, code
	}
	err = this.setFunction(function)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return function, nil, http.StatusOK
}

func (this *Controller) DeleteFunction(token string, id string) (error, int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return errors.New("token is not an admin"), http.StatusUnauthorized
	}
	err, code := this.ValidateFunctionDelete(id)
	if err != nil {
		return err, code
	}
	err = this.deleteFunction(id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) setFunctionSyncHandler(f models.Function) (err error) {
	return this.publisher.PublishFunction(f)
}

func (this *Controller) setFunction(function models.Function) error {
	ctx, _ := getTimeoutContext()
	return this.db.SetFunction(ctx, function, this.setFunctionSyncHandler)
}

func (this *Controller) deleteFunctionSyncHandler(f models.Function) (err error) {
	return this.publisher.PublishFunctionDelete(f.Id)
}

func (this *Controller) deleteFunction(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveFunction(ctx, id, this.deleteFunctionSyncHandler)
}

func (this *Controller) ListFunctions(options model.FunctionListOptions) (result []models.Function, total int64, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListFunctions(ctx, options)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) GetFunctionsByType(rdfType string) (result []models.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllFunctionsByType(ctx, rdfType)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

// returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
func (this *Controller) GetAspectNodesMeasuringFunctions(aspect string, ancestors bool, descendants bool) (result []models.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllMeasuringFunctionsByAspect(ctx, aspect, ancestors, descendants)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetDeviceClassesFunctions(deviceClass string) (result []models.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllFunctionsByDeviceClass(ctx, deviceClass)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetDeviceClassesControllingFunctions(deviceClass string) (result []models.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllControllingFunctionsByDeviceClass(ctx, deviceClass)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetFunction(id string) (result models.Function, err error, code int) {
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

func (this *Controller) ValidateFunction(function models.Function) (err error, code int) {
	if DisableFeaturesForTestEnv {
		return nil, http.StatusOK
	}
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

func (this *Controller) ValidateFunctionDelete(id string) (err error, code int) {
	ctx, _ := getTimeoutContext()
	isUsed, where, err := this.db.FunctionIsUsed(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if isUsed {
		return errors.New("still in use: " + strings.Join(where, ",")), http.StatusBadRequest
	}
	return nil, http.StatusOK
}
