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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
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

	conf, err := config.Load("../../../config.json")
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
	err = m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: "hid1", Name: "h1", DeviceIds: []string{"a", "b"}}})
	if err != nil {
		t.Fatal(err)
	}

	timeout, _ = context.WithTimeout(ctx, 10*time.Second)
	err = m.SetHub(timeout, model.HubWithConnectionState{Hub: models.Hub{Id: "hid2", Name: "h2", DeviceIds: []string{"b", "c"}}})
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
