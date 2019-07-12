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

package controller

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
	"time"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadDevice(id string, jwt jwt_http_router.Jwt) (result model.Device, err error, errCode int) {
	ok, err := this.security.CheckBool(jwt, this.config.DeviceTopic, id, model.READ)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	device, exists, err := this.db.GetDevice(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return device, nil, http.StatusOK
}

func (this *Controller) ValidateDevice(device model.Device) (err error, code int) {
	if device.Id == "" {
		return errors.New("missing device id"), http.StatusBadRequest
	}
	if device.Name == "" {
		return errors.New("missing device name"), http.StatusBadRequest
	}
	if device.LocalId == "" {
		return errors.New("missing device local id"), http.StatusBadRequest
	}
	if device.DeviceTypeId == "" {
		return errors.New("missing device type id"), http.StatusBadRequest
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, ok, err := this.db.GetDeviceType(ctx, device.DeviceTypeId)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("unknown device type id"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDevice(device model.Device, owner string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.SetDevice(ctx, device)
}

func (this *Controller) DeleteDevice(id string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.RemoveDevice(ctx, id)
}
