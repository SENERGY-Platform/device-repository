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
	"net/http"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib"
	"github.com/SENERGY-Platform/device-repository/v3/lib/client"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

// TestDeviceGroupCriteriaRecompute covers the admin trigger of SNRGY-4861: stale criteria of
// a generated and a manually created group are rebuilt from the device-type of their device,
// and the ids parameter limits the rebuild to the listed groups.
func TestDeviceGroupCriteriaRecompute(t *testing.T) {
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
		measuringFunction = models.URN_PREFIX + "measuring-function:getTemperature"
		airAspect         = models.URN_PREFIX + "aspect:air"
		deviceClass       = models.URN_PREFIX + "device-class:dc1"
	)

	var generatedGroupId string
	var manualGroup models.DeviceGroup
	var expected []string

	t.Run("create resources", func(t *testing.T) {
		_, err, _ := c.SetAspect(client.InternalAdminToken, models.Aspect{Id: airAspect, Name: "air"})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetFunction(client.InternalAdminToken, models.Function{
			Id:          measuringFunction,
			Name:        "getTemperature",
			DisplayName: "getTemperature",
			RdfType:     models.SES_ONTOLOGY_MEASURING_FUNCTION,
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetDeviceClass(client.InternalAdminToken, models.DeviceClass{Id: deviceClass, Name: "dc1"})
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
		_, err, _ = c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Id:            "dt1",
			Name:          "dt1",
			DeviceClassId: deviceClass,
			Services: []models.Service{{
				Id:          "dt1_s1",
				LocalId:     "dt1_s1",
				Name:        "s1",
				Interaction: models.EVENT_AND_REQUEST,
				ProtocolId:  "p1",
				Outputs: []models.Content{{
					Id:                "dt1_s1_c1",
					Serialization:     models.JSON,
					ProtocolSegmentId: "ps1",
					ContentVariable: models.ContentVariable{
						Id:         "dt1_cv1",
						Name:       "temperature",
						Type:       models.String,
						FunctionId: measuringFunction,
						AspectIds:  []string{airAspect},
					},
				}},
			}},
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		device, err, _ := c.CreateDevice(client.InternalAdminToken, models.Device{
			LocalId:      "d1",
			Name:         "d1",
			DeviceTypeId: "dt1",
		})
		if err != nil {
			t.Error(err)
			return
		}
		generatedGroupId = model.DeviceIdToGeneratedDeviceGroupId(device.Id)
		generated, err, _ := c.ReadDeviceGroup(generatedGroupId, client.InternalAdminToken, false)
		if err != nil {
			t.Error(err)
			return
		}
		expected = slices.Sorted(slices.Values(generated.CriteriaShort))
		if len(expected) == 0 {
			t.Error("expected the generated group to have criteria", generated)
			return
		}
		manualGroup, err, _ = c.SetDeviceGroup(client.InternalAdminToken, models.DeviceGroup{
			Name:      "manual",
			DeviceIds: []string{device.Id},
			Criteria:  generated.Criteria,
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	//the stale state is not producible through the api, so it is written past the controller
	t.Run("make criteria stale", func(t *testing.T) {
		db, err := mongo.New(config)
		if err != nil {
			t.Error(err)
			return
		}
		defer db.Disconnect()
		for _, id := range []string{generatedGroupId, manualGroup.Id} {
			dg, _, err := db.GetDeviceGroup(ctx, id)
			if err != nil {
				t.Error(err)
				return
			}
			user, _, err := db.GetDeviceGroupSyncUser(ctx, id)
			if err != nil {
				t.Error(err)
				return
			}
			dg.Criteria = []models.DeviceGroupFilterCriteria{}
			dg.SetShortCriteria()
			err = db.SetDeviceGroup(ctx, dg, func(models.DeviceGroup, string) error { return nil }, user)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	criteriaOf := func(t *testing.T, id string) []string {
		t.Helper()
		dg, err, _ := c.ReadDeviceGroup(id, client.InternalAdminToken, false)
		if err != nil {
			t.Error(err)
			return nil
		}
		return slices.Sorted(slices.Values(dg.CriteriaShort))
	}

	//the recompute runs in the background, so the result is polled
	waitForCriteria := func(t *testing.T, id string) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for !slices.Equal(criteriaOf(t, id), expected) {
			if time.Now().After(deadline) {
				t.Error("criteria not recomputed", id, criteriaOf(t, id), expected)
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}

	t.Run("non admin is rejected", func(t *testing.T) {
		err, code := c.RecomputeDeviceGroupCriteria(userToken, nil)
		if err == nil || code != http.StatusForbidden {
			t.Error("expected 403", err, code)
		}
	})

	t.Run("ids limit the recompute", func(t *testing.T) {
		err, _ := c.RecomputeDeviceGroupCriteria(client.InternalAdminToken, []string{manualGroup.Id})
		if err != nil {
			t.Error(err)
			return
		}
		waitForCriteria(t, manualGroup.Id)
		//the run is over once the only listed group is written, so the other one is final
		time.Sleep(time.Second)
		if len(criteriaOf(t, generatedGroupId)) != 0 {
			t.Error("expected the unlisted group to keep its stale criteria", criteriaOf(t, generatedGroupId))
		}
	})

	t.Run("without ids every group is recomputed", func(t *testing.T) {
		err, _ := c.RecomputeDeviceGroupCriteria(client.InternalAdminToken, nil)
		if err != nil {
			t.Error(err)
			return
		}
		waitForCriteria(t, generatedGroupId)
		waitForCriteria(t, manualGroup.Id)
	})
}
