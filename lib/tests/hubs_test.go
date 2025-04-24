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
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestHubResetOnDeviceDelete(t *testing.T) {
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
	config.DisableStrictValidationForTesting = true

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

	t.Run("create protocols", func(t *testing.T) {
		_, err, _ := c.SetProtocol(client.InternalAdminToken, models.Protocol{
			Id:      "p1",
			Name:    "p1",
			Handler: "p1",
			ProtocolSegments: []models.ProtocolSegment{{
				Id:   "ps1",
				Name: "ps1",
			}},
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("create device-types", func(t *testing.T) {
		_, err, _ := c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Id:   "dt1",
			Name: "dt1",
			Services: []models.Service{
				{
					Id:          "s1",
					LocalId:     "s1",
					Name:        "s1",
					Interaction: models.REQUEST,
					ProtocolId:  "p1",
				},
			},
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
	})

	deviceIds := []string{}
	t.Run("create devices", func(t *testing.T) {
		for i := range 100 {
			id := "d" + strconv.Itoa(i)
			deviceIds = append(deviceIds, id)
			_, err, _ := c.SetDevice(client.InternalAdminToken, models.Device{
				Id:           id,
				LocalId:      id,
				Name:         id,
				DeviceTypeId: "dt1",
			}, client.DeviceUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("create hubs", func(t *testing.T) {
		_, err, _ := c.SetHub(client.InternalAdminToken, models.Hub{
			Id:             "h1",
			Name:           "h1",
			Hash:           "foo",
			DeviceLocalIds: deviceIds[1:],
			DeviceIds:      deviceIds[1:],
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetHub(client.InternalAdminToken, models.Hub{
			Id:             "h2",
			Name:           "h2",
			Hash:           "bar",
			DeviceLocalIds: []string{deviceIds[0]},
			DeviceIds:      []string{deviceIds[0]},
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("delete devices", func(t *testing.T) {
		wg := sync.WaitGroup{}
		for _, id := range deviceIds[2:] {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				err, _ := c.DeleteDevice(client.InternalAdminToken, id)
				if err != nil {
					t.Error(err)
					return
				}
			}(id)
		}
		wg.Wait()
	})

	t.Run("check hubs", func(t *testing.T) {
		hubs, err, _ := c.ListHubs(client.InternalAdminToken, client.HubListOptions{SortBy: "name.asc"})
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(hubs, []models.Hub{
			{
				Id:             "h1",
				Name:           "h1",
				Hash:           "",
				DeviceLocalIds: []string{deviceIds[1]},
				DeviceIds:      []string{deviceIds[1]},
				OwnerId:        "dd69ea0d-f553-4336-80f3-7f4567f85c7b",
			},
			{
				Id:             "h2",
				Name:           "h2",
				Hash:           "bar",
				DeviceLocalIds: []string{deviceIds[0]},
				DeviceIds:      []string{deviceIds[0]},
				OwnerId:        "dd69ea0d-f553-4336-80f3-7f4567f85c7b",
			},
		}) {
			t.Errorf("%#v", hubs)
		}
	})
}
