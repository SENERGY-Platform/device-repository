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

// userToken carries the 'user' role but not 'admin'
const userToken = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`

func TestAspectClasses(t *testing.T) {
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
		fanClass   = models.URN_PREFIX + "aspect-class:fan"
	)

	t.Run("create with given id", func(t *testing.T) {
		result, err, _ := c.SetAspectClass(client.InternalAdminToken, models.AspectClass{Id: airClass, Name: "air"})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(result, models.AspectClass{Id: airClass, Name: "air"}) {
			t.Error("unexpected result", result)
		}
	})

	t.Run("create without id generates one", func(t *testing.T) {
		result, err, _ := c.SetAspectClass(client.InternalAdminToken, models.AspectClass{Name: "generated"})
		if err != nil {
			t.Error(err)
			return
		}
		if !strings.HasPrefix(result.Id, models.URN_PREFIX+"aspect-class:") {
			t.Error("unexpected generated id", result.Id)
			return
		}
		read, err, _ := c.GetAspectClass(result.Id)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(read, result) {
			t.Error("unexpected result", read, result)
		}
		err, _ = c.DeleteAspectClass(client.InternalAdminToken, result.Id)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("create remaining", func(t *testing.T) {
		for id, name := range map[string]string{waterClass: "water", fanClass: "fan"} {
			_, err, _ := c.SetAspectClass(client.InternalAdminToken, models.AspectClass{Id: id, Name: name})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("read", func(t *testing.T) {
		result, err, _ := c.GetAspectClass(airClass)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(result, models.AspectClass{Id: airClass, Name: "air"}) {
			t.Error("unexpected result", result)
		}
	})

	t.Run("read unknown", func(t *testing.T) {
		_, err, code := c.GetAspectClass(models.URN_PREFIX + "aspect-class:unknown")
		if err == nil {
			t.Error("expected error for unknown aspect-class")
			return
		}
		if code != http.StatusNotFound {
			t.Error("unexpected code", code)
		}
	})

	t.Run("update", func(t *testing.T) {
		_, err, _ := c.SetAspectClass(client.InternalAdminToken, models.AspectClass{Id: airClass, Name: "air updated"})
		if err != nil {
			t.Error(err)
			return
		}
		result, err, _ := c.GetAspectClass(airClass)
		if err != nil {
			t.Error(err)
			return
		}
		if result.Name != "air updated" {
			t.Error("unexpected result", result)
		}
	})

	t.Run("list sorted by name", func(t *testing.T) {
		result, total, err, _ := c.ListAspectClasses(model.AspectClassListOptions{SortBy: "name.asc"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 3 {
			t.Error("unexpected total", total)
		}
		if !reflect.DeepEqual(names(result), []string{"air updated", "fan", "water"}) {
			t.Error("unexpected result", names(result))
		}
	})

	t.Run("list sorted by name desc", func(t *testing.T) {
		result, _, err, _ := c.ListAspectClasses(model.AspectClassListOptions{SortBy: "name.desc"})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(names(result), []string{"water", "fan", "air updated"}) {
			t.Error("unexpected result", names(result))
		}
	})

	t.Run("list with search", func(t *testing.T) {
		result, total, err, _ := c.ListAspectClasses(model.AspectClassListOptions{Search: "wat", SortBy: "name.asc"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 1 {
			t.Error("unexpected total", total)
		}
		if !reflect.DeepEqual(names(result), []string{"water"}) {
			t.Error("unexpected result", names(result))
		}
	})

	t.Run("list with ids", func(t *testing.T) {
		result, total, err, _ := c.ListAspectClasses(model.AspectClassListOptions{Ids: []string{fanClass, waterClass}, SortBy: "name.asc"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 2 {
			t.Error("unexpected total", total)
		}
		if !reflect.DeepEqual(names(result), []string{"fan", "water"}) {
			t.Error("unexpected result", names(result))
		}
	})

	t.Run("list with paging", func(t *testing.T) {
		result, total, err, _ := c.ListAspectClasses(model.AspectClassListOptions{Limit: 1, Offset: 1, SortBy: "name.asc"})
		if err != nil {
			t.Error(err)
			return
		}
		//total counts all matching elements, not the returned page
		if total != 3 {
			t.Error("unexpected total", total)
		}
		if !reflect.DeepEqual(names(result), []string{"fan"}) {
			t.Error("unexpected result", names(result))
		}
	})

	t.Run("validate", func(t *testing.T) {
		err, _ := c.ValidateAspectClass(models.AspectClass{Id: airClass, Name: "air"})
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("validate missing id", func(t *testing.T) {
		err, code := c.ValidateAspectClass(models.AspectClass{Name: "air"})
		if err == nil {
			t.Error("expected error for missing id")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})

	t.Run("validate invalid id", func(t *testing.T) {
		err, code := c.ValidateAspectClass(models.AspectClass{Id: "air", Name: "air"})
		if err == nil {
			t.Error("expected error for id without urn prefix")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})

	t.Run("validate missing name", func(t *testing.T) {
		err, code := c.ValidateAspectClass(models.AspectClass{Id: airClass})
		if err == nil {
			t.Error("expected error for missing name")
			return
		}
		if code != http.StatusBadRequest {
			t.Error("unexpected code", code)
		}
	})

	t.Run("write without admin rights", func(t *testing.T) {
		_, err, code := c.SetAspectClass(userToken, models.AspectClass{Id: models.URN_PREFIX + "aspect-class:forbidden", Name: "forbidden"})
		if err == nil {
			t.Error("expected error for non-admin token")
			return
		}
		if code != http.StatusUnauthorized {
			t.Error("unexpected code", code)
		}
	})

	t.Run("delete without admin rights", func(t *testing.T) {
		err, code := c.DeleteAspectClass(userToken, fanClass)
		if err == nil {
			t.Error("expected error for non-admin token")
			return
		}
		if code != http.StatusUnauthorized {
			t.Error("unexpected code", code)
		}
	})

	t.Run("validate delete", func(t *testing.T) {
		err, _ := c.ValidateAspectClassDelete(fanClass)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		err, _ := c.DeleteAspectClass(client.InternalAdminToken, fanClass)
		if err != nil {
			t.Error(err)
			return
		}
		_, err, code := c.GetAspectClass(fanClass)
		if err == nil {
			t.Error("expected the aspect-class to be gone")
			return
		}
		if code != http.StatusNotFound {
			t.Error("unexpected code", code)
		}
		_, total, err, _ := c.ListAspectClasses(model.AspectClassListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 2 {
			t.Error("unexpected total", total)
		}
	})
}

func names(list []models.AspectClass) (result []string) {
	result = []string{}
	for _, e := range list {
		result = append(result, e.Name)
	}
	return result
}
