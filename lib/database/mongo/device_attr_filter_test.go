/*
 * Copyright 2024 InfAI (CC SES)
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
	"reflect"
	"sync"
	"testing"
)

func TestDeviceAttributeFilter(t *testing.T) {
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

	devices := []model.DeviceWithConnectionState{
		{
			Device: models.Device{
				Id:           "0",
				LocalId:      "0",
				Name:         "0",
				Attributes:   nil,
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
		{
			Device: models.Device{
				Id:           "1",
				LocalId:      "1",
				Name:         "1",
				Attributes:   []models.Attribute{},
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
		{
			Device: models.Device{
				Id:           "2",
				LocalId:      "2",
				Name:         "2",
				Attributes:   []models.Attribute{{Key: "key1"}},
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
		{
			Device: models.Device{
				Id:           "3",
				LocalId:      "3",
				Name:         "3",
				Attributes:   []models.Attribute{{Key: "key1", Value: "value1"}},
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
		{
			Device: models.Device{
				Id:           "4",
				LocalId:      "4",
				Name:         "4",
				Attributes:   []models.Attribute{{Key: "key1", Value: "value2"}},
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
		{
			Device: models.Device{
				Id:           "5",
				LocalId:      "5",
				Name:         "5",
				Attributes:   []models.Attribute{{Key: "key2", Value: "value1"}},
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
		{
			Device: models.Device{
				Id:           "6",
				LocalId:      "6",
				Name:         "6",
				Attributes:   []models.Attribute{{Key: "key2", Value: "value2"}},
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
		{
			Device: models.Device{
				Id:           "7",
				LocalId:      "7",
				Name:         "7",
				Attributes:   []models.Attribute{{Key: "key3", Value: "value1"}},
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
		{
			Device: models.Device{
				Id:           "8",
				LocalId:      "8",
				Name:         "8",
				Attributes:   []models.Attribute{{Key: "key3", Value: "value2"}, {Key: "key4", Value: "value1"}},
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
		{
			Device: models.Device{
				Id:           "9",
				LocalId:      "9",
				Name:         "9",
				Attributes:   []models.Attribute{{Key: "key3", Value: "value3"}, {Key: "key4", Value: "value4"}, {Key: "key5", Value: "value5"}},
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
		},
	}

	t.Run("create devices", func(t *testing.T) {
		for _, d := range devices {
			err = m.SetDevice(context.Background(), d)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	tests := []struct {
		name     string
		keys     []string
		values   []string
		expected []int
	}{
		{name: "no filter", keys: nil, values: nil, expected: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{name: "key1", keys: []string{"key1"}, values: nil, expected: []int{2, 3, 4}},
		{name: "key2", keys: []string{"key2"}, values: nil, expected: []int{5, 6}},
		{name: "key3", keys: []string{"key3"}, values: nil, expected: []int{7, 8, 9}},
		{name: "key1 value1", keys: []string{"key1"}, values: []string{"value1"}, expected: []int{3}},
		{name: "key1 value2", keys: []string{"key1"}, values: []string{"value2"}, expected: []int{4}},
		{name: "key2 value1", keys: []string{"key2"}, values: []string{"value1"}, expected: []int{5}},
		{name: "key2 value2", keys: []string{"key2"}, values: []string{"value2"}, expected: []int{6}},
		{name: "key3 value1", keys: []string{"key3"}, values: []string{"value1"}, expected: []int{
			7,
			8, //by limitation of the filters
		}},
		{name: "key3 value2", keys: []string{"key3"}, values: []string{"value2"}, expected: []int{8}},
		{name: "key3 key4 value1 value2", keys: []string{"key3", "key4"}, values: []string{"value1", "value2"}, expected: []int{7, 8}},
		{name: "key3 key4", keys: []string{"key3", "key4"}, values: nil, expected: []int{7, 8, 9}},
		{name: "key3 key4 key5", keys: []string{"key3", "key4", "key5"}, values: nil, expected: []int{7, 8, 9}},
		{name: "key3 key4 value3", keys: []string{"key3", "key4"}, values: []string{"value3"}, expected: []int{9}},
		{name: "key3 key4 value3 value4", keys: []string{"key3", "key4"}, values: []string{"value3", "value4"}, expected: []int{9}},
		{name: "key3 key4 value3 value5", keys: []string{"key3", "key4"}, values: []string{"value3", "value5"}, expected: []int{9}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expected := []model.DeviceWithConnectionState{}
			for _, i := range test.expected {
				d := devices[i]
				d.DisplayName = d.Name
				expected = append(expected, d)
			}
			result, _, err := m.ListDevices(context.Background(), model.DeviceListOptions{
				AttributeKeys:   test.keys,
				AttributeValues: test.values,
			}, false)
			if err != nil {
				t.Error(err)
				return
			}
			expectedIds := []string{}
			for _, d := range expected {
				expectedIds = append(expectedIds, d.Device.Id)
			}
			actualIds := []string{}
			for _, d := range result {
				actualIds = append(actualIds, d.Device.Id)
			}
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\n%#v\n%#v\n", result, expected)
				t.Errorf("\n%#v\n%#v\n", actualIds, expectedIds)
			}

		})
	}
}
