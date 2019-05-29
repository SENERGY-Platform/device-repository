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
		return result, err, http.StatusNotFound
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
