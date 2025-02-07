/*
 * Copyright 2025 InfAI (CC SES)
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

package repo_legacy

import (
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"testing"
)

func testCharacteristicList(t *testing.T, conf config.Config) {
	characteristics := []models.Characteristic{
		{
			Id:   "a1",
			Name: "a1",
		},
		{
			Id:   "a2",
			Name: "a2",
			SubCharacteristics: []models.Characteristic{
				{
					Id:   "a2.1",
					Name: "a2.1",
				},
				{
					Id:   "a2.2",
					Name: "a2.2",
					SubCharacteristics: []models.Characteristic{
						{
							Id:   "a2.2.1",
							Name: "a2.2.1",
						},
						{
							Id:   "a2.2.2",
							Name: "a2.2.2",
						},
					},
				},
			},
		},
		{
			Id:   "b3",
			Name: "b3",
		},
		{
			Id:   "c4",
			Name: "c4",
		},
	}

	t.Run("create characteristics", func(t *testing.T) {
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		for _, characteristic := range characteristics {
			_, err, _ := c.SetCharacteristic(AdminToken, characteristic)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
	t.Run("list all characteristics", func(t *testing.T) {
		list, total, err, _ := c.ListCharacteristics(client.CharacteristicListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 4 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, characteristics) {
			t.Error(list)
			return
		}
	})
	t.Run("list b characteristics", func(t *testing.T) {
		list, total, err, _ := c.ListCharacteristics(client.CharacteristicListOptions{Search: "b"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 1 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.Characteristic{characteristics[2]}) {
			t.Error(list)
			return
		}
	})
	t.Run("list a1,b3 characteristics", func(t *testing.T) {
		list, total, err, _ := c.ListCharacteristics(client.CharacteristicListOptions{Ids: []string{"a1", "b3"}})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 2 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.Characteristic{characteristics[0], characteristics[2]}) {
			t.Error(list)
			return
		}
	})
}

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
