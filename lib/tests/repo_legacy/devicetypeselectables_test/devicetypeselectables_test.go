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

package devicetypeselectables_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/repo_legacy/testenv"
	"github.com/SENERGY-Platform/models/go/models"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDeviceTypeSelectables(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("init metadata", createTestMetadata(conf, models.REQUEST))

	t.Run("toggle", func(t *testing.T) {
		toggleCriteria := []model.FilterCriteria{{
			FunctionId:    model.CONTROLLING_FUNCTION_PREFIX + "toggle",
			DeviceClassId: "toggle",
		}}

		toggleSelectable := model.DeviceTypeSelectable{
			DeviceTypeId: "toggle",
			Services: []models.Service{
				{
					Id:          "triggerToggle",
					Interaction: models.REQUEST,
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:         "void",
								IsVoid:     true,
								FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "toggle",
							},
						},
					},
				},
			},
			ServicePathOptions: map[string][]model.ServicePathOption{
				"triggerToggle": {
					{
						ServiceId:             "triggerToggle",
						Path:                  "prefix.",
						CharacteristicId:      "",
						AspectNode:            models.AspectNode{},
						FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "toggle",
						IsControllingFunction: true,
						IsVoid:                true,
					},
				},
			},
		}

		t.Run("find toggle", testDeviceTypeSelectablesWithoutConfigurables(conf, toggleCriteria, "prefix.", nil, []model.DeviceTypeSelectable{toggleSelectable}))

	})

	t.Run("interaction filter", func(t *testing.T) {
		waterProbeCriteria := []model.FilterCriteria{{
			FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
			AspectId:   "water",
		}}
		waterprobeSelectable := model.DeviceTypeSelectable{
			DeviceTypeId: "water-probe",
			Services: []models.Service{
				{
					Id:          "getTemperature",
					Interaction: models.REQUEST,
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
						AspectNode: models.AspectNode{
							Id:            "water",
							Name:          "",
							RootId:        "water",
							ParentId:      "",
							ChildIds:      []string{},
							AncestorIds:   []string{},
							DescendentIds: []string{},
						},
						FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		}

		t.Run("nil", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", nil, []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("empty", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{}, []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("event", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{models.EVENT}, []model.DeviceTypeSelectable{}))
		t.Run("request", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{models.REQUEST}, []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("event+request", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{models.EVENT, models.REQUEST}, []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("event_and_request", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{models.EVENT_AND_REQUEST}, []model.DeviceTypeSelectable{}))
		t.Run("event+request in criteria", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{{
			FunctionId:  model.MEASURING_FUNCTION_PREFIX + "getTemperature",
			AspectId:    "water",
			Interaction: models.EVENT_AND_REQUEST,
		}}, "prefix.", []models.Interaction{}, []model.DeviceTypeSelectable{}))
		t.Run("event in criteria", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{{
			FunctionId:  model.MEASURING_FUNCTION_PREFIX + "getTemperature",
			AspectId:    "water",
			Interaction: models.EVENT_AND_REQUEST,
		}}, "prefix.", []models.Interaction{}, []model.DeviceTypeSelectable{}))
		t.Run("request in criteria", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{{
			FunctionId:  model.MEASURING_FUNCTION_PREFIX + "getTemperature",
			AspectId:    "water",
			Interaction: models.REQUEST,
		}}, "prefix.", []models.Interaction{}, []model.DeviceTypeSelectable{waterprobeSelectable}))
	})
}

