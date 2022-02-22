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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDeviceTypeSelectablesInteractionFilter(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("init metadata", createTestMetadata(conf, model.REQUEST))

	waterProbeCriteria := []model.FilterCriteria{{
		FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
		AspectId:   "water",
	}}
	waterprobeSelectable := model.DeviceTypeSelectable{
		DeviceTypeId: "water-probe",
		Services: []model.Service{
			{
				Id:          "getTemperature",
				Interaction: model.REQUEST,
				Outputs: []model.Content{
					{
						ContentVariable: model.ContentVariable{
							Id:               "temperature",
							Name:             "temperature",
							FunctionId:       model.MEASURING_FUNCTION_PREFIX + "getTemperature",
							AspectId:         "water",
							CharacteristicId: "water-probe-test-characteristic",
						},
					},
				},
			},
		},
		ServicePathOptions: map[string][]model.ServicePathOption{
			"getTemperature": {
				{
					ServiceId:        "getTemperature",
					Path:             "prefix.temperature",
					CharacteristicId: "water-probe-test-characteristic",
					AspectNodeId:     "water",
					FunctionId:       model.MEASURING_FUNCTION_PREFIX + "getTemperature",
				},
			},
		},
	}

	t.Run("nil", testDeviceTypeSelectables(conf, waterProbeCriteria, "prefix.", nil, []model.DeviceTypeSelectable{waterprobeSelectable}))
	t.Run("empty", testDeviceTypeSelectables(conf, waterProbeCriteria, "prefix.", []model.Interaction{}, []model.DeviceTypeSelectable{waterprobeSelectable}))
	t.Run("event", testDeviceTypeSelectables(conf, waterProbeCriteria, "prefix.", []model.Interaction{model.EVENT}, []model.DeviceTypeSelectable{}))
	t.Run("request", testDeviceTypeSelectables(conf, waterProbeCriteria, "prefix.", []model.Interaction{model.REQUEST}, []model.DeviceTypeSelectable{waterprobeSelectable}))
	t.Run("event+request", testDeviceTypeSelectables(conf, waterProbeCriteria, "prefix.", []model.Interaction{model.EVENT, model.REQUEST}, []model.DeviceTypeSelectable{waterprobeSelectable}))
	t.Run("event_and_request", testDeviceTypeSelectables(conf, waterProbeCriteria, "prefix.", []model.Interaction{model.EVENT_AND_REQUEST}, []model.DeviceTypeSelectable{}))
}

func TestDeviceTypeMeasuringSelectables(t *testing.T) {
	t.Skip("not implemented") //TODO
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("init metadata", createTestMetadata(conf, model.EVENT_AND_REQUEST))

	t.Run("inside and outside temp", testDeviceTypeSelectables(conf, []model.FilterCriteria{
		{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "inside_air"},
		{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "outside_air"},
	}, "", nil, []model.DeviceTypeSelectable{}))

}

func TestDeviceTypeControllingSelectables(t *testing.T) {
	t.Skip("not implemented") //TODO

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("init metadata", createTestMetadata(conf, model.EVENT_AND_REQUEST))
}

func testDeviceTypeSelectables(config config.Config, criteria []model.FilterCriteria, pathPrefix string, interactionsFilter []model.Interaction, expectedResult []model.DeviceTypeSelectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := GetDeviceTypeSelectables(config, userjwt, pathPrefix, interactionsFilter, criteria)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func createTestMetadata(config config.Config, interaction model.Interaction) func(t *testing.T) {
	return func(t *testing.T) {
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
					{Id: "case"},
					{Id: "cooler",
						SubAspects: []model.Aspect{
							{Id: "cooler_radiator"},
							{Id: "cooler_water_reservoir"},
						},
					},
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
		}
		devicetypes := []model.DeviceType{
			{
				Id:            "thermostat",
				DeviceClassId: "thermostat",
				Services: []model.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
					{
						Id:          "getTargetTemperature",
						Interaction: interaction,
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermometer",
				DeviceClassId: "thermometer",
				Services: []model.Service{
					{
						Id:          "getInsideTemperature",
						Interaction: interaction,
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
					{
						Id:          "getOutsideTemperature",
						Interaction: interaction,
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "outside_air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "simple_thermometer",
				DeviceClassId: "thermometer",
				Services: []model.Service{
					{
						Id:          "getTemperature",
						Interaction: interaction,
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "water-probe",
				DeviceClassId: "thermometer",
				Services: []model.Service{
					{
						Id:          "getTemperature",
						Interaction: interaction,
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:               "temperature",
									Name:             "temperature",
									FunctionId:       model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:         "water",
									CharacteristicId: "water-probe-test-characteristic",
								},
							},
						},
					},
				},
			},
			{
				Id:            "pc_cooling_controller",
				DeviceClassId: "pc_cooling_controller",
				Services: []model.Service{
					{
						Id:          "getTemperatures",
						Interaction: interaction,
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:   "temperatures",
									Name: "temperatures",
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
						Id:          "getFanSpeeds",
						Interaction: interaction,
						Outputs: []model.Content{
							{
								ContentVariable: model.ContentVariable{
									Id:   "speeds",
									Name: "speeds",
									SubContentVariables: []model.ContentVariable{
										{
											Id:         "cpu_fan",
											Name:       "cpu_fan",
											FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
											AspectId:   "cpu_fan",
										},
										{
											Id:         "gpu_fan",
											Name:       "gpu_fan",
											FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
											AspectId:   "gpu_fan",
										},
										{
											Id:         "case_fan_1",
											Name:       "case_fan_1",
											FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
											AspectId:   "case_fan_1",
										},
										{
											Id:         "case_fan_2",
											Name:       "case_fan_2",
											FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
											AspectId:   "case_fan_2",
										},
									},
								},
							},
						},
					},
					{
						Id:          "setCase1FanSpeed",
						Interaction: interaction,
						Outputs: []model.Content{
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

func GetDeviceTypeSelectables(config config.Config, token string, prefix string, interactionsFilter []model.Interaction, descriptions []model.FilterCriteria) (result []model.DeviceTypeSelectable, err error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	payload := new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(descriptions)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	interactionsQuery := ""
	if interactionsFilter != nil {
		interactions := []string{}
		for _, v := range interactionsFilter {
			interactions = append(interactions, string(v))
		}
		interactionsQuery = "&interactions-filter=" + url.QueryEscape(strings.Join(interactions, ","))
	}
	req, err := http.NewRequest(
		"POST",
		"http://localhost:"+config.ServerPort+"/query/device-type-selectables?path-prefix="+url.QueryEscape(prefix)+interactionsQuery,
		payload,
	)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		debug.PrintStack()
		temp, _ := io.ReadAll(resp.Body)
		log.Println("ERROR: GetDeviceTypeSelectables():", resp.StatusCode, string(temp))
		return result, errors.New("unexpected statuscode")
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}

	return result, err
}
