/*
 * Copyright 2019 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"sync"
	"testing"
	"time"
)

func TestMongo_GetHubsByDeviceId(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := configuration.Load("../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	port, _, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	conf.MongoUrl = "mongodb://localhost:" + port
	m, err := New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	timeout, _ := context.WithTimeout(ctx, 10*time.Second)
	err = m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: "hid1", Name: "h1", DeviceIds: []string{"a", "b"}}}, func(_ model.HubWithConnectionState) error { return nil })
	if err != nil {
		t.Fatal(err)
	}

	timeout, _ = context.WithTimeout(ctx, 10*time.Second)
	err = m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: "hid2", Name: "h2", DeviceIds: []string{"b", "c"}}}, func(_ model.HubWithConnectionState) error { return nil })
	if err != nil {
		t.Fatal(err)
	}

	timeout, _ = context.WithTimeout(ctx, 10*time.Second)
	hubs, err := m.GetHubsByDeviceId(timeout, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(hubs) != 1 || hubs[0].Id != "hid1" {
		t.Fatal(hubs)
	}

	timeout, _ = context.WithTimeout(ctx, 10*time.Second)
	hubs, err = m.GetHubsByDeviceId(timeout, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(hubs) != 2 {
		t.Fatal(hubs)
	}
}

func TestMongo_SetHubConnectionStates_UpdatesStoredHubs(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := configuration.Load("../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	port, _, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	conf.MongoUrl = "mongodb://localhost:" + port
	m, err := New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	timeout, timeoutCancel := context.WithTimeout(ctx, 30*time.Second)
	defer timeoutCancel()

	for _, id := range []string{"hid1", "hid2"} {
		err = m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: id, Name: id}}, func(_ model.HubWithConnectionState) error { return nil })
		if err != nil {
			t.Fatal(err)
		}
	}

	err = m.SetHubConnectionStates(timeout, map[string]string{
		"hid1": models.ConnectionStateOnline,
		"hid2": models.ConnectionStateOffline,
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]models.ConnectionState{
		"hid1": models.ConnectionStateOnline,
		"hid2": models.ConnectionStateOffline,
	}
	for id, state := range expected {
		hub, exists, err := m.GetHub(timeout, id)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatalf("hub %v not found", id)
		}
		if hub.ConnectionState != state {
			t.Errorf("hub %v: expected connection state %#v, got %#v", id, state, hub.ConnectionState)
		}
	}
}
