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
	"encoding/base64"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib"
	"github.com/SENERGY-Platform/device-repository/v3/lib/client"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	permissions "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

// TestDeleteMany covers the DELETE /<resource> endpoints with a list of ids.
// The rights are removed in permissions-v2 with one request per topic, which needs a permissions-v2 image with DELETE /manage/{topic}.
func TestDeleteMany(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const ownerJwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`
	//the services do not validate the signature of tokens
	otherJwt := "Bearer " + base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`)) + "." +
		base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"otherOwner","roles":["user"],"realm_access":{"roles":["user"]}}`)) + "." +
		base64.RawURLEncoding.EncodeToString([]byte("signature"))

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	config.SyncLockDuration = time.Second.String()
	config.Debug = true

	config, err = docker.NewEnv(ctx, wg, config)
	if err != nil {
		t.Fatal(err)
	}
	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Fatal(err)
	}

	c := client.NewClient("http://localhost:"+config.ServerPort, client.Options{})
	permClient := permissions.New(config.PermissionsV2Url)
	bg := context.Background()

	protocol, err, _ := c.SetProtocol(bg, client.InternalAdminToken, models.Protocol{
		Name:             "p1",
		Handler:          "p1",
		ProtocolSegments: []models.ProtocolSegment{{Id: "ps1", Name: "ps1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	newDeviceType := func(t *testing.T, name string) models.DeviceType {
		t.Helper()
		dt, err, _ := c.SetDeviceType(bg, client.InternalAdminToken, models.DeviceType{
			Name: name,
			Services: []models.Service{{
				LocalId:     "s1",
				Name:        "s1",
				Interaction: models.REQUEST,
				ProtocolId:  protocol.Id,
			}},
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Fatal(err)
		}
		return dt
	}
	newDevice := func(t *testing.T, token string, dt models.DeviceType, name string) models.Device {
		t.Helper()
		d, err, _ := c.CreateDevice(bg, token, models.Device{LocalId: name, Name: name, DeviceTypeId: dt.Id})
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	expectCode := func(t *testing.T, err error, code int, expected int) {
		t.Helper()
		if err == nil {
			t.Fatalf("expected error with code %v, got none", expected)
		}
		if code != expected {
			t.Fatalf("expected code %v, got %v: %v", expected, code, err)
		}
	}
	deviceTypeForDevices := newDeviceType(t, "dt-for-devices")

	t.Run("things", func(t *testing.T) {
		ids := []string{}
		for _, name := range []string{"t1", "t2", "t3"} {
			thing, err, _ := c.SetThing(bg, ownerJwt, models.Thing{Name: name})
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, thing.Id)
		}
		kept, err, _ := c.SetThing(bg, ownerJwt, models.Thing{Name: "kept"})
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range ids {
			_, err, _ = permClient.GetResourceContext(bg, permissions.InternalAdminToken, config.ThingTopic, id)
			if err != nil {
				t.Fatalf("rights of %v are missing before the test: %v", id, err)
			}
		}

		//unknown ids and duplicates are ignored
		request := append(append([]string{}, ids...), "unknown-thing", ids[0])
		err, _ = c.DeleteThings(bg, ownerJwt, request)
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range ids {
			//without rights the read is denied before the missing thing is noticed
			_, err, _ := c.ReadThing(bg, ownerJwt, id)
			if err == nil {
				t.Fatalf("thing %v can still be read", id)
			}
			_, err, code := permClient.GetResourceContext(bg, permissions.InternalAdminToken, config.ThingTopic, id)
			expectCode(t, err, code, http.StatusNotFound)
		}
		_, err, _ = c.ReadThing(bg, ownerJwt, kept.Id)
		if err != nil {
			t.Fatalf("thing outside of the request was deleted: %v", err)
		}

		t.Run("empty list", func(t *testing.T) {
			err, _ := c.DeleteThings(bg, ownerJwt, []string{})
			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("devices with a foreign id", func(t *testing.T) {
		own1 := newDevice(t, ownerJwt, deviceTypeForDevices, "own1")
		own2 := newDevice(t, ownerJwt, deviceTypeForDevices, "own2")
		foreign := newDevice(t, otherJwt, deviceTypeForDevices, "foreign")

		err, code := c.DeleteDevices(bg, ownerJwt, []string{own1.Id, foreign.Id, own2.Id})
		expectCode(t, err, code, http.StatusForbidden)
		for _, d := range []struct {
			token string
			id    string
		}{{ownerJwt, own1.Id}, {ownerJwt, own2.Id}, {otherJwt, foreign.Id}} {
			_, err, _ := c.ReadDevice(bg, d.id, d.token, model.READ)
			if err != nil {
				t.Fatalf("%v was deleted by a rejected request: %v", d.id, err)
			}
		}

		err, _ = c.DeleteDevices(bg, ownerJwt, []string{own1.Id, own2.Id})
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{own1.Id, own2.Id} {
			_, err, _ := c.ReadDevice(bg, id, ownerJwt, model.READ)
			if err == nil {
				t.Fatalf("device %v can still be read", id)
			}
			_, err, code := permClient.GetResourceContext(bg, permissions.InternalAdminToken, config.DeviceTopic, id)
			expectCode(t, err, code, http.StatusNotFound)
		}
		_, err, _ = c.ReadDevice(bg, foreign.Id, otherJwt, model.READ)
		if err != nil {
			t.Fatalf("foreign device was deleted: %v", err)
		}
	})

	t.Run("devices as admin", func(t *testing.T) {
		foreign := newDevice(t, otherJwt, deviceTypeForDevices, "admin-foreign")
		withoutRights := newDevice(t, ownerJwt, deviceTypeForDevices, "admin-without-rights")
		//the single delete lets an admin delete a device that has no rights entry, the batch must do the same
		err, _ := permClient.RemoveResourceContext(bg, permissions.InternalAdminToken, config.DeviceTopic, withoutRights.Id)
		if err != nil {
			t.Fatal(err)
		}

		err, _ = c.DeleteDevices(bg, client.InternalAdminToken, []string{foreign.Id, withoutRights.Id})
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range []string{foreign.Id, withoutRights.Id} {
			_, err, code := c.ReadDevice(bg, id, client.InternalAdminToken, model.READ)
			expectCode(t, err, code, http.StatusNotFound)
		}
	})

	t.Run("device-types in use", func(t *testing.T) {
		free1 := newDeviceType(t, "free1")
		free2 := newDeviceType(t, "free2")
		used := newDeviceType(t, "used")
		device := newDevice(t, ownerJwt, used, "uses-device-type")
		ids := []string{free1.Id, used.Id, free2.Id}

		err, code := c.DeleteDeviceTypes(bg, ownerJwt, ids)
		expectCode(t, err, code, http.StatusForbidden)

		err, code = c.DeleteDeviceTypes(bg, client.InternalAdminToken, ids)
		expectCode(t, err, code, http.StatusBadRequest)
		for _, id := range ids {
			_, err, _ := c.ReadDeviceType(bg, id, client.InternalAdminToken)
			if err != nil {
				t.Fatalf("%v was deleted by a rejected request: %v", id, err)
			}
		}

		err, _ = c.DeleteDevice(bg, ownerJwt, device.Id)
		if err != nil {
			t.Fatal(err)
		}
		err, _ = c.DeleteDeviceTypes(bg, client.InternalAdminToken, ids)
		if err != nil {
			t.Fatal(err)
		}
		for _, id := range ids {
			_, err, code := c.ReadDeviceType(bg, id, client.InternalAdminToken)
			expectCode(t, err, code, http.StatusNotFound)
		}
	})

	t.Run("protocols without admin rights", func(t *testing.T) {
		//same answer as DELETE /protocols/{id}
		err, code := c.DeleteProtocols(bg, ownerJwt, []string{protocol.Id})
		expectCode(t, err, code, http.StatusUnauthorized)
		_, err, _ = c.ReadProtocol(bg, protocol.Id, client.InternalAdminToken)
		if err != nil {
			t.Fatalf("protocol was deleted by a rejected request: %v", err)
		}
	})

	t.Run("too many ids", func(t *testing.T) {
		kept, err, _ := c.SetThing(bg, ownerJwt, models.Thing{Name: "kept-by-limit"})
		if err != nil {
			t.Fatal(err)
		}
		ids := []string{kept.Id}
		for i := 1; i <= 10000; i++ {
			ids = append(ids, "unknown-"+strconv.Itoa(i))
		}
		err, code := c.DeleteThings(bg, ownerJwt, ids)
		expectCode(t, err, code, http.StatusBadRequest)
		_, err, _ = c.ReadThing(bg, ownerJwt, kept.Id)
		if err != nil {
			t.Fatalf("thing was deleted by a rejected request: %v", err)
		}

		//the limit is inclusive: exactly 10000 ids are accepted
		err, _ = c.DeleteThings(bg, ownerJwt, ids[:10000])
		if err != nil {
			t.Fatal(err)
		}
	})
}
