/*
 * Copyright 2019 InfAI (CC SES)
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

package semantic_legacy

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testenv"
	"github.com/SENERGY-Platform/models/go/models"
	"sync"
	"testing"
)

func TestConcepts(t *testing.T) {
	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()
	conf, ctrl, err := NewPartialMockEnv(ctx, wg, conf, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("testProduceValidConcept1withCharIdAndBaseCharId", testProduceValidConcept1withCharIdAndBaseCharId(conf))
	t.Run("testProduceValidConcept1withNoCharId", testProduceValidConcept1withNoCharId(conf))
	t.Run("testProduceValidConcept1withCharId", testProduceValidConcept1withCharId(conf))
	t.Run("testProduceValidCharacteristicDependencie", testProduceValidCharacteristicDependencie(conf))
	t.Run("testReadConcept1WithoutSubClass", testReadConcept1WithoutSubClass(ctrl))
	t.Run("testReadConcept1WithSubClass", testReadConcept1WithSubClass(ctrl))
	t.Run("testDeleteConcept1", testDeleteConcept1(conf))
}

func testProduceValidConcept1withCharIdAndBaseCharId(conf config.Config) func(t *testing.T) {
	return func(t *testing.T) {
		concept := models.Concept{}
		concept.Id = "urn:ses:infai:concept:1a1a1a1-28-11-2019"
		concept.Name = "color1"
		concept.BaseCharacteristicId = "urn:ses:infai:characteristic:544433333"
		concept.CharacteristicIds = []string{"urn:ses:infai:characteristic:544433333"}
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetConcept(testenv.AdminToken, concept)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testProduceValidConcept1withNoCharId(conf config.Config) func(t *testing.T) {
	return func(t *testing.T) {
		concept := models.Concept{}
		concept.Id = "urn:ses:infai:concept:1a1a1a"
		concept.Name = "color1"
		concept.CharacteristicIds = nil
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetConcept(testenv.AdminToken, concept)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testProduceValidConcept1withCharId(conf config.Config) func(t *testing.T) {
	return func(t *testing.T) {
		concept := models.Concept{}
		concept.Id = "urn:ses:infai:concept:1a1a1a"
		concept.Name = "color1"
		concept.CharacteristicIds = []string{"urn:ses:infai:characteristic:544433333"}
		concept.BaseCharacteristicId = "urn:ses:infai:characteristic:544433333"
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetConcept(testenv.AdminToken, concept)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testProduceValidCharacteristicDependencie(conf config.Config) func(t *testing.T) {
	return func(t *testing.T) {
		characteristic := models.Characteristic{}
		characteristic.Id = "urn:ses:infai:characteristic:544433333"
		characteristic.Name = "struct1"
		characteristic.Type = models.Structure
		characteristic.SubCharacteristics = []models.Characteristic{{
			Id:                 "urn:infai:ses:characteristic:123456789999",
			MinValue:           -2,
			MaxValue:           3,
			Value:              2.2,
			Type:               models.Float,
			Name:               "charFloat",
			SubCharacteristics: nil,
		}}
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetCharacteristic(testenv.AdminToken, characteristic)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testReadConcept1WithoutSubClass(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		concept, err, _ := con.GetConceptWithoutCharacteristics("urn:ses:infai:concept:1a1a1a")

		if err == nil {
			if concept.Id != "urn:ses:infai:concept:1a1a1a" {
				t.Fatal("wrong id")
			}
			if concept.Name != "color1" {
				t.Fatal("wrong name")
			}
			if concept.CharacteristicIds[0] != "urn:ses:infai:characteristic:544433333" {
				t.Fatal("wrong CharacteristicIds")
			}
			t.Log(concept)
		} else {
			t.Fatal(err)
		}
	}
}

func testReadConcept1WithSubClass(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		concept, err, _ := con.GetConceptWithCharacteristics("urn:ses:infai:concept:1a1a1a")
		if err == nil {
			if concept.Id != "urn:ses:infai:concept:1a1a1a" {
				t.Fatal("wrong id")
			}
			if concept.Name != "color1" {
				t.Fatal("wrong name")
			}
			if concept.BaseCharacteristicId != "urn:ses:infai:characteristic:544433333" {
				t.Fatal("wrong BaseCharacteristicId")
			}
			if concept.Characteristics[0].Id != "urn:ses:infai:characteristic:544433333" {
				t.Fatal("wrong Characteristics")
			}
			if concept.Characteristics[0].SubCharacteristics[0].Id != "urn:infai:ses:characteristic:123456789999" {
				t.Fatal("wrong SubCharacteristics")
			}
			if concept.Characteristics[0].SubCharacteristics[0].Name != "charFloat" {
				t.Fatal("wrong SubCharacteristics")
			}
			t.Log(concept)
		} else {
			t.Fatal(err)
		}
	}
}

func testDeleteConcept1(conf config.Config) func(t *testing.T) {
	return func(t *testing.T) {
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		err, _ := c.DeleteConcept(testenv.AdminToken, "urn:ses:infai:concept:1a1a1a")
		if err != nil {
			t.Fatal(err)
		}
	}
}
