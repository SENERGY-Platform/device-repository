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
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testenv"
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

func TestDeviceTypeSelectablesWithModifiedId(t *testing.T) {
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

	t.Run("v1", func(t *testing.T) {
		criteria := []model.FilterCriteria{{
			FunctionId: model.MEASURING_FUNCTION_PREFIX + "getPlugState",
			AspectId:   "plug",
		}}
		expectedSelectables := []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "plug-strip",
				Services: []models.Service{
					{
						Id:              "plug1",
						ServiceGroupKey: "sg1",
						Interaction:     models.REQUEST,
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
						Interaction:     models.REQUEST,
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"plug1": {
						{
							ServiceId:        "plug1",
							Path:             "prefix.state",
							CharacteristicId: "plug-state-characteristic",
							AspectNode: models.AspectNode{
								Id:            "plug",
								Name:          "",
								RootId:        "plug",
								ParentId:      "",
								ChildIds:      []string{},
								AncestorIds:   []string{},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getPlugState",
						},
					},
					"plug2": {
						{
							ServiceId:        "plug2",
							Path:             "prefix.state",
							CharacteristicId: "plug-state-characteristic",
							AspectNode: models.AspectNode{
								Id:            "plug",
								Name:          "",
								RootId:        "plug",
								ParentId:      "",
								ChildIds:      []string{},
								AncestorIds:   []string{},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getPlugState",
						},
					},
				},
			},
		}

		var expectedSeletablesWithModifiedIds []model.DeviceTypeSelectable = append(expectedSelectables, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "plug-strip" + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {"sg1"}}),
				Services: []models.Service{
					{
						Id:              "plug1",
						ServiceGroupKey: "sg1",
						Interaction:     models.REQUEST,
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"plug1": {
						{
							ServiceId:        "plug1",
							Path:             "prefix.state",
							CharacteristicId: "plug-state-characteristic",
							AspectNode: models.AspectNode{
								Id:            "plug",
								Name:          "",
								RootId:        "plug",
								ParentId:      "",
								ChildIds:      []string{},
								AncestorIds:   []string{},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getPlugState",
						},
					},
				},
			},
			{
				DeviceTypeId: "plug-strip" + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {"sg2"}}),
				Services: []models.Service{
					{
						Id:              "plug2",
						ServiceGroupKey: "sg2",
						Interaction:     models.REQUEST,
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"plug2": {
						{
							ServiceId:        "plug2",
							Path:             "prefix.state",
							CharacteristicId: "plug-state-characteristic",
							AspectNode: models.AspectNode{
								Id:            "plug",
								Name:          "",
								RootId:        "plug",
								ParentId:      "",
								ChildIds:      []string{},
								AncestorIds:   []string{},
								DescendentIds: []string{},
							},
							FunctionId: model.MEASURING_FUNCTION_PREFIX + "getPlugState",
						},
					},
				},
			},
		}...)

		t.Run("nil", testDeviceTypeSelectablesWithoutConfigurables(conf, criteria, "prefix.", nil, expectedSelectables))
		t.Run("empty", testDeviceTypeSelectablesWithoutConfigurables(conf, criteria, "prefix.", []models.Interaction{}, expectedSelectables))
		t.Run("event", testDeviceTypeSelectablesWithoutConfigurables(conf, criteria, "prefix.", []models.Interaction{models.EVENT}, []model.DeviceTypeSelectable{}))
		t.Run("request", testDeviceTypeSelectablesWithoutConfigurables(conf, criteria, "prefix.", []models.Interaction{models.REQUEST}, expectedSelectables))
		t.Run("event+request", testDeviceTypeSelectablesWithoutConfigurables(conf, criteria, "prefix.", []models.Interaction{models.EVENT, models.REQUEST}, expectedSelectables))
		t.Run("event_and_request", testDeviceTypeSelectablesWithoutConfigurables(conf, criteria, "prefix.", []models.Interaction{models.EVENT_AND_REQUEST}, []model.DeviceTypeSelectable{}))

		t.Run("nil modified", testDeviceTypeSelectablesWithoutConfigurablesIncludeModified(conf, criteria, "prefix.", nil, expectedSeletablesWithModifiedIds))
		t.Run("empty modified", testDeviceTypeSelectablesWithoutConfigurablesIncludeModified(conf, criteria, "prefix.", []models.Interaction{}, expectedSeletablesWithModifiedIds))
		t.Run("event modified", testDeviceTypeSelectablesWithoutConfigurablesIncludeModified(conf, criteria, "prefix.", []models.Interaction{models.EVENT}, []model.DeviceTypeSelectable{}))
		t.Run("request modified", testDeviceTypeSelectablesWithoutConfigurablesIncludeModified(conf, criteria, "prefix.", []models.Interaction{models.REQUEST}, expectedSeletablesWithModifiedIds))
		t.Run("event+request modified", testDeviceTypeSelectablesWithoutConfigurablesIncludeModified(conf, criteria, "prefix.", []models.Interaction{models.EVENT, models.REQUEST}, expectedSeletablesWithModifiedIds))
		t.Run("event_and_request modified", testDeviceTypeSelectablesWithoutConfigurablesIncludeModified(conf, criteria, "prefix.", []models.Interaction{models.EVENT_AND_REQUEST}, []model.DeviceTypeSelectable{}))

	})

	t.Run("v2", func(t *testing.T) {

		criteria := []model.FilterCriteria{{
			FunctionId: model.MEASURING_FUNCTION_PREFIX + "getPlugState",
			AspectId:   "plug",
		}}
		expectedSelectables := []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "plug-strip",
				Services: []models.Service{
					{
						Id:              "plug1",
						ServiceGroupKey: "sg1",
						Interaction:     models.REQUEST,
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
						Interaction:     models.REQUEST,
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"plug1": {
						{
							ServiceId:        "plug1",
							Path:             "prefix.state",
							CharacteristicId: "plug-state-characteristic",
							AspectNode: models.AspectNode{
								Id:            "plug",
								Name:          "",
								RootId:        "plug",
								ParentId:      "",
								ChildIds:      []string{},
								AncestorIds:   []string{},
								DescendentIds: []string{},
							},
							FunctionId:  model.MEASURING_FUNCTION_PREFIX + "getPlugState",
							Interaction: models.REQUEST,
						},
					},
					"plug2": {
						{
							ServiceId:        "plug2",
							Path:             "prefix.state",
							CharacteristicId: "plug-state-characteristic",
							AspectNode: models.AspectNode{
								Id:            "plug",
								Name:          "",
								RootId:        "plug",
								ParentId:      "",
								ChildIds:      []string{},
								AncestorIds:   []string{},
								DescendentIds: []string{},
							},
							FunctionId:  model.MEASURING_FUNCTION_PREFIX + "getPlugState",
							Interaction: models.REQUEST,
						},
					},
				},
			},
		}

		var expectedSeletablesWithModifiedIds []model.DeviceTypeSelectable = append(expectedSelectables, []model.DeviceTypeSelectable{
			{
				DeviceTypeId: "plug-strip" + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {"sg1"}}),
				Services: []models.Service{
					{
						Id:              "plug1",
						ServiceGroupKey: "sg1",
						Interaction:     models.REQUEST,
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"plug1": {
						{
							ServiceId:        "plug1",
							Path:             "prefix.state",
							CharacteristicId: "plug-state-characteristic",
							AspectNode: models.AspectNode{
								Id:            "plug",
								Name:          "",
								RootId:        "plug",
								ParentId:      "",
								ChildIds:      []string{},
								AncestorIds:   []string{},
								DescendentIds: []string{},
							},
							FunctionId:  model.MEASURING_FUNCTION_PREFIX + "getPlugState",
							Interaction: models.REQUEST,
						},
					},
				},
			},
			{
				DeviceTypeId: "plug-strip" + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {"sg2"}}),
				Services: []models.Service{
					{
						Id:              "plug2",
						ServiceGroupKey: "sg2",
						Interaction:     models.REQUEST,
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
				},
				ServicePathOptions: map[string][]model.ServicePathOption{
					"plug2": {
						{
							ServiceId:        "plug2",
							Path:             "prefix.state",
							CharacteristicId: "plug-state-characteristic",
							AspectNode: models.AspectNode{
								Id:            "plug",
								Name:          "",
								RootId:        "plug",
								ParentId:      "",
								ChildIds:      []string{},
								AncestorIds:   []string{},
								DescendentIds: []string{},
							},
							FunctionId:  model.MEASURING_FUNCTION_PREFIX + "getPlugState",
							Interaction: models.REQUEST,
						},
					},
				},
			},
		}...)

		t.Run("nil", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, criteria, "prefix.", expectedSelectables))
		t.Run("empty", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, criteria, "prefix.", expectedSelectables))
		t.Run("event", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, testAddInteractionToCriterias(criteria, []models.Interaction{models.EVENT}), "prefix.", []model.DeviceTypeSelectable{}))
		t.Run("request", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, testAddInteractionToCriterias(criteria, []models.Interaction{models.REQUEST}), "prefix.", expectedSelectables))
		t.Run("event+request", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, testAddInteractionToCriterias(criteria, []models.Interaction{models.EVENT, models.REQUEST}), "prefix.", []model.DeviceTypeSelectable{}))
		t.Run("event_and_request", testDeviceTypeSelectablesWithoutConfigurablesV2(conf, testAddInteractionToCriterias(criteria, []models.Interaction{models.EVENT_AND_REQUEST}), "prefix.", []model.DeviceTypeSelectable{}))

		t.Run("nil modified", testDeviceTypeSelectablesWithoutConfigurablesV2IncludeModified(conf, criteria, "prefix.", expectedSeletablesWithModifiedIds))
		t.Run("empty modified", testDeviceTypeSelectablesWithoutConfigurablesV2IncludeModified(conf, criteria, "prefix.", expectedSeletablesWithModifiedIds))
		t.Run("event modified", testDeviceTypeSelectablesWithoutConfigurablesV2IncludeModified(conf, testAddInteractionToCriterias(criteria, []models.Interaction{models.EVENT}), "prefix.", []model.DeviceTypeSelectable{}))
		t.Run("request modified", testDeviceTypeSelectablesWithoutConfigurablesV2IncludeModified(conf, testAddInteractionToCriterias(criteria, []models.Interaction{models.REQUEST}), "prefix.", expectedSeletablesWithModifiedIds))
		t.Run("event+request modified", testDeviceTypeSelectablesWithoutConfigurablesV2IncludeModified(conf, testAddInteractionToCriterias(criteria, []models.Interaction{models.EVENT, models.REQUEST}), "prefix.", []model.DeviceTypeSelectable{}))
		t.Run("event_and_request modified", testDeviceTypeSelectablesWithoutConfigurablesV2IncludeModified(conf, testAddInteractionToCriterias(criteria, []models.Interaction{models.EVENT_AND_REQUEST}), "prefix.", []model.DeviceTypeSelectable{}))

	})
}

