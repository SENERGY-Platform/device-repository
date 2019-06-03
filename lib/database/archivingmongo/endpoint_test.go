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

package archivingmongo

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"github.com/ory/dockertest"
	"testing"
	"time"
)

func TestMongoEndpointDeleteBeforeCreate(t *testing.T) {
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
	err = m.RemoveEndpoint(ctx, "rm")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err := m.ListEndpoints(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 0 {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetEndpoint(ctx, model.Endpoint{Id: "rm"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err = m.ListEndpoints(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 0 {
		t.Error("unexpected result", result)
		return
	}

}

func TestMongoEndpoint(t *testing.T) {

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
	result, err := m.ListEndpoints(ctx, listoptions.New().Set("device", "does_not_exist").Strict())
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 0 {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetEndpoint(ctx, model.Endpoint{Id: "foobar", Device: "d1"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err = m.ListEndpoints(ctx, listoptions.New().Set("device", "d1").Strict())
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 1 || result[0].Device != "d1" || result[0].Id != "foobar" {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.RemoveEndpoint(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err = m.ListEndpoints(ctx, listoptions.New().Set("device", "does_not_exist").Strict())
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 0 {
		t.Error("unexpected result", result)
		return
	}
}
