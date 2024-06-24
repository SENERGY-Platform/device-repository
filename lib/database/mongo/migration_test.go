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
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"slices"
	"sync"
	"testing"
)

type MockProducer struct {
	Mux                         sync.Mutex
	PublishedDeviceUpdates      map[string][]models.Device
	PublishedHubUpdates         map[string][]models.Hub
	PublishedDeviceRightsUpdate map[string][]model.ResourceRights
}

func (this *MockProducer) PublishHub(hub models.Hub, userId string) (err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	if this.PublishedHubUpdates == nil {
		this.PublishedHubUpdates = map[string][]models.Hub{}
	}
	this.PublishedHubUpdates[hub.Id] = append(this.PublishedHubUpdates[hub.Id], hub)
	return nil
}

func (this *MockProducer) PublishDevice(device models.Device, userId string) (err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	if this.PublishedDeviceUpdates == nil {
		this.PublishedDeviceUpdates = map[string][]models.Device{}
	}
	this.PublishedDeviceUpdates[device.Id] = append(this.PublishedDeviceUpdates[device.Id], device)
	return nil
}

func (this *MockProducer) PublishDeviceRights(deviceId string, userId string, rights model.ResourceRights) (err error) {
	this.Mux.Lock()
	defer this.Mux.Unlock()
	if this.PublishedDeviceRightsUpdate == nil {
		this.PublishedDeviceRightsUpdate = map[string][]model.ResourceRights{}
	}
	this.PublishedDeviceRightsUpdate[deviceId] = append(this.PublishedDeviceRightsUpdate[deviceId], rights)
	return nil
}

