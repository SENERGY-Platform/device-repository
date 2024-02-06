/*
 * Copyright 2024 InfAI (CC SES)
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
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"sync"
	"testing"
)

func TestQueryDeviceTypeReferences(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	aspects, functions, deviceTypes := getTestConfigurableMetadata()
	deviceTypes = append(deviceTypes, models.DeviceType{
		Id:   "test2",
		Name: "test2",
		Services: []models.Service{
			{
				Id:          "ts1",
				Name:        "ts1",
				Interaction: models.EVENT_AND_REQUEST,
				Inputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id:       "tc1",
							Name:     "tc1",
							Type:     models.String,
							AspectId: "air",
						},
					},
				},
				Outputs: []models.Content{},
			},
			{
				Id:          "ts2",
				Name:        "ts2",
				Interaction: models.EVENT_AND_REQUEST,
				Outputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id:         "tc2",
							Name:       "tc2",
							Type:       models.String,
							AspectId:   "cpu",
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
				},
				Inputs: []models.Content{},
			},
		},
	})

	t.Run("init metadata", createTestConfigurableMetadataBase(conf, aspects, functions, deviceTypes))

	clientUrl := "http://localhost:" + conf.ServerPort

	c := client.NewClient(clientUrl)

	query := model.UsedInDeviceTypeQuery{
		Resource: "aspects",
		Ids: []string{
			"unknown",
			"air",
			"inside_air",
			"outside_air",
			"morning_outside_air",
			"evening_outside_air",
			"water",
			"device",
			"cpu",
			"gpu",
			"case",
			"fan",
			"cpu_fan",
			"gpu_fan",
			"case_fan",
			"case_fan_1",
			"case_fan_2",
			"case_fan_3",
			"case_fan_4",
			"case_fan_5",
		},
	}

	t.Run("query default", func(t *testing.T) {
		result, err, _ := c.GetUsedInDeviceType(query)
		if err != nil {
			t.Error(err)
			return
		}

		expected := map[string]model.RefInDeviceTypeResponseElement{
			"air": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "test2",
						Name: "test2",
					},
				},
			},
			"case": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"case_fan": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_1": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"case_fan_2": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"case_fan_3": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_4": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_5": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"cpu": {
				Count: 2,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
					{
						Id:   "test2",
						Name: "test2",
					},
				},
			},
			"cpu_fan": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"device": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"evening_outside_air": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"fan": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"gpu": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"gpu_fan": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"inside_air": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"morning_outside_air": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"outside_air": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"unknown": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"water": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("\n%#v\n%#v\n", expected, result)
		}
	})

	t.Run("query count device-type", func(t *testing.T) {
		query.CountBy = "device-type"
		result, err, _ := c.GetUsedInDeviceType(query)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(result, map[string]model.RefInDeviceTypeResponseElement{
			"air": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "test2",
						Name: "test2",
					},
				},
			},
			"case": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"case_fan": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_1": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"case_fan_2": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"case_fan_3": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_4": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_5": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"cpu": {
				Count: 2,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
					{
						Id:   "test2",
						Name: "test2",
					},
				},
			},
			"cpu_fan": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"device": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"evening_outside_air": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"fan": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"gpu": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"gpu_fan": {
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"inside_air": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"morning_outside_air": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"outside_air": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"unknown": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"water": {
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
		}) {
			t.Errorf("%#v", result)
		}
	})

	t.Run("query count service", func(t *testing.T) {
		query.CountBy = "service"
		result, err, _ := c.GetUsedInDeviceType(query)
		if err != nil {
			t.Error(err)
			return
		}
		expected := map[string]model.RefInDeviceTypeResponseElement{
			"air": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "test2",
						Name: "test2",
					},
				},
			},
			"case": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"case_fan": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_1": model.RefInDeviceTypeResponseElement{
				Count: 3,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"case_fan_2": model.RefInDeviceTypeResponseElement{
				Count: 3,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"case_fan_3": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_4": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_5": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"cpu": model.RefInDeviceTypeResponseElement{
				Count: 2,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					}, model.DeviceTypeReference{
						Id:   "test2",
						Name: "test2",
					},
				},
			},
			"cpu_fan": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"device": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"evening_outside_air": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"fan": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"gpu": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"gpu_fan": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
					},
				},
			},
			"inside_air": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"morning_outside_air": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"outside_air": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"unknown": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"water": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("\n%#v\n%#v\n", expected, result)
		}
	})

	t.Run("query with service", func(t *testing.T) {
		query.CountBy = "service"
		query.With = "service"
		result, err, _ := c.GetUsedInDeviceType(query)
		if err != nil {
			t.Error(err)
			return
		}

		expected := map[string]model.RefInDeviceTypeResponseElement{
			"air": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "test2",
						Name: "test2",
						UsedIn: []model.ServiceReference{
							{
								Id:   "ts1",
								Name: "ts1",
							},
						},
					},
				},
			},
			"case": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
						UsedIn: []model.ServiceReference{
							{
								Id:   "getTemperatures",
								Name: "",
							},
						},
					},
				},
			},
			"case_fan": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_1": model.RefInDeviceTypeResponseElement{
				Count: 3,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
						UsedIn: []model.ServiceReference{
							{
								Id:   "getCaseFan1Speed",
								Name: "",
							},
							{
								Id:   "setCaseFanSpeed",
								Name: "",
							},
							{
								Id:   "setCaseFan1Speed",
								Name: "",
							},
						},
					},
				},
			},
			"case_fan_2": model.RefInDeviceTypeResponseElement{
				Count: 3,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
						UsedIn: []model.ServiceReference{
							{
								Id:   "getCaseFan2Speed",
								Name: "",
							},
							{
								Id:   "setCaseFanSpeed",
								Name: "",
							},
							{
								Id:   "setCaseFan2Speed",
								Name: "",
							},
						},
					},
				},
			},
			"case_fan_3": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_4": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"case_fan_5": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"cpu": model.RefInDeviceTypeResponseElement{
				Count: 2,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
						UsedIn: []model.ServiceReference{
							{
								Id:   "getTemperatures",
								Name: "",
							},
						},
					}, model.DeviceTypeReference{
						Id:   "test2",
						Name: "test2",
						UsedIn: []model.ServiceReference{
							{
								Id:   "ts2",
								Name: "ts2",
							},
						},
					},
				},
			},
			"cpu_fan": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
						UsedIn: []model.ServiceReference{
							{
								Id:   "getCpuSpeed",
								Name: "",
							},
						},
					},
				},
			},
			"device": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"evening_outside_air": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"fan": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"gpu": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
						UsedIn: []model.ServiceReference{
							{
								Id:   "getTemperatures",
								Name: "",
							},
						},
					},
				},
			},
			"gpu_fan": model.RefInDeviceTypeResponseElement{
				Count: 1,
				UsedIn: []model.DeviceTypeReference{
					{
						Id:   "pc_cooling_controller",
						Name: "pc_cooling_controller_name",
						UsedIn: []model.ServiceReference{
							{
								Id:   "getGpuSpeed",
								Name: "",
							},
						},
					},
				},
			},
			"inside_air": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"morning_outside_air": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"outside_air": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"unknown": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
			"water": model.RefInDeviceTypeResponseElement{
				Count:  0,
				UsedIn: []model.DeviceTypeReference{},
			},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("\n%#v\n%#v\n", expected, result)
		}
	})

}