func TestDeviceTypeSelectables2(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("init metadata", createTestMetadata(conf, models.EVENT_AND_REQUEST))

	t.Run("interaction filter", func(t *testing.T) {
		waterProbeCriteria := []model.FilterCriteria{{
			FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
			AspectId:   "water",
		}}
		waterprobeSelectable := model.DeviceTypeSelectable{
			DeviceTypeId: "water-probe",
			Services: []models.Service{
				{
					Id:          "getTemperature",
					Interaction: models.EVENT_AND_REQUEST,
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
						AspectNode: models.AspectNode{
							Id:            "water",
							Name:          "",
							RootId:        "water",
							ParentId:      "",
							ChildIds:      []string{},
							AncestorIds:   []string{},
							DescendentIds: []string{},
						},
						FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
					},
				},
			},
		}

		t.Run("nil", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", nil, []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("empty", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{}, []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("event", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{models.EVENT}, []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("request", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{models.REQUEST}, []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("event+request", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{models.EVENT, models.REQUEST}, []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("event_and_request", testDeviceTypeSelectablesWithoutConfigurables(conf, waterProbeCriteria, "prefix.", []models.Interaction{models.EVENT_AND_REQUEST}, []model.DeviceTypeSelectable{waterprobeSelectable}))

	})

	t.Run("interaction filter v2", func(t *testing.T) {
		waterProbeCriteria := []model.FilterCriteria{{
			FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
			AspectId:   "water",
		}}
		waterprobeSelectable := model.DeviceTypeSelectable{
			DeviceTypeId: "water-probe",
			Services: []models.Service{
				{
					Id:          "getTemperature",
					Interaction: models.EVENT_AND_REQUEST,
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
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
						AspectNode: models.AspectNode{
							Id:            "water",
							Name:          "",
							RootId:        "water",
							ParentId:      "",
							ChildIds:      []string{},
							AncestorIds:   []string{},
							DescendentIds: []string{},
						},
						FunctionId:  model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						Interaction: models.EVENT_AND_REQUEST,
					},
				},
			},
		}

		t.Run("nil", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, waterProbeCriteria, "prefix.", []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("empty", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, waterProbeCriteria, "prefix.", []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("event", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, testAddInteractionToCriterias(waterProbeCriteria, []models.Interaction{models.EVENT}), "prefix.", []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("request", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, testAddInteractionToCriterias(waterProbeCriteria, []models.Interaction{models.REQUEST}), "prefix.", []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("event+request", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, testAddInteractionToCriterias(waterProbeCriteria, []models.Interaction{models.EVENT, models.REQUEST}), "prefix.", []model.DeviceTypeSelectable{waterprobeSelectable}))
		t.Run("event_and_request", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, testAddInteractionToCriterias(waterProbeCriteria, []models.Interaction{models.EVENT_AND_REQUEST}), "prefix.", []model.DeviceTypeSelectable{waterprobeSelectable}))

		expectedNoCriteriaResultStr := `[{"device_type_id":"pc_cooling_controller","services":[{"id":"getFanSpeeds","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"speeds","name":"speeds","is_void":false,"omit_empty":false,"type":"","sub_content_variables":[{"id":"cpu_fan","name":"cpu_fan","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getFanSpeed","aspect_id":"cpu_fan"},{"id":"gpu_fan","name":"gpu_fan","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getFanSpeed","aspect_id":"gpu_fan"},{"id":"case_fan_1","name":"case_fan_1","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getFanSpeed","aspect_id":"case_fan_1"},{"id":"case_fan_2","name":"case_fan_2","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getFanSpeed","aspect_id":"case_fan_2"}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":""},{"id":"getTemperatures","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"temperatures","name":"temperatures","is_void":false,"omit_empty":false,"type":"","sub_content_variables":[{"id":"cpu","name":"cpu","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getTemperature","aspect_id":"cpu"},{"id":"gpu","name":"gpu","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getTemperature","aspect_id":"gpu"},{"id":"case","name":"case","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getTemperature","aspect_id":"case"}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":""},{"id":"setCaseFan1Speed","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"speed","name":"speed","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setFanSpeed","aspect_id":"case_fan_1"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""},{"id":"setCaseFan2Speed","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"speed","name":"speed","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setFanSpeed","aspect_id":"case_fan_2"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""},{"id":"setCpuSpeed","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"speed","name":"speed","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setFanSpeed","aspect_id":"cpu_fan"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""},{"id":"setGpuSpeed","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"speed","name":"speed","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setFanSpeed","aspect_id":"gpu_fan"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""}],"service_path_options":{"getFanSpeeds":[{"service_id":"getFanSpeeds","path":"prefix.speeds.case_fan_1","characteristic_id":"","aspect_node":{"id":"case_fan_1","name":"","root_id":"fan","parent_id":"case_fan","child_ids":[],"ancestor_ids":["case_fan","fan"],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getFanSpeed","is_void":false,"is_controlling_function":false,"interaction":"event+request"},{"service_id":"getFanSpeeds","path":"prefix.speeds.case_fan_2","characteristic_id":"","aspect_node":{"id":"case_fan_2","name":"","root_id":"fan","parent_id":"case_fan","child_ids":[],"ancestor_ids":["case_fan","fan"],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getFanSpeed","is_void":false,"is_controlling_function":false,"interaction":"event+request"},{"service_id":"getFanSpeeds","path":"prefix.speeds.cpu_fan","characteristic_id":"","aspect_node":{"id":"cpu_fan","name":"","root_id":"fan","parent_id":"fan","child_ids":[],"ancestor_ids":["fan"],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getFanSpeed","is_void":false,"is_controlling_function":false,"interaction":"event+request"},{"service_id":"getFanSpeeds","path":"prefix.speeds.gpu_fan","characteristic_id":"","aspect_node":{"id":"gpu_fan","name":"","root_id":"fan","parent_id":"fan","child_ids":[],"ancestor_ids":["fan"],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getFanSpeed","is_void":false,"is_controlling_function":false,"interaction":"event+request"}],"getTemperatures":[{"service_id":"getTemperatures","path":"prefix.temperatures.case","characteristic_id":"","aspect_node":{"id":"case","name":"","root_id":"device","parent_id":"device","child_ids":[],"ancestor_ids":["device"],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getTemperature","is_void":false,"is_controlling_function":false,"interaction":"event+request"},{"service_id":"getTemperatures","path":"prefix.temperatures.cpu","characteristic_id":"","aspect_node":{"id":"cpu","name":"","root_id":"device","parent_id":"device","child_ids":[],"ancestor_ids":["device"],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getTemperature","is_void":false,"is_controlling_function":false,"interaction":"event+request"},{"service_id":"getTemperatures","path":"prefix.temperatures.gpu","characteristic_id":"","aspect_node":{"id":"gpu","name":"","root_id":"device","parent_id":"device","child_ids":[],"ancestor_ids":["device"],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getTemperature","is_void":false,"is_controlling_function":false,"interaction":"event+request"}],"setCaseFan1Speed":[{"service_id":"setCaseFan1Speed","path":"prefix.speed","characteristic_id":"","aspect_node":{"id":"case_fan_1","name":"","root_id":"fan","parent_id":"case_fan","child_ids":[],"ancestor_ids":["case_fan","fan"],"descendent_ids":[]},"function_id":"urn:infai:ses:controlling-function:setFanSpeed","is_void":false,"is_controlling_function":true,"interaction":"event+request"}],"setCaseFan2Speed":[{"service_id":"setCaseFan2Speed","path":"prefix.speed","characteristic_id":"","aspect_node":{"id":"case_fan_2","name":"","root_id":"fan","parent_id":"case_fan","child_ids":[],"ancestor_ids":["case_fan","fan"],"descendent_ids":[]},"function_id":"urn:infai:ses:controlling-function:setFanSpeed","is_void":false,"is_controlling_function":true,"interaction":"event+request"}],"setCpuSpeed":[{"service_id":"setCpuSpeed","path":"prefix.speed","characteristic_id":"","aspect_node":{"id":"cpu_fan","name":"","root_id":"fan","parent_id":"fan","child_ids":[],"ancestor_ids":["fan"],"descendent_ids":[]},"function_id":"urn:infai:ses:controlling-function:setFanSpeed","is_void":false,"is_controlling_function":true,"interaction":"event+request"}],"setGpuSpeed":[{"service_id":"setGpuSpeed","path":"prefix.speed","characteristic_id":"","aspect_node":{"id":"gpu_fan","name":"","root_id":"fan","parent_id":"fan","child_ids":[],"ancestor_ids":["fan"],"descendent_ids":[]},"function_id":"urn:infai:ses:controlling-function:setFanSpeed","is_void":false,"is_controlling_function":true,"interaction":"event+request"}]}},{"device_type_id":"plug-strip","services":[{"id":"plug1","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"state1","name":"state","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"plug-state-characteristic","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getPlugState","aspect_id":"plug"},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":"sg1"},{"id":"plug2","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"state2","name":"state","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"plug-state-characteristic","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getPlugState","aspect_id":"plug"},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":"sg2"},{"id":"plugs","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"states","name":"states","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"plug-state-list-characteristic","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getPlugStates","aspect_id":"plug"},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":""}],"service_path_options":{"plug1":[{"service_id":"plug1","path":"prefix.state","characteristic_id":"plug-state-characteristic","aspect_node":{"id":"plug","name":"","root_id":"plug","parent_id":"","child_ids":[],"ancestor_ids":[],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getPlugState","is_void":false,"is_controlling_function":false,"interaction":"event+request"}],"plug2":[{"service_id":"plug2","path":"prefix.state","characteristic_id":"plug-state-characteristic","aspect_node":{"id":"plug","name":"","root_id":"plug","parent_id":"","child_ids":[],"ancestor_ids":[],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getPlugState","is_void":false,"is_controlling_function":false,"interaction":"event+request"}],"plugs":[{"service_id":"plugs","path":"prefix.states","characteristic_id":"plug-state-list-characteristic","aspect_node":{"id":"plug","name":"","root_id":"plug","parent_id":"","child_ids":[],"ancestor_ids":[],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getPlugStates","is_void":false,"is_controlling_function":false,"interaction":"event+request"}]}},{"device_type_id":"simple_thermometer","services":[{"id":"getTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getTemperature","aspect_id":"air"},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":""}],"service_path_options":{"getTemperature":[{"service_id":"getTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"air","name":"","root_id":"air","parent_id":"","child_ids":["inside_air","outside_air"],"ancestor_ids":[],"descendent_ids":["evening_outside_air","inside_air","morning_outside_air","outside_air"]},"function_id":"urn:infai:ses:measuring-function:getTemperature","is_void":false,"is_controlling_function":false,"interaction":"event+request"}]}},{"device_type_id":"thermometer","services":[{"id":"getInsideTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getTemperature","aspect_id":"inside_air"},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":""},{"id":"getOutsideTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getTemperature","aspect_id":"outside_air"},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":""}],"service_path_options":{"getInsideTemperature":[{"service_id":"getInsideTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"inside_air","name":"","root_id":"air","parent_id":"air","child_ids":[],"ancestor_ids":["air"],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getTemperature","is_void":false,"is_controlling_function":false,"interaction":"event+request"}],"getOutsideTemperature":[{"service_id":"getOutsideTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"outside_air","name":"","root_id":"air","parent_id":"air","child_ids":["evening_outside_air","morning_outside_air"],"ancestor_ids":["air"],"descendent_ids":["evening_outside_air","morning_outside_air"]},"function_id":"urn:infai:ses:measuring-function:getTemperature","is_void":false,"is_controlling_function":false,"interaction":"event+request"}]}},{"device_type_id":"thermostat","services":[{"id":"getTargetTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getTemperature","aspect_id":"inside_air"},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":""},{"id":"setTargetTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setTemperature","aspect_id":"inside_air"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""}],"service_path_options":{"getTargetTemperature":[{"service_id":"getTargetTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"inside_air","name":"","root_id":"air","parent_id":"air","child_ids":[],"ancestor_ids":["air"],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getTemperature","is_void":false,"is_controlling_function":false,"interaction":"event+request"}],"setTargetTemperature":[{"service_id":"setTargetTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"inside_air","name":"","root_id":"air","parent_id":"air","child_ids":[],"ancestor_ids":["air"],"descendent_ids":[]},"function_id":"urn:infai:ses:controlling-function:setTemperature","is_void":false,"is_controlling_function":true,"interaction":"event+request"}]}},{"device_type_id":"thermostat_without_get","services":[{"id":"setTargetTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setTemperature","aspect_id":"inside_air"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""}],"service_path_options":{"setTargetTemperature":[{"service_id":"setTargetTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"inside_air","name":"","root_id":"air","parent_id":"air","child_ids":[],"ancestor_ids":["air"],"descendent_ids":[]},"function_id":"urn:infai:ses:controlling-function:setTemperature","is_void":false,"is_controlling_function":true,"interaction":"event+request"}]}},{"device_type_id":"thermostat_without_get_base","services":[{"id":"setTargetTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setTemperature","aspect_id":"air"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""}],"service_path_options":{"setTargetTemperature":[{"service_id":"setTargetTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"air","name":"","root_id":"air","parent_id":"","child_ids":["inside_air","outside_air"],"ancestor_ids":[],"descendent_ids":["evening_outside_air","inside_air","morning_outside_air","outside_air"]},"function_id":"urn:infai:ses:controlling-function:setTemperature","is_void":false,"is_controlling_function":true,"interaction":"event+request"}]}},{"device_type_id":"thermostat_without_get_multiservice","services":[{"id":"setInsideTargetTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setTemperature","aspect_id":"inside_air"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""},{"id":"setOutsideTargetTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setTemperature","aspect_id":"outside_air"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""}],"service_path_options":{"setInsideTargetTemperature":[{"service_id":"setInsideTargetTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"inside_air","name":"","root_id":"air","parent_id":"air","child_ids":[],"ancestor_ids":["air"],"descendent_ids":[]},"function_id":"urn:infai:ses:controlling-function:setTemperature","is_void":false,"is_controlling_function":true,"interaction":"event+request"}],"setOutsideTargetTemperature":[{"service_id":"setOutsideTargetTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"outside_air","name":"","root_id":"air","parent_id":"air","child_ids":["evening_outside_air","morning_outside_air"],"ancestor_ids":["air"],"descendent_ids":["evening_outside_air","morning_outside_air"]},"function_id":"urn:infai:ses:controlling-function:setTemperature","is_void":false,"is_controlling_function":true,"interaction":"event+request"}]}},{"device_type_id":"thermostat_without_get_multivalue","services":[{"id":"setTargetTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":[{"id":"inside","name":"inside","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setTemperature","aspect_id":"inside_air"},{"id":"outside","name":"outside","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setTemperature","aspect_id":"outside_air"}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""}],"service_path_options":{"setTargetTemperature":[{"service_id":"setTargetTemperature","path":"prefix.temperature.inside","characteristic_id":"","aspect_node":{"id":"inside_air","name":"","root_id":"air","parent_id":"air","child_ids":[],"ancestor_ids":["air"],"descendent_ids":[]},"function_id":"urn:infai:ses:controlling-function:setTemperature","is_void":false,"is_controlling_function":true,"interaction":"event+request"},{"service_id":"setTargetTemperature","path":"prefix.temperature.outside","characteristic_id":"","aspect_node":{"id":"outside_air","name":"","root_id":"air","parent_id":"air","child_ids":["evening_outside_air","morning_outside_air"],"ancestor_ids":["air"],"descendent_ids":["evening_outside_air","morning_outside_air"]},"function_id":"urn:infai:ses:controlling-function:setTemperature","is_void":false,"is_controlling_function":true,"interaction":"event+request"}]}},{"device_type_id":"thermostat_without_get_without_aspect","services":[{"id":"setTargetTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:setTemperature"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""}],"service_path_options":{"setTargetTemperature":[{"service_id":"setTargetTemperature","path":"prefix.temperature","characteristic_id":"","aspect_node":{"id":"","name":"","root_id":"","parent_id":"","child_ids":null,"ancestor_ids":null,"descendent_ids":null},"function_id":"urn:infai:ses:controlling-function:setTemperature","is_void":false,"is_controlling_function":true,"interaction":"event+request"}]}},{"device_type_id":"toggle","services":[{"id":"triggerToggle","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":[{"id":"","content_variable":{"id":"void","name":"","is_void":true,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"","value":null,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:toggle"},"serialization":"","protocol_segment_id":""}],"outputs":null,"attributes":null,"service_group_key":""}],"service_path_options":{"triggerToggle":[{"service_id":"triggerToggle","path":"prefix.","characteristic_id":"","aspect_node":{"id":"","name":"","root_id":"","parent_id":"","child_ids":null,"ancestor_ids":null,"descendent_ids":null},"function_id":"urn:infai:ses:controlling-function:toggle","is_void":true,"is_controlling_function":true,"interaction":"event+request"}]}},{"device_type_id":"water-probe","services":[{"id":"getTemperature","local_id":"","name":"","description":"","interaction":"event+request","protocol_id":"","inputs":null,"outputs":[{"id":"","content_variable":{"id":"temperature","name":"temperature","is_void":false,"omit_empty":false,"type":"","sub_content_variables":null,"characteristic_id":"water-probe-test-characteristic","value":null,"serialization_options":null,"function_id":"urn:infai:ses:measuring-function:getTemperature","aspect_id":"water"},"serialization":"","protocol_segment_id":""}],"attributes":null,"service_group_key":""}],"service_path_options":{"getTemperature":[{"service_id":"getTemperature","path":"prefix.temperature","characteristic_id":"water-probe-test-characteristic","aspect_node":{"id":"water","name":"","root_id":"water","parent_id":"","child_ids":[],"ancestor_ids":[],"descendent_ids":[]},"function_id":"urn:infai:ses:measuring-function:getTemperature","is_void":false,"is_controlling_function":false,"interaction":"event+request"}]}}]`
		expectedNoCriteriaResult := []model.DeviceTypeSelectable{}
		err = json.Unmarshal([]byte(expectedNoCriteriaResultStr), &expectedNoCriteriaResult)
		if err != nil {
			t.Error(err)
		}

		t.Run("no criteria", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, nil, "prefix.", expectedNoCriteriaResult))
	})

	t.Run("measuring", func(t *testing.T) {
		interaction := models.EVENT_AND_REQUEST
		t.Run("inside and outside temp", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "inside_air"},
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "outside_air"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "thermometer",
				Services: []models.Service{
					{
						Id:          "getInsideTemperature",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
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
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "outside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getInsideTemperature": {
						{
							ServiceId:        "getInsideTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
					"getOutsideTemperature": {
						{
							ServiceId:        "getOutsideTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "outside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
				},
			},
		}))

		t.Run("inside temp", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "inside_air"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "thermometer",
				Services: []models.Service{
					{
						Id:          "getInsideTemperature",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getInsideTemperature": {
						{
							ServiceId:        "getInsideTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat",
				Services: []models.Service{
					{
						Id:          "getTargetTemperature",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getTargetTemperature": {
						{
							ServiceId:        "getTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
				},
			},
		}))

		t.Run("air temperature", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "air"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "simple_thermometer",
				Services: []models.Service{
					{
						Id:          "getTemperature",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getTemperature": {
						{
							ServiceId:        "getTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "air",
								RootId:        "air",
								ParentId:      "",
								ChildIds:      []string{"inside_air", "outside_air"},
								AncestorIds:   []string{},
								DescendentIds: []string{"evening_outside_air", "inside_air", "morning_outside_air", "outside_air"},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
				},
			},
			{
				DeviceTypeId: "thermometer",
				Services: []models.Service{
					{
						Id:          "getInsideTemperature",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
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
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "outside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getInsideTemperature": {
						{
							ServiceId:        "getInsideTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
					"getOutsideTemperature": {
						{
							ServiceId:        "getOutsideTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "outside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat",
				Services: []models.Service{
					{
						Id:          "getTargetTemperature",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getTargetTemperature": {
						{
							ServiceId:        "getTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
				},
			},
		}))

		t.Run("device temperature", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "device"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "pc_cooling_controller",
				Services: []models.Service{
					{
						Id:          "getTemperatures",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "temperatures",
									Name: "temperatures",
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
							Path:             "temperatures.case",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "case",
								RootId:        "device",
								ParentId:      "device",
								ChildIds:      []string{},
								AncestorIds:   []string{"device"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
						{
							ServiceId:        "getTemperatures",
							Path:             "temperatures.cpu",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "cpu",
								RootId:        "device",
								ParentId:      "device",
								ChildIds:      []string{},
								AncestorIds:   []string{"device"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
						{
							ServiceId:        "getTemperatures",
							Path:             "temperatures.gpu",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "gpu",
								RootId:        "device",
								ParentId:      "device",
								ChildIds:      []string{},
								AncestorIds:   []string{"device"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
				},
			},
		}))

		t.Run("cpu temperature", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature", AspectId: "cpu"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "pc_cooling_controller",
				Services: []models.Service{
					{
						Id:          "getTemperatures",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "temperatures",
									Name: "temperatures",
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
							Path:             "temperatures.cpu",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "cpu",
								RootId:        "device",
								ParentId:      "device",
								ChildIds:      []string{},
								AncestorIds:   []string{"device"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getTemperature",
						},
					},
				},
			},
		}))

		t.Run("fan speed", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed", AspectId: "fan"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "pc_cooling_controller",
				Services: []models.Service{
					{
						Id:          "getFanSpeeds",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "speeds",
									Name: "speeds",
									SubContentVariables: []models.ContentVariable{
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getFanSpeeds": {
						{
							ServiceId:        "getFanSpeeds",
							Path:             "speeds.case_fan_1",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "case_fan_1",
								RootId:        "fan",
								ParentId:      "case_fan",
								ChildIds:      []string{},
								AncestorIds:   []string{"case_fan", "fan"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
						},
						{
							ServiceId:        "getFanSpeeds",
							Path:             "speeds.case_fan_2",
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
						},
						{
							ServiceId:        "getFanSpeeds",
							Path:             "speeds.cpu_fan",
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
						},
						{
							ServiceId:        "getFanSpeeds",
							Path:             "speeds.gpu_fan",
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
						},
					},
				},
			},
		}))

		t.Run("cpu fan speed", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed", AspectId: "cpu_fan"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "pc_cooling_controller",
				Services: []models.Service{
					{
						Id:          "getFanSpeeds",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "speeds",
									Name: "speeds",
									SubContentVariables: []models.ContentVariable{
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getFanSpeeds": {
						{
							ServiceId:        "getFanSpeeds",
							Path:             "speeds.cpu_fan",
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
						},
					},
				},
			},
		}))

		t.Run("case fan speed", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed", AspectId: "case_fan"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "pc_cooling_controller",
				Services: []models.Service{
					{
						Id:          "getFanSpeeds",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "speeds",
									Name: "speeds",
									SubContentVariables: []models.ContentVariable{
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getFanSpeeds": {
						{
							ServiceId:        "getFanSpeeds",
							Path:             "speeds.case_fan_1",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "case_fan_1",
								RootId:        "fan",
								ParentId:      "case_fan",
								ChildIds:      []string{},
								AncestorIds:   []string{"case_fan", "fan"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
						},
						{
							ServiceId:        "getFanSpeeds",
							Path:             "speeds.case_fan_2",
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
						},
					},
				},
			},
		}))

		t.Run("case fan speed", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed", AspectId: "case_fan_1"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "pc_cooling_controller",
				Services: []models.Service{
					{
						Id:          "getFanSpeeds",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "speeds",
									Name: "speeds",
									SubContentVariables: []models.ContentVariable{
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"getFanSpeeds": {
						{
							ServiceId:        "getFanSpeeds",
							Path:             "speeds.case_fan_1",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "case_fan_1",
								RootId:        "fan",
								ParentId:      "case_fan",
								ChildIds:      []string{},
								AncestorIds:   []string{"case_fan", "fan"},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed",
						},
					},
				},
			},
		}))
	})

	t.Run("controlling", func(t *testing.T) {
		interaction := models.EVENT_AND_REQUEST
		t.Run("thermostat", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature", DeviceClassId: "thermostat"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "thermostat",
				Services: []models.Service{{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				}},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId: "setTargetTemperature",
							Path:      "temperature",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get_base",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "air",
								RootId:        "air",
								ParentId:      "",
								ChildIds:      []string{"inside_air", "outside_air"},
								AncestorIds:   []string{},
								DescendentIds: []string{"evening_outside_air", "inside_air", "morning_outside_air", "outside_air"},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get_multiservice",
				Services: []models.Service{
					{
						Id:          "setInsideTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
					{
						Id:          "setOutsideTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "outside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setInsideTargetTemperature": {
						{
							ServiceId:        "setInsideTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
					"setOutsideTargetTemperature": {
						{
							ServiceId:        "setOutsideTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "outside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get_multivalue",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "temperature",
									Name: "temperature",
									SubContentVariables: []models.ContentVariable{
										{
											Id:         "inside",
											Name:       "inside",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "inside_air",
										},
										{
											Id:         "outside",
											Name:       "outside",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "outside_air",
										},
									},
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature.inside",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature.outside",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "outside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get_without_aspect",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId:             "setTargetTemperature",
							Path:                  "temperature",
							CharacteristicId:      "",
							AspectNode:            models.AspectNode{},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
		}))

		t.Run("thermostat air", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature", DeviceClassId: "thermostat", AspectId: "air"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "thermostat",
				Services: []models.Service{{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				}},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId: "setTargetTemperature",
							Path:      "temperature",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get_base",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "air",
								RootId:        "air",
								ParentId:      "",
								ChildIds:      []string{"inside_air", "outside_air"},
								AncestorIds:   []string{},
								DescendentIds: []string{"evening_outside_air", "inside_air", "morning_outside_air", "outside_air"},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get_multiservice",
				Services: []models.Service{
					{
						Id:          "setInsideTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
					{
						Id:          "setOutsideTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "outside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setInsideTargetTemperature": {
						{
							ServiceId:        "setInsideTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
					"setOutsideTargetTemperature": {
						{
							ServiceId:        "setOutsideTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "outside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get_multivalue",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "temperature",
									Name: "temperature",
									SubContentVariables: []models.ContentVariable{
										{
											Id:         "inside",
											Name:       "inside",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "inside_air",
										},
										{
											Id:         "outside",
											Name:       "outside",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "outside_air",
										},
									},
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature.inside",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature.outside",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "outside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{"evening_outside_air", "morning_outside_air"},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{"evening_outside_air", "morning_outside_air"},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
		}))

		t.Run("thermostat inside air", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature", DeviceClassId: "thermostat", AspectId: "inside_air"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "thermostat",
				Services: []models.Service{{
					Id:          "setTargetTemperature",
					Interaction: interaction,
					Inputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Id:         "temperature",
								Name:       "temperature",
								FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
								AspectId:   "inside_air",
							},
						},
					},
				}},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId: "setTargetTemperature",
							Path:      "temperature",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get_multiservice",
				Services: []models.Service{
					{
						Id:          "setInsideTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setInsideTargetTemperature": {
						{
							ServiceId:        "setInsideTargetTemperature",
							Path:             "temperature",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
			{
				DeviceTypeId: "thermostat_without_get_multivalue",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "temperature",
									Name: "temperature",
									SubContentVariables: []models.ContentVariable{
										{
											Id:         "inside",
											Name:       "inside",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "inside_air",
										},
										{
											Id:         "outside",
											Name:       "outside",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "outside_air",
										},
									},
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"setTargetTemperature": {
						{
							ServiceId:        "setTargetTemperature",
							Path:             "temperature.inside",
							CharacteristicId: "",
							AspectNode: models.AspectNode{
								Id:            "inside_air",
								RootId:        "air",
								ParentId:      "air",
								ChildIds:      []string{},
								AncestorIds:   []string{"air"},
								DescendentIds: []string{},
							},
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
							IsControllingFunction: true,
						},
					},
				},
			},
		}))

		t.Run("pc_cooling_controller fan_speed", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed", DeviceClassId: "pc_cooling_controller"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "pc_cooling_controller",
				Services: []models.Service{
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
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "case_fan_2",
								},
							},
						},
					},
					{
						Id:          "setCpuSpeed",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "cpu_fan",
								},
							},
						},
					},
					{
						Id:          "setGpuSpeed",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "gpu_fan",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
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
						},
					},
					"setCaseFan2Speed": {
						{
							ServiceId:        "setCaseFan2Speed",
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
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
							IsControllingFunction: true,
						},
					},
					"setCpuSpeed": {
						{
							ServiceId:        "setCpuSpeed",
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
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
							IsControllingFunction: true,
						},
					},
					"setGpuSpeed": {
						{
							ServiceId:        "setGpuSpeed",
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
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
							IsControllingFunction: true,
						},
					},
				},
			},
		}))

		t.Run("pc_cooling_controller fan_speed case_fan", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed", DeviceClassId: "pc_cooling_controller", AspectId: "case_fan"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "pc_cooling_controller",
				Services: []models.Service{
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
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "case_fan_2",
								},
							},
						},
					},
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
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
						},
					},
					"setCaseFan2Speed": {
						{
							ServiceId:        "setCaseFan2Speed",
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
							FunctionId:            model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
							IsControllingFunction: true,
						},
					},
				},
			},
		}))

		t.Run("pc_cooling_controller fan_speed case_fan_1", testDeviceTypeSelectablesWithoutConfigurables(conf, []model.FilterCriteria{
			{FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed", DeviceClassId: "pc_cooling_controller", AspectId: "case_fan_1"},
		}, "", nil, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "pc_cooling_controller",
				Services: []models.Service{
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
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
						},
					},
				},
			},
		}))
	})

}

