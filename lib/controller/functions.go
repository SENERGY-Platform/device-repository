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
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func (this *Controller) SetFunction(ctx context.Context, token string, function models.Function) (result models.Function, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return result, errors.New("token is not an admin"), http.StatusUnauthorized
	}

	model.SetFunctionRdfType(&function)
	//ensure ids
	function.GenerateId()
	if !this.config.DisableStrictValidationForTesting {
		err, code = this.ValidateFunction(ctx, function)
		if err != nil {
			return result, err, code
		}
	}

	err = this.setFunction(ctx, function)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return function, nil, http.StatusOK
}

func (this *Controller) DeleteFunction(ctx context.Context, token string, id string) (error, int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return errors.New("token is not an admin"), http.StatusUnauthorized
	}
	err, code := this.ValidateFunctionDelete(ctx, id)
	if err != nil {
		return err, code
	}
	err = this.deleteFunction(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) setFunctionSyncHandler(ctx context.Context, f models.Function) (err error) {
	return this.publisher.PublishFunction(f)
}

func (this *Controller) setFunction(ctx context.Context, function models.Function) error {
	ctx, _ = getTimeoutContext(ctx)
	return this.db.SetFunction(ctx, function, withContext(ctx, this.setFunctionSyncHandler))
}

func (this *Controller) deleteFunctionSyncHandler(ctx context.Context, f models.Function) (err error) {
	return this.publisher.PublishFunctionDelete(f.Id)
}

func (this *Controller) deleteFunction(ctx context.Context, id string) error {
	ctx, _ = getTimeoutContext(ctx)
	return this.db.RemoveFunction(ctx, id, withContext(ctx, this.deleteFunctionSyncHandler))
}

func (this *Controller) ListFunctions(ctx context.Context, options model.FunctionListOptions) (result []models.Function, total int64, err error, errCode int) {
	ctx, _ = getTimeoutContext(ctx)
	result, total, err = this.db.ListFunctions(ctx, options)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) GetFunctionsByType(ctx context.Context, rdfType string) (result []models.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.ListAllFunctionsByType(ctx, rdfType)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

// returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
func (this *Controller) GetAspectNodesMeasuringFunctions(ctx context.Context, aspect string, ancestors bool, descendants bool) (result []models.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.ListAllMeasuringFunctionsByAspect(ctx, aspect, ancestors, descendants)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

// returns all controlling functions used in combination with given aspect (and optional its descendants and ancestors)
func (this *Controller) GetAspectNodesControllingFunctions(ctx context.Context, aspect string, ancestors bool, descendants bool) (result []models.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.ListAllControllingFunctionsByAspect(ctx, aspect, ancestors, descendants)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetDeviceClassesFunctions(ctx context.Context, deviceClass string) (result []models.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.ListAllFunctionsByDeviceClass(ctx, deviceClass)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetDeviceClassesControllingFunctions(ctx context.Context, deviceClass string) (result []models.Function, err error, code int) {
	code = http.StatusOK
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.ListAllControllingFunctionsByDeviceClass(ctx, deviceClass)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetFunction(ctx context.Context, id string) (result models.Function, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	result, exists, err := this.db.GetFunction(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ValidateFunction(ctx context.Context, function models.Function) (err error, code int) {
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

func (this *Controller) ValidateFunctionDelete(ctx context.Context, id string) (err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	isUsed, where, err := this.db.FunctionIsUsed(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if isUsed {
		return errors.New("still in use: " + strings.Join(where, ",")), http.StatusBadRequest
	}
	return nil, http.StatusOK
}
