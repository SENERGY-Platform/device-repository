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

package tests

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/source/consumer"
	"github.com/SENERGY-Platform/device-repository/lib/source/producer"
	"github.com/SENERGY-Platform/device-repository/lib/source/util"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestPermissionMigration(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := config.Load("../../config.json")
	if err != nil {
		log.Println("ERROR: unable to load config: ", err)
		t.Error(err)
		return
	}
	conf.Debug = true

	_, ip, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	conf.MongoUrl = "mongodb://" + ip + ":27017"

	_, zkIp, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	conf.KafkaUrl, err = docker.Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}

	err = util.InitTopic(conf.KafkaUrl,
		"concepts",
		"device-groups",
		"aspects",
		"characteristics",
		"processmodel",
		"device-types",
		"hubs",
		"devices",
		"device-classes",
		"functions",
		"protocols",
		"import-types",
		"locations",
		"smart_service_releases",
		"gateway_log",
		"device_log")
	if err != nil {
		t.Error(err)
		return
	}

	_, permV2Ip, err := docker.PermissionsV2(ctx, wg, conf.MongoUrl, conf.KafkaUrl)
	if err != nil {
		t.Error(err)
		return
	}

	permissions := map[string]model.ResourceRights{
		"1": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: false, Execute: false, Administrate: true}},
			GroupRights:          nil,
			KeycloakGroupsRights: nil,
		},
		"2": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: false, Administrate: true}},
			GroupRights:          nil,
			KeycloakGroupsRights: nil,
		},
		"3": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          nil,
			KeycloakGroupsRights: nil,
		},
		"4": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          nil,
			KeycloakGroupsRights: nil,
		},
		"5": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          map[string]model.Right{"group": {Read: true, Write: false, Execute: false, Administrate: false}},
			KeycloakGroupsRights: nil,
		},
		"6": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          map[string]model.Right{"group": {Read: true, Write: true, Execute: false, Administrate: false}},
			KeycloakGroupsRights: nil,
		},
		"7": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          map[string]model.Right{"group": {Read: true, Write: true, Execute: true, Administrate: false}},
			KeycloakGroupsRights: nil,
		},
		"8": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          map[string]model.Right{"group": {Read: true, Write: true, Execute: true, Administrate: true}},
			KeycloakGroupsRights: nil,
		},
		"9": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          map[string]model.Right{"group": {Read: true, Write: true, Execute: true, Administrate: true}},
			KeycloakGroupsRights: map[string]model.Right{"kcg": {Read: true, Write: false, Execute: false, Administrate: false}},
		},
		"10": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          map[string]model.Right{"group": {Read: true, Write: true, Execute: true, Administrate: true}},
			KeycloakGroupsRights: map[string]model.Right{"kcg": {Read: true, Write: true, Execute: false, Administrate: false}},
		},
		"11": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          map[string]model.Right{"group": {Read: true, Write: true, Execute: true, Administrate: true}},
			KeycloakGroupsRights: map[string]model.Right{"kcg": {Read: true, Write: true, Execute: true, Administrate: false}},
		},
		"12": {
			UserRights:           map[string]model.Right{"user": {Read: true, Write: true, Execute: true, Administrate: true}},
			GroupRights:          map[string]model.Right{"group": {Read: true, Write: true, Execute: true, Administrate: true}},
			KeycloakGroupsRights: map[string]model.Right{"kcg": {Read: true, Write: true, Execute: true, Administrate: true}},
		},
	}

	devices := []model.DeviceWithConnectionState{}
	hubs := []model.HubWithConnectionState{}
	deviceGroups := []models.DeviceGroup{}
	for istr := range permissions {
		devices = append(devices, model.DeviceWithConnectionState{
			Device: models.Device{
				Id:           istr,
				LocalId:      istr,
				Name:         istr,
				DeviceTypeId: "dt",
				OwnerId:      "owner",
			},
			DisplayName: istr,
		})
		hubs = append(hubs, model.HubWithConnectionState{
			Hub: models.Hub{
				Id:      istr,
				Name:    istr,
				OwnerId: "owner",
			},
		})
		deviceGroups = append(deviceGroups, models.DeviceGroup{
			Id:   istr,
			Name: istr,
		})
	}

	t.Run("init", func(t *testing.T) {
		db, err := mongo.New(conf)
		if err != nil {
			t.Error(err)
			return
		}
		defer db.Disconnect()

		for _, device := range devices {
			timeout, _ := getTimeoutContext()
			err = db.SetDevice(timeout, device)
			if err != nil {
				t.Error(err)
				return
			}
			err = db.SetRights(conf.DeviceTopic, device.Id, permissions[device.Id])
			if err != nil {
				t.Error(err)
				return
			}
		}
		for _, hub := range hubs {
			timeout, _ := getTimeoutContext()
			err = db.SetHub(timeout, hub)
			if err != nil {
				t.Error(err)
				return
			}
			err = db.SetRights(conf.HubTopic, hub.Id, permissions[hub.Id])
			if err != nil {
				t.Error(err)
				return
			}
		}
		for _, dg := range deviceGroups {
			timeout, _ := getTimeoutContext()
			err = db.SetDeviceGroup(timeout, dg)
			if err != nil {
				t.Error(err)
				return
			}
			err = db.SetRights(conf.DeviceGroupTopic, dg.Id, permissions[dg.Id])
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	var db *mongo.Mongo
	t.Run("migrate", func(t *testing.T) {
		conf.PermissionsV2Url = "http://" + permV2Ip + ":8080"
		db, err = mongo.New(conf)
		if err != nil {
			t.Error(err)
			return
		}
		err = db.RunStartupMigrations()
		if err != nil {
			t.Error(err)
			return
		}
		p, err := producer.New(conf)
		if err != nil {
			t.Error(err)
			return
		}
		ctrl, err := controller.New(conf, db, p, nil)
		if err != nil {
			t.Error(err)
			return
		}
		err = consumer.Start(ctx, conf, ctrl)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		permv2 := client.New(conf.PermissionsV2Url)

		f := func(t *testing.T, topic string) {
			permList, err, _ := permv2.ListResourcesWithAdminPermission(client.InternalAdminToken, topic, client.ListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(permList) != len(permissions) {
				t.Error(len(permList), len(permissions))
				return
			}
			for _, perm := range permList {
				expectedRaw, ok := permissions[perm.Id]
				if !ok {
					t.Error("unexpected permission id", perm.Id)
					return
				}
				if expectedRaw.KeycloakGroupsRights == nil {
					expectedRaw.KeycloakGroupsRights = map[string]model.Right{}
				}
				if expectedRaw.GroupRights == nil {
					expectedRaw.GroupRights = map[string]model.Right{}
				}
				expected := expectedRaw.ToPermV2Permissions()
				if !reflect.DeepEqual(perm.ResourcePermissions, expected) {
					t.Errorf("\n%#v\n%#v\n", expected, perm.ResourcePermissions)
					return
				}

				actualDbEntry, err := db.GetRights(topic, perm.Id)
				if err != nil {
					t.Error(err)
					return
				}
				if !reflect.DeepEqual(actualDbEntry, expectedRaw) {
					t.Errorf("\n%#v\n%#v\n", expectedRaw, actualDbEntry)
					return
				}
			}
		}

		t.Run("check devices", func(t *testing.T) {
			f(t, conf.DeviceTopic)
		})

		t.Run("check hubs", func(t *testing.T) {
			f(t, conf.HubTopic)
		})

		t.Run("check device-groups", func(t *testing.T) {
			f(t, conf.DeviceGroupTopic)
		})
	})

}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
