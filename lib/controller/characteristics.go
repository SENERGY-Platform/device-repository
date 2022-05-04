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
)

func (this *Controller) SetCharacteristic(characteristic model.Characteristic, owner string) error {
	ctx, _ := getTimeoutContext()
	return this.db.SetCharacteristic(ctx, characteristic)
}

func (this *Controller) DeleteCharacteristic(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveCharacteristic(ctx, id)
}

func (this *Controller) GetLeafCharacteristics() (result []model.Characteristic, err error, code int) {
	ctx, _ := getTimeoutContext()
	temp, err := this.db.ListAllCharacteristics(ctx)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	result = getLeafCharacteristics(temp)
	return result, nil, http.StatusOK
}

func getLeafCharacteristics(list []model.Characteristic) (result []model.Characteristic) {
	for _, element := range list {
		if len(element.SubCharacteristics) == 0 {
			result = append(result, element)
		} else {
			result = append(result, getLeafCharacteristics(element.SubCharacteristics)...)
		}
	}
	return result
}

func (this *Controller) GetCharacteristic(id string) (result model.Characteristic, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetCharacteristic(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ValidateCharacteristics(characteristic model.Characteristic) (err error, code int) {
	if characteristic.Id == "" {
		return errors.New("missing characteristic id"), http.StatusBadRequest
	}
	if characteristic.Name == "" {
		return errors.New("missing characteristic name"), http.StatusBadRequest
	}

	if characteristic.Type != model.String &&
		characteristic.Type != model.Integer &&
		characteristic.Type != model.Float &&
		characteristic.Type != model.Boolean &&
		characteristic.Type != model.List &&
		characteristic.Type != model.Structure {
		return errors.New("wrong characteristic type"), http.StatusBadRequest
	}

	err, code = this.validateSubCharacteristics(characteristic.SubCharacteristics)
	if err != nil {
		return err, code
	}

	return nil, http.StatusOK
}

func (this *Controller) validateSubCharacteristics(characteristics []model.Characteristic) (error, int) {
	knownName := map[string]bool{}
	for _, characteristic := range characteristics {
		if knownName[characteristic.Name] {
			return errors.New("duplicate sub characteristic name: " + characteristic.Name), http.StatusBadRequest
		}
		knownName[characteristic.Name] = true
		err, code := this.ValidateCharacteristics(characteristic)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) ValidateCharacteristicDelete(id string) (err error, code int) {
	ctx, _ := getTimeoutContext()
	isUsed, err := this.db.CharacteristicIsUsed(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if isUsed {
		return errors.New("still in use"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}
