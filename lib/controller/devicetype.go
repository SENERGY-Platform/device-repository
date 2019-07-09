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

func (this *Controller) ListDeviceTypes(jwt jwt_http_router.Jwt, limit int64, offset int64, sort string) (result []model.DeviceType, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err = this.db.ListDeviceTypes(ctx, limit, offset, sort)
	return
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDeviceType(deviceType model.DeviceType, owner string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.SetDeviceType(ctx, deviceType)
}

func (this *Controller) DeleteDeviceType(id string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.RemoveDeviceType(ctx, id)
}