func TestHubOwnerMigrationWithMockProducer(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	conf.RunStartupMigrations = true
	conf.SecurityImpl = "db"

	port, _, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	conf.MongoUrl = "mongodb://localhost:" + port
	db, err := New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	var devicesWithOwnerA, devicesWithOwnerB, devicesWithOwnerC, devicesWithOwnerD []models.Device
	var hub1, hub2, hub3, hub4, hub5, hub6, hub7, hub8 models.Hub

	t.Run("init test data", func(t *testing.T) {
		devicesWithOwnerA = createTestDevices(t, db, 20, "a", []string{"a"})     //20 devices with owner a
		devicesWithOwnerB = createTestDevices(t, db, 5, "b", []string{"a", "b"}) //5 devices with owner b (and owner a as admin)
		devicesWithOwnerC = createTestDevices(t, db, 5, "c", []string{"c"})      //5 devices with owner c (no other admins)
		devicesWithOwnerD = createTestDevices(t, db, 5, "", []string{"d"})       //5 devices without owner (user d has admin right --> will get owner d through deviceMigration) --> publish to devices

		hub1 = creatTestHub(t, db, "hub1", "", []string{"a"}, []models.Device{})                                              //hub without devices --> use hub admin as owner; publish to hubs
		hub2 = creatTestHub(t, db, "hub2", "a", []string{"a"}, []models.Device{})                                             //hub with owner and no devices --> no changes
		hub3 = creatTestHub(t, db, "hub3", "a", []string{"a"}, devicesWithOwnerA[0:2])                                        //hub with owner a and 2 owner a devices --> no changes
		hub4 = creatTestHub(t, db, "hub4", "", []string{"a"}, devicesWithOwnerA[2:7])                                         //hub with 5 owner a devices --> use device owner (a) as owner; publish to hubs
		hub5 = creatTestHub(t, db, "hub5", "", []string{"a"}, concatSlices(devicesWithOwnerA[7:10], devicesWithOwnerB[0:2]))  //hub with 3 owner a devices and 2 owner b devices --> use 'a' as owner for hub and 'b' devices; publish to hubs; publish to devices
		hub6 = creatTestHub(t, db, "hub6", "", []string{"a"}, concatSlices(devicesWithOwnerA[10:13], devicesWithOwnerC[0:2])) //hub with 3 owner a devices and 2 owner c devices --> use 'a' as owner and admin for hub and 'c' devices; publish to hubs; publish to devices; publish to device rights
		hub7 = creatTestHub(t, db, "hub7", "", []string{"a"}, concatSlices(devicesWithOwnerA[13:16], devicesWithOwnerD[0:2])) //hub with 3 owner a devices and 2 devices without owner --> use 'a' as owner and admin for hub and 'd' devices; publish to hubs; publish to devices; publish to device rights
		hub8 = creatTestHub(t, db, "hub8", "", []string{"a"}, devicesWithOwnerC[2:4])                                         //hub with 2 owner c devices --> use 'a' as owner and admin for hub and 'c' devices; publish to hubs; publish to devices; publish to device rights
	})

	publisher := &MockProducer{}

	t.Run("migrate", func(t *testing.T) {
		err = db.RunStartupMigrations(publisher)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check hubs", func(t *testing.T) {
		checkTestHubAndHubDevices(t, db, "a", hub1)
		checkTestHubAndHubDevices(t, db, "a", hub2)
		checkTestHubAndHubDevices(t, db, "a", hub3)
		checkTestHubAndHubDevices(t, db, "a", hub4)
		checkTestHubAndHubDevices(t, db, "a", hub5)
		checkTestHubAndHubDevices(t, db, "a", hub6)
		checkTestHubAndHubDevices(t, db, "a", hub7)
		checkTestHubAndHubDevices(t, db, "a", hub8)
	})

	t.Run("check hub publishes", func(t *testing.T) {
		t.Run("hub1", func(t *testing.T) {
			updates := publisher.PublishedHubUpdates[hub1.Id]
			if len(updates) != 1 {
				t.Error("expected 1 update")
				return
			}
			if updates[0].OwnerId != "a" {
				t.Error("expected owner update to a")
				return
			}
		})
		t.Run("hub2", func(t *testing.T) {
			updates := publisher.PublishedHubUpdates[hub2.Id]
			if len(updates) != 0 {
				t.Error("expected 0 update")
				return
			}
		})
		t.Run("hub3", func(t *testing.T) {
			updates := publisher.PublishedHubUpdates[hub3.Id]
			if len(updates) != 0 {
				t.Error("expected 0 update")
				return
			}
		})
		t.Run("hub4", func(t *testing.T) {
			updates := publisher.PublishedHubUpdates[hub4.Id]
			if len(updates) != 1 {
				t.Error("expected 1 update")
				return
			}
			if updates[0].OwnerId != "a" {
				t.Error("expected owner update to a")
				return
			}
		})
		t.Run("hub5", func(t *testing.T) {
			updates := publisher.PublishedHubUpdates[hub5.Id]
			if len(updates) != 1 {
				t.Error("expected 1 update")
				return
			}
			if updates[0].OwnerId != "a" {
				t.Error("expected owner update to a")
				return
			}
		})
		t.Run("hub6", func(t *testing.T) {
			updates := publisher.PublishedHubUpdates[hub6.Id]
			if len(updates) != 1 {
				t.Error("expected 1 update")
				return
			}
			if updates[0].OwnerId != "a" {
				t.Error("expected owner update to a")
				return
			}
		})
		t.Run("hub7", func(t *testing.T) {
			updates := publisher.PublishedHubUpdates[hub7.Id]
			if len(updates) != 1 {
				t.Error("expected 1 update")
				return
			}
			if updates[0].OwnerId != "a" {
				t.Error("expected owner update to a")
				return
			}
		})
		t.Run("hub8", func(t *testing.T) {
			updates := publisher.PublishedHubUpdates[hub8.Id]
			if len(updates) != 1 {
				t.Error("expected 1 update")
				return
			}
			if updates[0].OwnerId != "a" {
				t.Error("expected owner update to a")
				return
			}
		})
	})

	t.Run("check device publishes", func(t *testing.T) {
		t.Run("devicesWithOwnerA", func(t *testing.T) {
			for i, device := range devicesWithOwnerA {
				updates := publisher.PublishedDeviceUpdates[device.Id]
				if len(updates) != 0 {
					t.Error("expected 0 updates in devicesWithOwnerA", i)
					return
				}
				rightUpdates := publisher.PublishedDeviceRightsUpdate[device.Id]
				if len(rightUpdates) != 0 {
					t.Error("expected 0 right updates in devicesWithOwnerA", i)
					return
				}
			}
		})
		t.Run("devicesWithOwnerB", func(t *testing.T) {
			for i, device := range devicesWithOwnerB[:2] {
				updates := publisher.PublishedDeviceUpdates[device.Id]
				if len(updates) == 0 {
					t.Error("expected updates in devicesWithOwnerB", i)
					return
				}
				if updates[len(updates)-1].OwnerId != "a" {
					t.Error("expected a as owner in devicesWithOwnerB", i)
					return
				}
				rightUpdates := publisher.PublishedDeviceRightsUpdate[device.Id]
				if len(rightUpdates) != 0 {
					t.Error("expected 0 right updates in devicesWithOwnerB", i)
					return
				}
			}
			for i, device := range devicesWithOwnerB[2:] {
				updates := publisher.PublishedDeviceUpdates[device.Id]
				if len(updates) != 0 {
					t.Error("expected 0 updates in devicesWithOwnerB", i+2)
					return
				}
			}
		})
		t.Run("devicesWithOwnerC", func(t *testing.T) {
			for i, device := range devicesWithOwnerC[:4] {
				updates := publisher.PublishedDeviceUpdates[device.Id]
				if len(updates) == 0 {
					t.Error("expected updates in devicesWithOwnerC", i)
					return
				}
				if updates[len(updates)-1].OwnerId != "a" {
					t.Error("expected a as owner in devicesWithOwnerC", i)
					return
				}
				rightUpdates := publisher.PublishedDeviceRightsUpdate[device.Id]
				if len(rightUpdates) == 0 {
					t.Error("expected right updates in devicesWithOwnerC", i)
					return
				}
				if !reflect.DeepEqual(rightUpdates[len(rightUpdates)-1].UserRights, map[string]model.Right{
					"a": {Read: true, Write: true, Execute: true, Administrate: true},
					"c": {Read: true, Write: true, Execute: true, Administrate: true},
				}) {
					t.Errorf("expected a as admin in devicesWithOwnerC %v %v got %#v\n", i, device.Id, rightUpdates[len(rightUpdates)-1])
					return
				}
			}
			for i, device := range devicesWithOwnerC[4:] {
				updates := publisher.PublishedDeviceUpdates[device.Id]
				if len(updates) != 0 {
					t.Error("expected 0 updates in devicesWithOwnerC", i+4)
					return
				}
			}
		})
		t.Run("devicesWithOwnerD", func(t *testing.T) {
			for i, device := range devicesWithOwnerD[:2] {
				updates := publisher.PublishedDeviceUpdates[device.Id]
				if len(updates) == 0 {
					t.Error("expected updates in devicesWithOwnerD", i)
					return
				}
				if updates[len(updates)-1].OwnerId != "a" {
					t.Error("expected a as owner in devicesWithOwnerD", i)
					return
				}
				rightUpdates := publisher.PublishedDeviceRightsUpdate[device.Id]
				if len(rightUpdates) == 0 {
					t.Error("expected right updates in devicesWithOwnerD", i)
					return
				}
				if !reflect.DeepEqual(rightUpdates[len(rightUpdates)-1].UserRights, map[string]model.Right{
					"a": {Read: true, Write: true, Execute: true, Administrate: true},
					"d": {Read: true, Write: true, Execute: true, Administrate: true},
				}) {
					t.Error("expected a as admin in devicesWithOwnerD", i, device.Id)
					return
				}
			}
			for i, device := range devicesWithOwnerD[2:] {
				updates := publisher.PublishedDeviceUpdates[device.Id]
				if len(updates) == 0 {
					t.Error("expected updates in devicesWithOwnerD", i+2)
					return
				}
				if updates[len(updates)-1].OwnerId != "d" {
					t.Error("expected d as owner in devicesWithOwnerD", i+2)
					return
				}
				rightUpdates := publisher.PublishedDeviceRightsUpdate[device.Id]
				if len(rightUpdates) != 0 {
					t.Error("expected 0 right updates in devicesWithOwnerD", i+2)
					return
				}
			}
		})
	})
}

