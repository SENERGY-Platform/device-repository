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
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

// TestAspectClassIdRequired covers the config flag that turns the aspect-class
// assignment into a requirement. It is off by default, which is why every other test
// can write aspects without one.
func TestAspectClassIdRequired(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	if config.AspectClassIdRequired {
		t.Error("expected aspect_class_id_required to default to false")
		return
	}
	config.AspectClassIdRequired = true
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

	const aspectClass = models.URN_PREFIX + "aspect-class:required"

	t.Run("create aspect-class", func(t *testing.T) {
		_, err, _ := c.SetAspectClass(client.InternalAdminToken, models.AspectClass{Id: aspectClass, Name: "required"})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("aspect without a class is rejected", func(t *testing.T) {
		_, err, code := c.SetAspect(client.InternalAdminToken, models.Aspect{
			Id:   models.URN_PREFIX + "aspect:without_class",
			Name: "without_class",
		})
		if err == nil {
			t.Error("expected an error while aspect_class_id_required is on")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})

	t.Run("aspect with a class is accepted", func(t *testing.T) {
		result, err, _ := c.SetAspect(client.InternalAdminToken, models.Aspect{
			Id:            models.URN_PREFIX + "aspect:with_class",
			Name:          "with_class",
			AspectClassId: aspectClass,
			SubAspects: []models.Aspect{
				{Id: models.URN_PREFIX + "aspect:with_class_sub", Name: "with_class_sub"},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		//the requirement is satisfied for the whole hierarchy by the root, through inheritance
		if result.SubAspects[0].AspectClassId != aspectClass {
			t.Error("sub aspect did not inherit", result.SubAspects[0].AspectClassId)
		}
	})

	t.Run("validation endpoint agrees", func(t *testing.T) {
		err, code := c.ValidateAspect(models.Aspect{
			Id:   models.URN_PREFIX + "aspect:validate_without_class",
			Name: "validate_without_class",
		})
		if err == nil {
			t.Error("expected an error")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})
}
