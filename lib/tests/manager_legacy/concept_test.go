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

package tests

import (
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/tests/manager_legacy/helper"
	"github.com/SENERGY-Platform/models/go/models"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func testConcepts(t *testing.T, conf config.Config) {
	resp, err := helper.Jwtput(adminjwt, "http://localhost:"+conf.ServerPort+"/characteristics/urn:infai:ses:characteristic:4711a?wait=true", models.Characteristic{
		Id:          "urn:infai:ses:characteristic:4711a",
		Name:        "urn:infai:ses:characteristic:4711a",
		DisplayUnit: "urn:infai:ses:characteristic:4711a",
		Type:        models.String,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	time.Sleep(10 * time.Second)

	createConcept := models.Concept{
		Name:                 "c1",
		CharacteristicIds:    []string{"urn:infai:ses:characteristic:4711a"},
		BaseCharacteristicId: "urn:infai:ses:characteristic:4711a",
	}
	resp, err = helper.Jwtpost(adminjwt, "http://localhost:"+conf.ServerPort+"/concepts?wait=true", createConcept)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	concept := models.Concept{}
	err = json.NewDecoder(resp.Body).Decode(&concept)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("create: ids are set", func(t *testing.T) {
		conceptWithIds(t, concept)
	})

	t.Run("create: concept preserved structure", func(t *testing.T) {
		conceptHasStructure(t, concept, createConcept)
	})

	t.Run("create: concept exists at semantic repo", func(t *testing.T) {
		checkConcept(t, conf, concept.Id, createConcept)
	})

	resp, err = helper.Jwtput(adminjwt, "http://localhost:"+conf.ServerPort+"/characteristics/urn:infai:ses:characteristic:4712322a?wait=true", models.Characteristic{
		Id:          "urn:infai:ses:characteristic:4712322a",
		Name:        "urn:infai:ses:characteristic:4712322a",
		DisplayUnit: "urn:infai:ses:characteristic:4712322a",
		Type:        models.String,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	time.Sleep(10 * time.Second)

	updateConcept := models.Concept{
		Id:                   concept.Id,
		Name:                 "c2",
		CharacteristicIds:    []string{"urn:infai:ses:characteristic:4712322a"},
		BaseCharacteristicId: "urn:infai:ses:characteristic:4712322a",
	}
	resp, err = helper.Jwtput(adminjwt, "http://localhost:"+conf.ServerPort+"/concepts/"+url.PathEscape(concept.Id)+"?wait=true", updateConcept)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	concept2 := models.Concept{}
	err = json.NewDecoder(resp.Body).Decode(&concept2)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("update: ids are set", func(t *testing.T) {
		conceptWithIds(t, concept2)
	})

	t.Run("update: concept preserved structure", func(t *testing.T) {
		conceptHasStructure(t, concept2, updateConcept)
	})

	t.Run("update: concept exists at semantic repo", func(t *testing.T) {
		checkConcept(t, conf, concept2.Id, updateConcept)
	})

	resp, err = helper.Jwtdelete(adminjwt, "http://localhost:"+conf.ServerPort+"/concepts/"+url.PathEscape(concept.Id)+"?wait=true")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	t.Run("delete: concept removed at semantic repo", func(t *testing.T) {
		checkConceptDelete(t, conf, concept2.Id)
	})
}

func checkConceptDelete(t *testing.T, conf config.Config, id string) {
	resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/concepts/"+url.PathEscape(id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}
}

func checkConcept(t *testing.T, conf config.Config, id string, expected models.Concept) {
	resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/concepts/"+url.PathEscape(id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	concept := models.Concept{}
	err = json.NewDecoder(resp.Body).Decode(&concept)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("concept preserved structure", func(t *testing.T) {
		conceptHasStructure(t, concept, expected)
	})
}

func conceptHasStructure(t *testing.T, concept models.Concept, expected models.Concept) {
	expected = removeIdsFromConcept(expected)
	concept = removeIdsFromConcept(concept)
	if !reflect.DeepEqual(concept, expected) {
		t.Fatal(concept, expected)
	}
}

func conceptWithIds(t *testing.T, concept models.Concept) {
	if concept.Id == "" {
		t.Fatal(concept)
	}
	for i, characteristic := range concept.CharacteristicIds {
		t.Run("concept characteristics "+strconv.Itoa(i), func(t *testing.T) {
			characteristicWithId(t, characteristic)
		})
	}
}

func characteristicWithId(t *testing.T, characteristicId string) {
	if characteristicId == "" {
		t.Fatal(characteristicId)
	}
}

func removeIdsFromConcept(concept models.Concept) models.Concept {
	concept.Id = ""

	return concept
}