func concatSlices[T any](a []T, b []T) []T {
	result := []T{}
	for _, v := range a {
		result = append(result, v)
	}
	for _, v := range b {
		result = append(result, v)
	}
	return result
}

func createTestDevices(t *testing.T, db *Mongo, count int, owner string, admins []string) (devices []models.Device) {
	t.Helper()
	for range count {
		id := uuid.NewString()
		device := models.Device{
			Id:      id,
			LocalId: id,
			Name:    id,
			OwnerId: owner,
		}
		devices = append(devices, device)
		err := db.SetDevice(context.Background(), device)
		if err != nil {
			t.Error(err)
			return
		}
		userRights := map[string]model.Right{}
		for _, admin := range admins {
			userRights[admin] = model.Right{
				Read:         true,
				Write:        true,
				Execute:      true,
				Administrate: true,
			}
		}
		if len(userRights) > 0 {
			err = db.SetRights(db.config.DeviceTopic, id, model.ResourceRights{
				UserRights:  userRights,
				GroupRights: map[string]model.Right{},
			})
			if err != nil {
				t.Error(err)
				return
			}
		}
	}
	return devices
}

func creatTestHub(t *testing.T, db *Mongo, name string, owner string, admins []string, devices []models.Device) (hub models.Hub) {
	t.Helper()
	id := uuid.NewString()
	hub = models.Hub{
		Id:      id,
		Name:    name,
		OwnerId: owner,
	}
	for _, device := range devices {
		hub.DeviceIds = append(hub.DeviceIds, device.Id)
		hub.DeviceLocalIds = append(hub.DeviceLocalIds, device.Id)
	}
	err := db.SetHub(context.Background(), hub)
	if err != nil {
		t.Error(err)
		return
	}
	userRights := map[string]model.Right{}
	for _, admin := range admins {
		userRights[admin] = model.Right{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		}
	}
	if len(userRights) > 0 {
		err = db.SetRights(db.config.HubTopic, id, model.ResourceRights{
			UserRights:  userRights,
			GroupRights: map[string]model.Right{},
		})
		if err != nil {
			t.Error(err)
			return
		}
	}
	return hub
}

