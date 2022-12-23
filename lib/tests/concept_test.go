/*
 * Copyright 2022 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/ory/dockertest/v3"
	"log"
	"sync"
	"testing"
)

func TestConceptValidation(t *testing.T) {
	conf, err := config.Load("../../config.json")
	if err != nil {
		log.Println("ERROR: unable to load config: ", err)
		t.Error(err)
		return
	}
	conf.FatalErrHandler = t.Fatal
	conf.MongoReplSet = false
	conf.Debug = true

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Println("Could not connect to docker:", err)
		t.Error(err)
		return
	}

	_, ip, err := docker.MongoDB(pool, ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	conf.MongoUrl = "mongodb://" + ip + ":27017"

	db, err := database.New(conf)
	if err != nil {
		log.Println("ERROR: unable to connect to database", err)
		t.Error(err)
		return
	}
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		<-ctx.Done()
		db.Disconnect()
		if wg != nil {
			wg.Done()
		}
	}()

	ctrl, err := controller.New(conf, db, nil, nil)
	if err != nil {
		db.Disconnect()
		log.Println("ERROR: unable to start control", err)
		t.Error(err)
		return
	}

	err = ctrl.SetCharacteristic(models.Characteristic{
		Id:   "c",
		Name: "c",
	}, "")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("no characteristic", testValidateConcept(ctrl, false, models.Concept{
		Id:                   "v",
		Name:                 "v",
		CharacteristicIds:    nil,
		BaseCharacteristicId: "",
	}))

	t.Run("missing base characteristic", testValidateConcept(ctrl, true, models.Concept{
		Id:                   "v",
		Name:                 "v",
		CharacteristicIds:    []string{"c"},
		BaseCharacteristicId: "",
	}))

	t.Run("with base characteristic", testValidateConcept(ctrl, false, models.Concept{
		Id:                   "v",
		Name:                 "v",
		CharacteristicIds:    []string{"c"},
		BaseCharacteristicId: "c",
	}))

	t.Run("unknown characteristic", testValidateConcept(ctrl, true, models.Concept{
		Id:                   "v",
		Name:                 "v",
		CharacteristicIds:    []string{"c", "unknown"},
		BaseCharacteristicId: "c",
	}))

	t.Run("base characteristic not in list", testValidateConcept(ctrl, true, models.Concept{
		Id:                   "v",
		Name:                 "v",
		CharacteristicIds:    []string{"c"},
		BaseCharacteristicId: "unknown",
	}))
}

func testValidateConcept(ctrl *controller.Controller, expectError bool, concept models.Concept) func(t *testing.T) {
	return func(t *testing.T) {
		err, _ := ctrl.ValidateConcept(concept)
		if (err != nil) != expectError {
			t.Error(expectError, err)
		}
	}
}
