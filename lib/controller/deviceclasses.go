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
	"net/http"
	"strings"
)

func (this *Controller) ListDeviceClasses(listOptions model.DeviceClassListOptions) (result []models.DeviceClass, total int64, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListDeviceClasses(ctx, listOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) SetDeviceClass(class models.DeviceClass, owner string) error {
	ctx, _ := getTimeoutContext()
	return this.db.SetDeviceClass(ctx, class)
}

func (this *Controller) DeleteDeviceClass(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveDeviceClass(ctx, id)
}

func (this *Controller) GetDeviceClasses() (result []models.DeviceClass, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllDeviceClasses(ctx)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetDeviceClass(id string) (result models.DeviceClass, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetDeviceClass(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ValidateDeviceClass(deviceClass models.DeviceClass) (err error, code int) {
	if deviceClass.Id == "" {
		return errors.New("missing device class id"), http.StatusBadRequest
	}
	if !strings.HasPrefix(deviceClass.Id, model.URN_PREFIX) {
		return errors.New("invalid deviceClass id"), http.StatusBadRequest
	}
	if deviceClass.Name == "" {
		return errors.New("missing device class name"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func (this *Controller) GetDeviceClassesWithControllingFunctions() (result []models.DeviceClass, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllDeviceClassesUsedWithControllingFunctions(ctx)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) ValidateDeviceClassDelete(id string) (err error, code int) {
	ctx, _ := getTimeoutContext()
	isUsed, where, err := this.db.DeviceClassIsUsed(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if isUsed {
		return errors.New("still in use: " + strings.Join(where, ",")), http.StatusBadRequest
	}
	return nil, http.StatusOK
}
