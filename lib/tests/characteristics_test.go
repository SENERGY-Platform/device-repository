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

package tests

import (
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/models/go/models"
	"testing"
)

func TestCharacteristicValidation(t *testing.T) {
	t.Run("struct duplicate sub characteristic name", testValidateCharacteristic(true, models.Characteristic{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubCharacteristics: []models.Characteristic{
			{
				Id:   "v",
				Name: "v",
				Type: models.String,
			},
			{
				Id:   "v2",
				Name: "v",
				Type: models.Integer,
			},
		},
	}))

	t.Run("list duplicate sub characteristic name", testValidateCharacteristic(true, models.Characteristic{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubCharacteristics: []models.Characteristic{
			{
				Id:   "v",
				Name: "0",
				Type: models.String,
			},
			{
				Id:   "v2",
				Name: "0",
				Type: models.String,
			},
		},
	}))
}

func testValidateCharacteristic(expectError bool, characteristic models.Characteristic) func(t *testing.T) {
	return func(t *testing.T) {
		err, _ := controller.ValidateCharacteristicsWithoutDbAccess(characteristic)
		if (err != nil) != expectError {
			t.Error(expectError, err)
		}
	}
}
