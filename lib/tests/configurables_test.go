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
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestConfigurables(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	interaction := model.EVENT_AND_REQUEST

	t.Run("init metadata", createTestConfigurableMetadata(conf))

	getTemperaturesAverageTimeConfigurables := []model.Configurable{
		{
			Path:             "duration.sec",
			CharacteristicId: "",
			AspectNode:       model.AspectNode{},
			FunctionId:       "",
			Value:            30.0,
			Type:             model.Integer,
		},
		{
			Path:             "duration.ms",
			CharacteristicId: "ms",
			AspectNode:       model.AspectNode{},
			FunctionId:       "",
			Value:            32.0,
			Type:             model.Integer,
		},
	}

	t.Run("measuring temperature with config input", testDeviceTypeSelectables(conf, []model.FilterCriteria{
		{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "device"},
	}, "", nil, []model.DeviceTypeSelectable{
		{
			DeviceTypeId: "pc_cooling_controller",
			Services: []model.Service{
				{
					Id:          "getTemperatures",
					Interaction: interaction,
					Inputs: []model.Content{
						{
							ContentVariable: model.ContentVariable{
								Id:   "duration",
								Name: "duration",
								Type: model.Structure,
								SubContentVariables: []model.ContentVariable{
									{
										Id:               "sec",
										Name:             "sec",
										Type:             model.Integer,
										CharacteristicId: "",
										Value:            30.0,
									},
									{
										Id:               "ms",
										Name:             "ms",
										Type:             model.Integer,
										CharacteristicId: "ms",
										Value:            32.0,
									},
								},
							},
						},
					},
					Outputs: []model.Content{
						{
							ContentVariable: model.ContentVariable{
								Id:   "avg_temperatures",
								Name: "avg_temperatures",
								SubContentVariables: []model.ContentVariable{
									{
										Id:         "cpu",
										Name:       "cpu",
										FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
										AspectId:   "cpu",
									},
									{
										Id:         "gpu",
										Name:       "gpu",
										FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
										AspectId:   "gpu",
									},
									{
										Id:         "case",
										Name:       "case",
										FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
										AspectId:   "case",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.ServicePathOption{
				"getTemperatures": {
					{
						ServiceId:        "getTemperatures",
						Path:             "avg_temperatures.case",
						CharacteristicId: "",
						AspectNode: model.AspectNode{
							Id:            "case",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId:    model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						Configurables: getTemperaturesAverageTimeConfigurables,
					},
					{
						ServiceId:        "getTemperatures",
						Path:             "avg_temperatures.cpu",
						CharacteristicId: "",
						AspectNode: model.AspectNode{
							Id:            "cpu",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId:    model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						Configurables: getTemperaturesAverageTimeConfigurables,
					},
					{
						ServiceId:        "getTemperatures",
						Path:             "avg_temperatures.gpu",
						CharacteristicId: "",
						AspectNode: model.AspectNode{
							Id:            "gpu",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId:    model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						Configurables: getTemperaturesAverageTimeConfigurables,
					},
				},
			},
		},
	}))

	t.Run("measuring cpu temperature with config input", testDeviceTypeSelectables(conf, []model.FilterCriteria{
		{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "cpu"},
	}, "", nil, []model.DeviceTypeSelectable{
		{
			DeviceTypeId: "pc_cooling_controller",
			Services: []model.Service{
				{
					Id:          "getTemperatures",
					Interaction: interaction,
					Inputs: []model.Content{
						{
							ContentVariable: model.ContentVariable{
								Id:   "duration",
								Name: "duration",
								Type: model.Structure,
								SubContentVariables: []model.ContentVariable{
									{
										Id:               "sec",
										Name:             "sec",
										Type:             model.Integer,
										CharacteristicId: "",
										Value:            30.0,
									},
									{
										Id:               "ms",
										Name:             "ms",
										Type:             model.Integer,
										CharacteristicId: "ms",
										Value:            32.0,
									},
								},
							},
						},
					},
					Outputs: []model.Content{
						{
							ContentVariable: model.ContentVariable{
								Id:   "avg_temperatures",
								Name: "avg_temperatures",
								SubContentVariables: []model.ContentVariable{
									{
										Id:         "cpu",
										Name:       "cpu",
										FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
										AspectId:   "cpu",
									},
									{
										Id:         "gpu",
										Name:       "gpu",
										FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
										AspectId:   "gpu",
									},
									{
										Id:         "case",
										Name:       "case",
										FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
										AspectId:   "case",
									},
								},
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.ServicePathOption{
				"getTemperatures": {
					{
						ServiceId:        "getTemperatures",
						Path:             "avg_temperatures.cpu",
						CharacteristicId: "",
						AspectNode: model.AspectNode{
							Id:            "cpu",
							RootId:        "device",
							ParentId:      "device",
							ChildIds:      []string{},
							AncestorIds:   []string{"device"},
							DescendentIds: []string{},
						},
						FunctionId:    model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						Configurables: getTemperaturesAverageTimeConfigurables,
					},
				},
			},
		},
	}))
}

func testDeviceTypeSelectables(config config.Config, criteria []model.FilterCriteria, pathPrefix string, interactionsFilter []model.Interaction, expectedResult []model.DeviceTypeSelectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := GetDeviceTypeSelectables(config, userjwt, pathPrefix, interactionsFilter, criteria)
		if err != nil {
			t.Error(err)
			return
		}
		expectedResult = sortServices(expectedResult)
		result = sortServices(result)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func createTestConfigurableMetadata(config config.Config) func(t *testing.T) {
	return func(t *testing.T) {
		interaction := model.EVENT_AND_REQUEST
		aspects := []model.Aspect{
			{
				Id: "air",
				SubAspects: []model.Aspect{
					{Id: "inside_air"},
					{Id: "outside_air",
						SubAspects: []model.Aspect{
							{Id: "morning_outside_air"},
							{Id: "evening_outside_air"},
						},
					},
				},
			},
			{
				Id: "water",
			},
			{
				Id: "device",
				SubAspects: []model.Aspect{
					{Id: "cpu"},
					{Id: "gpu"},
					{Id: "case"},
				},
			},
			{
				Id: "fan",
				SubAspects: []model.Aspect{
					{Id: "cpu_fan"},
					{Id: "gpu_fan"},
					{Id: "case_fan",
						SubAspects: []model.Aspect{
							{Id: "case_fan_1"},
							{Id: "case_fan_2"},
							{Id: "case_fan_3"},
							{Id: "case_fan_4"},
							{Id: "case_fan_5"},
						},
					},
				},
			},
		}
		functions := []model.Function{
			{Id: model.MEASURING_FUNCTION_PREFIX + "getTemperature"},
			{Id: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature"},
			{Id: model.MEASURING_FUNCTION_PREFIX + "getVolume"},
			{Id: model.CONTROLLING_FUNCTION_PREFIX + "setVolume"},
			{Id: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed"},
			{Id: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed"},
			{Id: model.CONTROLLING_FUNCTION_PREFIX + "toggle"},
			{Id: model.CONTROLLING_FUNCTION_PREFIX + "setMeasuringTime"},
		}
		devicetypes := []model.DeviceType{
			{
				Id:            "pc_cooling_controller",
				DeviceClassId: "pc_cooling_controller",
				Services: []model.Service{
					{
						Id:          "getTemperatures",
						Interaction: interaction,
						Inputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:   "duration",
									Name: "duration",
									Type: model.Structure,
									SubContentVariables: []model.ContentVariable{
										{
											Id:               "sec",
											Name:             "sec",
											Type:             model.Integer,
											CharacteristicId: "",
											Value:            30,
										},
										{
											Id:               "ms",
											Name:             "ms",
											Type:             model.Integer,
											CharacteristicId: "ms",
											Value:            32,
										},
									},
								},
							},
						},
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:   "avg_temperatures",
									Name: "avg_temperatures",
									SubContentVariables: []model.ContentVariable{
										{
											Id:         "cpu",
											Name:       "cpu",
											FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
											AspectId:   "cpu",
										},
										{
											Id:         "gpu",
											Name:       "gpu",
											FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
											AspectId:   "gpu",
										},
										{
											Id:         "case",
											Name:       "case",
											FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
											AspectId:   "case",
										},
									},
								},
							},
						},
					},

					{
						Id:          "getCaseFan1Speed",
						Interaction: interaction,
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
									AspectId:   "case_fan_1",
								},
							},
						},
					},
					{
						Id:          "getCaseFan2Speed",
						Interaction: interaction,
						Inputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:               "sec",
									Name:             "sec",
									Type:             model.Integer,
									CharacteristicId: "",
									Value:            24,
								},
							},
						},
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
									AspectId:   "case_fan_2",
								},
							},
						},
					},
					{
						Id:          "getCpuSpeed",
						Interaction: interaction,
						Inputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:               "sec",
									Name:             "sec",
									Type:             model.Integer,
									CharacteristicId: "sec",
									Value:            24,
								},
							},
						},
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
									AspectId:   "cpu_fan",
								},
							},
						},
					},
					{
						Id:          "getGpuSpeed",
						Interaction: interaction,
						Inputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:               "sec",
									Name:             "sec",
									Type:             model.Integer,
									CharacteristicId: "sec",
									FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "setMeasuringTime",
									Value:            24,
								},
							},
						},
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
									AspectId:   "gpu_fan",
								},
							},
						},
					},

					{
						Id:          "setCaseFanSpeed",
						Interaction: interaction,
						Inputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:   "speed",
									Name: "speed",
									SubContentVariables: []model.ContentVariable{
										{
											Id:         "speed",
											Name:       "speed",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
											AspectId:   "case_fan_1",
											Value:      13,
										},
										{
											Id:         "speed",
											Name:       "speed",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
											AspectId:   "case_fan_2",
											Value:      14,
										},
									},
								},
							},
						},
					},
					{
						Id:          "setCaseFan1Speed",
						Interaction: interaction,
						Inputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "case_fan_1",
								},
							},
						},
					},
					{
						Id:          "setCaseFan2Speed",
						Interaction: interaction,
						Inputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "case_fan_2",
								},
							},
						},
					},
				},
			},
		}

		producer, err := testutils.NewPublisher(config)
		if err != nil {
			t.Error(err)
			return
		}

		for _, aspect := range aspects {
			err = producer.PublishAspect(aspect, userid)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, function := range functions {
			err = producer.PublishFunction(function, userid)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, dt := range devicetypes {
			err = producer.PublishDeviceType(dt, userid)
			if err != nil {
				t.Error(err)
				return
			}
		}

		time.Sleep(5 * time.Second)
	}
}
