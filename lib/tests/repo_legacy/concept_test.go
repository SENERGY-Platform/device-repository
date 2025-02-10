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

package repo_legacy

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/device-repository/lib/tests/repo_legacy/testenv"
	"github.com/SENERGY-Platform/models/go/models"
	permclient "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"log"
	"reflect"
	"sync"
	"testing"
)

func testConceptList(t *testing.T, conf configuration.Config) {
	characteristics := []models.Characteristic{
		{
			Id:   "c1",
			Name: "c1",
		},
		{
			Id:   "c2",
			Name: "c2",
		},
		{
			Id:   "c3",
			Name: "c3",
		},
		{
			Id:   "c4",
			Name: "c4",
		},
		{
			Id:   "c5",
			Name: "c5",
		},
	}

	concepts := []models.Concept{
		{
			Id:   "co1",
			Name: "co1",
		},
		{
			Id:                   "co2",
			Name:                 "co2",
			CharacteristicIds:    []string{"c2", "c3"},
			BaseCharacteristicId: "c2",
		},
		{
			Id:                   "co3",
			Name:                 "co3",
			CharacteristicIds:    []string{"c3", "c4"},
			BaseCharacteristicId: "c3",
		},
		{
			Id:                   "co4",
			Name:                 "co4",
			CharacteristicIds:    []string{"c1"},
			BaseCharacteristicId: "c1",
		},
	}

	conceptsWithCharacteristics := []models.ConceptWithCharacteristics{
		{
			Id:              "co1",
			Name:            "co1",
			Characteristics: []models.Characteristic{},
		},
		{
			Id:                   "co2",
			Name:                 "co2",
			Characteristics:      []models.Characteristic{characteristics[1], characteristics[2]},
			BaseCharacteristicId: "c2",
		},
		{
			Id:                   "co3",
			Name:                 "co3",
			Characteristics:      []models.Characteristic{characteristics[2], characteristics[3]},
			BaseCharacteristicId: "c3",
		},
		{
			Id:                   "co4",
			Name:                 "co4",
			Characteristics:      []models.Characteristic{characteristics[0]},
			BaseCharacteristicId: "c1",
		},
	}

	c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
	t.Run("create characteristics", func(t *testing.T) {
		for _, characteristic := range characteristics {
			_, err, _ := c.SetCharacteristic(AdminToken, characteristic)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	t.Run("create concepts", func(t *testing.T) {
		for _, concept := range concepts {
			_, err, _ := c.SetConcept(AdminToken, concept)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	t.Run("list all concepts", func(t *testing.T) {
		list, total, err, _ := c.ListConcepts(client.ConceptListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 4 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, concepts) {
			t.Error(list)
			return
		}
	})
	t.Run("list all concepts with characteristics", func(t *testing.T) {
		list, total, err, _ := c.ListConceptsWithCharacteristics(client.ConceptListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 4 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, conceptsWithCharacteristics) {
			t.Errorf("\na=%#v\ne=%#v\n", list, conceptsWithCharacteristics)
			return
		}
	})

	t.Run("search co3 concepts", func(t *testing.T) {
		list, total, err, _ := c.ListConcepts(client.ConceptListOptions{Search: "co3"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 1 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.Concept{concepts[2]}) {
			t.Error(list)
			return
		}
	})
	t.Run("search co3 concepts with characteristic", func(t *testing.T) {
		list, total, err, _ := c.ListConceptsWithCharacteristics(client.ConceptListOptions{Search: "co3"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 1 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.ConceptWithCharacteristics{conceptsWithCharacteristics[2]}) {
			t.Error(list)
			return
		}
	})

	t.Run("list co2,co4 concepts", func(t *testing.T) {
		list, total, err, _ := c.ListConcepts(client.ConceptListOptions{Ids: []string{"co2", "co4"}})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 2 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.Concept{concepts[1], concepts[3]}) {
			t.Error(list)
			return
		}
	})
	t.Run("list co2,co4 concepts with characteristics", func(t *testing.T) {
		list, total, err, _ := c.ListConceptsWithCharacteristics(client.ConceptListOptions{Ids: []string{"co2", "co4"}})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 2 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.ConceptWithCharacteristics{conceptsWithCharacteristics[1], conceptsWithCharacteristics[3]}) {
			t.Error(list)
			return
		}
	})
}

func TestConceptValidation(t *testing.T) {
	conf, err := configuration.Load("../../../config.json")
	if err != nil {
		log.Println("ERROR: unable to load config: ", err)
		t.Error(err)
		return
	}
	conf.Debug = true

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, ip, err := docker.MongoDB(ctx, wg)
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

	pc, err := permclient.NewTestClient(ctx)
	if err != nil {
		db.Disconnect()
		log.Println("ERROR: unable to start test permission v2", err)
		t.Error(err)
		return
	}
	ctrl, err := controller.New(conf, db, testenv.VoidProducerMock{}, pc)
	if err != nil {
		db.Disconnect()
		log.Println("ERROR: unable to start control", err)
		t.Error(err)
		return
	}

	_, err, _ = ctrl.SetCharacteristic(AdminToken, models.Characteristic{
		Id:   "c",
		Name: "c",
		Type: models.String,
	})
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
