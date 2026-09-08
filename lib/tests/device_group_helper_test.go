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
	"encoding/json"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v2/lib"
	"github.com/SENERGY-Platform/device-repository/v2/lib/client"
	"github.com/SENERGY-Platform/device-repository/v2/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v2/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/v2/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

// TestDeviceGroupHelper covers POST /device-group-helper: the criteria a device-group of the
// given devices would have, and the devices that may still be added to it.
//
// The criteria of a group are the intersection of the criteria of its devices by Short(). A
// function reaches the group only if every device answers it; an aspect only if every device
// carries it or a descendant of it, which is why a criteria over an aspect list is kept next
// to the criteria over the single aspects.
func TestDeviceGroupHelper(t *testing.T) {
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
		getTemperature = models.URN_PREFIX + "measuring-function:getTemperature"
		getHumidity    = models.URN_PREFIX + "measuring-function:getHumidity"
		setState       = models.URN_PREFIX + "controlling-function:setState"
		deviceClass    = models.URN_PREFIX + "device-class:thermostat"
	)

	//flat aspects without a hierarchy, so that no ancestor criteria join the result
	const (
		aspectA = models.URN_PREFIX + "aspect:a"
		aspectB = models.URN_PREFIX + "aspect:b"
		aspectC = models.URN_PREFIX + "aspect:c"
		aspectD = models.URN_PREFIX + "aspect:d"
	)

	//and a separate hierarchy for the cases below: q descends from p, r stands beside it.
	//These are their own aspects, so that the flat cases above are unaffected by the hierarchy.
	const (
		aspectP = models.URN_PREFIX + "aspect:p"
		aspectQ = models.URN_PREFIX + "aspect:q"
		aspectR = models.URN_PREFIX + "aspect:r"
	)

	const serviceGroupOne = "sg1"
	const serviceGroupTwo = "sg2"

	t.Run("create metadata", func(t *testing.T) {
		aspects := []models.Aspect{
			{Id: aspectA, Name: "a"},
			{Id: aspectB, Name: "b"},
			{Id: aspectC, Name: "c"},
			{Id: aspectD, Name: "d"},
			{Id: aspectP, Name: "p", SubAspects: []models.Aspect{{Id: aspectQ, Name: "q"}}},
			{Id: aspectR, Name: "r"},
		}
		for _, aspect := range aspects {
			_, err, _ := c.SetAspect(client.InternalAdminToken, aspect)
			if err != nil {
				t.Error(err)
				return
			}
		}
		_, err, _ := c.SetFunction(client.InternalAdminToken, models.Function{
			Id:          getTemperature,
			Name:        "getTemperature",
			DisplayName: "getTemperature",
			RdfType:     models.SES_ONTOLOGY_MEASURING_FUNCTION,
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetFunction(client.InternalAdminToken, models.Function{
			Id:          getHumidity,
			Name:        "getHumidity",
			DisplayName: "getHumidity",
			RdfType:     models.SES_ONTOLOGY_MEASURING_FUNCTION,
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetFunction(client.InternalAdminToken, models.Function{
			Id:          setState,
			Name:        "setState",
			DisplayName: "setState",
			RdfType:     models.SES_ONTOLOGY_CONTROLLING_FUNCTION,
		})
		if err != nil {
			t.Error(err)
			return
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
	})

	//measuringService answers one measuring function over one content variable carrying all the
	//aspects
	measuringService := func(deviceTypeId string, suffix string, serviceGroupKey string, functionId string, aspectIds []string) models.Service {
		return models.Service{
			Id:              deviceTypeId + "_s_" + suffix,
			LocalId:         deviceTypeId + "_s_" + suffix,
			Name:            "s_" + suffix,
			Interaction:     models.REQUEST,
			ProtocolId:      "p1",
			ServiceGroupKey: serviceGroupKey,
			Outputs: []models.Content{{
				Id:                deviceTypeId + "_c_" + suffix,
				Serialization:     models.JSON,
				ProtocolSegmentId: "ps1",
				ContentVariable: models.ContentVariable{
					Id:         deviceTypeId + "_cv_" + suffix,
					Name:       "measured_" + suffix,
					Type:       models.Float,
					FunctionId: functionId,
					AspectIds:  aspectIds,
				},
			}},
		}
	}

	//one device-type per aspect set, each with a single service measuring the same function
	measuringDeviceTypeAspects := map[string][]string{
		"dt_ab":       {aspectA, aspectB},
		"dt_bc":       {aspectB, aspectC},
		"dt_ab_again": {aspectA, aspectB},
		"dt_d":        {aspectD},
		"dt_pr":       {aspectP, aspectR},
		"dt_qr":       {aspectQ, aspectR},
	}

	//dt_sg carries the same function on two service-groups, so it has a modified variant per
	//group: dt_sg$service_group_selection=sg1 measures b, sg2 measures d
	serviceGroupDeviceType := models.DeviceType{
		Id:            "dt_sg",
		Name:          "dt_sg",
		DeviceClassId: deviceClass,
		ServiceGroups: []models.ServiceGroup{
			{Key: serviceGroupOne, Name: "one"},
			{Key: serviceGroupTwo, Name: "two"},
		},
		Services: []models.Service{
			measuringService("dt_sg", "one", serviceGroupOne, getTemperature, []string{aspectB}),
			measuringService("dt_sg", "two", serviceGroupTwo, getTemperature, []string{aspectD}),
		},
	}

	//dt_two_functions answers both measuring functions over [a b], dt_humidity only the second.
	//A group needs criteria over two functions before a block list can tell them apart.
	twoFunctionDeviceType := models.DeviceType{
		Id:            "dt_two_functions",
		Name:          "dt_two_functions",
		DeviceClassId: deviceClass,
		Services: []models.Service{
			measuringService("dt_two_functions", "temp", "", getTemperature, []string{aspectA, aspectB}),
			measuringService("dt_two_functions", "hum", "", getHumidity, []string{aspectA, aspectB}),
		},
	}

	humidityDeviceType := models.DeviceType{
		Id:            "dt_humidity",
		Name:          "dt_humidity",
		DeviceClassId: deviceClass,
		Services: []models.Service{
			measuringService("dt_humidity", "hum", "", getHumidity, []string{aspectA, aspectB}),
		},
	}

	//dt_control answers a controlling function, whose criteria carry the device-class instead
	//of an aspect
	controllingDeviceType := models.DeviceType{
		Id:            "dt_control",
		Name:          "dt_control",
		DeviceClassId: deviceClass,
		Services: []models.Service{{
			Id:          "dt_control_s",
			LocalId:     "dt_control_s",
			Name:        "s",
			Interaction: models.REQUEST,
			ProtocolId:  "p1",
			Inputs: []models.Content{{
				Id:                "dt_control_c",
				Serialization:     models.JSON,
				ProtocolSegmentId: "ps1",
				ContentVariable: models.ContentVariable{
					Id:         "dt_control_cv",
					Name:       "state",
					Type:       models.Boolean,
					FunctionId: setState,
				},
			}},
		}},
	}

	deviceIdByDeviceTypeId := map[string]string{}

	t.Run("create device-types and devices", func(t *testing.T) {
		deviceTypes := []models.DeviceType{serviceGroupDeviceType, controllingDeviceType, twoFunctionDeviceType, humidityDeviceType}
		for _, deviceTypeId := range slices.Sorted(maps.Keys(measuringDeviceTypeAspects)) {
			deviceTypes = append(deviceTypes, models.DeviceType{
				Id:            deviceTypeId,
				Name:          deviceTypeId,
				DeviceClassId: deviceClass,
				Services:      []models.Service{measuringService(deviceTypeId, "m", "", getTemperature, measuringDeviceTypeAspects[deviceTypeId])},
			})
		}
		for _, deviceType := range deviceTypes {
			_, err, _ := c.SetDeviceType(client.InternalAdminToken, deviceType, client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(deviceType.Id, err)
				return
			}
			deviceId := strings.Replace(deviceType.Id, "dt_", "device_", 1)
			_, err, _ = c.SetDevice(client.InternalAdminToken, models.Device{
				Id:           deviceId,
				LocalId:      deviceId,
				Name:         deviceId,
				DeviceTypeId: deviceType.Id,
			}, client.DeviceUpdateOptions{})
			if err != nil {
				t.Error(deviceType.Id, err)
				return
			}
			deviceIdByDeviceTypeId[deviceType.Id] = deviceId
		}
	})

	measuringFn := func(functionId string, aspectIds ...string) models.DeviceGroupFilterCriteria {
		sorted := slices.Sorted(slices.Values(aspectIds))
		return models.DeviceGroupFilterCriteria{
			FunctionId:  functionId,
			AspectId:    sorted[0], //the alphabetically first, which is the deprecated alias
			AspectIds:   sorted,
			Interaction: models.REQUEST,
		}
	}

	measuring := func(aspectIds ...string) models.DeviceGroupFilterCriteria {
		return measuringFn(getTemperature, aspectIds...)
	}

	controlling := models.DeviceGroupFilterCriteria{
		FunctionId:    setState,
		DeviceClassId: deviceClass,
		Interaction:   models.REQUEST,
	}

	//a single device keeps the criteria over its whole aspect list next to the single aspects
	t.Run("criteria: one device with two aspects", testDeviceGroupHelperCriteria(c, []string{"device_ab"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectA, aspectB),
		measuring(aspectA),
		measuring(aspectB),
	}))

	t.Run("criteria: one device with one aspect", testDeviceGroupHelperCriteria(c, []string{"device_d"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectD),
	}))

	//the intersection case: [a b] and [b c] leave b, because b is the only aspect that stands
	//as a criteria of its own on both sides
	t.Run("criteria: intersection keeps only the common aspect", testDeviceGroupHelperCriteria(c, []string{"device_ab", "device_bc"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectB),
	}))

	t.Run("criteria: device order does not matter", testDeviceGroupHelperCriteria(c, []string{"device_bc", "device_ab"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectB),
	}))

	//two devices carrying the same pair keep the pair, because the list criteria survives too
	t.Run("criteria: the same aspect pair keeps the pair", testDeviceGroupHelperCriteria(c, []string{"device_ab", "device_ab_again"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectA, aspectB),
		measuring(aspectA),
		measuring(aspectB),
	}))

	//the aspect list of a variable is not an alternative: [a b] does not cover a group member
	//that only measures d, so the group is left without a criteria for the shared function
	t.Run("criteria: no common aspect", testDeviceGroupHelperCriteria(c, []string{"device_ab", "device_d"}, []models.DeviceGroupFilterCriteria{}))

	t.Run("criteria: three devices narrow to the aspect all of them carry", testDeviceGroupHelperCriteria(c, []string{"device_ab", "device_bc", "device_ab_again"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectB),
	}))

	//An aspect criteria covers the subtree of its node, so a variable carrying [q r] is found
	//by a query over [p r] as well. The criteria of that device therefore hold [p r] next to
	//[q r], and the ancestor of a single aspect stands on its own the way it always did.
	t.Run("criteria: one device with a descendant aspect", testDeviceGroupHelperCriteria(c, []string{"device_qr"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectQ, aspectR),
		measuring(aspectP, aspectR),
		measuring(aspectQ),
		measuring(aspectP),
		measuring(aspectR),
	}))

	t.Run("criteria: one device with the ancestor aspect", testDeviceGroupHelperCriteria(c, []string{"device_pr"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectP, aspectR),
		measuring(aspectP),
		measuring(aspectR),
	}))

	//the case a hierarchy adds: a device carrying [p r] and one carrying [q r] keep [p r],
	//because p covers q. Without the ancestor of the list, the group would fall back to the
	//single p and r and lose that one variable carries both.
	t.Run("criteria: intersection over a descendant keeps the ancestor pair", testDeviceGroupHelperCriteria(c, []string{"device_pr", "device_qr"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectP, aspectR),
		measuring(aspectP),
		measuring(aspectR),
	}))

	t.Run("criteria: intersection over a descendant, device order reversed", testDeviceGroupHelperCriteria(c, []string{"device_qr", "device_pr"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectP, aspectR),
		measuring(aspectP),
		measuring(aspectR),
	}))

	//a controlling function has no aspects; its criteria carries the device-class
	t.Run("criteria: a controlling function carries the device-class", testDeviceGroupHelperCriteria(c, []string{"device_control"}, []models.DeviceGroupFilterCriteria{
		controlling,
	}))

	t.Run("criteria: no devices, no criteria", testDeviceGroupHelperCriteria(c, []string{}, []models.DeviceGroupFilterCriteria{}))

	//the modified variants of dt_sg answer only the service of their service-group
	modifiedId := func(deviceId string, serviceGroupKey string) string {
		return deviceId + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{
			"service_group_selection": {serviceGroupKey},
		})
	}
	deviceSgOne := modifiedId("device_sg", serviceGroupOne)
	deviceSgTwo := modifiedId("device_sg", serviceGroupTwo)

	t.Run("criteria: the unmodified device answers both service-groups", testDeviceGroupHelperCriteria(c, []string{"device_sg"}, []models.DeviceGroupFilterCriteria{
		measuring(aspectB),
		measuring(aspectD),
	}))

	t.Run("criteria: a modified device answers only its service-group", testDeviceGroupHelperCriteria(c, []string{deviceSgOne}, []models.DeviceGroupFilterCriteria{
		measuring(aspectB),
	}))

	t.Run("criteria: a modified device in a group", testDeviceGroupHelperCriteria(c, []string{"device_ab", deviceSgOne}, []models.DeviceGroupFilterCriteria{
		measuring(aspectB),
	}))

	t.Run("a device the requesting user may not read is refused", func(t *testing.T) {
		//every device of this test belongs to the admin, so the plain user token reaches none
		const userjwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`
		_, err, code := c.DeviceGroupHelper(userjwt, []string{"device_ab"}, client.DeviceGroupHelperOptions{})
		if err == nil {
			t.Error("expected the request to be refused")
			return
		}
		if code != http.StatusForbidden {
			t.Error("expected 403, got", code, err)
		}
	})

	t.Run("options", func(t *testing.T) {
		result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, []string{"device_ab"}, client.DeviceGroupHelperOptions{})
		if err != nil {
			t.Error(err, code)
			return
		}
		options := map[string]client.DeviceGroupOption{}
		for _, option := range result.Options {
			options[option.Device.Id] = option
		}

		t.Run("a device of the group is no option", func(t *testing.T) {
			if _, ok := options["device_ab"]; ok {
				t.Error("expected the group member to be left out", optionIds(result.Options))
			}
		})

		t.Run("an unrelated device is an option that removes every criteria", func(t *testing.T) {
			option, ok := options["device_d"]
			if !ok {
				t.Error("expected device_d to be listed", optionIds(result.Options))
				return
			}
			//device_d measures d only, so none of the three criteria of the group survive
			assertSameCriteria(t, option.RemovesCriteria, []models.DeviceGroupFilterCriteria{
				measuring(aspectA, aspectB),
				measuring(aspectA),
				measuring(aspectB),
			})
			if option.MaintainsGroupUsability {
				t.Error("expected device_d to leave the group without a criteria")
			}
		})

		t.Run("a device sharing an aspect keeps the group usable", func(t *testing.T) {
			option, ok := options["device_bc"]
			if !ok {
				t.Error("expected device_bc to be listed", optionIds(result.Options))
				return
			}
			//b survives, the criteria naming a do not
			assertSameCriteria(t, option.RemovesCriteria, []models.DeviceGroupFilterCriteria{
				measuring(aspectA, aspectB),
				measuring(aspectA),
			})
			if !option.MaintainsGroupUsability {
				t.Error("expected device_bc to keep the criteria over b")
			}
		})

		t.Run("a device of the same device-type removes nothing", func(t *testing.T) {
			option, ok := options["device_ab_again"]
			if !ok {
				t.Error("expected device_ab_again to be listed", optionIds(result.Options))
				return
			}
			if len(option.RemovesCriteria) != 0 {
				t.Error("expected no criteria to fall away", option.RemovesCriteria)
			}
			if !option.MaintainsGroupUsability {
				t.Error("expected device_ab_again to keep the group usable")
			}
		})

		t.Run("a controlling device removes every measuring criteria", func(t *testing.T) {
			option, ok := options["device_control"]
			if !ok {
				t.Error("expected device_control to be listed", optionIds(result.Options))
				return
			}
			if option.MaintainsGroupUsability {
				t.Error("expected device_control to leave the group without a criteria")
			}
		})

		t.Run("modified devices are options too", func(t *testing.T) {
			one, ok := options[deviceSgOne]
			if !ok {
				t.Error("expected the modified device of service-group one to be listed", optionIds(result.Options))
				return
			}
			//it measures b, so the criteria over b survives
			if !one.MaintainsGroupUsability {
				t.Error("expected the modified device measuring b to keep the group usable")
			}
			two, ok := options[deviceSgTwo]
			if !ok {
				t.Error("expected the modified device of service-group two to be listed", optionIds(result.Options))
				return
			}
			//it measures d only
			if two.MaintainsGroupUsability {
				t.Error("expected the modified device measuring d to leave the group without a criteria")
			}
		})

		t.Run("the unmodified device behind the modified ones is an option as well", func(t *testing.T) {
			if _, ok := options["device_sg"]; !ok {
				t.Error("expected device_sg to be listed", optionIds(result.Options))
			}
		})
	})

	t.Run("options: maintains_group_usability drops the unusable ones", func(t *testing.T) {
		result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, []string{"device_ab"}, client.DeviceGroupHelperOptions{
			MaintainsGroupUsability: true,
		})
		if err != nil {
			t.Error(err, code)
			return
		}
		ids := optionIds(result.Options)
		expected := []string{"device_ab_again", "device_bc", "device_sg", "device_two_functions", deviceSgOne}
		slices.Sort(expected)
		if !slices.Equal(ids, expected) {
			t.Error("\na=", ids, "\ne=", expected)
		}
		for _, option := range result.Options {
			if !option.MaintainsGroupUsability {
				t.Error("expected every listed option to keep the group usable", option.Device.Id)
			}
		}
	})

	//A group over two functions is what a block list can act on. The caller says which functions
	//are too common to make a device a meaningful member; a device answering only those adds
	//nothing, even though it technically leaves a criteria standing.
	t.Run("function_block_list", func(t *testing.T) {
		group := []string{"device_two_functions"}

		t.Run("the group carries both functions", testDeviceGroupHelperCriteria(c, group, []models.DeviceGroupFilterCriteria{
			measuringFn(getTemperature, aspectA, aspectB),
			measuringFn(getTemperature, aspectA),
			measuringFn(getTemperature, aspectB),
			measuringFn(getHumidity, aspectA, aspectB),
			measuringFn(getHumidity, aspectA),
			measuringFn(getHumidity, aspectB),
		}))

		t.Run("without a block list both functions count", func(t *testing.T) {
			result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, group, client.DeviceGroupHelperOptions{})
			if err != nil {
				t.Error(err, code)
				return
			}
			options := optionsById(result.Options)
			//device_ab answers getTemperature only, so the getHumidity criteria fall away
			if !options["device_ab"].MaintainsGroupUsability {
				t.Error("expected device_ab to keep the group usable")
			}
			if !options["device_humidity"].MaintainsGroupUsability {
				t.Error("expected device_humidity to keep the group usable")
			}
		})

		t.Run("a blocked function stops counting", func(t *testing.T) {
			result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, group, client.DeviceGroupHelperOptions{
				FunctionBlockList: []string{getTemperature},
			})
			if err != nil {
				t.Error(err, code)
				return
			}
			options := optionsById(result.Options)

			//device_ab now keeps only blocked criteria, so it no longer counts as usable
			deviceAb, ok := options["device_ab"]
			if !ok {
				t.Error("expected device_ab to be listed, only flagged differently", optionIds(result.Options))
				return
			}
			if deviceAb.MaintainsGroupUsability {
				t.Error("expected device_ab to stop counting once getTemperature is blocked")
			}
			//RemovesCriteria reports what actually falls away, blocked or not
			assertSameCriteria(t, deviceAb.RemovesCriteria, []models.DeviceGroupFilterCriteria{
				measuringFn(getHumidity, aspectA, aspectB),
				measuringFn(getHumidity, aspectA),
				measuringFn(getHumidity, aspectB),
			})

			//device_humidity keeps the criteria of the function that still counts
			if !options["device_humidity"].MaintainsGroupUsability {
				t.Error("expected device_humidity to keep counting")
			}
		})

		t.Run("the block list narrows the filtered listing", func(t *testing.T) {
			result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, group, client.DeviceGroupHelperOptions{
				MaintainsGroupUsability: true,
				FunctionBlockList:       []string{getTemperature},
			})
			if err != nil {
				t.Error(err, code)
				return
			}
			//only the devices answering getHumidity are left
			expected := []string{"device_humidity"}
			if !slices.Equal(optionIds(result.Options), expected) {
				t.Error("\na=", optionIds(result.Options), "\ne=", expected)
			}
		})

		t.Run("blocking every function of the group leaves no option", func(t *testing.T) {
			result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, group, client.DeviceGroupHelperOptions{
				MaintainsGroupUsability: true,
				FunctionBlockList:       []string{getTemperature, getHumidity},
			})
			if err != nil {
				t.Error(err, code)
				return
			}
			if len(result.Options) != 0 {
				t.Error("expected no option to count", optionIds(result.Options))
			}
			//the criteria of the group itself are unaffected: the block list is about which
			//devices are worth offering, not about what the group can do
			if len(result.Criteria) != 6 {
				t.Error("expected the block list to leave the group criteria alone", len(result.Criteria))
			}
		})

		t.Run("an unknown blocked function changes nothing", func(t *testing.T) {
			withBlock, err, code := c.DeviceGroupHelper(client.InternalAdminToken, group, client.DeviceGroupHelperOptions{
				MaintainsGroupUsability: true,
				FunctionBlockList:       []string{models.URN_PREFIX + "measuring-function:notUsedHere"},
			})
			if err != nil {
				t.Error(err, code)
				return
			}
			without, err, code := c.DeviceGroupHelper(client.InternalAdminToken, group, client.DeviceGroupHelperOptions{
				MaintainsGroupUsability: true,
			})
			if err != nil {
				t.Error(err, code)
				return
			}
			if !slices.Equal(optionIds(withBlock.Options), optionIds(without.Options)) {
				t.Error("\na=", optionIds(withBlock.Options), "\ne=", optionIds(without.Options))
			}
		})

		//the block list reaches the service as one comma-separated query parameter, so an entry
		//padded with spaces has to survive the split
		t.Run("entries are trimmed", func(t *testing.T) {
			result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, group, client.DeviceGroupHelperOptions{
				MaintainsGroupUsability: true,
				FunctionBlockList:       []string{" " + getTemperature + " "},
			})
			if err != nil {
				t.Error(err, code)
				return
			}
			if !slices.Equal(optionIds(result.Options), []string{"device_humidity"}) {
				t.Error("expected the padded entry to block the same function", optionIds(result.Options))
			}
		})
	})

	t.Run("options: search filters the listing", func(t *testing.T) {
		result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, []string{"device_ab"}, client.DeviceGroupHelperOptions{
			Search: "device_bc",
		})
		if err != nil {
			t.Error(err, code)
			return
		}
		if !slices.Equal(optionIds(result.Options), []string{"device_bc"}) {
			t.Error("expected only the searched device", optionIds(result.Options))
		}
	})

	t.Run("options: limit caps the listing", func(t *testing.T) {
		result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, []string{"device_ab"}, client.DeviceGroupHelperOptions{
			Limit: 2,
		})
		if err != nil {
			t.Error(err, code)
			return
		}
		//the limit applies to the unmodified listing; the modified variants are read by id from
		//the same page, so the result may hold more elements than the limit but not the full set
		if len(result.Options) == 0 {
			t.Error("expected options")
		}
		if len(result.Options) >= len(deviceIdByDeviceTypeId) {
			t.Error("expected the limit to cut the listing", optionIds(result.Options))
		}
	})

	t.Run("options: an empty group offers every device that has a criteria", func(t *testing.T) {
		result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, []string{}, client.DeviceGroupHelperOptions{
			MaintainsGroupUsability: true,
		})
		if err != nil {
			t.Error(err, code)
			return
		}
		if len(result.Criteria) != 0 {
			t.Error("expected no criteria for an empty group", result.Criteria)
		}
		for _, deviceId := range deviceIdByDeviceTypeId {
			if !slices.Contains(optionIds(result.Options), deviceId) {
				t.Error("expected every device to be an option", deviceId, optionIds(result.Options))
				return
			}
		}
	})
}

func testDeviceGroupHelperCriteria(c client.Interface, deviceIds []string, expected []models.DeviceGroupFilterCriteria) func(t *testing.T) {
	return func(t *testing.T) {
		result, err, code := c.DeviceGroupHelper(client.InternalAdminToken, deviceIds, client.DeviceGroupHelperOptions{})
		if err != nil {
			t.Error(err, code)
			return
		}
		assertSameCriteria(t, result.Criteria, expected)
	}
}

// assertSameCriteria compares by Short(), which is the key the intersection uses and therefore
// total: sorting by AspectId alone leaves the criteria over an aspect list next to the one over
// its first aspect in the order the answer happened to carry them.
func assertSameCriteria(t *testing.T, actual []models.DeviceGroupFilterCriteria, expected []models.DeviceGroupFilterCriteria) {
	t.Helper()
	a := sortCriteriaByShort(actual)
	e := sortCriteriaByShort(expected)
	if !reflect.DeepEqual(a, e) {
		actualJson, _ := json.Marshal(a)
		expectedJson, _ := json.Marshal(e)
		t.Error("\na=", string(actualJson), "\ne=", string(expectedJson))
	}
}

func sortCriteriaByShort(criteria []models.DeviceGroupFilterCriteria) []models.DeviceGroupFilterCriteria {
	result := slices.Clone(criteria)
	if result == nil {
		result = []models.DeviceGroupFilterCriteria{}
	}
	slices.SortFunc(result, func(a, b models.DeviceGroupFilterCriteria) int {
		return strings.Compare(a.Short(), b.Short())
	})
	return result
}

func optionsById(options []client.DeviceGroupOption) map[string]client.DeviceGroupOption {
	result := map[string]client.DeviceGroupOption{}
	for _, option := range options {
		result[option.Device.Id] = option
	}
	return result
}

func optionIds(options []client.DeviceGroupOption) []string {
	result := []string{}
	for _, option := range options {
		result = append(result, option.Device.Id)
	}
	slices.Sort(result)
	return result
}
