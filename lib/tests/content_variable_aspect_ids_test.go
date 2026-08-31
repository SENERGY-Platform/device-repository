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
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

// TestContentVariableAspectIds covers the migration from the deprecated
// models.ContentVariable.AspectId to models.ContentVariable.AspectIds:
// the write path adds AspectId to AspectIds, the read path derives AspectId
// from AspectIds, and everything in between works on AspectIds only.
func TestContentVariableAspectIds(t *testing.T) {
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

	const measuringFunction = models.URN_PREFIX + "measuring-function:getTemperature"
	const (
		airAspect   = models.URN_PREFIX + "aspect:air"
		waterAspect = models.URN_PREFIX + "aspect:water"
		deviceClass = models.URN_PREFIX + "device-class:dc1"
	)

	t.Run("create metadata", func(t *testing.T) {
		for _, aspectId := range []string{airAspect, waterAspect} {
			_, err, _ := c.SetAspect(client.InternalAdminToken, models.Aspect{Id: aspectId, Name: aspectId})
			if err != nil {
				t.Error(err)
				return
			}
		}
		_, err, _ := c.SetFunction(client.InternalAdminToken, models.Function{
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
	})

	deviceType := func(id string, variable models.ContentVariable) models.DeviceType {
		return models.DeviceType{
			Id:            id,
			Name:          id,
			DeviceClassId: deviceClass,
			Services: []models.Service{{
				Id:          id + "_s1",
				LocalId:     id + "_s1",
				Name:        "s1",
				Interaction: models.EVENT_AND_REQUEST,
				ProtocolId:  "p1",
				Outputs: []models.Content{{
					Id:                id + "_s1_c1",
					Serialization:     models.JSON,
					ProtocolSegmentId: "ps1",
					ContentVariable:   variable,
				}},
			}},
		}
	}

	readVariable := func(t *testing.T, deviceTypeId string) models.ContentVariable {
		t.Helper()
		dt, err, _ := c.ReadDeviceType(deviceTypeId, client.InternalAdminToken)
		if err != nil {
			t.Error(err)
			return models.ContentVariable{}
		}
		if len(dt.Services) != 1 || len(dt.Services[0].Outputs) != 1 {
			t.Error("unexpected device-type", dt)
			return models.ContentVariable{}
		}
		return dt.Services[0].Outputs[0].ContentVariable
	}

	t.Run("write deprecated aspect_id only", func(t *testing.T) {
		_, err, _ := c.SetDeviceType(client.InternalAdminToken, deviceType("dt_deprecated", models.ContentVariable{
			Id:         "dt_deprecated_cv1",
			Name:       "temperature",
			Type:       models.String,
			FunctionId: measuringFunction,
			AspectId:   airAspect,
		}), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		variable := readVariable(t, "dt_deprecated")
		if !reflect.DeepEqual(variable.AspectIds, []string{airAspect}) {
			t.Error("expected aspect_id to be added to aspect_ids", variable.AspectIds)
		}
		if variable.AspectId != airAspect {
			t.Error("unexpected aspect_id", variable.AspectId)
		}
	})

	t.Run("write aspect_ids only", func(t *testing.T) {
		_, err, _ := c.SetDeviceType(client.InternalAdminToken, deviceType("dt_aspectids", models.ContentVariable{
			Id:         "dt_aspectids_cv1",
			Name:       "temperature",
			Type:       models.String,
			FunctionId: measuringFunction,
			AspectIds:  []string{waterAspect, airAspect},
		}), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		variable := readVariable(t, "dt_aspectids")
		if !reflect.DeepEqual(variable.AspectIds, []string{waterAspect, airAspect}) {
			t.Error("unexpected aspect_ids", variable.AspectIds)
		}
		//the deprecated field gets the alphabetically first entry, independent of the stored order
		if variable.AspectId != airAspect {
			t.Error("unexpected aspect_id", variable.AspectId)
		}
	})

	t.Run("write both with unlisted aspect_id", func(t *testing.T) {
		_, err, _ := c.SetDeviceType(client.InternalAdminToken, deviceType("dt_both", models.ContentVariable{
			Id:         "dt_both_cv1",
			Name:       "temperature",
			Type:       models.String,
			FunctionId: measuringFunction,
			AspectId:   waterAspect,
			AspectIds:  []string{airAspect},
		}), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		variable := readVariable(t, "dt_both")
		if !reflect.DeepEqual(variable.AspectIds, []string{airAspect, waterAspect}) {
			t.Error("expected aspect_id to be appended to aspect_ids", variable.AspectIds)
		}
		if variable.AspectId != airAspect {
			t.Error("unexpected aspect_id", variable.AspectId)
		}
	})

	t.Run("write both with listed aspect_id", func(t *testing.T) {
		_, err, _ := c.SetDeviceType(client.InternalAdminToken, deviceType("dt_both_known", models.ContentVariable{
			Id:         "dt_both_known_cv1",
			Name:       "temperature",
			Type:       models.String,
			FunctionId: measuringFunction,
			AspectId:   waterAspect,
			AspectIds:  []string{airAspect, waterAspect},
		}), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		variable := readVariable(t, "dt_both_known")
		if !reflect.DeepEqual(variable.AspectIds, []string{airAspect, waterAspect}) {
			t.Error("expected no duplicate in aspect_ids", variable.AspectIds)
		}
	})

	t.Run("sub content variables", func(t *testing.T) {
		_, err, _ := c.SetDeviceType(client.InternalAdminToken, deviceType("dt_sub", models.ContentVariable{
			Id:   "dt_sub_cv1",
			Name: "temperatures",
			Type: models.Structure,
			SubContentVariables: []models.ContentVariable{{
				Id:         "dt_sub_cv2",
				Name:       "inside",
				Type:       models.String,
				FunctionId: measuringFunction,
				AspectId:   airAspect,
			}, {
				Id:         "dt_sub_cv3",
				Name:       "outside",
				Type:       models.String,
				FunctionId: measuringFunction,
				AspectIds:  []string{waterAspect},
			}},
		}), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		variable := readVariable(t, "dt_sub")
		if len(variable.SubContentVariables) != 2 {
			t.Error("unexpected sub content variables", variable.SubContentVariables)
			return
		}
		if !reflect.DeepEqual(variable.SubContentVariables[0].AspectIds, []string{airAspect}) {
			t.Error("unexpected aspect_ids", variable.SubContentVariables[0].AspectIds)
		}
		if variable.SubContentVariables[1].AspectId != waterAspect {
			t.Error("unexpected aspect_id", variable.SubContentVariables[1].AspectId)
		}
	})

	t.Run("unknown aspect in aspect_ids is rejected", func(t *testing.T) {
		err, code := c.ValidateDeviceType(deviceType("dt_invalid", models.ContentVariable{
			Id:         "dt_invalid_cv1",
			Name:       "temperature",
			Type:       models.String,
			FunctionId: measuringFunction,
			AspectIds:  []string{airAspect, models.URN_PREFIX + "aspect:unknown"},
		}), client.DeviceTypeValidationOptions{})
		if err == nil {
			t.Error("expected validation error for unknown aspect id")
			return
		}
		if code != 400 {
			t.Error("unexpected code", code)
		}
	})

	t.Run("every aspect of aspect_ids is selectable", func(t *testing.T) {
		for _, aspectId := range []string{waterAspect, airAspect} {
			selectables, err, _ := c.GetDeviceTypeSelectablesV2([]model.FilterCriteria{{
				FunctionId: measuringFunction,
				AspectId:   aspectId,
			}}, "", false, false)
			if err != nil {
				t.Error(err)
				return
			}
			found := slices.ContainsFunc(selectables, func(selectable model.DeviceTypeSelectable) bool {
				return selectable.DeviceTypeId == "dt_aspectids"
			})
			if !found {
				t.Error("dt_aspectids not selectable by aspect", aspectId, selectables)
			}
		}
	})

	t.Run("every aspect of aspect_ids is used-in", func(t *testing.T) {
		result, err, _ := c.GetUsedInDeviceType(model.UsedInDeviceTypeQuery{
			Resource: "aspects",
			Ids:      []string{waterAspect, airAspect},
		})
		if err != nil {
			t.Error(err)
			return
		}
		for _, aspectId := range []string{waterAspect, airAspect} {
			found := slices.ContainsFunc(result[aspectId].UsedIn, func(ref model.DeviceTypeReference) bool {
				return ref.Id == "dt_aspectids"
			})
			if !found {
				t.Error("dt_aspectids not used-in aspect", aspectId, result[aspectId])
			}
		}
	})

	//device-types stored before the migration carry only the deprecated AspectId; writing one
	//through the database layer skips the controller and reproduces exactly that state
	t.Run("device-type stored with aspect_id only", func(t *testing.T) {
		db, err := mongo.New(config)
		if err != nil {
			t.Error(err)
			return
		}
		defer db.Disconnect()
		legacy := deviceType("dt_legacy", models.ContentVariable{
			Id:         "dt_legacy_cv1",
			Name:       "temperature",
			Type:       models.String,
			FunctionId: measuringFunction,
			AspectId:   waterAspect,
		})
		err = db.SetDeviceType(context.Background(), legacy, func(models.DeviceType) error { return nil })
		if err != nil {
			t.Error(err)
			return
		}

		t.Run("read fills aspect_ids", func(t *testing.T) {
			variable := readVariable(t, "dt_legacy")
			if !reflect.DeepEqual(variable.AspectIds, []string{waterAspect}) {
				t.Error("expected aspect_ids to be derived from aspect_id", variable.AspectIds)
			}
			if variable.AspectId != waterAspect {
				t.Error("unexpected aspect_id", variable.AspectId)
			}
		})

		t.Run("generated device-group keeps the aspect", func(t *testing.T) {
			device, err, _ := c.CreateDevice(client.InternalAdminToken, models.Device{
				LocalId:      "d_legacy",
				Name:         "d_legacy",
				DeviceTypeId: "dt_legacy",
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
			found := slices.ContainsFunc(group.Criteria, func(criteria models.DeviceGroupFilterCriteria) bool {
				return criteria.FunctionId == measuringFunction && criteria.AspectId == waterAspect
			})
			if !found {
				t.Error("generated device-group lost the aspect", group.Criteria)
			}
		})
	})
}