//go:embed a5t_testcase/devicetypes.json
var nousA5tDeviceTypeStr string

//go:embed a5t_testcase/aspects.json
var aspectsStr string

//go:embed a5t_testcase/functions.json
var functionsStr string

//go:embed a5t_testcase/expected.json
var expectedStr string

//go:embed a5t_testcase/expected_with_modified.json
var expectedWithModifiedStr string

//go:embed a5t_testcase/expected_service_must_match.json
var expectedServiceMustMatchStr string

//go:embed a5t_testcase/expected_service_must_match_short.json
var expectedServiceMustMatchShortStr string

func TestDeviceTypeSelectablesV2WithModifiedIdNousA5T(t *testing.T) {
	if testing.Short() {
		t.Skip("skip in short test run")
		return
	}
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("init metadata", createTestMetadataFromString(conf, nousA5tDeviceTypeStr, aspectsStr, functionsStr))

	time.Sleep(5 * time.Second)

	criteriaStr := `[{"device_class_id":"urn:infai:ses:device-class:79de1bd9-b933-412d-b98e-4cfe19aa3250","function_id":"urn:infai:ses:controlling-function:79e7914b-f303-4a7d-90af-dee70db05fd9","interaction":"request"},{"interaction":"request","function_id":"urn:infai:ses:controlling-function:2f35150b-9df7-4cad-95bc-165fa00219fd","device_class_id":"urn:infai:ses:device-class:79de1bd9-b933-412d-b98e-4cfe19aa3250"},{"function_id":"urn:infai:ses:measuring-function:20d3c1d3-77d7-4181-a9f3-b487add58cd0","aspect_id":"urn:infai:ses:aspect:861227f6-1523-46a7-b8ab-a4e76f0bdd32"},{"function_id":"urn:infai:ses:measuring-function:1c7c90fb-73b6-4690-aac2-72e9735e68d0","aspect_id":"urn:infai:ses:aspect:74a7b913-73ac-42b7-9b35-573f2c1e97cf","interaction":"event"}]`
	criteria := []model.FilterCriteria{}
	err = json.Unmarshal([]byte(criteriaStr), &criteria)
	if err != nil {
		t.Error(err)
		return
	}

	expectedSelectables := []model.DeviceTypeSelectable{}
	err = json.Unmarshal([]byte(expectedStr), &expectedSelectables)
	if err != nil {
		t.Error(err)
		return
	}

	expectedSeletablesWithModifiedIds := []model.DeviceTypeSelectable{}
	err = json.Unmarshal([]byte(expectedWithModifiedStr), &expectedSeletablesWithModifiedIds)
	if err != nil {
		t.Error(err)
		return
	}

	expectedSeletablesServiceMustMatch := []model.DeviceTypeSelectable{}
	err = json.Unmarshal([]byte(expectedServiceMustMatchStr), &expectedSeletablesServiceMustMatch)
	if err != nil {
		t.Error(err)
		return
	}

	expectedSeletablesServiceMustMatchShort := []model.DeviceTypeSelectable{}
	err = json.Unmarshal([]byte(expectedServiceMustMatchShortStr), &expectedSeletablesServiceMustMatchShort)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("without modified", clientTestDeviceTypeSelectables(conf, criteria, "prefix.", false, false, expectedSelectables))
	t.Run("with modified", clientTestDeviceTypeSelectables(conf, criteria, "prefix.", true, false, expectedSeletablesWithModifiedIds))
	t.Run("servicesMustMatchAllCriteria ", clientTestDeviceTypeSelectables(conf, criteria, "prefix.", true, true, expectedSeletablesServiceMustMatch))
	t.Run("one criteria servicesMustMatchAllCriteria ", clientTestDeviceTypeSelectables(conf, []model.FilterCriteria{{
		Interaction:   "request",
		FunctionId:    "urn:infai:ses:controlling-function:79e7914b-f303-4a7d-90af-dee70db05fd9",
		DeviceClassId: "urn:infai:ses:device-class:79de1bd9-b933-412d-b98e-4cfe19aa3250",
	}}, "prefix.", true, true, expectedSeletablesServiceMustMatchShort))

}

