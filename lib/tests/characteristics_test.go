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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"log"
	"testing"
)

func TestCharacteristicValidation(t *testing.T) {
	ctrl, err := controller.New(config.Config{}, nil, nil, nil)
	if err != nil {
		log.Println("ERROR: unable to start control", err)
		t.Error(err)
		return
	}

	t.Run("struct duplicate sub characteristic name", testValidateCharacteristic(ctrl, true, model.Characteristic{
		Id:   "root",
		Name: "root",
		Type: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "v",
				Name: "v",
				Type: model.String,
			},
			{
				Id:   "v2",
				Name: "v",
				Type: model.Integer,
			},
		},
	}))

	t.Run("list duplicate sub characteristic name", testValidateCharacteristic(ctrl, true, model.Characteristic{
		Id:   "root",
		Name: "root",
		Type: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "v",
				Name: "0",
				Type: model.String,
			},
			{
				Id:   "v2",
				Name: "0",
				Type: model.String,
			},
		},
	}))
}

func testValidateCharacteristic(ctrl *controller.Controller, expectError bool, characteristic model.Characteristic) func(t *testing.T) {
	return func(t *testing.T) {
		err, _ := ctrl.ValidateCharacteristics(characteristic)
		if (err != nil) != expectError {
			t.Error(expectError, err)
		}
	}
}
