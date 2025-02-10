/*
 * Copyright 2021 InfAI (CC SES)
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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib"
	devicerepo "github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/device-repository/lib/tests/manager_legacy/auth"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"log"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"testing"
)

func TestUserDelete(t *testing.T) {
	conf, err := configuration.Load("./../../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user1, err := auth.CreateToken("test", "user1")
	if err != nil {
		t.Error(err)
		return
	}
	user2, err := auth.CreateToken("test", "user2")
	if err != nil {
		t.Error(err)
		return
	}
	user1a, err := auth.CreateTokenWithRoles("test", "user1", []string{"admin"})
	if err != nil {
		t.Error(err)
		return
	}
	user2a, err := auth.CreateTokenWithRoles("test", "user2", []string{"admin"})
	if err != nil {
		t.Error(err)
		return
	}

	conf, err = docker.NewEnv(ctx, wg, conf)
	if err != nil {
		t.Error(err)
		return
	}
	conf.DisableStrictValidationForTesting = false

	oldBatchSize := controller.ResourcesEffectedByUserDelete_BATCH_SIZE
	controller.ResourcesEffectedByUserDelete_BATCH_SIZE = 5
	defer func() {
		controller.ResourcesEffectedByUserDelete_BATCH_SIZE = oldBatchSize
	}()

	conf.Debug = true

	err = lib.Start(ctx, wg, conf)
	if err != nil {
		t.Error(err)
		return
	}

	cache := &map[string]client.ResourcePermissions{}

	c := devicerepo.NewClient("http://localhost:"+conf.ServerPort, nil)

	dt := models.DeviceType{}
	t.Run("create device-type", func(t *testing.T) {
		protocol, err, _ := c.SetProtocol(user1a.Jwt(), models.Protocol{
			Name:             "p2",
			Handler:          "ph1",
			ProtocolSegments: []models.ProtocolSegment{{Name: "ps2"}},
		})
		if err != nil {
			t.Error(err)
			return
		}

		dt, err, _ = c.SetDeviceType(user1a.Jwt(), models.DeviceType{
			Name:          "foo",
			DeviceClassId: "dc1",
			Services: []models.Service{
				{
					Name:    "s1name",
					LocalId: "lid1",
					Inputs: []models.Content{
						{
							ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
							Serialization:     "json",
							ContentVariable: models.ContentVariable{
								Name: "v1name",
								Type: models.String,
							},
						},
					},

					ProtocolId: protocol.Id,
				},
			},
		}, model.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("create devices", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			id := strconv.Itoa(i)
			device := models.Device{
				Id:           id,
				LocalId:      id,
				Name:         id + "_name",
				Attributes:   nil,
				DeviceTypeId: dt.Id,
			}
			_, err, _ = c.SetDevice(user1a.Jwt(), device, model.DeviceUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			(*cache)[conf.DeviceTopic+"."+device.Id] = client.ResourcePermissions{
				UserPermissions: map[string]client.PermissionsMap{user1a.GetUserId(): {
					Read:         true,
					Write:        true,
					Execute:      true,
					Administrate: true,
				}},
				RolePermissions: map[string]client.PermissionsMap{"admin": {Read: true, Write: true, Execute: true, Administrate: true}},
			}
		}
		for i := 20; i < 40; i++ {
			id := strconv.Itoa(i)
			device := models.Device{
				Id:           id,
				LocalId:      id,
				Name:         id + "_name",
				Attributes:   nil,
				DeviceTypeId: dt.Id,
			}
			log.Println("test create device", id)
			_, err, _ = c.SetDevice(user2a.Jwt(), device, model.DeviceUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			(*cache)[conf.DeviceTopic+"."+device.Id] = client.ResourcePermissions{
				UserPermissions: map[string]client.PermissionsMap{user2a.GetUserId(): {
					Read:         true,
					Write:        true,
					Execute:      true,
					Administrate: true,
				}},
				RolePermissions: map[string]client.PermissionsMap{"admin": {Read: true, Write: true, Execute: true, Administrate: true}},
			}
		}
	})

	t.Run("change permissions", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			id := strconv.Itoa(i)
			err = setPermission(c.GetPermissionsClient(), user2.GetUserId(), conf.DeviceTopic, id, "rwxa", cache)
			if err != nil {
				t.Error(err)
				return
			}
		}
		for i := 20; i < 30; i++ {
			id := strconv.Itoa(i)
			err = setPermission(c.GetPermissionsClient(), user1.GetUserId(), conf.DeviceTopic, id, "rwxa", cache)
			if err != nil {
				t.Error(err)
				return
			}
		}
		for i := 5; i < 10; i++ {
			id := strconv.Itoa(i)
			err = setPermission(c.GetPermissionsClient(), user1.GetUserId(), conf.DeviceTopic, id, "rx", cache)
			if err != nil {
				t.Error(err)
				return
			}
		}
		for i := 25; i < 30; i++ {
			id := strconv.Itoa(i)
			err = setPermission(c.GetPermissionsClient(), user2.GetUserId(), conf.DeviceTopic, id, "rx", cache)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("check user1 before delete", checkUserDevices(conf, user1, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29}))
	t.Run("check user2 before delete", checkUserDevices(conf, user2, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39}))

	t.Run("delete user1", func(t *testing.T) {
		err, _ = c.DeleteUser(client.InternalAdminToken, user1.GetUserId())
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check user1 after delete", checkUserDevices(conf, user1, []int{}))
	t.Run("check user2 after delete", checkUserDevices(conf, user2, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 20, 21, 22, 23, 24, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39}))

}

func checkUserDevices(conf configuration.Config, token auth.Token, expectedDeviceIdsAsInt []int) func(t *testing.T) {
	return func(t *testing.T) {
		devices, err, _ := devicerepo.NewClient("http://localhost:"+conf.ServerPort, nil).ListDevices(token.Jwt(), devicerepo.DeviceListOptions{Limit: 100})
		if err != nil {
			t.Error(err)
			return
		}
		actualIds := []string{}
		for _, device := range devices {
			actualIds = append(actualIds, device.Id)
		}
		sort.Strings(actualIds)

		expectedIds := []string{}
		for _, intId := range expectedDeviceIdsAsInt {
			expectedIds = append(expectedIds, strconv.Itoa(intId))
		}
		sort.Strings(expectedIds)
		if !reflect.DeepEqual(actualIds, expectedIds) {
			t.Errorf("\na=%#v\ne=%#v\n", actualIds, expectedIds)
			return
		}
	}
}

func setPermission(com client.Client, userId string, kind string, id string, right string, cache *map[string]client.ResourcePermissions) error {
	userRight := client.PermissionsMap{
		Read:         false,
		Write:        false,
		Execute:      false,
		Administrate: false,
	}
	for _, r := range right {
		switch r {
		case 'r':
			userRight.Read = true
		case 'w':
			userRight.Write = true
		case 'a':
			userRight.Administrate = true
		case 'x':
			userRight.Execute = true
		default:
			return errors.New("unknown right in " + right)
		}
	}
	cacheKey := kind + "." + id
	msg, ok := (*cache)[cacheKey]
	if !ok {
		msg = client.ResourcePermissions{
			UserPermissions: map[string]client.PermissionsMap{},
			RolePermissions: map[string]client.PermissionsMap{"admin": {Read: true, Write: true, Execute: true, Administrate: true}},
		}
	}
	msg.UserPermissions[userId] = userRight
	(*cache)[cacheKey] = msg

	_, err, _ := com.SetPermission(client.InternalAdminToken, kind, id, msg)
	return err
}
