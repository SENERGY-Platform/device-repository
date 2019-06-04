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

func TestMongoValueTypeDeleteBeforeCreate(t *testing.T) {
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
	err = m.RemoveValueType(ctx, "rm")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err := m.GetValueType(ctx, "rm")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("vt should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetValueType(ctx, model.ValueType{Id: "rm", Name: "foo", Fields: []model.FieldType{{Type: model.ValueType{Id: "sub1"}}}})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err = m.GetValueType(ctx, "rm")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("vt should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err := m.ListValueTypes(ctx, listoptions.New())
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 0 {
		t.Error("vt should not exist", result)
		return
	}

	result, err = m.ListValueTypesUsingValueType(ctx, "sub1")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 0 {
		t.Error("unexpected result", result)
		return
	}
}

func TestMongoValueType(t *testing.T) {

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
	_, exists, err := m.GetValueType(ctx, "does_not_exist")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("vt should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetValueType(ctx, model.ValueType{
		Id:     "foobar1",
		Name:   "foo1",
		Fields: []model.FieldType{{Type: model.ValueType{Id: "sub1"}}},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	vt, exists, err := m.GetValueType(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("vt should exist")
		return
	}
	if vt.Id != "foobar1" || vt.Name != "foo1" {
		t.Error("unexpected result", vt)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err := m.ListValueTypesUsingValueType(ctx, "sub1")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 1 {
		t.Error("unexpected result", result)
		return
	}
	if result[0].Id != "foobar1" {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.RemoveValueType(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err = m.ListValueTypesUsingValueType(ctx, "sub1")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 0 {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err = m.GetValueType(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("vt should not exist")
		return
	}
}