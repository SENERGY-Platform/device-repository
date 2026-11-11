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

package mongo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
)

func startConnectionStateTestMongo(t *testing.T, ctx context.Context, wg *sync.WaitGroup) *Mongo {
	t.Helper()
	conf, err := configuration.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	port, _, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	conf.MongoUrl = "mongodb://localhost:" + port
	m, err := New(conf)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// setReplaceHook installs f as the hook between read and replace for the given id and removes it when the test ends.
func setReplaceHook(t *testing.T, id string, f func(attempt int)) {
	t.Helper()
	attempt := 0
	beforeConnectionStateReplace = func(hookId string) {
		if hookId == id {
			f(attempt)
			attempt++
		}
	}
	t.Cleanup(func() { beforeConnectionStateReplace = func(string) {} })
}

func TestMongo_SetDevice_ConnectionState(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := startConnectionStateTestMongo(t, ctx, wg)

	timeout, timeoutCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer timeoutCancel()

	noSync := func(model.DeviceWithConnectionState, model.DeviceWithConnectionState) error { return nil }

	countDevices := func(t *testing.T, id string) int64 {
		t.Helper()
		count, err := m.deviceCollection().CountDocuments(timeout, bson.M{DeviceBson.Id: id})
		if err != nil {
			t.Fatal(err)
		}
		return count
	}
	getDevice := func(t *testing.T, id string) model.DeviceWithConnectionState {
		t.Helper()
		device, exists, err := m.GetDevice(timeout, id)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatalf("device %v not found", id)
		}
		return device
	}
	createDevice := func(t *testing.T, id string, state models.ConnectionState) {
		t.Helper()
		err := m.SetDevice(timeout, model.DeviceWithConnectionState{Device: models.Device{Id: id, Name: "before", OwnerId: "owner"}, ConnectionState: state}, noSync)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("new device is inserted with the given state", func(t *testing.T) {
		createDevice(t, "new", models.ConnectionStateOnline)
		device := getDevice(t, "new")
		if device.ConnectionState != models.ConnectionStateOnline || device.Name != "before" || device.DisplayName != "before" {
			t.Fatalf("%#v", device)
		}
		if count := countDevices(t, "new"); count != 1 {
			t.Fatal(count)
		}
	})

	t.Run("stale state from the caller does not overwrite the stored state", func(t *testing.T) {
		createDevice(t, "stale", models.ConnectionStateOnline)
		err := m.SetDeviceConnectionState(timeout, "stale", models.ConnectionStateOffline)
		if err != nil {
			t.Fatal(err)
		}
		err = m.SetDevice(timeout, model.DeviceWithConnectionState{Device: models.Device{Id: "stale", Name: "after", OwnerId: "owner"}, ConnectionState: models.ConnectionStateOnline}, noSync)
		if err != nil {
			t.Fatal(err)
		}
		device := getDevice(t, "stale")
		if device.ConnectionState != models.ConnectionStateOffline || device.Name != "after" || device.DisplayName != "after" {
			t.Fatalf("%#v", device)
		}
		if count := countDevices(t, "stale"); count != 1 {
			t.Fatal(count)
		}
	})

	t.Run("state change between read and replace is kept", func(t *testing.T) {
		createDevice(t, "race", models.ConnectionStateOnline)
		calls := 0
		setReplaceHook(t, "race", func(attempt int) {
			calls++
			if attempt == 0 {
				err := m.SetDeviceConnectionState(timeout, "race", models.ConnectionStateOffline)
				if err != nil {
					t.Error(err)
				}
			}
		})
		var syncedOld, syncedNew model.DeviceWithConnectionState
		err := m.SetDevice(timeout, model.DeviceWithConnectionState{Device: models.Device{Id: "race", Name: "after", OwnerId: "owner"}, ConnectionState: models.ConnectionStateOnline}, func(old model.DeviceWithConnectionState, new model.DeviceWithConnectionState) error {
			syncedOld, syncedNew = old, new
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		device := getDevice(t, "race")
		if device.ConnectionState != models.ConnectionStateOffline || device.Name != "after" {
			t.Fatalf("%#v", device)
		}
		if calls != 2 {
			t.Fatalf("expected one retry, got %v attempts", calls)
		}
		if syncedOld.Name != "before" || syncedOld.ConnectionState != models.ConnectionStateOffline {
			t.Fatalf("old passed to sync handler: %#v", syncedOld)
		}
		if syncedNew.Name != "after" || syncedNew.ConnectionState != models.ConnectionStateOffline {
			t.Fatalf("new passed to sync handler: %#v", syncedNew)
		}
		if count := countDevices(t, "race"); count != 1 {
			t.Fatal(count)
		}
	})

	t.Run("gives up after bounded attempts without writing", func(t *testing.T) {
		createDevice(t, "flapping", models.ConnectionStateOnline)
		calls := 0
		setReplaceHook(t, "flapping", func(attempt int) {
			calls++
			state := models.ConnectionStateOffline
			if attempt%2 == 1 {
				state = models.ConnectionStateOnline
			}
			err := m.SetDeviceConnectionState(timeout, "flapping", state)
			if err != nil {
				t.Error(err)
			}
		})
		err := m.SetDevice(timeout, model.DeviceWithConnectionState{Device: models.Device{Id: "flapping", Name: "after", OwnerId: "owner"}, ConnectionState: models.ConnectionStateOnline}, noSync)
		if err == nil {
			t.Fatal("expected error")
		}
		if calls != connectionStateWriteAttempts {
			t.Fatalf("expected %v attempts, got %v", connectionStateWriteAttempts, calls)
		}
		device := getDevice(t, "flapping")
		if device.Name != "before" || device.ConnectionState != models.ConnectionStateOffline {
			t.Fatalf("%#v", device)
		}
		if count := countDevices(t, "flapping"); count != 1 {
			t.Fatal(count)
		}
	})

	t.Run("document without connection_state field is updated", func(t *testing.T) {
		_, err := m.deviceCollection().InsertOne(timeout, bson.M{DeviceBson.Id: "legacy", DeviceBson.Name: "before"})
		if err != nil {
			t.Fatal(err)
		}
		err = m.SetDevice(timeout, model.DeviceWithConnectionState{Device: models.Device{Id: "legacy", Name: "after", OwnerId: "owner"}}, noSync)
		if err != nil {
			t.Fatal(err)
		}
		device := getDevice(t, "legacy")
		if device.Name != "after" || device.ConnectionState != models.ConnectionStateUnknown {
			t.Fatalf("%#v", device)
		}
		if count := countDevices(t, "legacy"); count != 1 {
			t.Fatal(count)
		}
	})

	// a recreated document carries the state the caller passed, as an insert always does
	t.Run("document deleted between read and replace is written again", func(t *testing.T) {
		createDevice(t, "deleted", models.ConnectionStateOnline)
		setReplaceHook(t, "deleted", func(attempt int) {
			if attempt == 0 {
				_, err := m.deviceCollection().DeleteOne(timeout, bson.M{DeviceBson.Id: "deleted"})
				if err != nil {
					t.Error(err)
				}
			}
		})
		var syncedOld model.DeviceWithConnectionState
		err := m.SetDevice(timeout, model.DeviceWithConnectionState{Device: models.Device{Id: "deleted", Name: "after", OwnerId: "owner"}, ConnectionState: models.ConnectionStateOffline}, func(old model.DeviceWithConnectionState, _ model.DeviceWithConnectionState) error {
			syncedOld = old
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		device := getDevice(t, "deleted")
		if device.Name != "after" || device.ConnectionState != models.ConnectionStateOffline {
			t.Fatalf("%#v", device)
		}
		if syncedOld.Id != "" {
			t.Fatalf("old passed to sync handler: %#v", syncedOld)
		}
		if count := countDevices(t, "deleted"); count != 1 {
			t.Fatal(count)
		}
	})

	t.Run("soft deleted document is replaced", func(t *testing.T) {
		createDevice(t, "softdeleted", models.ConnectionStateOnline)
		err := m.setDeleted(timeout, m.deviceCollection(), DeviceBson.Id, "softdeleted")
		if err != nil {
			t.Fatal(err)
		}
		err = m.SetDevice(timeout, model.DeviceWithConnectionState{Device: models.Device{Id: "softdeleted", Name: "after", OwnerId: "owner"}}, noSync)
		if err != nil {
			t.Fatal(err)
		}
		device := getDevice(t, "softdeleted")
		if device.Name != "after" {
			t.Fatalf("%#v", device)
		}
		if count := countDevices(t, "softdeleted"); count != 1 {
			t.Fatal(count)
		}
	})
}

func TestMongo_SetHub_ConnectionState(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := startConnectionStateTestMongo(t, ctx, wg)

	timeout, timeoutCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer timeoutCancel()

	noSync := func(model.HubWithConnectionState) error { return nil }

	countHubs := func(t *testing.T, id string) int64 {
		t.Helper()
		count, err := m.hubCollection().CountDocuments(timeout, bson.M{HubBson.Id: id})
		if err != nil {
			t.Fatal(err)
		}
		return count
	}
	getHub := func(t *testing.T, id string) model.HubWithConnectionState {
		t.Helper()
		hub, exists, err := m.GetHub(timeout, id)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatalf("hub %v not found", id)
		}
		return hub
	}
	createHub := func(t *testing.T, id string, state models.ConnectionState) {
		t.Helper()
		err := m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: id, Name: "before", OwnerId: "owner"}, ConnectionState: state}, noSync)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("new hub is inserted with the given state", func(t *testing.T) {
		createHub(t, "new", models.ConnectionStateOnline)
		hub := getHub(t, "new")
		if hub.ConnectionState != models.ConnectionStateOnline || hub.Name != "before" {
			t.Fatalf("%#v", hub)
		}
		if count := countHubs(t, "new"); count != 1 {
			t.Fatal(count)
		}
	})

	t.Run("stale state from the caller does not overwrite the stored state", func(t *testing.T) {
		createHub(t, "stale", models.ConnectionStateOnline)
		err := m.SetHubConnectionState(timeout, "stale", models.ConnectionStateOffline)
		if err != nil {
			t.Fatal(err)
		}
		err = m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: "stale", Name: "after", OwnerId: "owner"}, ConnectionState: models.ConnectionStateOnline}, noSync)
		if err != nil {
			t.Fatal(err)
		}
		hub := getHub(t, "stale")
		if hub.ConnectionState != models.ConnectionStateOffline || hub.Name != "after" {
			t.Fatalf("%#v", hub)
		}
		if count := countHubs(t, "stale"); count != 1 {
			t.Fatal(count)
		}
	})

	t.Run("state change between read and replace is kept", func(t *testing.T) {
		createHub(t, "race", models.ConnectionStateOnline)
		calls := 0
		setReplaceHook(t, "race", func(attempt int) {
			calls++
			if attempt == 0 {
				err := m.SetHubConnectionState(timeout, "race", models.ConnectionStateOffline)
				if err != nil {
					t.Error(err)
				}
			}
		})
		var synced model.HubWithConnectionState
		err := m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: "race", Name: "after", OwnerId: "owner"}, ConnectionState: models.ConnectionStateOnline}, func(hub model.HubWithConnectionState) error {
			synced = hub
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		hub := getHub(t, "race")
		if hub.ConnectionState != models.ConnectionStateOffline || hub.Name != "after" {
			t.Fatalf("%#v", hub)
		}
		if calls != 2 {
			t.Fatalf("expected one retry, got %v attempts", calls)
		}
		if synced.Name != "after" || synced.ConnectionState != models.ConnectionStateOffline {
			t.Fatalf("hub passed to sync handler: %#v", synced)
		}
		if count := countHubs(t, "race"); count != 1 {
			t.Fatal(count)
		}
	})

	t.Run("gives up after bounded attempts without writing", func(t *testing.T) {
		createHub(t, "flapping", models.ConnectionStateOnline)
		calls := 0
		setReplaceHook(t, "flapping", func(attempt int) {
			calls++
			state := models.ConnectionStateOffline
			if attempt%2 == 1 {
				state = models.ConnectionStateOnline
			}
			err := m.SetHubConnectionState(timeout, "flapping", state)
			if err != nil {
				t.Error(err)
			}
		})
		err := m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: "flapping", Name: "after", OwnerId: "owner"}, ConnectionState: models.ConnectionStateOnline}, noSync)
		if err == nil {
			t.Fatal("expected error")
		}
		if calls != connectionStateWriteAttempts {
			t.Fatalf("expected %v attempts, got %v", connectionStateWriteAttempts, calls)
		}
		hub := getHub(t, "flapping")
		if hub.Name != "before" || hub.ConnectionState != models.ConnectionStateOffline {
			t.Fatalf("%#v", hub)
		}
		if count := countHubs(t, "flapping"); count != 1 {
			t.Fatal(count)
		}
	})

	t.Run("document without connection state field is updated", func(t *testing.T) {
		_, err := m.hubCollection().InsertOne(timeout, bson.M{HubBson.Id: "legacy", HubBson.Name: "before"})
		if err != nil {
			t.Fatal(err)
		}
		err = m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: "legacy", Name: "after", OwnerId: "owner"}}, noSync)
		if err != nil {
			t.Fatal(err)
		}
		hub := getHub(t, "legacy")
		if hub.Name != "after" || hub.ConnectionState != models.ConnectionStateUnknown {
			t.Fatalf("%#v", hub)
		}
		if count := countHubs(t, "legacy"); count != 1 {
			t.Fatal(count)
		}
	})

	// a recreated document carries the state the caller passed, as an insert always does
	t.Run("document deleted between read and replace is written again", func(t *testing.T) {
		createHub(t, "deleted", models.ConnectionStateOnline)
		setReplaceHook(t, "deleted", func(attempt int) {
			if attempt == 0 {
				_, err := m.hubCollection().DeleteOne(timeout, bson.M{HubBson.Id: "deleted"})
				if err != nil {
					t.Error(err)
				}
			}
		})
		err := m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: "deleted", Name: "after", OwnerId: "owner"}, ConnectionState: models.ConnectionStateOffline}, noSync)
		if err != nil {
			t.Fatal(err)
		}
		hub := getHub(t, "deleted")
		if hub.Name != "after" || hub.ConnectionState != models.ConnectionStateOffline {
			t.Fatalf("%#v", hub)
		}
		if count := countHubs(t, "deleted"); count != 1 {
			t.Fatal(count)
		}
	})
}
