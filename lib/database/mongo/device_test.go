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

func TestMongoDeviceUpsert(t *testing.T) {
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
	m.Disconnect()

	//test multiple connect to same server
	m, err = New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err := m.GetDevice(ctx, "does_not_exist")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDevice(ctx, model.DeviceInstance{Id: "foobar", Name: "foo", Url: "bar", DeviceType: "footype"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	device, exists, err := m.GetDevice(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("device should exist")
		return
	}
	if device.Id != "foobar" || device.Name != "foo" || device.Url != "bar" || device.DeviceType != "footype" {
		t.Error("unexpected result", device)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDevice(ctx, model.DeviceInstance{Id: "foobar", Name: "foo2", Url: "bar2", DeviceType: "footype2"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	device, exists, err = m.GetDevice(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("device should exist")
		return
	}
	if device.Id != "foobar" || device.Name != "foo2" || device.Url != "bar2" || device.DeviceType != "footype2" {
		t.Error("unexpected result", device)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.RemoveDevice(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err = m.GetDevice(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device should not exist")
		return
	}
}

func TestMongoDeviceUrl(t *testing.T) {
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
	_, exists, err := m.GetDeviceByUri(ctx, "does_not_exist")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDevice(ctx, model.DeviceInstance{Id: "foobar", Name: "foo", Url: "bar", DeviceType: "footype"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	device, exists, err := m.GetDeviceByUri(ctx, "bar")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("device should exist")
		return
	}
	if device.Id != "foobar" || device.Name != "foo" || device.Url != "bar" || device.DeviceType != "footype" {
		t.Error("unexpected result", device)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDevice(ctx, model.DeviceInstance{Id: "foobar", Name: "foo2", Url: "bar2", DeviceType: "footype2"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err = m.GetDeviceByUri(ctx, "bar")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	device, exists, err = m.GetDeviceByUri(ctx, "bar2")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("device should exist")
		return
	}
	if device.Id != "foobar" || device.Name != "foo2" || device.Url != "bar2" || device.DeviceType != "footype2" {
		t.Error("unexpected result", device)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.RemoveDevice(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err = m.GetDeviceByUri(ctx, "bar2")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device should not exist")
		return
	}
}

func TestMongoDeviceList(t *testing.T) {
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
	m.Disconnect()

	//test multiple connect to same server
	m, err = New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	result, err := m.ListDevicesOfDeviceType(ctx, "foo")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 0 {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err = m.ListDevicesWithHub(ctx, "foo")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 0 {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDevice(ctx, model.DeviceInstance{Id: "foobar", Name: "foo", Url: "bar", DeviceType: "footype", Gateway: "foohub"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDevice(ctx, model.DeviceInstance{Id: "foobar2", Name: "foo", Url: "bar2", DeviceType: "footype2", Gateway: "foohub2"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err = m.ListDevicesOfDeviceType(ctx, "footype")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 1 || result[0].Id != "foobar" {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err = m.ListDevicesWithHub(ctx, "foohub")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 1 || result[0].Id != "foobar" {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDevice(ctx, model.DeviceInstance{Id: "foobar3", Name: "foo", Url: "bar3", DeviceType: "footype", Gateway: "foohub"})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err = m.ListDevicesOfDeviceType(ctx, "footype")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 2 || (result[0].Id != "foobar" && result[1].Id != "foobar") || (result[0].Id != "foobar3" && result[1].Id != "foobar3") {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err = m.ListDevicesWithHub(ctx, "foohub")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 2 || (result[0].Id != "foobar" && result[1].Id != "foobar") || (result[0].Id != "foobar3" && result[1].Id != "foobar3") {
		t.Error("unexpected result", result)
		return
	}
}

func TestMongoDeviceTransaction(t *testing.T) {
	skipMsg := `needs a prepared clean database with replSet configured:
		docker run --name mongo -p 27017:27017 -d mongo:4.1.11 mongod --replSet rs0
		docker exec -it mongo mongo
		> rs.initiate({"_id" : "rs0","members" : [{"_id" : 0,"host" : "localhost:27017"}]})
	`
	t.Skip(skipMsg)
	t.Parallel()
	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	conf.MongoUrl = "mongodb://localhost:27017" //expect prepared mongodb server with replSet on this address
	m, err := New(conf)
	if err != nil {
		t.Error(err)
		return
	}
	m.Disconnect()

	//test multiple connect to same server
	m, err = New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	transaction, finish, err := m.Transaction(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	err = m.SetDevice(transaction, model.DeviceInstance{Id: "foobar", Name: "foo", Url: "bar", DeviceType: "footype"})
	if err != nil {
		t.Error(err)
		return
	}
	err = finish(false)
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err := m.GetDevice(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device should not exist (rollback)")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	transaction, finish, err = m.Transaction(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	err = m.SetDevice(transaction, model.DeviceInstance{Id: "foobar", Name: "foo", Url: "bar", DeviceType: "footype"})
	if err != nil {
		t.Error(err)
		return
	}
	err = finish(true)
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err = m.GetDevice(ctx, "foobar")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("device should exist (commit)")
		return
	}
}
