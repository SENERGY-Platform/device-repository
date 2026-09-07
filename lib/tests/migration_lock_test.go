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
	"sync"
	"testing"
	"time"

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

// TestStartupMigrationLock covers the lock that keeps the startup migrations of several
// instances apart. Kubernetes starts them at once, and runConceptFunctionsMigration records
// itself only after it succeeded — without the lock every instance passes that check and
// creates its own function pair per concept.
func TestStartupMigrationLock(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	//the migrations are the only thing under test here, so the test runs against mongodb
	//alone: no api, no kafka, no permissions
	config.Debug = true
	config.InitPermissionsTopics = false

	_, mongoIp, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MongoUrl = "mongodb://" + mongoIp + ":27017"

	const instances = 3
	dbs := []*mongo.Mongo{}
	controllers := []*controller.Controller{}
	for range instances {
		db, err := mongo.New(config)
		if err != nil {
			t.Error(err)
			return
		}
		defer db.Disconnect()
		ctrl, err := controller.New(config, db, publisher.Void{}, nil)
		if err != nil {
			t.Error(err)
			return
		}
		dbs = append(dbs, db)
		controllers = append(controllers, ctrl)
	}

	noConceptSync := func(models.Concept) error { return nil }
	concepts := map[string]string{
		models.URN_PREFIX + "concept:lock-volume": "Volume",
		models.URN_PREFIX + "concept:lock-energy": "Energy",
		models.URN_PREFIX + "concept:lock-power":  "Power",
	}
	for id, name := range concepts {
		err = dbs[0].SetConcept(ctx, models.Concept{Id: id, Name: name}, noConceptSync)
		if err != nil {
			t.Error(err)
			return
		}
	}

	conceptFunctions := func(t *testing.T, conceptId string, rdfType string) []models.Function {
		t.Helper()
		functions, _, err := dbs[0].ListFunctions(ctx, model.FunctionListOptions{
			ConceptIds: []string{conceptId},
			RdfType:    rdfType,
			Limit:      1000,
		})
		if err != nil {
			t.Error(err)
			return nil
		}
		return functions
	}

	raw, err := mongodriver.Connect(ctx, mongooptions.Client().ApplyURI(config.MongoUrl))
	if err != nil {
		t.Error(err)
		return
	}
	defer raw.Disconnect(ctx)
	lockCollection := raw.Database(config.MongoTable).Collection(config.MongoMigrationLockCollection)
	migrationState := raw.Database(config.MongoTable).Collection(config.MongoMigrationStateCollection)

	//every instance starts its migrations at the same moment
	start := make(chan struct{})
	runs := &sync.WaitGroup{}
	errs := make([]error, instances)
	for i := range instances {
		runs.Add(1)
		go func() {
			defer runs.Done()
			<-start
			errs[i] = dbs[i].RunStartupMigrations(controllers[i])
		}()
	}
	close(start)
	runs.Wait()
	for i, err := range errs {
		if err != nil {
			t.Error("instance", i, err)
		}
	}

	t.Run("every concept has exactly one function pair", func(t *testing.T) {
		for id, name := range concepts {
			for _, rdfType := range model.ConceptFunctionRdfTypes {
				functions := conceptFunctions(t, id, rdfType)
				if len(functions) != 1 {
					t.Error("expected one function per concept and type, not one per instance", name, rdfType, len(functions))
					continue
				}
				expected := model.ConceptFunctionName(rdfType, name)
				if functions[0].Name != expected {
					t.Error("unexpected function name", functions[0].Name, expected)
				}
			}
		}
	})

	t.Run("the migration is recorded once", func(t *testing.T) {
		count, err := migrationState.CountDocuments(ctx, bson.M{"name": "concept-functions"})
		if err != nil {
			t.Error(err)
			return
		}
		if count != 1 {
			t.Error("expected exactly one record for the migration", count)
		}
	})

	t.Run("the lock is released afterwards", func(t *testing.T) {
		count, err := lockCollection.CountDocuments(ctx, bson.M{})
		if err != nil {
			t.Error(err)
			return
		}
		if count != 0 {
			t.Error("expected the lock to be gone after the last instance is done", count)
		}
	})

	t.Run("a start fails rather than migrate beside a running instance", func(t *testing.T) {
		_, err := lockCollection.InsertOne(ctx, bson.M{
			"id":             "startup-migrations",
			"holder":         "another-instance",
			"unix_timestamp": time.Now().Unix(),
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer func() {
			_, err = lockCollection.DeleteMany(ctx, bson.M{})
			if err != nil {
				t.Error(err)
			}
		}()

		//the wait is bounded, and exceeding it fails the start so that kubernetes retries
		impatient := config
		impatient.MigrationLockTimeout = "1s"
		db, err := mongo.New(impatient)
		if err != nil {
			t.Error(err)
			return
		}
		defer db.Disconnect()
		err = db.RunStartupMigrations(controllers[0])
		if err == nil {
			t.Error("expected the start to fail while another instance holds the lock")
		}
	})

	t.Run("the lock of a dead instance is taken over", func(t *testing.T) {
		//a heartbeat this old is what an instance leaves behind that died mid migration
		_, err := lockCollection.InsertOne(ctx, bson.M{
			"id":             "startup-migrations",
			"holder":         "dead-instance",
			"unix_timestamp": time.Now().Add(-time.Hour).Unix(),
		})
		if err != nil {
			t.Error(err)
			return
		}
		err = dbs[0].RunStartupMigrations(controllers[0])
		if err != nil {
			t.Error(err)
			return
		}
		count, err := lockCollection.CountDocuments(ctx, bson.M{})
		if err != nil {
			t.Error(err)
			return
		}
		if count != 0 {
			t.Error("expected the expired lock to be replaced and released", count)
		}
	})
}
