/*
 * Copyright 2025 InfAI (CC SES)
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

package tests

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestDefaultDeviceAttribute(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.SyncLockDuration = time.Second.String()
	config.Debug = true

	config, err = docker.NewEnv(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	c := client.NewClient("http://localhost:"+config.ServerPort, nil)

	var dt models.DeviceType

	t.Run("create device-type", func(t *testing.T) {
		// Create Protocol
		protocol, err, _ := c.SetProtocol(client.InternalAdminToken, models.Protocol{
			Id:      "urn:test:protocol:test",
			Name:    "test",
			Handler: "test",
			ProtocolSegments: []models.ProtocolSegment{
				{
					Id:   "urn:test:protocol:test:1",
					Name: "test",
				},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}

		dt, err, _ = c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Name: "test",
			Services: []models.Service{
				{
					LocalId:     "s",
					Name:        "s",
					Interaction: models.REQUEST,
					ProtocolId:  protocol.Id,
				},
			},
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
	})

	var d1, d2, d3, d4 models.Device

	t.Run("create devices", func(t *testing.T) {
		d1, err, _ = c.CreateDevice(client.InternalAdminToken, models.Device{
			LocalId:      "d1",
			Name:         "d1",
			Attributes:   nil,
			DeviceTypeId: dt.Id,
		})
		if err != nil {
			t.Error(err)
			return
		}

		d2, err, _ = c.CreateDevice(client.InternalAdminToken, models.Device{
			LocalId:      "d2",
			Name:         "d2",
			DeviceTypeId: dt.Id,
			Attributes: []models.Attribute{
				{
					Key:    "a1",
					Value:  "1",
					Origin: "test",
				},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}

		d3, err, _ = c.CreateDevice(client.InternalAdminToken, models.Device{
			LocalId:      "d3",
			Name:         "d3",
			DeviceTypeId: dt.Id,
			Attributes: []models.Attribute{
				{
					Key:    "a2",
					Value:  "2",
					Origin: "test",
				},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}

		const SecondOwnerToken = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJzZWNvbmRPd25lciIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwibm9uY2UiOiI5MmM0M2M5NS03NWIwLTQ2Y2YtODBhZS00NWRkOTczYjRiN2YiLCJhdXRoX3RpbWUiOjE1NDY1MDcwMDksInNlc3Npb25fc3RhdGUiOiI1ZGY5MjhmNC0wOGYwLTRlYjktOWI2MC0zYTBhZTIyZWZjNzMiLCJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbIioiXSwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbInVzZXIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJtYXN0ZXItcmVhbG0iOnsicm9sZXMiOlsidmlldy1yZWFsbSIsInZpZXctaWRlbnRpdHktcHJvdmlkZXJzIiwibWFuYWdlLWlkZW50aXR5LXByb3ZpZGVycyIsImltcGVyc29uYXRpb24iLCJjcmVhdGUtY2xpZW50IiwibWFuYWdlLXVzZXJzIiwicXVlcnktcmVhbG1zIiwidmlldy1hdXRob3JpemF0aW9uIiwicXVlcnktY2xpZW50cyIsInF1ZXJ5LXVzZXJzIiwibWFuYWdlLWV2ZW50cyIsIm1hbmFnZS1yZWFsbSIsInZpZXctZXZlbnRzIiwidmlldy11c2VycyIsInZpZXctY2xpZW50cyIsIm1hbmFnZS1hdXRob3JpemF0aW9uIiwibWFuYWdlLWNsaWVudHMiLCJxdWVyeS1ncm91cHMiXX0sImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInJvbGVzIjpbInVzZXIiXX0.cq8YeUuR0jSsXCEzp634fTzNbGkq_B8KbVrwBPgceJ4`

		d4, err, _ = c.CreateDevice(SecondOwnerToken, models.Device{
			LocalId:      "d4",
			Name:         "d4",
			Attributes:   nil,
			DeviceTypeId: dt.Id,
		})
		if err != nil {
			t.Error(err)
			return
		}

	})

	t.Run("set default device attribute", func(t *testing.T) {
		err, _ = c.SetDefaultDeviceAttributes(client.InternalAdminToken, []models.Attribute{
			{
				Key:   "a2",
				Value: "da2",
			},
			{
				Key:    "a3",
				Value:  "da3",
				Origin: "default",
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	expected := []models.Device{
		{
			Id:           d1.Id,
			LocalId:      d1.LocalId,
			Name:         d1.Name,
			DeviceTypeId: d1.DeviceTypeId,
			OwnerId:      d1.OwnerId,
			Attributes: []models.Attribute{
				{
					Key:    "a2",
					Value:  "da2",
					Origin: "default",
				},
				{
					Key:    "a3",
					Value:  "da3",
					Origin: "default",
				},
			},
		},
		{
			Id:           d2.Id,
			LocalId:      d2.LocalId,
			Name:         d2.Name,
			DeviceTypeId: d2.DeviceTypeId,
			OwnerId:      d2.OwnerId,
			Attributes: []models.Attribute{
				{
					Key:    "a1",
					Value:  "1",
					Origin: "test",
				},
				{
					Key:    "a2",
					Value:  "da2",
					Origin: "default",
				},
				{
					Key:    "a3",
					Value:  "da3",
					Origin: "default",
				},
			},
		},
		{
			Id:           d3.Id,
			LocalId:      d3.LocalId,
			Name:         d3.Name,
			DeviceTypeId: d3.DeviceTypeId,
			OwnerId:      d3.OwnerId,
			Attributes: []models.Attribute{
				{
					Key:    "a2",
					Value:  "2",
					Origin: "test",
				},
				{
					Key:    "a3",
					Value:  "da3",
					Origin: "default",
				},
			},
		},
		{
			Id:           d4.Id,
			LocalId:      d4.LocalId,
			Name:         d4.Name,
			DeviceTypeId: d4.DeviceTypeId,
			OwnerId:      d4.OwnerId,
		},
	}
	expectedExtended := []models.ExtendedDevice{
		{
			Device:         expected[0],
			DisplayName:    d1.Name,
			DeviceTypeName: dt.Name,
			Shared:         false,
			Permissions:    models.Permissions{Read: true, Write: true, Execute: true, Administrate: true},
		},
		{
			Device:         expected[1],
			DisplayName:    d2.Name,
			DeviceTypeName: dt.Name,
			Shared:         false,
			Permissions:    models.Permissions{Read: true, Write: true, Execute: true, Administrate: true},
		},
		{
			Device:         expected[2],
			DisplayName:    d3.Name,
			DeviceTypeName: dt.Name,
			Shared:         false,
			Permissions:    models.Permissions{Read: true, Write: true, Execute: true, Administrate: true},
		},
		{
			Device:         expected[3],
			DisplayName:    d4.Name,
			DeviceTypeName: dt.Name,
			Shared:         true,
			Permissions:    models.Permissions{Read: true, Write: true, Execute: true, Administrate: true},
		},
	}

	t.Run("check default device attribute", func(t *testing.T) {
		t.Run("list extended", func(t *testing.T) {
			result, _, err, _ := c.ListExtendedDevices(client.InternalAdminToken, client.ExtendedDeviceListOptions{SortBy: "name.asc"})
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(result, expectedExtended) {
				t.Errorf("\nr=%#v\ne=%#v\n", result, expectedExtended)
				return
			}
		})
		t.Run("list", func(t *testing.T) {
			result, err, _ := c.ListDevices(client.InternalAdminToken, client.DeviceListOptions{SortBy: "name.asc"})
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\nr=%#v\ne=%#v\n", result, expected)
				return
			}
		})
		t.Run("get", func(t *testing.T) {
			for i, d := range expected {
				t.Run(d.Name, func(t *testing.T) {
					result, err, _ := c.ReadDevice(d.Id, client.InternalAdminToken, models.Read)
					if err != nil {
						t.Error(err)
						return
					}
					if !reflect.DeepEqual(result, expected[i]) {
						t.Errorf("\nr=%#v\ne=%#v\n", result, expected)
						return
					}
				})
			}
		})
		t.Run("get extended", func(t *testing.T) {
			for i, d := range expected {
				t.Run(d.Name, func(t *testing.T) {
					result, err, _ := c.ReadExtendedDevice(d.Id, client.InternalAdminToken, models.Read, false)
					if err != nil {
						t.Error(err)
						return
					}
					if !reflect.DeepEqual(result, expectedExtended[i]) {
						t.Errorf("\nr=%#v\ne=%#v\n", result, expected)
						return
					}
				})
			}
		})
	})
}
