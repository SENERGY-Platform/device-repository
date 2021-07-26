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
	"github.com/ory/dockertest/v3"
	"testing"
	"time"
)

func TestMongo_GetHubsByDeviceLocalId(t *testing.T) {
	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Error("Could not connect to docker: ", err)
		return
	}
	closer, port, _, err := MongoTestServer(pool)
	if err != nil {
		t.Error(err)
		return
	}
	if true {
		defer closer()
	}
	conf.MongoUrl = "mongodb://localhost:" + port
	m, err := New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = m.SetHub(ctx, model.Hub{Id: "hid1", Name: "h1", DeviceLocalIds: []string{"a", "b"}})
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = m.SetHub(ctx, model.Hub{Id: "hid2", Name: "h2", DeviceLocalIds: []string{"b", "c"}})
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	hubs, err := m.GetHubsByDeviceLocalId(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(hubs) != 1 || hubs[0].Id != "hid1" {
		t.Fatal(hubs)
	}

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	hubs, err = m.GetHubsByDeviceLocalId(ctx, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(hubs) != 2 {
		t.Fatal(hubs)
	}
}
