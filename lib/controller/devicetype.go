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
	"net/http"
	"time"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadDeviceType(id string, token string) (result model.DeviceType, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	deviceType, exists, err := this.db.GetDeviceType(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return deviceType, nil, http.StatusOK
}

func (this *Controller) ListDeviceTypes(token string, limit int64, offset int64, sort string) (result []model.DeviceType, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListDeviceTypes(ctx, limit, offset, sort)
	return
}

func (this *Controller) ValidateDeviceType(dt model.DeviceType) (err error, code int) {
	if dt.Id == "" {
		return errors.New("missing device-type id"), http.StatusBadRequest
	}
	if dt.Name == "" {
		return errors.New("missing device-type name"), http.StatusBadRequest
	}
	if len(dt.Services) == 0 {
		return errors.New("expect at least one service"), http.StatusBadRequest
	}
	for _, service := range dt.Services {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
		deviceTypes, err := this.db.GetDeviceTypesByServiceId(ctx, service.Id)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if len(deviceTypes) > 1 {
			return errors.New("reused service id"), http.StatusBadRequest
		}
		if len(deviceTypes) == 1 && deviceTypes[0].Id != dt.Id {
			return errors.New("reused service id"), http.StatusBadRequest
		}
		err, code = this.ValidateService(service)
		if err != nil {
			return err, code
		}
	}
	err = ValidateServiceGroups(dt.ServiceGroups, dt.Services)
	if err != nil {
		return err, http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func ValidateServiceGroups(groups []model.ServiceGroup, services []model.Service) error {
	groupIndex := map[string]bool{}
	for _, g := range groups {
		if _, ok := groupIndex[g.Key]; ok {
			return errors.New("duplicate service-group key: " + g.Key)
		}
		groupIndex[g.Key] = true
	}
	for _, s := range services {
		if s.ServiceGroupKey != "" {
			_, ok := groupIndex[s.ServiceGroupKey]
			if !ok {
				return errors.New("unknown service-group key: " + s.ServiceGroupKey)
			}
		}
	}
	return nil
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDeviceType(deviceType model.DeviceType, owner string) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.SetDeviceType(ctx, deviceType)
}

func (this *Controller) DeleteDeviceType(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveDeviceType(ctx, id)
}
