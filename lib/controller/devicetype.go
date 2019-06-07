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
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
	"time"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadDeviceType(id string, jwt jwt_http_router.Jwt) (result model.DeviceType, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	deviceType, exists, err := this.db.GetDeviceType(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return deviceType, nil, http.StatusOK
}

func (this *Controller) ListDeviceTypes(jwt jwt_http_router.Jwt, options listoptions.ListOptions) (result []model.DeviceType, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err = this.db.ListDeviceTypes(ctx, options)
	opterr := options.EvalStrict()
	if opterr != nil {
		return result, opterr, http.StatusBadRequest
	}
	return
}

func (this *Controller) PublishDeviceTypeUpdate(jwt jwt_http_router.Jwt, id string, dt model.DeviceType) (result model.DeviceType, err error, errCode int) {
	if err, errCode = this.validateDeviceTypeUpdate(dt, id); err != nil {
		return result, err, errCode
	}
	dt.Id = id
	allowed, err := this.security.CheckBool(jwt, this.config.DeviceTypeTopic, dt.Id, model.WRITE)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !allowed {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	err = this.source.PublishDeviceType(dt, jwt.UserId)
	if err != nil {
		errCode = http.StatusInternalServerError
	}
	result = dt
	return
}

func (this *Controller) PublishDeviceTypeCreate(jwt jwt_http_router.Jwt, dt model.DeviceType) (result model.DeviceType, err error, errCode int) {
	if err, errCode = this.validateDeviceTypeCreate(dt); err != nil {
		return result, err, errCode
	}
	dt.Id = generateId()
	err = this.source.PublishDeviceType(dt, jwt.UserId)
	if err != nil {
		errCode = http.StatusInternalServerError
	}
	result = dt
	return
}

func (this *Controller) PublishDeviceTypeDelete(jwt jwt_http_router.Jwt, id string) (err error, errCode int) {
	allowed, err := this.security.CheckBool(jwt, this.config.DeviceTypeTopic, id, model.ADMINISTRATE)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !allowed {
		return errors.New("access denied"), http.StatusForbidden
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	devices, err := this.db.ListDevicesOfDeviceType(ctx, id, listoptions.New().Limit(1))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if len(devices) > 0 {
		return errors.New("found devices using this device-type"), http.StatusBadRequest
	}
	err = this.source.PublishDeviceTypeDelete(id)
	if err != nil {
		errCode = http.StatusInternalServerError
	}
	return
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDeviceType(deviceType model.DeviceType, owner string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	transaction, finish, err := this.db.Transaction(ctx)
	if err != nil {
		return err
	}
	err = this.publishMissingValueTypesOfDeviceType(transaction, deviceType, owner)
	if err != nil {
		_ = finish(false)
		return err
	}

	old, exists, err := this.db.GetDeviceType(transaction, deviceType.Id)
	if err != nil {
		_ = finish(false)
		return
	}

	err = this.updateEndpointsOfDeviceType(transaction, old, deviceType)
	if err != nil {
		_ = finish(false)
		return err
	}

	if exists && old.ImgUrl != deviceType.ImgUrl {
		err = this.updateDefaultDeviceImages(transaction, deviceType.Id, old.ImgUrl, deviceType.ImgUrl)
		if err != nil {
			_ = finish(false)
			return err
		}
	}

	err = this.db.SetDeviceType(transaction, deviceType)
	if err != nil {
		_ = finish(false)
		return err
	}
	return finish(true)
}

func (this *Controller) DeleteDeviceType(id string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.RemoveDeviceType(ctx, id)
}

func (this *Controller) validateDeviceTypeCreate(dt model.DeviceType) (error, int) {
	if dt.Id != "" {
		return errors.New("expected empty dt.id"), http.StatusBadRequest
	}
	if dt.Name == "" {
		return errors.New("missing expected dt.name"), http.StatusBadRequest
	}
	if valid, msg := dt.IsValid(); !valid {
		return errors.New(msg), http.StatusBadRequest
	}
	return nil, 200
}

func (this *Controller) validateDeviceTypeUpdate(dt model.DeviceType, id string) (error, int) {
	if dt.Id != id {
		return errors.New("dt.id different from update id"), http.StatusBadRequest
	}
	if dt.Name == "" {
		return errors.New("missing expected dt.name"), http.StatusBadRequest
	}
	if valid, msg := dt.IsValid(); !valid {
		return errors.New(msg), http.StatusBadRequest
	}
	return nil, 200
}
