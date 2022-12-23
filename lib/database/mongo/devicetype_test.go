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
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/ory/dockertest/v3"
	"testing"
	"time"
)

func TestMongoDeviceType(t *testing.T) {

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
	_, exists, err := m.GetDeviceType(ctx, "does_not_exist")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device type should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDeviceType(ctx, models.DeviceType{
		Id:   "foobar1",
		Name: "foo1",
		Services: []models.Service{
			{
				Id: "s1",
				Inputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval1",
						},
					},
				},
				Outputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval2",
							SubContentVariables: []models.ContentVariable{
								{
									Id:               "sub1",
									Name:             "sub1_name",
									CharacteristicId: "something",
								},
								{
									Id:            "sub2",
									Name:          "sub2_name",
									UnitReference: "sub1_name",
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDeviceType(ctx, models.DeviceType{
		Id:   "foobar2",
		Name: "foo2",
		Services: []models.Service{
			{
				Id: "s2",
				Inputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval1",
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	device, exists, err := m.GetDeviceType(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("device should exist")
		return
	}
	if device.Id != "foobar1" || device.Name != "foo1" {
		t.Error("unexpected result", device)
		return
	}
	if device.Services[0].Outputs[0].ContentVariable.SubContentVariables[1].UnitReference !=
		device.Services[0].Outputs[0].ContentVariable.SubContentVariables[0].Name {
		t.Error("unexpected result", device)
		return
	}

	err = m.SetDeviceType(ctx, models.DeviceType{
		Id:   "foobar1",
		Name: "foo1changed",
		Services: []models.Service{
			{
				Id: "s1",
				Inputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval1",
						},
					},
				},
				Outputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval2",
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	device, exists, err = m.GetDeviceType(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("device should exist")
		return
	}
	if device.Id != "foobar1" || device.Name != "foo1changed" {
		t.Error("unexpected result", device)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err := m.ListDeviceTypes(ctx, 100, 0, "name.asc", nil, nil, false)
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 2 {
		t.Error("unexpected result", result)
		return
	}
	if (result[0].Id != "foobar1" && result[1].Id != "foobar1") || (result[0].Id != "foobar2" && result[1].Id != "foobar2") {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.RemoveDeviceType(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	dt, exists, err := m.GetDeviceType(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("dt should not exist", dt)
		return
	}
}

func TestMongoDeviceTypeByService(t *testing.T) {
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
	_, exists, err := m.GetDeviceType(ctx, "does_not_exist")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device type should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDeviceType(ctx, models.DeviceType{
		Id:   "foobar1",
		Name: "foo1",
		Services: []models.Service{
			{
				Id: "s1",
				Inputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval1",
						},
					},
				},
				Outputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval2",
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetDeviceType(ctx, models.DeviceType{
		Id:   "foobar2",
		Name: "foo2",
		Services: []models.Service{
			{
				Id: "s2",
				Inputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval1",
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = m.SetDeviceType(ctx, models.DeviceType{
		Id:   "foobar1",
		Name: "foo1changed",
		Services: []models.Service{
			{
				Id: "s1",
				Inputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval1",
						},
					},
				},
				Outputs: []models.Content{
					{
						ContentVariable: models.ContentVariable{
							Id: "fooval2",
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	devicetypes, err := m.GetDeviceTypesByServiceId(ctx, "s1")
	if err != nil {
		t.Error(err)
		return
	}
	if len(devicetypes) != 1 {
		t.Fatal(devicetypes)
	}
	if devicetypes[0].Id != "foobar1" {
		t.Fatal(devicetypes)
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	devicetypes, err = m.GetDeviceTypesByServiceId(ctx, "s2")
	if err != nil {
		t.Error(err)
		return
	}
	if len(devicetypes) != 1 {
		t.Fatal(devicetypes)
	}
	if devicetypes[0].Id != "foobar2" {
		t.Fatal(devicetypes)
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	devicetypes, err = m.GetDeviceTypesByServiceId(ctx, "s3")
	if err != nil {
		t.Error(err)
		return
	}
	if len(devicetypes) != 0 {
		t.Fatal(devicetypes)
	}
}
