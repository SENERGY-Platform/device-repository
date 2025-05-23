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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestMongoDeviceType(t *testing.T) {
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

	timeout, _ := context.WithTimeout(ctx, 2*time.Second)
	_, exists, err := m.GetDeviceType(timeout, "does_not_exist")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device type should not exist")
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	err = m.SetDeviceType(timeout, models.DeviceType{
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
	}, func(deviceType models.DeviceType) error { return nil })
	if err != nil {
		t.Error(err)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	err = m.SetDeviceType(timeout, models.DeviceType{
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
	}, func(deviceType models.DeviceType) error { return nil })
	if err != nil {
		t.Error(err)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	device, exists, err := m.GetDeviceType(timeout, "foobar1")
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

	err = m.SetDeviceType(timeout, models.DeviceType{
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
	}, func(deviceType models.DeviceType) error { return nil })
	if err != nil {
		t.Error(err)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	device, exists, err = m.GetDeviceType(timeout, "foobar1")
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

	var listDeviceTypesV2 = func(ctx context.Context, limit int64, offset int64, sort string, filterCriteria []model.FilterCriteria, includeModified bool) (result []models.DeviceType, err error) {
		result, err = m.ListDeviceTypesV2(ctx, limit, offset, sort, filterCriteria, includeModified)
		if err != nil {
			return result, err
		}
		result2, _, err := m.ListDeviceTypesV3(ctx, model.DeviceTypeListOptions{
			Limit:           limit,
			Offset:          offset,
			SortBy:          sort,
			Criteria:        filterCriteria,
			IncludeModified: includeModified,
		})
		if err != nil {
			return result, err
		}
		if !reflect.DeepEqual(result, result2) {
			return result, errors.New("result != ListDeviceTypesV3()")
		}
		return result, nil
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	result, err := m.ListDeviceTypes(timeout, 100, 0, "name.asc", nil, nil, false)
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 2 {
		t.Error("unexpected result", result)
		return
	}
	if result[0].Id != "foobar1" && result[1].Id != "foobar2" {
		t.Error("unexpected result", result)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	result, err = m.ListDeviceTypes(timeout, 100, 0, "name.desc", nil, nil, false)
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 2 {
		t.Error("unexpected result", result)
		return
	}
	if result[1].Id != "foobar1" && result[0].Id != "foobar2" {
		t.Error("unexpected result", result)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	result, err = listDeviceTypesV2(timeout, 100, 0, "name.asc", nil, false)
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 2 {
		t.Error("unexpected result", result)
		return
	}
	if result[0].Id != "foobar1" && result[1].Id != "foobar2" {
		t.Error("unexpected result", result)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	result, err = listDeviceTypesV2(timeout, 100, 0, "name.desc", nil, false)
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 2 {
		t.Error("unexpected result", result)
		return
	}
	if result[1].Id != "foobar1" && result[0].Id != "foobar2" {
		t.Error("unexpected result", result)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	err = m.RemoveDeviceType(timeout, "foobar1", func(deviceType models.DeviceType) error { return nil })
	if err != nil {
		t.Error(err)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	dt, exists, err := m.GetDeviceType(timeout, "foobar1")
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

	timeout, _ := context.WithTimeout(ctx, 2*time.Second)
	_, exists, err := m.GetDeviceType(timeout, "does_not_exist")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("device type should not exist")
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	err = m.SetDeviceType(timeout, models.DeviceType{
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
	}, func(deviceType models.DeviceType) error { return nil })
	if err != nil {
		t.Error(err)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	err = m.SetDeviceType(timeout, models.DeviceType{
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
	}, func(deviceType models.DeviceType) error { return nil })
	if err != nil {
		t.Error(err)
		return
	}

	err = m.SetDeviceType(timeout, models.DeviceType{
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
	}, func(deviceType models.DeviceType) error { return nil })
	if err != nil {
		t.Error(err)
		return
	}

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	devicetypes, err := m.GetDeviceTypesByServiceId(timeout, "s1")
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

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	devicetypes, err = m.GetDeviceTypesByServiceId(timeout, "s2")
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

	timeout, _ = context.WithTimeout(ctx, 2*time.Second)
	devicetypes, err = m.GetDeviceTypesByServiceId(timeout, "s3")
	if err != nil {
		t.Error(err)
		return
	}
	if len(devicetypes) != 0 {
		t.Fatal(devicetypes)
	}
}
