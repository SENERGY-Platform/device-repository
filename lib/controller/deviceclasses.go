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

func (this *Controller) SetDeviceClass(class model.DeviceClass, owner string) error {
	ctx, _ := getTimeoutContext()
	return this.db.SetDeviceClass(ctx, class)
}

func (this *Controller) DeleteDeviceClass(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveDeviceClass(ctx, id)
}

func (this *Controller) GetDeviceClasses() (result []model.DeviceClass, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllDeviceClasses(ctx)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}

func (this *Controller) GetDeviceClass(id string) (result model.DeviceClass, err error, errCode int) {
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

func (this *Controller) ValidateDeviceClass(deviceClass model.DeviceClass) (err error, code int) {
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

func (this *Controller) GetDeviceClassesWithControllingFunctions() (result []model.DeviceClass, err error, code int) {
	code = http.StatusOK
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListAllDeviceClassesUsedWithControllingFunctions(ctx)
	if err != nil {
		code = http.StatusInternalServerError
	}
	return
}
