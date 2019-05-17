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
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"github.com/ory/dockertest"
	"testing"
	"time"
)

func TestMongoHub(t *testing.T) {
	t.Parallel()
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

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err := m.GetHub(ctx, "does_not_exist")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("hub should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetHub(ctx, model.Hub{Id: "foobar", Name: "foo"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	hub, exists, err := m.GetHub(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("hub should exist")
		return
	}
	if hub.Id != "foobar" || hub.Name != "foo" {
		t.Error("unexpected result", hub)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetHub(ctx, model.Hub{Id: "foobar", Name: "foo2"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	hub, exists, err = m.GetHub(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("hub should exist")
		return
	}
	if hub.Id != "foobar" || hub.Name != "foo2" {
		t.Error("unexpected result", hub)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.RemoveHub(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err = m.GetHub(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("hub should not exist")
		return
	}
}
