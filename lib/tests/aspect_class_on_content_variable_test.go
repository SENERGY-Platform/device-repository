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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

// TestAspectClassOnContentVariable covers the rule that a content variable carries at
// most one aspect per aspect-class.
func TestAspectClassOnContentVariable(t *testing.T) {
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
		airClass   = models.URN_PREFIX + "aspect-class:air"
		waterClass = models.URN_PREFIX + "aspect-class:water"

		insideAir  = models.URN_PREFIX + "aspect:inside_air"
		outsideAir = models.URN_PREFIX + "aspect:outside_air"
		lake       = models.URN_PREFIX + "aspect:lake"
		freeA      = models.URN_PREFIX + "aspect:free_a"
		freeB      = models.URN_PREFIX + "aspect:free_b"

		measuringFunction = models.URN_PREFIX + "measuring-function:getTemperature"
		deviceClass       = models.URN_PREFIX + "device-class:dc1"
	)

	t.Run("create metadata", func(t *testing.T) {
		for _, id := range []string{airClass, waterClass} {
			_, err, _ := c.SetAspectClass(client.InternalAdminToken, models.AspectClass{Id: id, Name: id})
			if err != nil {
				t.Error(err)
				return
			}
		}
		//two aspects of the same class, sitting in one hierarchy
		_, err, _ := c.SetAspect(client.InternalAdminToken, models.Aspect{
			Id:            models.URN_PREFIX + "aspect:air",
			Name:          "air",
			AspectClassId: airClass,
			SubAspects: []models.Aspect{
				{Id: insideAir, Name: "inside_air"},
				{Id: outsideAir, Name: "outside_air"},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		//one aspect of a second class
		_, err, _ = c.SetAspect(client.InternalAdminToken, models.Aspect{
			Id: models.URN_PREFIX + "aspect:water", Name: "water", AspectClassId: waterClass,
			SubAspects: []models.Aspect{{Id: lake, Name: "lake"}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		//two aspects without any class
		for _, id := range []string{freeA, freeB} {
			_, err, _ = c.SetAspect(client.InternalAdminToken, models.Aspect{Id: id, Name: id})
			if err != nil {
				t.Error(err)
				return
			}
		}
		_, err, _ = c.SetFunction(client.InternalAdminToken, models.Function{
			Id: measuringFunction, Name: "getTemperature", DisplayName: "getTemperature",
			RdfType: models.SES_ONTOLOGY_MEASURING_FUNCTION,
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
			Id: "p1", Name: "p1", Handler: "p1",
			ProtocolSegments: []models.ProtocolSegment{{Id: "ps1", Name: "ps1"}},
		})
		if err != nil {
			t.Error(err)
		}
	})

	deviceTypeWithAspects := func(id string, aspectIds []string) models.DeviceType {
		return models.DeviceType{
			Id:            id,
			Name:          id,
			DeviceClassId: deviceClass,
			Services: []models.Service{{
				Id: id + "_s1", LocalId: id + "_s1", Name: "s1",
				Interaction: models.EVENT_AND_REQUEST, ProtocolId: "p1",
				Outputs: []models.Content{{
					Id: id + "_s1_c1", Serialization: models.JSON, ProtocolSegmentId: "ps1",
					ContentVariable: models.ContentVariable{
						Id: id + "_cv1", Name: "temperature", Type: models.String,
						FunctionId: measuringFunction, AspectIds: aspectIds,
					},
				}},
			}},
		}
	}

	t.Run("two aspects of different classes", func(t *testing.T) {
		_, err, _ := c.SetDeviceType(client.InternalAdminToken,
			deviceTypeWithAspects("dt_two_classes", []string{insideAir, lake}), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("two aspects of the same class", func(t *testing.T) {
		_, err, code := c.SetDeviceType(client.InternalAdminToken,
			deviceTypeWithAspects("dt_same_class", []string{insideAir, outsideAir}), client.DeviceTypeUpdateOptions{})
		if err == nil {
			t.Error("expected an error for two aspects of one aspect-class")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
		for _, expected := range []string{insideAir, outsideAir, airClass} {
			if !strings.Contains(err.Error(), expected) {
				t.Error("error does not name", expected, "|", err.Error())
			}
		}
	})

	t.Run("two aspects without a class", func(t *testing.T) {
		//no class means no collision
		_, err, _ := c.SetDeviceType(client.InternalAdminToken,
			deviceTypeWithAspects("dt_classless", []string{freeA, freeB}), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("the same aspect twice", func(t *testing.T) {
		//a repeated id shares its own aspect-class with itself, so the rule catches it —
		//with a message that says "listed twice" rather than naming two aspects
		_, err, code := c.SetDeviceType(client.InternalAdminToken,
			deviceTypeWithAspects("dt_repeat", []string{insideAir, insideAir}), client.DeviceTypeUpdateOptions{})
		if err == nil {
			t.Error("expected an error for the same aspect listed twice")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
		if !strings.Contains(err.Error(), "listed twice") {
			t.Error("unexpected message", err.Error())
		}
	})

	t.Run("the same unclassified aspect twice", func(t *testing.T) {
		//without a class there is nothing to collide on, so this is not caught here
		_, err, _ := c.SetDeviceType(client.InternalAdminToken,
			deviceTypeWithAspects("dt_repeat_classless", []string{freeA, freeA}), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("sub content variables are checked too", func(t *testing.T) {
		dt := deviceTypeWithAspects("dt_sub", nil)
		dt.Services[0].Outputs[0].ContentVariable = models.ContentVariable{
			Id: "dt_sub_cv1", Name: "temperatures", Type: models.Structure,
			SubContentVariables: []models.ContentVariable{{
				Id: "dt_sub_cv2", Name: "inside", Type: models.String,
				FunctionId: measuringFunction, AspectIds: []string{insideAir, outsideAir},
			}},
		}
		_, err, code := c.SetDeviceType(client.InternalAdminToken, dt, client.DeviceTypeUpdateOptions{})
		if err == nil {
			t.Error("expected an error for two aspects of one aspect-class in a sub content variable")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})
}
