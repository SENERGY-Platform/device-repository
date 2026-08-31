/*
 * Copyright 2026 InfAI (CC SES)
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
	"net/http"
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func (this *Controller) ListAspectClasses(listOptions model.AspectClassListOptions) (result []models.AspectClass, total int64, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListAspectClasses(ctx, listOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) GetAspectClass(id string) (result models.AspectClass, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetAspectClass(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) SetAspectClass(token string, aspectClass models.AspectClass) (result models.AspectClass, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return result, errors.New("token is not an admin"), http.StatusUnauthorized
	}

	//ensure ids
	aspectClass.GenerateId()
	if !this.config.DisableStrictValidationForTesting {
		err, code = this.ValidateAspectClass(aspectClass)
		if err != nil {
			return result, err, code
		}
	}
	err = this.setAspectClass(aspectClass)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return aspectClass, nil, http.StatusOK
}

func (this *Controller) setAspectClass(aspectClass models.AspectClass) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.SetAspectClass(ctx, aspectClass, this.setAspectClassSyncHandler)
}

func (this *Controller) setAspectClassSyncHandler(aspectClass models.AspectClass) error {
	return this.publisher.PublishAspectClass(aspectClass)
}

func (this *Controller) DeleteAspectClass(token string, id string) (error, int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return errors.New("token is not an admin"), http.StatusUnauthorized
	}
	err, code := this.ValidateAspectClassDelete(id)
	if err != nil {
		return err, code
	}
	ctx, _ := getTimeoutContext()
	err = this.db.RemoveAspectClass(ctx, id, this.deleteAspectClassSyncHandler)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteAspectClassSyncHandler(aspectClass models.AspectClass) error {
	return this.publisher.PublishAspectClassDelete(aspectClass.Id)
}

func (this *Controller) ValidateAspectClass(aspectClass models.AspectClass) (err error, code int) {
	if aspectClass.Id == "" {
		return errors.New("missing aspect class id"), http.StatusBadRequest
	}
	if !strings.HasPrefix(aspectClass.Id, model.URN_PREFIX) {
		return errors.New("invalid aspect class id"), http.StatusBadRequest
	}
	if aspectClass.Name == "" {
		return errors.New("missing aspect class name"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

// ValidateAspectClassDelete has nothing to check yet: aspects reference an aspect-class
// through Aspect.AspectClassId, but that field is not evaluated by this service so far.
// The usage check belongs here once it is.
func (this *Controller) ValidateAspectClassDelete(id string) (err error, code int) {
	return nil, http.StatusOK
}