func TestDeviceTypeFilterWithModifiedId(t *testing.T) {
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

	dt := models.DeviceType{
		Id:            "plug-strip",
		Name:          "dt",
		DeviceClassId: "toggle",
		ServiceGroups: []models.ServiceGroup{{Key: "sg1", Name: "sg1"}, {Key: "sg2", Name: "sg2"}},
		Services: []models.Service{
			{
				Id:              "plug1",
				ServiceGroupKey: "sg1",
				Interaction:     models.REQUEST,
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
				Interaction:     models.REQUEST,
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
				Interaction: models.REQUEST,
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
	}
	dtSg1 := models.DeviceType{
		Id:            "plug-strip" + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {"sg1"}}),
		DeviceClassId: "toggle",
		Name:          "dt sg1",
		ServiceGroups: []models.ServiceGroup{{Key: "sg1", Name: "sg1"}, {Key: "sg2", Name: "sg2"}},
		Services: []models.Service{
			{
				Id:              "plug1",
				ServiceGroupKey: "sg1",
				Interaction:     models.REQUEST,
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
				Id:          "plugs",
				Interaction: models.REQUEST,
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
	}
	dtSg2 := models.DeviceType{
		Id:            "plug-strip" + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {"sg2"}}),
		DeviceClassId: "toggle",
		Name:          "dt sg2",
		ServiceGroups: []models.ServiceGroup{{Key: "sg1", Name: "sg1"}, {Key: "sg2", Name: "sg2"}},
		Services: []models.Service{
			{
				Id:              "plug2",
				ServiceGroupKey: "sg2",
				Interaction:     models.REQUEST,
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
				Interaction: models.REQUEST,
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
	}

	criteria := []model.FilterCriteria{{
		FunctionId: model.MEASURING_FUNCTION_PREFIX + "getPlugState",
		AspectId:   "plug",
	}}
	criteriaJson, err := json.Marshal(criteria)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("without modify", testGetRequest(testenv.Userjwt, conf, "/device-types?filter="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dt}))
	t.Run("with modify", testGetRequest(testenv.Userjwt, conf, "/device-types?interactions-filter=request&include_id_modified=true&filter="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dt, dtSg1, dtSg2}))
	t.Run("with modify v2", testGetRequest(testenv.Userjwt, conf, "/device-types?include_id_modified=true&filter="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dt, dtSg1, dtSg2}))

	t.Run("modified only", testGetRequest(testenv.Userjwt, conf, "/device-types?include_id_modified=true&include_id_unmodified=false&filter="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dtSg1, dtSg2}))
	t.Run("unfiltered modified only", testGetRequest(testenv.Userjwt, conf, "/device-types?include_id_modified=true&include_id_unmodified=false", []models.DeviceType{dtSg1, dtSg2}))

	t.Run("sort name asc", testGetRequest(testenv.Userjwt, conf, "/device-types?sort=name.asc&interactions-filter=request&include_id_modified=true&filter="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dt, dtSg1, dtSg2}))
	t.Run("sort name desc", testGetRequest(testenv.Userjwt, conf, "/device-types?sort=name.desc&interactions-filter=request&include_id_modified=true&filter="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dtSg2, dtSg1, dt}))

	t.Run("v3 without modify", testGetRequest(testenv.Userjwt, conf, "/v3/device-types?criteria="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dt}))
	t.Run("v3 with modify v2", testGetRequest(testenv.Userjwt, conf, "/v3/device-types?include-modified=true&criteria="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dt, dtSg1, dtSg2}))

	t.Run("v3 modified only", testGetRequest(testenv.Userjwt, conf, "/v3/device-types?include-modified=true&ignore-unmodified=true&criteria="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dtSg1, dtSg2}))
	t.Run("v3 unfiltered modified only", testGetRequest(testenv.Userjwt, conf, "/v3/device-types?include-modified=true&ignore-unmodified=true", []models.DeviceType{dtSg1, dtSg2}))

	t.Run("v3 sort name asc", testGetRequest(testenv.Userjwt, conf, "/v3/device-types?sort=name.asc&include-modified=true&criteria="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dt, dtSg1, dtSg2}))
	t.Run("v3 sort name desc", testGetRequest(testenv.Userjwt, conf, "/v3/device-types?sort=name.desc&include-modified=true&criteria="+url.QueryEscape(string(criteriaJson)), []models.DeviceType{dtSg2, dtSg1, dt}))

}

