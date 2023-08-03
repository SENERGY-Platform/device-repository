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
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"strings"
)

func (this *Controller) SetCharacteristic(characteristic models.Characteristic, owner string) error {
	ctx, _ := getTimeoutContext()
	return this.db.SetCharacteristic(ctx, characteristic)
}

func (this *Controller) DeleteCharacteristic(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveCharacteristic(ctx, id)
}

func (this *Controller) GetLeafCharacteristics() (result []models.Characteristic, err error, code int) {
	ctx, _ := getTimeoutContext()
	temp, err := this.db.ListAllCharacteristics(ctx)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	result = getLeafCharacteristics(temp)
	return result, nil, http.StatusOK
}

func getLeafCharacteristics(list []models.Characteristic) (result []models.Characteristic) {
	for _, element := range list {
		if len(element.SubCharacteristics) == 0 {
			result = append(result, element)
		} else {
			result = append(result, getLeafCharacteristics(element.SubCharacteristics)...)
		}
	}
	return result
}

func (this *Controller) GetCharacteristic(id string) (result models.Characteristic, err error, errCode int) {
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

func characteristicIdToBaseCharacteristic(characteristic models.Characteristic) (result map[string]models.Characteristic) {
	result = map[string]models.Characteristic{characteristic.Id: characteristic}
	for _, sub := range characteristic.SubCharacteristics {
		for key, _ := range characteristicIdToBaseCharacteristic(sub) {
			result[key] = characteristic
		}
	}
	return result
}

func (this *Controller) ValidateCharacteristics(characteristic models.Characteristic) (err error, code int) {
	ctx, _ := getTimeoutContext()
	knownCharacteristics, err := this.db.ListAllCharacteristics(ctx)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	err, code = validateCharacteristicIdReuse(characteristic, knownCharacteristics)
	if err != nil {
		return err, code
	}
	return ValidateCharacteristicsWithoutDbAccess(characteristic)
}

func validateCharacteristicIdReuse(characteristic models.Characteristic, knownCharacteristics []models.Characteristic) (err error, code int) {
	usedCharacteristicIds := characteristicIdToBaseCharacteristic(characteristic)
	for _, e := range knownCharacteristics {
		existingSubCharacteristics := characteristicIdToBaseCharacteristic(e)
		for usedSubId, _ := range usedCharacteristicIds {
			if existingBaseCharacteristic, found := existingSubCharacteristics[usedSubId]; found && existingBaseCharacteristic.Id != characteristic.Id {
				return errors.New("characteristic references existing sub characteristic"), http.StatusBadRequest
			}
		}
	}
	return nil, http.StatusOK
}

func ValidateCharacteristicsWithoutDbAccess(characteristic models.Characteristic) (err error, code int) {
	if characteristic.Id == "" {
		return errors.New("missing characteristic id"), http.StatusBadRequest
	}
	if characteristic.Name == "" {
		return errors.New("missing characteristic name"), http.StatusBadRequest
	}

	if characteristic.Type != models.String &&
		characteristic.Type != models.Integer &&
		characteristic.Type != models.Float &&
		characteristic.Type != models.Boolean &&
		characteristic.Type != models.List &&
		characteristic.Type != models.Structure {
		return errors.New("wrong characteristic type"), http.StatusBadRequest
	}

	err, code = validateSubCharacteristics(characteristic.SubCharacteristics)
	if err != nil {
		return err, code
	}

	return nil, http.StatusOK
}

func validateSubCharacteristics(characteristics []models.Characteristic) (error, int) {
	knownName := map[string]bool{}
	for _, characteristic := range characteristics {
		if knownName[characteristic.Name] {
			return errors.New("duplicate sub characteristic name: " + characteristic.Name), http.StatusBadRequest
		}
		knownName[characteristic.Name] = true
		err, code := ValidateCharacteristicsWithoutDbAccess(characteristic)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) ValidateCharacteristicDelete(id string) (err error, code int) {
	ctx, _ := getTimeoutContext()
	isUsed, where, err := this.db.CharacteristicIsUsed(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if isUsed {
		return errors.New("still in use: " + strings.Join(where, ",")), http.StatusBadRequest
	}
	return nil, http.StatusOK
}
