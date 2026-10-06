/*
 * Copyright 2026 InfAI (CC SES)
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
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib"
	"github.com/SENERGY-Platform/device-repository/v3/lib/client"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

// A controlling function is combined with aspects the way a measuring function is, and keeps
// its combination with the device-class next to that.
func TestControllingFunctionAspects(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.SyncLockDuration = time.Second.String()
	config.Debug = true

	config, err = docker.NewEnv(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	c := client.NewClient("http://localhost:"+config.ServerPort, nil)

	const (
		setTemperature = models.URN_PREFIX + "controlling-function:setTemperature"
		setOn          = models.URN_PREFIX + "controlling-function:setOn"
		getTemperature = models.URN_PREFIX + "measuring-function:getTemperature"
		airAspect      = models.URN_PREFIX + "aspect:air"
		insideAir      = models.URN_PREFIX + "aspect:inside_air"
		waterAspect    = models.URN_PREFIX + "aspect:water"
		deviceClass    = models.URN_PREFIX + "device-class:thermostat"
	)

	t.Run("create resources", func(t *testing.T) {
		_, err, _ := c.SetAspect(client.InternalAdminToken, models.Aspect{
			Id:         airAspect,
			Name:       "air",
			SubAspects: []models.Aspect{{Id: insideAir, Name: "inside_air"}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetAspect(client.InternalAdminToken, models.Aspect{Id: waterAspect, Name: "water"})
		if err != nil {
			t.Error(err)
			return
		}
		for _, f := range []models.Function{
			{Id: setTemperature, Name: "setTemperature", DisplayName: "setTemperature", RdfType: models.SES_ONTOLOGY_CONTROLLING_FUNCTION},
			{Id: setOn, Name: "setOn", DisplayName: "setOn", RdfType: models.SES_ONTOLOGY_CONTROLLING_FUNCTION},
			{Id: getTemperature, Name: "getTemperature", DisplayName: "getTemperature", RdfType: models.SES_ONTOLOGY_MEASURING_FUNCTION},
		} {
			_, err, _ = c.SetFunction(client.InternalAdminToken, f)
			if err != nil {
				t.Error(err)
				return
			}
		}
		_, err, _ = c.SetDeviceClass(client.InternalAdminToken, models.DeviceClass{Id: deviceClass, Name: "thermostat"})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetProtocol(client.InternalAdminToken, models.Protocol{
			Id:               "p1",
			Name:             "p1",
			Handler:          "p1",
			ProtocolSegments: []models.ProtocolSegment{{Id: "ps1", Name: "ps1"}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		input := func(id string, functionId string, aspectIds []string) []models.Content {
			return []models.Content{{
				Id:                id + "_c",
				Serialization:     models.JSON,
				ProtocolSegmentId: "ps1",
				ContentVariable: models.ContentVariable{
					Id:         id + "_cv",
					Name:       "value",
					Type:       models.String,
					FunctionId: functionId,
					AspectIds:  aspectIds,
				},
			}}
		}
		_, err, _ = c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Id:            "dt1",
			Name:          "dt1",
			DeviceClassId: deviceClass,
			Services: []models.Service{
				{
					Id:          "dt1_set_temperature",
					LocalId:     "set_temperature",
					Name:        "set_temperature",
					Interaction: models.REQUEST,
					ProtocolId:  "p1",
					Inputs:      input("dt1_set_temperature", setTemperature, []string{insideAir}),
				},
				{
					Id:          "dt1_set_on",
					LocalId:     "set_on",
					Name:        "set_on",
					Interaction: models.REQUEST,
					ProtocolId:  "p1",
					Inputs:      input("dt1_set_on", setOn, nil),
				},
				{
					Id:          "dt1_get_temperature",
					LocalId:     "get_temperature",
					Name:        "get_temperature",
					Interaction: models.EVENT,
					ProtocolId:  "p1",
					Outputs:     input("dt1_get_temperature", getTemperature, []string{waterAspect}),
				},
			},
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
	})

	nodeIds := func(nodes []models.AspectNode) []string {
		result := []string{}
		for _, node := range nodes {
			result = append(result, node.Id)
		}
		return slices.Sorted(slices.Values(result))
	}
	functionIds := func(functions []models.Function) []string {
		result := []string{}
		for _, f := range functions {
			result = append(result, f.Id)
		}
		return slices.Sorted(slices.Values(result))
	}

	t.Run("aspect-nodes with controlling function", func(t *testing.T) {
		nodes, err, _ := c.GetAspectNodesWithControllingFunction(false, false)
		if err != nil {
			t.Error(err)
			return
		}
		if got := nodeIds(nodes); !slices.Equal(got, []string{insideAir}) {
			t.Error("unexpected aspect-nodes", got)
		}
		//descendants=true adds the nodes that have a used node among their descendants
		nodes, err, _ = c.GetAspectNodesWithControllingFunction(false, true)
		if err != nil {
			t.Error(err)
			return
		}
		if got := nodeIds(nodes); !slices.Equal(got, []string{airAspect, insideAir}) {
			t.Error("unexpected aspect-nodes with descendants", got)
		}
	})

	t.Run("aspect-nodes with measuring function are unchanged", func(t *testing.T) {
		nodes, err, _ := c.GetAspectNodesWithMeasuringFunction(false, false)
		if err != nil {
			t.Error(err)
			return
		}
		if got := nodeIds(nodes); !slices.Equal(got, []string{waterAspect}) {
			t.Error("unexpected aspect-nodes", got)
		}
	})

	t.Run("aspects with controlling function", func(t *testing.T) {
		aspects, err, _ := c.GetAspectsWithControllingFunction(false, true)
		if err != nil {
			t.Error(err)
			return
		}
		if len(aspects) != 1 || aspects[0].Id != airAspect {
			t.Error("unexpected aspects", aspects)
		}
	})

	t.Run("controlling functions of an aspect-node", func(t *testing.T) {
		functions, err, _ := c.GetAspectNodesControllingFunctions(insideAir, false, false)
		if err != nil {
			t.Error(err)
			return
		}
		if got := functionIds(functions); !slices.Equal(got, []string{setTemperature}) {
			t.Error("unexpected functions", got)
		}
		functions, err, _ = c.GetAspectNodesControllingFunctions(airAspect, false, false)
		if err != nil {
			t.Error(err)
			return
		}
		if got := functionIds(functions); len(got) != 0 {
			t.Error("the parent alone should not find a function of its child", got)
		}
		functions, err, _ = c.GetAspectNodesControllingFunctions(airAspect, false, true)
		if err != nil {
			t.Error(err)
			return
		}
		if got := functionIds(functions); !slices.Equal(got, []string{setTemperature}) {
			t.Error("unexpected functions with descendants", got)
		}
		functions, err, _ = c.GetAspectNodesMeasuringFunctions(insideAir, false, true)
		if err != nil {
			t.Error(err)
			return
		}
		if got := functionIds(functions); len(got) != 0 {
			t.Error("measuring listing returned a controlling function", got)
		}
	})

	t.Run("generated device-group criteria", func(t *testing.T) {
		device, err, _ := c.CreateDevice(client.InternalAdminToken, models.Device{
			LocalId:      "d1",
			Name:         "d1",
			DeviceTypeId: "dt1",
		})
		if err != nil {
			t.Error(err)
			return
		}
		group, err, _ := c.ReadDeviceGroup(model.DeviceIdToGeneratedDeviceGroupId(device.Id), client.InternalAdminToken, false)
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.DeviceGroupFilterCriteria{
			{FunctionId: setTemperature, DeviceClassId: deviceClass, Interaction: models.REQUEST},
			{FunctionId: setTemperature, AspectIds: []string{insideAir}, Interaction: models.REQUEST},
			{FunctionId: setTemperature, AspectIds: []string{airAspect}, Interaction: models.REQUEST},
			{FunctionId: setOn, DeviceClassId: deviceClass, Interaction: models.REQUEST},
			{FunctionId: getTemperature, AspectIds: []string{waterAspect}, Interaction: models.EVENT},
		}
		expectedShort := []string{}
		for _, criteria := range expected {
			expectedShort = append(expectedShort, criteria.Short())
		}
		slices.Sort(expectedShort)
		actualShort := slices.Sorted(slices.Values(group.CriteriaShort))
		if !slices.Equal(actualShort, expectedShort) {
			t.Errorf("unexpected criteria\n%#v\n%#v", actualShort, expectedShort)
		}
	})
}