func GetDeviceTypeSelectablesIncludeModified(config config.Config, token string, prefix string, interactionsFilter []models.Interaction, descriptions []model.FilterCriteria) (result []model.DeviceTypeSelectable, err error) {
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
		"http://localhost:"+config.ServerPort+"/query/device-type-selectables?include_id_modified=true&path-prefix="+url.QueryEscape(prefix)+interactionsQuery,
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

func GetDeviceTypeSelectablesV2IncludeModified(config config.Config, token string, prefix string, descriptions []model.FilterCriteria) (result []model.DeviceTypeSelectable, err error) {
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
		"http://localhost:"+config.ServerPort+"/v2/query/device-type-selectables?include_id_modified=true&path-prefix="+url.QueryEscape(prefix),
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

func testDeviceTypeSelectablesWithoutConfigurablesIncludeModified(config config.Config, criteria []model.FilterCriteria, pathPrefix string, interactionsFilter []models.Interaction, expectedResult []model.DeviceTypeSelectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := GetDeviceTypeSelectablesIncludeModified(config, testenv.Userjwt, pathPrefix, interactionsFilter, criteria)
		if err != nil {
			t.Error(err)
			return
		}
		expectedResult = removeConfigurables(expectedResult)
		expectedResult = sortServices(expectedResult)
		result = removeConfigurables(result)
		result = sortServices(result)
		sort.Slice(result, func(i, j int) bool {
			return result[i].DeviceTypeId < result[j].DeviceTypeId
		})
		sort.Slice(expectedResult, func(i, j int) bool {
			return expectedResult[i].DeviceTypeId < expectedResult[j].DeviceTypeId
		})
		if !reflect.DeepEqual(result, expectedResult) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expectedResult)
			t.Error("\n", string(resultJson), "\n", string(expectedJson))
		}
	}
}

