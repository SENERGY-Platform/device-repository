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
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"sync"
	"testing"
)

func TestMigration(t *testing.T) {
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

	t.Run("migrate", func(t *testing.T) {
		err = db.RunStartupMigrations()
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check", func(t *testing.T) {
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
				t.Errorf("expected: %v\nactual: %v\n", d, actual)
			}
		}
	})

}
