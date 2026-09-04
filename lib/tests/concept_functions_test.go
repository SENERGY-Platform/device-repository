/*
 * Copyright 2026 InfAI (CC SES)
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
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/controller/publisher"
	"github.com/SENERGY-Platform/device-repository/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
)

// TestConceptFunctions covers the pair of functions a concept owns: created with the
// concept, renamed with it while the name still is the one the convention gave, and
// established once for the existing stock by the startup migration.
func TestConceptFunctions(t *testing.T) {
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

	conceptFunctions := func(t *testing.T, conceptId string, rdfType string) []models.Function {
		t.Helper()
		functions, _, err, _ := c.ListFunctions(client.FunctionListOptions{RdfType: rdfType, Limit: 1000})
		if err != nil {
			t.Error(err)
			return nil
		}
		result := []models.Function{}
		for _, function := range functions {
			if function.ConceptId == conceptId {
				result = append(result, function)
			}
		}
		return result
	}

	// one measuring and one controlling function, both named after the concept
	assertPair := func(t *testing.T, conceptId string, conceptName string) {
		t.Helper()
		for _, rdfType := range model.ConceptFunctionRdfTypes {
			functions := conceptFunctions(t, conceptId, rdfType)
			if len(functions) != 1 {
				t.Error("expected exactly one function", rdfType, functions)
				continue
			}
			expected := model.ConceptFunctionName(rdfType, conceptName)
			if functions[0].Name != expected || functions[0].DisplayName != expected {
				t.Error("unexpected function name", rdfType, functions[0].Name, functions[0].DisplayName, expected)
			}
		}
	}

	t.Run("a new concept gets its function pair", func(t *testing.T) {
		concept, err, _ := c.SetConcept(client.InternalAdminToken, models.Concept{Name: "Temperature"})
		if err != nil {
			t.Error(err)
			return
		}
		assertPair(t, concept.Id, "Temperature")

		t.Run("the measuring function is a Get and the controlling one a Set", func(t *testing.T) {
			measuring := conceptFunctions(t, concept.Id, models.SES_ONTOLOGY_MEASURING_FUNCTION)
			controlling := conceptFunctions(t, concept.Id, models.SES_ONTOLOGY_CONTROLLING_FUNCTION)
			if len(measuring) != 1 || len(controlling) != 1 {
				t.Error("expected one of each", measuring, controlling)
				return
			}
			if measuring[0].Name != "Get-Temperature" {
				t.Error("unexpected measuring function name", measuring[0].Name)
			}
			if controlling[0].Name != "Set-Temperature" {
				t.Error("unexpected controlling function name", controlling[0].Name)
			}
		})

		t.Run("renaming the concept renames the pair", func(t *testing.T) {
			concept.Name = "Humidity"
			_, err, _ := c.SetConcept(client.InternalAdminToken, concept)
			if err != nil {
				t.Error(err)
				return
			}
			assertPair(t, concept.Id, "Humidity")
		})
	})

	t.Run("a name the user set is not overwritten", func(t *testing.T) {
		concept, err, _ := c.SetConcept(client.InternalAdminToken, models.Concept{Name: "Pressure"})
		if err != nil {
			t.Error(err)
			return
		}
		measuring := conceptFunctions(t, concept.Id, models.SES_ONTOLOGY_MEASURING_FUNCTION)
		if len(measuring) != 1 {
			t.Error("expected one measuring function", measuring)
			return
		}
		renamed := measuring[0]
		renamed.Name = "read the barometer"
		renamed.DisplayName = "read the barometer"
		_, err, _ = c.SetFunction(client.InternalAdminToken, renamed)
		if err != nil {
			t.Error(err)
			return
		}

		concept.Name = "AirPressure"
		_, err, _ = c.SetConcept(client.InternalAdminToken, concept)
		if err != nil {
			t.Error(err)
			return
		}

		measuring = conceptFunctions(t, concept.Id, models.SES_ONTOLOGY_MEASURING_FUNCTION)
		if len(measuring) != 1 || measuring[0].Name != "read the barometer" {
			t.Error("the concept rename must not touch a name the user chose", measuring)
		}
		//the controlling function still carried the convention name, so it follows
		controlling := conceptFunctions(t, concept.Id, models.SES_ONTOLOGY_CONTROLLING_FUNCTION)
		if len(controlling) != 1 || controlling[0].Name != "Set-AirPressure" {
			t.Error("unexpected controlling function", controlling)
		}
	})

	t.Run("a display name of its own survives the rename", func(t *testing.T) {
		concept, err, _ := c.SetConcept(client.InternalAdminToken, models.Concept{Name: "Brightness"})
		if err != nil {
			t.Error(err)
			return
		}
		measuring := conceptFunctions(t, concept.Id, models.SES_ONTOLOGY_MEASURING_FUNCTION)
		if len(measuring) != 1 {
			t.Error("expected one measuring function", measuring)
			return
		}
		withDisplayName := measuring[0]
		withDisplayName.DisplayName = "Helligkeit ablesen"
		_, err, _ = c.SetFunction(client.InternalAdminToken, withDisplayName)
		if err != nil {
			t.Error(err)
			return
		}

		concept.Name = "Luminance"
		_, err, _ = c.SetConcept(client.InternalAdminToken, concept)
		if err != nil {
			t.Error(err)
			return
		}

		measuring = conceptFunctions(t, concept.Id, models.SES_ONTOLOGY_MEASURING_FUNCTION)
		if len(measuring) != 1 {
			t.Error("expected one measuring function", measuring)
			return
		}
		if measuring[0].Name != "Get-Luminance" {
			t.Error("the name still matched the convention and has to follow", measuring[0].Name)
		}
		if measuring[0].DisplayName != "Helligkeit ablesen" {
			t.Error("the display name was the user's and has to stay", measuring[0].DisplayName)
		}
	})

	t.Run("startup migration", func(t *testing.T) {
		testConceptFunctionsMigration(t, config, conceptFunctions, assertPair)
	})
}

// testConceptFunctionsMigration builds the three states the migration distinguishes. The
// concepts are written through the database layer, because a concept written through the
// controller already gets its pair and none of the three would arise.
func testConceptFunctionsMigration(
	t *testing.T,
	config configuration.Config,
	conceptFunctions func(t *testing.T, conceptId string, rdfType string) []models.Function,
	assertPair func(t *testing.T, conceptId string, conceptName string),
) {
	ctx := context.Background()
	db, err := mongo.New(config)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Disconnect()

	noSync := func(models.Concept) error { return nil }
	noFunctionSync := func(models.Function) error { return nil }

	const (
		withoutFunctions     = models.URN_PREFIX + "concept:migration-none"
		withOneFunction      = models.URN_PREFIX + "concept:migration-one"
		withTwoFunctions     = models.URN_PREFIX + "concept:migration-two"
		withMatchingFunction = models.URN_PREFIX + "concept:migration-matching"
	)
	const matchingFunctionId = models.URN_PREFIX + "measuring-function:migration-matching"
	for id, name := range map[string]string{
		withoutFunctions:     "Volume",
		withOneFunction:      "Energy",
		withTwoFunctions:     "Power",
		withMatchingFunction: "Flow",
	} {
		err = db.SetConcept(ctx, models.Concept{Id: id, Name: name}, noSync)
		if err != nil {
			t.Error(err)
			return
		}
	}

	//one function, named by whoever created it rather than by the convention
	err = db.SetFunction(ctx, models.Function{
		Id:        models.URN_PREFIX + "measuring-function:migration-one",
		Name:      "getEnergy",
		ConceptId: withOneFunction,
		RdfType:   models.SES_ONTOLOGY_MEASURING_FUNCTION,
	}, noFunctionSync)
	if err != nil {
		t.Error(err)
		return
	}
	//two of them, so that neither can be picked without guessing
	for _, suffix := range []string{"a", "b"} {
		err = db.SetFunction(ctx, models.Function{
			Id:          models.URN_PREFIX + "measuring-function:migration-two-" + suffix,
			Name:        "getPower" + suffix,
			DisplayName: "getPower" + suffix,
			ConceptId:   withTwoFunctions,
			RdfType:     models.SES_ONTOLOGY_MEASURING_FUNCTION,
		}, noFunctionSync)
		if err != nil {
			t.Error(err)
			return
		}
	}

	//one that already carries the name the convention would give it, so the migration has
	//nothing to do for it
	err = db.SetFunction(ctx, models.Function{
		Id:          matchingFunctionId,
		Name:        "Get-Flow",
		DisplayName: "Get-Flow",
		ConceptId:   withMatchingFunction,
		RdfType:     models.SES_ONTOLOGY_MEASURING_FUNCTION,
	}, noFunctionSync)
	if err != nil {
		t.Error(err)
		return
	}

	//lib.Start ran the migration on an empty database and recorded it. Dropping that record
	//reproduces a deployment that comes up on data it has not seen.
	raw, err := mongodriver.Connect(ctx, mongooptions.Client().ApplyURI(config.MongoUrl))
	if err != nil {
		t.Error(err)
		return
	}
	defer raw.Disconnect(ctx)
	migrationState := raw.Database(config.MongoTable).Collection(config.MongoMigrationStateCollection)
	_, err = migrationState.DeleteMany(ctx, bson.M{})
	if err != nil {
		t.Error(err)
		return
	}

	migrationController, err := controller.New(config, db, publisher.Void{}, nil)
	if err != nil {
		t.Error(err)
		return
	}
	err = db.RunStartupMigrations(migrationController)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("a concept without functions gets the pair", func(t *testing.T) {
		assertPair(t, withoutFunctions, "Volume")
	})

	t.Run("a single function off the convention is renamed", func(t *testing.T) {
		assertPair(t, withOneFunction, "Energy")
		measuring := conceptFunctions(t, withOneFunction, models.SES_ONTOLOGY_MEASURING_FUNCTION)
		if len(measuring) != 1 || measuring[0].Id != models.URN_PREFIX+"measuring-function:migration-one" {
			t.Error("expected the existing function to be renamed, not a second one to appear", measuring)
		}
	})

	t.Run("a function already named by the convention is left alone", func(t *testing.T) {
		measuring := conceptFunctions(t, withMatchingFunction, models.SES_ONTOLOGY_MEASURING_FUNCTION)
		if len(measuring) != 1 {
			t.Error("the migration must not add a second one next to it", measuring)
			return
		}
		if measuring[0].Id != matchingFunctionId || measuring[0].Name != "Get-Flow" {
			t.Error("unexpected function", measuring[0])
		}
	})

	t.Run("several functions are deprecated and a new one is created", func(t *testing.T) {
		measuring := conceptFunctions(t, withTwoFunctions, models.SES_ONTOLOGY_MEASURING_FUNCTION)
		if len(measuring) != 3 {
			t.Error("expected the two existing ones next to the new one", measuring)
			return
		}
		names := []string{}
		for _, function := range measuring {
			names = append(names, function.Name)
		}
		slices.Sort(names)
		if !slices.Equal(names, []string{"Get-Power", "getPowera-deprecated", "getPowerb-deprecated"}) {
			t.Error("unexpected names", names)
		}
		//the deprecated ones keep their id, so no device-type loses its reference
		for _, function := range measuring {
			if function.Name == "Get-Power" {
				continue
			}
			if function.Id != models.URN_PREFIX+"measuring-function:migration-two-a" &&
				function.Id != models.URN_PREFIX+"measuring-function:migration-two-b" {
				t.Error("unexpected id on a deprecated function", function.Id)
			}
		}
		//the controlling side had none at all and is created like any other
		controlling := conceptFunctions(t, withTwoFunctions, models.SES_ONTOLOGY_CONTROLLING_FUNCTION)
		if len(controlling) != 1 || controlling[0].Name != "Set-Power" {
			t.Error("unexpected controlling function", controlling)
		}
	})

	t.Run("the run is recorded in the database", func(t *testing.T) {
		count, err := migrationState.CountDocuments(ctx, bson.M{"name": "concept-functions"})
		if err != nil {
			t.Error(err)
			return
		}
		if count != 1 {
			t.Error("expected exactly one record for the migration", count)
		}
	})

	t.Run("the migration runs only once", func(t *testing.T) {
		before := map[string][]models.Function{}
		for _, conceptId := range []string{withoutFunctions, withOneFunction, withTwoFunctions} {
			before[conceptId] = conceptFunctions(t, conceptId, models.SES_ONTOLOGY_MEASURING_FUNCTION)
		}
		err := db.RunStartupMigrations(migrationController)
		if err != nil {
			t.Error(err)
			return
		}
		for conceptId, functions := range before {
			after := conceptFunctions(t, conceptId, models.SES_ONTOLOGY_MEASURING_FUNCTION)
			if len(after) != len(functions) {
				t.Error("a second run must not create anything", conceptId, len(functions), len(after))
			}
		}
	})
}
