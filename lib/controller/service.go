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
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
	"time"
)

func (this *Controller) ReadService(id string, jwt jwt_http_router.Jwt) (result model.Service, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	deviceType, exists, err := this.db.GetDeviceTypeWithService(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	for _, service := range deviceType.Services {
		if service.Id == id {
			return service, nil, http.StatusOK
		}
	}
	return result, errors.New("found device-type without service in search for service"), http.StatusInternalServerError
}

func (this *Controller) setServiceIds(in model.Service) (out model.Service) {
	out = in
	if out.Id == "" {
		out.Id = generateId()
	}
	for i, field := range out.Input {
		if field.Id == "" {
			field.Id = generateId()
		}
		field.Type = this.setValueTypeIds(field.Type)
		out.Input[i] = field
	}
	for i, field := range out.Output {
		if field.Id == "" {
			field.Id = generateId()
		}
		field.Type = this.setValueTypeIds(field.Type)
		out.Output[i] = field
	}
	return out
}

func (this *Controller) validateService(service model.Service) error {
	for _, inp := range service.Input {
		if inp.Name == "" {
			return errors.New("missing input.Name")
		}
		if inp.Id == "" {
			return errors.New("missing input.id")
		}
		if inp.Format == "" {
			return errors.New("missing input.Format")
		}
		if err := this.validateValueType(inp.Type); err != nil {
			return err
		}
	}
	for _, outp := range service.Output {
		if outp.Name == "" {
			return errors.New("missing output.Name")
		}
		if outp.Id == "" {
			return errors.New("missing output.id")
		}
		if outp.Format == "" {
			return errors.New("missing output.Format")
		}
		if err := this.validateValueType(outp.Type); err != nil {
			return err
		}
	}
	return nil
}