func testDeviceTypeSelectablesWithoutConfigurablesV2IncludeModified(config config.Config, criteria []model.FilterCriteria, pathPrefix string, expectedResult []model.DeviceTypeSelectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err := GetDeviceTypeSelectablesV2IncludeModified(config, testenv.Userjwt, pathPrefix, criteria)
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

func clientTestDeviceTypeSelectables(config config.Config, criteria []model.FilterCriteria, pathPrefix string, includeModified bool, servicesMustMatchAllCriteria bool, expectedResult []model.DeviceTypeSelectable) func(t *testing.T) {
	return func(t *testing.T) {
		result, err, _ := client.NewClient("http://localhost:"+config.ServerPort, nil).GetDeviceTypeSelectablesV2(criteria, pathPrefix, includeModified, servicesMustMatchAllCriteria)
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

func testGetRequest(token string, conf config.Config, path string, expected interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://localhost:"+conf.ServerPort+path, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error("unexpected response", path, resp.Status, resp.StatusCode, string(b))
			return
		}

		var result interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
		}
		expectedNormalized := normalize(expected)
		if !reflect.DeepEqual(expectedNormalized, result) {
			eJson, _ := json.Marshal(expectedNormalized)
			rJson, _ := json.Marshal(result)
			t.Error("unexpected result", expectedNormalized, result, "\n", string(eJson), "\n", string(rJson))
			return
		}
	}
}

func normalize(expected interface{}) (result interface{}) {
	temp, _ := json.Marshal(expected)
	json.Unmarshal(temp, &result)
	return
}
