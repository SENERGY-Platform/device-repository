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
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"sync"
	"testing"
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
	conf.InitTopics = true

	interaction := models.EVENT_AND_REQUEST

	t.Run("init metadata", createTestConfigurableMetadata(conf))

	getTemperaturesAverageTimeConfigurables := []model.Configurable{
		{
			Path:             "duration.sec",
			CharacteristicId: "",
			AspectNode:       models.AspectNode{},
			FunctionId:       "",
			Value:            30.0,
			Type:             models.Integer,
		},
		{
			Path:             "duration.ms",
			CharacteristicId: "ms",
			AspectNode:       models.AspectNode{},
			FunctionId:       "",
			Value:            32.0,
			Type:             models.Integer,
		},
	}

	t.Run("measuring temperature with config input", testDeviceTypeSelectables(conf, []model.FilterCriteria{
		{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "device"},
	}, "", nil, []model.DeviceTypeSelectable{
		{
			DeviceTypeId: "pc_cooling_controller",
			Services: []models.Service{
				{
					Id:          "getTemperatures",
					Interaction: interaction,
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:   "duration",
								Name: "duration",
								Type: models.Structure,
								SubContentVariables: []models.ContentVariable{
									{
										Id:               "sec",
										Name:             "sec",
										Type:             models.Integer,
										CharacteristicId: "",
										Value:            30.0,
									},
									{
										Id:               "ms",
										Name:             "ms",
										Type:             models.Integer,
										CharacteristicId: "ms",
										Value:            32.0,
									},
								},
							},
						},
					},
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:   "avg_temperatures",
								Name: "avg_temperatures",
								SubContentVariables: []models.ContentVariable{
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
						AspectNode: models.AspectNode{
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
						AspectNode: models.AspectNode{
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
						AspectNode: models.AspectNode{
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
			Services: []models.Service{
				{
					Id:          "getTemperatures",
					Interaction: interaction,
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:   "duration",
								Name: "duration",
								Type: models.Structure,
								SubContentVariables: []models.ContentVariable{
									{
										Id:               "sec",
										Name:             "sec",
										Type:             models.Integer,
										CharacteristicId: "",
										Value:            30.0,
									},
									{
										Id:               "ms",
										Name:             "ms",
										Type:             models.Integer,
										CharacteristicId: "ms",
										Value:            32.0,
									},
								},
							},
						},
					},
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:   "avg_temperatures",
								Name: "avg_temperatures",
								SubContentVariables: []models.ContentVariable{
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
						AspectNode: models.AspectNode{
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

	t.Run("measuring fan speed", testDeviceTypeSelectables(conf, []model.FilterCriteria{
		{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed", AspectId: "fan"},
	}, "", nil, []model.DeviceTypeSelectable{
		{
			DeviceTypeId: "pc_cooling_controller",
			Services: []models.Service{
				{
					Id:          "getCaseFan1Speed",
					Interaction: interaction,
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:               "sec",
								Name:             "sec",
								Type:             models.Integer,
								CharacteristicId: "",
								Value:            24.0,
							},
						},
					},
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:               "sec",
								Name:             "sec",
								Type:             models.Integer,
								CharacteristicId: "sec",
								Value:            24.0,
							},
						},
					},
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:               "sec",
								Name:             "sec",
								Type:             models.Integer,
								CharacteristicId: "sec",
								FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "setMeasuringTime",
								Value:            24.0,
							},
						},
					},
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:         "speed",
								Name:       "speed",
								FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
								AspectId:   "gpu_fan",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.ServicePathOption{
				"getCaseFan1Speed": {
					{
						ServiceId:        "getCaseFan1Speed",
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: models.AspectNode{
							Id:            "case_fan_1",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId:    model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
						Configurables: []model.Configurable{},
					},
				},
				"getCaseFan2Speed": {
					{
						ServiceId:        "getCaseFan2Speed",
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: models.AspectNode{
							Id:            "case_fan_2",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
						Configurables: []model.Configurable{
							{
								Path:             "sec",
								CharacteristicId: "",
								AspectNode:       models.AspectNode{},
								FunctionId:       "",
								Value:            24.0,
								Type:             models.Integer,
							},
						},
					},
				},
				"getCpuSpeed": {
					{
						ServiceId:        "getCpuSpeed",
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: models.AspectNode{
							Id:            "cpu_fan",
							RootId:        "fan",
							ParentId:      "fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"fan"},
							DescendentIds: []string{},
						},
						FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
						Configurables: []model.Configurable{
							{
								Path:             "sec",
								CharacteristicId: "sec",
								AspectNode:       models.AspectNode{},
								FunctionId:       "",
								Value:            24.0,
								Type:             models.Integer,
							},
						},
					},
				},
				"getGpuSpeed": {
					{
						ServiceId:        "getGpuSpeed",
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: models.AspectNode{
							Id:            "gpu_fan",
							RootId:        "fan",
							ParentId:      "fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"fan"},
							DescendentIds: []string{},
						},
						FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
						Configurables: []model.Configurable{
							{
								Path:             "sec",
								CharacteristicId: "sec",
								AspectNode:       models.AspectNode{},
								FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "setMeasuringTime",
								Value:            24.0,
								Type:             models.Integer,
							},
						},
					},
				},
			},
		},
	}))

	t.Run("set fan speed", testDeviceTypeSelectables(conf, []model.FilterCriteria{
		{FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed", AspectId: "fan"},
	}, "", nil, []model.DeviceTypeSelectable{
		{
			DeviceTypeId: "pc_cooling_controller",
			Services: []models.Service{
				{
					Id:          "setCaseFanSpeed",
					Interaction: interaction,
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:   "speed",
								Name: "speed",
								SubContentVariables: []models.ContentVariable{
									{
										Id:         "1",
										Name:       "1",
										FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
										AspectId:   "case_fan_1",
										Value:      13.0,
									},
									{
										Id:         "2",
										Name:       "2",
										FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
										AspectId:   "case_fan_2",
										Value:      14.0,
									},
								},
							},
						},
					},
				},
				{
					Id:          "setCaseFan1Speed",
					Interaction: interaction,
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:    "header",
								Name:  "header",
								Type:  models.String,
								Value: "auth",
							},
						},
						{
							ContentVariable: models.ContentVariable{
								Id:               "speed",
								Name:             "speed",
								CharacteristicId: "foo",
								FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
								AspectId:         "case_fan_2",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.ServicePathOption{
				"setCaseFanSpeed": {
					{
						ServiceId:        "setCaseFanSpeed",
						Path:             "speed.1",
						CharacteristicId: "",
						AspectNode: models.AspectNode{
							Id:            "case_fan_1",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
						Value:                 13.0,
						IsControllingFunction: true,
						Configurables: []model.Configurable{
							{
								Path:             "speed.2",
								CharacteristicId: "",
								AspectNode: models.AspectNode{
									Id:            "case_fan_2",
									RootId:        "fan",
									ParentId:      "case_fan",
									ChildIds:      []string{},
									AncestorIds:   []string{"case_fan", "fan"},
									DescendentIds: []string{},
								},
								FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
								Value:      14.0,
							},
						},
					},
					{
						ServiceId:        "setCaseFanSpeed",
						Path:             "speed.2",
						CharacteristicId: "",
						AspectNode: models.AspectNode{
							Id:            "case_fan_2",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
						Value:                 14.0,
						IsControllingFunction: true,
						Configurables: []model.Configurable{
							{
								Path:             "speed.1",
								CharacteristicId: "",
								AspectNode: models.AspectNode{
									Id:            "case_fan_1",
									RootId:        "fan",
									ParentId:      "case_fan",
									ChildIds:      []string{},
									AncestorIds:   []string{"case_fan", "fan"},
									DescendentIds: []string{},
								},
								FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
								Value:      13.0,
							},
						},
					},
				},
				"setCaseFan1Speed": {
					{
						ServiceId:        "setCaseFan1Speed",
						Path:             "speed",
						CharacteristicId: "",
						AspectNode: models.AspectNode{
							Id:            "case_fan_1",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
						IsControllingFunction: true,
						Configurables:         []model.Configurable{},
						Type:                  "",
					},
				},
				"setCaseFan2Speed": {
					{
						ServiceId:        "setCaseFan2Speed",
						Path:             "speed",
						CharacteristicId: "foo",
						AspectNode: models.AspectNode{
							Id:            "case_fan_2",
							RootId:        "fan",
							ParentId:      "case_fan",
							ChildIds:      []string{},
							AncestorIds:   []string{"case_fan", "fan"},
							DescendentIds: []string{},
						},
						FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
						IsControllingFunction: true,
						Configurables: []model.Configurable{
							{
								Path:             "header",
								CharacteristicId: "",
								AspectNode:       models.AspectNode{},
								FunctionId:       "",
								Value:            "auth",
								Type:             models.String,
							},
						},
						Type: "",
					},
				},
			},
		},
	}))
}

func testDeviceTypeSelectables(config configuration.Config, criteria []model.FilterCriteria, pathPrefix string, interactionsFilter []models.Interaction, expectedResult []model.DeviceTypeSelectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := GetDeviceTypeSelectables(config, userjwt, pathPrefix, interactionsFilter, criteria)
		if err != nil {
			t.Error(err)
			return
		}
		expectedResult = sortServices(expectedResult)
		result = sortServices(result)
		if !reflect.DeepEqual(normalize(result), normalize(expectedResult)) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func getTestConfigurableMetadata() (aspects []models.Aspect, functions []models.Function, devicetypes []models.DeviceType) {
	interaction := models.EVENT_AND_REQUEST
	aspects = []models.Aspect{
		{
			Id: "air",
			SubAspects: []models.Aspect{
				{Id: "inside_air"},
				{Id: "outside_air",
					SubAspects: []models.Aspect{
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
			SubAspects: []models.Aspect{
				{Id: "cpu"},
				{Id: "gpu"},
				{Id: "case"},
			},
		},
		{
			Id: "fan",
			SubAspects: []models.Aspect{
				{Id: "cpu_fan"},
				{Id: "gpu_fan"},
				{Id: "case_fan",
					SubAspects: []models.Aspect{
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
	functions = []models.Function{
		{Id: model.MEASURING_FUNCTION_PREFIX + "getTemperature"},
		{Id: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature"},
		{Id: model.MEASURING_FUNCTION_PREFIX + "getVolume"},
		{Id: model.CONTROLLING_FUNCTION_PREFIX + "setVolume"},
		{Id: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed"},
		{Id: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed"},
		{Id: model.CONTROLLING_FUNCTION_PREFIX + "toggle"},
		{Id: model.CONTROLLING_FUNCTION_PREFIX + "setMeasuringTime"},
	}
	devicetypes = []models.DeviceType{
		{
			Id:            "pc_cooling_controller",
			Name:          "pc_cooling_controller_name",
			DeviceClassId: "pc_cooling_controller",
			Services: []models.Service{
				{
					Id:          "getTemperatures",
					Interaction: interaction,
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:   "duration",
								Name: "duration",
								Type: models.Structure,
								SubContentVariables: []models.ContentVariable{
									{
										Id:               "sec",
										Name:             "sec",
										Type:             models.Integer,
										CharacteristicId: "",
										Value:            30,
									},
									{
										Id:               "ms",
										Name:             "ms",
										Type:             models.Integer,
										CharacteristicId: "ms",
										Value:            32,
									},
								},
							},
						},
					},
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:   "avg_temperatures",
								Name: "avg_temperatures",
								SubContentVariables: []models.ContentVariable{
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
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:               "sec",
								Name:             "sec",
								Type:             models.Integer,
								CharacteristicId: "",
								Value:            24,
							},
						},
					},
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:               "sec",
								Name:             "sec",
								Type:             models.Integer,
								CharacteristicId: "sec",
								Value:            24,
							},
						},
					},
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:               "sec",
								Name:             "sec",
								Type:             models.Integer,
								CharacteristicId: "sec",
								FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "setMeasuringTime",
								Value:            24,
							},
						},
					},
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:   "speed",
								Name: "speed",
								SubContentVariables: []models.ContentVariable{
									{
										Id:         "1",
										Name:       "1",
										FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
										AspectId:   "case_fan_1",
										Value:      13,
									},
									{
										Id:         "2",
										Name:       "2",
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:    "header",
								Name:  "header",
								Type:  models.String,
								Value: "auth",
							},
						},
						{
							ContentVariable: models.ContentVariable{
								Id:               "speed",
								Name:             "speed",
								FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
								AspectId:         "case_fan_2",
								CharacteristicId: "foo",
							},
						},
					},
				},
			},
		},
	}

	return
}

func createTestConfigurableMetadata(config configuration.Config) func(t *testing.T) {
	aspects, functions, deviceTypes := getTestConfigurableMetadata()
	return createTestConfigurableMetadataBase(config, aspects, functions, deviceTypes)
}

func createTestConfigurableMetadataBase(config configuration.Config, aspects []models.Aspect, functions []models.Function, devicetypes []models.DeviceType) func(t *testing.T) {
	return func(t *testing.T) {
		c := client.NewClient("http://localhost:"+config.ServerPort, nil)
		for _, aspect := range aspects {
			_, err, _ := c.SetAspect(AdminToken, aspect)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, function := range functions {
			_, err, _ := c.SetFunction(AdminToken, function)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, dt := range devicetypes {
			_, err, _ := c.SetDeviceType(AdminToken, dt, client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
		}

	}
}
