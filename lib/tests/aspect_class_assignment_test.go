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
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

// TestAspectClassAssignment covers how an aspect hierarchy gets its aspect-class:
// only the root assigns one, every aspect below inherits it, and the class cannot
// be deleted while an aspect still carries it.
func TestAspectClassAssignment(t *testing.T) {
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
		envClass   = models.URN_PREFIX + "aspect-class:environment"
		deviceCls  = models.URN_PREFIX + "aspect-class:device"
		unusedCls  = models.URN_PREFIX + "aspect-class:unused"
		unknownCls = models.URN_PREFIX + "aspect-class:unknown"
	)

	t.Run("create aspect-classes", func(t *testing.T) {
		for _, id := range []string{envClass, deviceCls, unusedCls} {
			_, err, _ := c.SetAspectClass(client.InternalAdminToken, models.AspectClass{Id: id, Name: id})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	airAspect := models.Aspect{
		Id:            models.URN_PREFIX + "aspect:air",
		Name:          "air",
		AspectClassId: envClass,
		SubAspects: []models.Aspect{
			{Id: models.URN_PREFIX + "aspect:inside_air", Name: "inside_air"},
			{
				Id:   models.URN_PREFIX + "aspect:outside_air",
				Name: "outside_air",
				SubAspects: []models.Aspect{
					{Id: models.URN_PREFIX + "aspect:morning_air", Name: "morning_air"},
				},
			},
		},
	}

	t.Run("sub aspects inherit the root class", func(t *testing.T) {
		result, err, _ := c.SetAspect(client.InternalAdminToken, airAspect)
		if err != nil {
			t.Error(err)
			return
		}
		if result.SubAspects[0].AspectClassId != envClass {
			t.Error("direct sub aspect did not inherit", result.SubAspects[0].AspectClassId)
		}
		if result.SubAspects[1].SubAspects[0].AspectClassId != envClass {
			t.Error("nested sub aspect did not inherit", result.SubAspects[1].SubAspects[0].AspectClassId)
		}
	})

	t.Run("aspect-nodes carry the class", func(t *testing.T) {
		for _, id := range []string{
			models.URN_PREFIX + "aspect:air",
			models.URN_PREFIX + "aspect:inside_air",
			models.URN_PREFIX + "aspect:morning_air",
		} {
			node, err, _ := c.GetAspectNode(id)
			if err != nil {
				t.Error(err)
				return
			}
			if node.AspectClassId != envClass {
				t.Error("aspect-node without class", id, node.AspectClassId)
			}
		}
	})

	t.Run("sub aspect may repeat the root class", func(t *testing.T) {
		repeated := models.Aspect{
			Id:            models.URN_PREFIX + "aspect:repeat_root",
			Name:          "repeat_root",
			AspectClassId: deviceCls,
			SubAspects: []models.Aspect{
				{Id: models.URN_PREFIX + "aspect:repeat_sub", Name: "repeat_sub", AspectClassId: deviceCls},
			},
		}
		_, err, _ := c.SetAspect(client.InternalAdminToken, repeated)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("sub aspect may not set a different class", func(t *testing.T) {
		conflicting := models.Aspect{
			Id:            models.URN_PREFIX + "aspect:conflict_root",
			Name:          "conflict_root",
			AspectClassId: envClass,
			SubAspects: []models.Aspect{
				{Id: models.URN_PREFIX + "aspect:conflict_sub", Name: "conflict_sub", AspectClassId: deviceCls},
			},
		}
		_, err, code := c.SetAspect(client.InternalAdminToken, conflicting)
		if err == nil {
			t.Error("expected an error for two classes in one hierarchy")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})

	t.Run("sub aspect may not classify an unclassified hierarchy", func(t *testing.T) {
		fromBelow := models.Aspect{
			Id:   models.URN_PREFIX + "aspect:unclassified_root",
			Name: "unclassified_root",
			SubAspects: []models.Aspect{
				{Id: models.URN_PREFIX + "aspect:classified_sub", Name: "classified_sub", AspectClassId: envClass},
			},
		}
		_, err, code := c.SetAspect(client.InternalAdminToken, fromBelow)
		if err == nil {
			t.Error("expected an error: only the root assigns the class")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})

	t.Run("unknown class is rejected", func(t *testing.T) {
		_, err, code := c.SetAspect(client.InternalAdminToken, models.Aspect{
			Id:            models.URN_PREFIX + "aspect:unknown_class",
			Name:          "unknown_class",
			AspectClassId: unknownCls,
		})
		if err == nil {
			t.Error("expected an error for an unknown aspect class")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})

	t.Run("hierarchy without class stays allowed", func(t *testing.T) {
		_, err, _ := c.SetAspect(client.InternalAdminToken, models.Aspect{
			Id:   models.URN_PREFIX + "aspect:classless",
			Name: "classless",
			SubAspects: []models.Aspect{
				{Id: models.URN_PREFIX + "aspect:classless_sub", Name: "classless_sub"},
			},
		})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("filter aspects by aspect-class", func(t *testing.T) {
		result, total, err, _ := c.ListAspects(model.AspectListOptions{AspectClassIds: []string{envClass}})
		if err != nil {
			t.Error(err)
			return
		}
		//aspects are stored as trees, so only the root of the air hierarchy is listed
		if total != 1 || len(result) != 1 || result[0].Id != models.URN_PREFIX+"aspect:air" {
			t.Error("unexpected result", total, result)
		}
	})

	t.Run("filter aspect-nodes by aspect-class", func(t *testing.T) {
		result, total, err, _ := c.ListAspectNodes(model.AspectListOptions{AspectClassIds: []string{envClass}})
		if err != nil {
			t.Error(err)
			return
		}
		//every aspect of the hierarchy has its own node
		if total != 4 {
			t.Error("unexpected total", total)
		}
		ids := []string{}
		for _, node := range result {
			ids = append(ids, node.Id)
		}
		expected := []string{
			models.URN_PREFIX + "aspect:air",
			models.URN_PREFIX + "aspect:inside_air",
			models.URN_PREFIX + "aspect:morning_air",
			models.URN_PREFIX + "aspect:outside_air",
		}
		if !reflect.DeepEqual(ids, expected) {
			t.Error("unexpected result", ids)
		}
	})

	t.Run("filter by two aspect-classes", func(t *testing.T) {
		_, total, err, _ := c.ListAspectNodes(model.AspectListOptions{AspectClassIds: []string{envClass, deviceCls}})
		if err != nil {
			t.Error(err)
			return
		}
		//four of the air hierarchy plus the two of the repeat_root hierarchy
		if total != 6 {
			t.Error("unexpected total", total)
		}
	})

	t.Run("filter by an empty aspect-class list", func(t *testing.T) {
		_, total, err, _ := c.ListAspectNodes(model.AspectListOptions{AspectClassIds: []string{}})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 0 {
			t.Error("an empty filter list returns nothing", total)
		}
	})

	t.Run("without the filter nothing is filtered", func(t *testing.T) {
		_, total, err, _ := c.ListAspectNodes(model.AspectListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total < 6 {
			t.Error("unexpected total", total)
		}
	})

	t.Run("delete a class still in use", func(t *testing.T) {
		err, code := c.DeleteAspectClass(client.InternalAdminToken, envClass)
		if err == nil {
			t.Error("expected an error for an aspect-class still in use")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
		//the error names every offending aspect with name and id
		for _, expected := range []string{
			"air (" + models.URN_PREFIX + "aspect:air)",
			"inside_air (" + models.URN_PREFIX + "aspect:inside_air)",
			"morning_air (" + models.URN_PREFIX + "aspect:morning_air)",
			"outside_air (" + models.URN_PREFIX + "aspect:outside_air)",
		} {
			if !strings.Contains(err.Error(), expected) {
				t.Error("error does not name", expected, "|", err.Error())
			}
		}
	})

	t.Run("validate delete of a class still in use", func(t *testing.T) {
		err, code := c.ValidateAspectClassDelete(envClass)
		if err == nil {
			t.Error("expected an error")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})

	t.Run("delete an unused class", func(t *testing.T) {
		err, _ := c.DeleteAspectClass(client.InternalAdminToken, unusedCls)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("delete becomes possible once the aspects are gone", func(t *testing.T) {
		err, _ := c.DeleteAspect(client.InternalAdminToken, models.URN_PREFIX+"aspect:air")
		if err != nil {
			t.Error(err)
			return
		}
		err, _ = c.DeleteAspectClass(client.InternalAdminToken, envClass)
		if err != nil {
			t.Error(err)
		}
	})
}