func testAddInteractionToCriterias(criteria []model.FilterCriteria, interactions []models.Interaction) (result []model.FilterCriteria) {
	for _, interaction := range interactions {
		for _, c := range criteria {
			c.Interaction = interaction
			result = append(result, c)
		}
	}
	return result
}

func testDeviceTypeSelectablesWithoutConfigurables(config config.Config, criteria []model.FilterCriteria, pathPrefix string, interactionsFilter []models.Interaction, expectedResult []model.DeviceTypeSelectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := GetDeviceTypeSelectables(config, testenv.Userjwt, pathPrefix, interactionsFilter, criteria)
		if err != nil {
			t.Error(err)
			return
		}
		expectedResult = removeConfigurables(expectedResult)
		expectedResult = sortServices(expectedResult)
		result = removeConfigurables(result)
		result = sortServices(result)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func testDeviceTypeSelectablesWithoutConfigurablesV2(config config.Config, criteria []model.FilterCriteria, pathPrefix string, expectedResult []model.DeviceTypeSelectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := GetDeviceTypeSelectablesV2(config, testenv.Userjwt, pathPrefix, criteria)
		if err != nil {
			t.Error(err)
			return
		}
		expectedResult = removeConfigurables(expectedResult)
		expectedResult = sortServices(expectedResult)
		result = removeConfigurables(result)
		result = sortServices(result)
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func removeConfigurables(list []model.DeviceTypeSelectable) (result []model.DeviceTypeSelectable) {
	result = []model.DeviceTypeSelectable{}
	for _, e := range list {
		for sid, pathoptions := range e.ServicePathOptions {
			temp := []model.ServicePathOption{}
			for _, option := range pathoptions {
				option.Configurables = nil
				temp = append(temp, option)
			}
			e.ServicePathOptions[sid] = temp
		}
		result = append(result, e)
	}
	return
}

func sortServices(list []model.DeviceTypeSelectable) (result []model.DeviceTypeSelectable) {
	result = []model.DeviceTypeSelectable{}
	for _, e := range list {
		sort.Slice(e.Services, func(i, j int) bool {
			return e.Services[i].Id < e.Services[j].Id
		})
		result = append(result, e)
	}
	return
}

func createTestMetadata(config config.Config, interaction models.Interaction) func(t *testing.T) {
	return func(t *testing.T) {
		aspects := []models.Aspect{
			{
				Id: "plug",
			},
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
		functions := []models.Function{
			{Id: model.MEASURING_FUNCTION_PREFIX + "getPlugState"},
			{Id: model.MEASURING_FUNCTION_PREFIX + "getPlugStates"},
			{Id: model.MEASURING_FUNCTION_PREFIX + "getTemperature"},
			{Id: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature"},
			{Id: model.MEASURING_FUNCTION_PREFIX + "getVolume"},
			{Id: model.CONTROLLING_FUNCTION_PREFIX + "setVolume"},
			{Id: model.MEASURING_FUNCTION_PREFIX + "getFanSpeed"},
			{Id: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed"},
			{Id: model.CONTROLLING_FUNCTION_PREFIX + "toggle"},
		}
		devicetypes := []models.DeviceType{
			{
				Id:            "plug-strip",
				DeviceClassId: "toggle",
				Name:          "dt",
				ServiceGroups: []models.ServiceGroup{{Key: "sg1", Name: "sg1"}, {Key: "sg2", Name: "sg2"}},
				Services: []models.Service{
					{
						Id:              "plug1",
						ServiceGroupKey: "sg1",
						Interaction:     interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:               "state1",
									Name:             "state",
									FunctionId:       model.MEASURING_FUNCTION_PREFIX + "getPlugState",
									AspectId:         "plug",
									CharacteristicId: "plug-state-characteristic",
								},
							},
						},
					},
					{
						Id:              "plug2",
						ServiceGroupKey: "sg2",
						Interaction:     interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:               "state2",
									Name:             "state",
									FunctionId:       model.MEASURING_FUNCTION_PREFIX + "getPlugState",
									AspectId:         "plug",
									CharacteristicId: "plug-state-characteristic",
								},
							},
						},
					},
					{
						Id:          "plugs",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:               "states",
									Name:             "states",
									FunctionId:       model.MEASURING_FUNCTION_PREFIX + "getPlugStates",
									AspectId:         "plug",
									CharacteristicId: "plug-state-list-characteristic",
								},
							},
						},
					},
				},
			},
			{
				Id:            "toggle",
				DeviceClassId: "toggle",
				Services: []models.Service{
					{
						Id:          "triggerToggle",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "void",
									IsVoid:     true,
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "toggle",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get",
				DeviceClassId: "thermostat",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get_base",
				DeviceClassId: "thermostat",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get_without_aspect",
				DeviceClassId: "thermostat",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get_multivalue",
				DeviceClassId: "thermostat",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "temperature",
									Name: "temperature",
									SubContentVariables: []models.ContentVariable{
										{
											Id:         "inside",
											Name:       "inside",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "inside_air",
										},
										{
											Id:         "outside",
											Name:       "outside",
											FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
											AspectId:   "outside_air",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat_without_get_multiservice",
				DeviceClassId: "thermostat",
				Services: []models.Service{
					{
						Id:          "setInsideTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "inside_air",
								},
							},
						},
					},
					{
						Id:          "setOutsideTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "temperature",
									Name:       "temperature",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setTemperature",
									AspectId:   "outside_air",
								},
							},
						},
					},
				},
			},
			{
				Id:            "thermostat",
				DeviceClassId: "thermostat",
				Services: []models.Service{
					{
						Id:          "setTargetTemperature",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
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
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
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
				Services: []models.Service{
					{
						Id:          "getInsideTemperature",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
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
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
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
				Services: []models.Service{
					{
						Id:          "getTemperature",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
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
				Services: []models.Service{
					{
						Id:          "getTemperature",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
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
				Services: []models.Service{
					{
						Id:          "getTemperatures",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "temperatures",
									Name: "temperatures",
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
						Id:          "getFanSpeeds",
						Interaction: interaction,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:   "speeds",
									Name: "speeds",
									SubContentVariables: []models.ContentVariable{
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
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "case_fan_2",
								},
							},
						},
					},
					{
						Id:          "setCpuSpeed",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "cpu_fan",
								},
							},
						},
					},
					{
						Id:          "setGpuSpeed",
						Interaction: interaction,
						Inputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Id:         "speed",
									Name:       "speed",
									FunctionId: model.CONTROLLING_FUNCTION_PREFIX + "setFanSpeed",
									AspectId:   "gpu_fan",
								},
							},
						},
					},
				},
			},
		}

		c := client.NewClient("http://localhost:"+config.ServerPort, nil)

		for _, aspect := range aspects {
			_, err, _ := c.SetAspect(testenv.AdminToken, aspect)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, function := range functions {
			_, err, _ := c.SetFunction(testenv.AdminToken, function)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, dt := range devicetypes {
			_, err, _ := c.SetDeviceType(testenv.AdminToken, dt, client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
		}
	}
}

func createTestMetadataFromString(config config.Config, deviceTypesStr string, aspectsStr string, functionsStr string) func(t *testing.T) {
	return func(t *testing.T) {
		aspects := []models.Aspect{}
		functions := []models.Function{}
		devicetypes := []models.DeviceType{}

		err := json.Unmarshal([]byte(deviceTypesStr), &devicetypes)
		if err != nil {
			t.Error(err)
			return
		}

		err = json.Unmarshal([]byte(functionsStr), &functions)
		if err != nil {
			t.Error(err)
			return
		}

		err = json.Unmarshal([]byte(aspectsStr), &aspects)
		if err != nil {
			t.Error(err)
			return
		}

		c := client.NewClient("http://localhost:"+config.ServerPort, nil)

		for _, aspect := range aspects {
			_, err, _ = c.SetAspect(testenv.AdminToken, aspect)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, function := range functions {
			_, err, _ = c.SetFunction(testenv.AdminToken, function)
			if err != nil {
				t.Error(err)
				return
			}
		}

		for _, dt := range devicetypes {
			_, err, _ = c.SetDeviceType(testenv.AdminToken, dt, client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
		}
	}
}

func GetDeviceTypeSelectables(config config.Config, token string, prefix string, interactionsFilter []models.Interaction, descriptions []model.FilterCriteria) (result []model.DeviceTypeSelectable, err error) {
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

func GetDeviceTypeSelectablesV2(config config.Config, token string, prefix string, descriptions []model.FilterCriteria) (result []model.DeviceTypeSelectable, err error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	payload := new(bytes.Buffer)
	err = json.NewEncoder(payload).Encode(descriptions)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req, err := http.NewRequest(
		"POST",
		"http://localhost:"+config.ServerPort+"/v2/query/device-type-selectables?path-prefix="+url.QueryEscape(prefix),
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
