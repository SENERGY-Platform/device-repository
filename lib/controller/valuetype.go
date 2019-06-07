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

func (this *Controller) ReadValueType(id string, jwt jwt_http_router.Jwt) (result model.ValueType, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	deviceType, exists, err := this.db.GetValueType(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return deviceType, nil, http.StatusOK
}

func (this *Controller) ListValueTypes(jwt jwt_http_router.Jwt, options listoptions.ListOptions) (result []model.ValueType, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err = this.db.ListValueTypes(ctx, options)
	opterr := options.EvalStrict()
	if opterr != nil {
		return result, opterr, http.StatusBadRequest
	}
	return
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetValueType(valueType model.ValueType, owner string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	transaction, finish, err := this.db.Transaction(ctx)
	if err != nil {
		return err
	}
	for _, field := range valueType.Fields {
		err := this.source.PublishValueType(field.Type, owner)
		if err != nil {
			_ = finish(false)
			return err
		}
	}

	//cascade valuetype changes to devicetypes
	deviceTypes, err := this.db.ListDeviceTypesUsingValueType(transaction, valueType.Id)
	if err != nil {
		_ = finish(false)
		return err
	}
	for _, dt := range deviceTypes {
		for serviceIndex, service := range dt.Services {
			for assignmentIndex, assignment := range service.Input {
				if assignment.Type.Id == valueType.Id {
					dt.Services[serviceIndex].Input[assignmentIndex].Type = valueType
				}
			}
			for assignmentIndex, assignment := range service.Output {
				if assignment.Type.Id == valueType.Id {
					dt.Services[serviceIndex].Output[assignmentIndex].Type = valueType
				}
			}
		}
		err = this.db.SetDeviceType(transaction, dt)
		if err != nil {
			_ = finish(false)
			return err
		}
	}

	//cascade valuetype changes to other valueTypes
	valuetypes, err := this.db.ListValueTypesUsingValueType(transaction, valueType.Id)
	if err != nil {
		_ = finish(false)
		return err
	}
	for _, vt := range valuetypes {
		for fieldIndex, field := range vt.Fields {
			if field.Type.Id == valueType.Id {
				vt.Fields[fieldIndex].Type = valueType
			}
		}
		err = this.db.SetValueType(transaction, vt)
		if err != nil {
			_ = finish(false)
			return err
		}
	}

	err = this.db.SetValueType(transaction, valueType)
	if err != nil {
		_ = finish(false)
		return err
	}
	return finish(true)
}

func (this *Controller) DeleteValueType(id string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.RemoveValueType(ctx, id)
}

func (this *Controller) publishMissingValueTypesOfDeviceType(ctx context.Context, deviceType model.DeviceType, owner string) error {
	for _, service := range deviceType.Services {
		for _, assignment := range service.Input {
			err := this.publishMissingValueType(ctx, assignment.Type, owner)
			if err != nil {
				return err
			}
		}
		for _, assignment := range service.Output {
			err := this.publishMissingValueType(ctx, assignment.Type, owner)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *Controller) publishMissingValueType(ctx context.Context, valueType model.ValueType, owner string) error {
	_, exists, err := this.db.GetValueType(ctx, valueType.Id)
	if err != nil {
		return err
	}
	if !exists {
		return this.source.PublishValueType(valueType, owner)
	}
	return nil
}
