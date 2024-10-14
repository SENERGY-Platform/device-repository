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
	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/source/util"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestGeneratedDeviceGroupMigration(t *testing.T) {
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

	//models.URN_PREFIX + "device-group:" + strings.TrimPrefix(deviceId, models.URN_PREFIX+"device:")
	devices := []model.DeviceWithConnectionState{
		{
			Device: models.Device{
				Id:           models.URN_PREFIX + "device:1",
				LocalId:      "1",
				Name:         "1",
				DeviceTypeId: "dt1",
				OwnerId:      userid,
			},
		},
		{
			Device: models.Device{
				Id:           models.URN_PREFIX + "device:2",
				LocalId:      "2",
				Name:         "2",
				DeviceTypeId: "dt1",
				OwnerId:      userid,
			},
		},
		{
			Device: models.Device{
				Id:           models.URN_PREFIX + "device:3",
				LocalId:      "3",
				Name:         "3",
				DeviceTypeId: "dt1",
				OwnerId:      userid,
			},
		},
		{
			Device: models.Device{
				Id:           models.URN_PREFIX + "device:4",
				LocalId:      "4",
				Name:         "4n",
				DeviceTypeId: "dt1",
				OwnerId:      userid,
			},
		},
	}
	deviceGroups := []models.DeviceGroup{
		{
			Id:   models.URN_PREFIX + "device-group:unrelated1",
			Name: "unrelated1",
		},
		{
			Id:        models.URN_PREFIX + "device-group:unrelated2",
			Name:      "unrelated2",
			DeviceIds: []string{models.URN_PREFIX + "device:1"},
		},
		{
			Id:        models.URN_PREFIX + "device-group:2",
			Name:      "2_group",
			DeviceIds: []string{models.URN_PREFIX + "device:1"},
		},
		{
			Id:        models.URN_PREFIX + "device-group:3",
			Name:      "3_group",
			DeviceIds: []string{},
		},
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
			err = db.SetRights(conf.DeviceTopic, device.Id, model.ResourceRights{
				UserRights:  map[string]model.Right{userid: {Read: true, Write: true, Execute: true, Administrate: true}},
				GroupRights: map[string]model.Right{"admin": {Read: true, Write: true, Execute: true, Administrate: true}},
			})
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
			err = db.SetRights(conf.DeviceGroupTopic, dg.Id, model.ResourceRights{
				UserRights:  map[string]model.Right{userid: {Read: true, Write: true, Execute: true, Administrate: true}},
				GroupRights: map[string]model.Right{"admin": {Read: true, Write: true, Execute: true, Administrate: true}},
			})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("migrate", func(t *testing.T) {
		conf.PermissionsV2Url = "http://" + permV2Ip + ":8080"
		err = lib.Start(ctx, wg, conf)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check", func(t *testing.T) {
		time.Sleep(10 * time.Second)
		db, err := mongo.New(conf)
		if err != nil {
			t.Error(err)
			return
		}
		defer db.Disconnect()
		list, _, err := db.ListDeviceGroups(ctx, model.DeviceGroupListOptions{
			Limit:  100,
			Offset: 0,
			SortBy: "name.asc",
		})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.DeviceGroup{
			{
				Id:        models.URN_PREFIX + "device-group:1",
				Name:      "1_group",
				DeviceIds: []string{models.URN_PREFIX + "device:1"},
				Attributes: []models.Attribute{{
					Key:    "platform/generated",
					Value:  "true",
					Origin: "device-repository",
				}},
				AutoGeneratedByDevice: models.URN_PREFIX + "device:1",
			},
			{
				Id:        models.URN_PREFIX + "device-group:2",
				Name:      "2_group",
				DeviceIds: []string{models.URN_PREFIX + "device:1"},
			},
			{
				Id:        models.URN_PREFIX + "device-group:3",
				Name:      "3_group",
				DeviceIds: []string{},
			},
			{
				Id:        models.URN_PREFIX + "device-group:4",
				Name:      "4n_group",
				DeviceIds: []string{models.URN_PREFIX + "device:4"},
				Attributes: []models.Attribute{{
					Key:    "platform/generated",
					Value:  "true",
					Origin: "device-repository",
				}},
				AutoGeneratedByDevice: models.URN_PREFIX + "device:4",
			},
			{
				Id:   models.URN_PREFIX + "device-group:unrelated1",
				Name: "unrelated1",
			},
			{
				Id:        models.URN_PREFIX + "device-group:unrelated2",
				Name:      "unrelated2",
				DeviceIds: []string{models.URN_PREFIX + "device:1"},
			},
		}
		if !reflect.DeepEqual(list, expected) {
			t.Errorf("len(list)=%v len(expected)==%v\n%#v\n%#v\n", len(list), len(expected), list, expected)
			return
		}
	})

}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