func checkTestHubAndHubDevices(t *testing.T, db *Mongo, expectedOwner string, hubToCheck models.Hub) {
	t.Helper()
	actual, exists, err := db.GetHub(context.Background(), hubToCheck.Id)
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("hub does not exist")
		return
	}
	//updated hub must have expected owner
	if actual.OwnerId != expectedOwner {
		t.Errorf("got owner %s, expected %s", actual.OwnerId, expectedOwner)
		return
	}
	//updated hub must still contain the same devices
	if !reflect.DeepEqual(hubToCheck.DeviceIds, actual.DeviceIds) {
		t.Errorf("got device ids %v, expected %v", actual.DeviceIds, hubToCheck.DeviceIds)
		return
	}
	if !reflect.DeepEqual(hubToCheck.DeviceIds, actual.DeviceIds) {
		t.Errorf("got device ids %v, expected %v", actual.DeviceIds, hubToCheck.DeviceIds)
		return
	}
	if !reflect.DeepEqual(hubToCheck.DeviceLocalIds, actual.DeviceLocalIds) {
		t.Errorf("got device local-ids %v, expected %v", actual.DeviceLocalIds, hubToCheck.DeviceLocalIds)
		return
	}
	//hub devices mus have expectedOwner as owner and admin
	for _, id := range hubToCheck.DeviceIds {
		device, exists, err := db.GetDevice(context.Background(), id)
		if err != nil {
			t.Error(err)
			return
		}
		if !exists {
			t.Error("hub does not exist")
			return
		}
		if !slices.Contains(hubToCheck.DeviceLocalIds, device.LocalId) {
			t.Errorf("unable to find device.LocalId (%v) in hub device local id list", device.LocalId)
			return
		}
		if device.OwnerId != expectedOwner {
			t.Errorf("got hub %s device %s owner %s, expected %s", hubToCheck.Name, device.Id, device.OwnerId, expectedOwner)
			return
		}
		kind, err := db.getInternalKind(db.config.DeviceTopic)
		if err != nil {
			t.Error(err)
			return
		}
		rights, err := db.getRights(kind, device.Id)
		if err != nil {
			t.Error(err)
			return
		}
		if !slices.Contains(rights.AdminUsers, expectedOwner) {
			t.Error("expect owner in device admins")
			return
		}
	}
}

