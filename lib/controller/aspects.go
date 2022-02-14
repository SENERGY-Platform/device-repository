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

func (this *Controller) SetAspect(aspect model.Aspect, owner string) error {
	ctx, _ := getTimeoutContext()
	return this.db.SetAspect(ctx, aspect)
}

func (this *Controller) DeleteAspect(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveAspect(ctx, id)
}

func (this *Controller) GetAspects() (result []model.Aspect, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllAspects(ctx)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetAspect(id string) (result model.Aspect, err error, code int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetAspect(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) GetAspectsWithMeasuringFunction() (result []model.Aspect, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAspectsWithMeasuringFunction(ctx)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) ValidateAspect(aspect model.Aspect) (err error, code int) {
	if aspect.Id == "" {
		return errors.New("missing aspect id"), http.StatusBadRequest
	}
	if !strings.HasPrefix(aspect.Id, model.URN_PREFIX) {
		return errors.New("invalid aspect id"), http.StatusBadRequest
	}
	if aspect.Name == "" {
		return errors.New("missing aspect name"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}