func TestDeviceOwnerMigrationWithMockProducer(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	conf.RunStartupMigrations = true
	conf.SecurityImpl = "db"

	port, _, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	conf.MongoUrl = "mongodb://localhost:" + port
	db, err := New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("init test data", func(t *testing.T) {
		err = db.SetDevice(context.Background(), models.Device{
			Id:      "withOwner",
			LocalId: "withOwner",
			Name:    "withOwner",
			OwnerId: "withOwner",
		})
		if err != nil {
			t.Error(err)
			return
		}

		err = db.EnsureInitialRights(conf.DeviceTopic, "withOwner", "withOwner")
		if err != nil {
			t.Error(err)
			return
		}

		err = db.SetDevice(context.Background(), models.Device{
			Id:      "withOtherOwner",
			LocalId: "withOtherOwner",
			Name:    "withOtherOwner",
			OwnerId: "withOtherOwner",
		})
		if err != nil {
			t.Error(err)
			return
		}

		err = db.EnsureInitialRights(conf.DeviceTopic, "withOtherOwner", "initialOwner")
		if err != nil {
			t.Error(err)
			return
		}

		err = db.SetDevice(context.Background(), models.Device{
			Id:      "withEmptyOwner",
			LocalId: "withEmptyOwner",
			Name:    "withEmptyOwner",
			OwnerId: "",
		})
		if err != nil {
			t.Error(err)
			return
		}

		err = db.EnsureInitialRights(conf.DeviceTopic, "withEmptyOwner", "withEmptyOwner")
		if err != nil {
			t.Error(err)
			return
		}

		_, err = db.deviceCollection().ReplaceOne(ctx, bson.M{deviceIdKey: "withNullOwner"}, struct {
			Id           string `json:"id"`
			LocalId      string `json:"local_id"`
			Name         string `json:"name"`
			DeviceTypeId string `json:"device_type_id"`
		}{
			Id:      "withNullOwner",
			LocalId: "withNullOwner",
			Name:    "withNullOwner",
		}, options.Replace().SetUpsert(true))
		if err != nil {
			t.Error(err)
			return
		}

		err = db.EnsureInitialRights(conf.DeviceTopic, "withNullOwner", "withNullOwner")
		if err != nil {
			t.Error(err)
			return
		}
	})

	publisher := &MockProducer{}

	t.Run("migrate", func(t *testing.T) {
		err = db.RunStartupMigrations(publisher)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check device publishes", func(t *testing.T) {
		publisher.Mux.Lock()
		defer publisher.Mux.Unlock()
		if len(publisher.PublishedHubUpdates) != 0 {
			t.Error(publisher.PublishedHubUpdates)
			return
		}
		if len(publisher.PublishedDeviceRightsUpdate) != 0 {
			t.Error(publisher.PublishedDeviceRightsUpdate)
			return
		}

		if !reflect.DeepEqual(publisher.PublishedDeviceUpdates, map[string][]models.Device{
			"withEmptyOwner": {
				{
					Id:      "withEmptyOwner",
					LocalId: "withEmptyOwner",
					Name:    "withEmptyOwner",
					OwnerId: "withEmptyOwner",
				},
			},
			"withNullOwner": {
				{
					Id:      "withNullOwner",
					LocalId: "withNullOwner",
					Name:    "withNullOwner",
					OwnerId: "withNullOwner",
				},
			},
		}) {
			t.Errorf("%#v\n", publisher.PublishedDeviceUpdates)
			return
		}
	})

	t.Run("check devices", func(t *testing.T) {
		expected := []models.Device{
			{
				Id:      "withOwner",
				LocalId: "withOwner",
				Name:    "withOwner",
				OwnerId: "withOwner",
			},
			{
				Id:      "withOtherOwner",
				LocalId: "withOtherOwner",
				Name:    "withOtherOwner",
				OwnerId: "withOtherOwner",
			},
			{
				Id:      "withEmptyOwner",
				LocalId: "withEmptyOwner",
				Name:    "withEmptyOwner",
				OwnerId: "withEmptyOwner",
			},
			{
				Id:      "withNullOwner",
				LocalId: "withNullOwner",
				Name:    "withNullOwner",
				OwnerId: "withNullOwner",
			},
		}
		for _, d := range expected {
			actual, _, err := db.GetDevice(context.Background(), d.Id)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(actual, d) {
				t.Errorf("expected: %#v\nactual: %#v\n", d, actual)
			}
		}
	})

}
