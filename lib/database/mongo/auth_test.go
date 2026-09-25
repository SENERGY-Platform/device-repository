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

package mongo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v2/lib/configuration"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestStartAuthenticates needs a throwaway server with access control; MONGO_AUTH_TEST_USER and
// MONGO_AUTH_TEST_PASSWORD are root credentials, used to create and remove the test users.
func TestStartAuthenticates(t *testing.T) {
	url, rootUser, rootPassword := os.Getenv("MONGO_AUTH_TEST_URL"), os.Getenv("MONGO_AUTH_TEST_USER"), os.Getenv("MONGO_AUTH_TEST_PASSWORD")
	if testing.Short() || url == "" || rootUser == "" || rootPassword == "" {
		t.Skip("needs MONGO_AUTH_TEST_URL, MONGO_AUTH_TEST_USER and MONGO_AUTH_TEST_PASSWORD, not in -short")
	}
	ctx := context.Background()
	root, err := mongo.Connect(ctx, options.Client().ApplyURI(url).SetAuth(options.Credential{Username: rootUser, Password: rootPassword, AuthSource: "admin"}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Disconnect(ctx) })

	suffix := randomHex(t)
	testDB, otherDB := "device_repository_auth_test_"+suffix, "device_repository_auth_other_"+suffix
	svcUser, svcPassword := "device-repository-test-"+suffix, randomHex(t)
	otherUser, otherPassword := "device-repository-other-"+suffix, randomHex(t)
	readUser, readPassword := "device-repository-read-"+suffix, randomHex(t)
	createUser(t, root, svcUser, svcPassword, "readWrite", testDB)
	createUser(t, root, otherUser, otherPassword, "readWrite", otherDB)
	createUser(t, root, readUser, readPassword, "read", testDB)
	passwords := []string{svcPassword, otherPassword, readPassword, rootPassword}

	base, err := configuration.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	config := func(user, password string) configuration.Config {
		conf := base
		conf.MongoUrl = url
		conf.MongoUser = user
		conf.MongoPassword = password
		conf.MongoAuthSource = "admin"
		conf.MongoDatabase = testDB
		conf.RunStartupMigrations = true
		conf.SkipDeviceGroupMigration = true
		return conf
	}

	t.Run("New and startup migrations with correct credentials", func(t *testing.T) {
		db, err := New(config(svcUser, svcPassword))
		if err != nil {
			t.Fatal(err)
		}
		defer db.Disconnect()
		if _, _, err = db.GetDevice(ctx, "urn:infai:ses:device:none"); err != nil {
			t.Errorf("query as the service user: %v", err)
		}
		if err = db.RunStartupMigrations(unusedMigrationMethods{t}); err != nil {
			t.Errorf("startup migrations as the service user: %v", err)
		}
		indexes, err := root.Database(testDB).Collection(base.MongoDeviceCollection).Indexes().ListSpecifications(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(indexes) < 2 {
			t.Errorf("expected the collection hooks to create indexes on %s, have %d", base.MongoDeviceCollection, len(indexes))
		}
	})

	cases := []struct {
		name, user, password string
		wantErr              string
	}{
		{"correct credentials", svcUser, svcPassword, ""},
		{"no credentials", "", "", "mongo startup check failed: "},
		{"user of another database", otherUser, otherPassword, "mongo startup check failed: "},
		{"wrong password", svcUser, svcPassword + "-wrong", "mongo startup check failed: "},
		// listCollections passes, index creation in the collection hooks does not.
		{"read-only user", readUser, readPassword, "Unauthorized"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := config(c.user, c.password)
			if err := validateConfig(conf); err != nil {
				t.Fatal(err)
			}
			pools := &poolCounter{}
			db, err := start(conf, clientOptions(conf).SetPoolMonitor(pools.monitor()), startupCheckTimeout)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				db.Disconnect()
				pools.assertAllClosed(t)
				return
			}
			if err == nil {
				db.Disconnect()
				t.Fatalf("expected an error containing %q", c.wantErr)
			}
			if db != nil {
				t.Error("expected no db on failure")
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("unexpected error: %v", err)
			}
			for _, pw := range passwords {
				if strings.Contains(err.Error(), pw) {
					t.Error("error text contains a password")
				}
			}
			pools.assertAllClosed(t)
		})
	}
}

// unusedMigrationMethods fails the test when called; on an empty database no migration needs the controller.
type unusedMigrationMethods struct{ t *testing.T }

func (this unusedMigrationMethods) DeviceIdToGeneratedDeviceGroupId(string) string {
	this.t.Error("unexpected call of DeviceIdToGeneratedDeviceGroupId")
	return ""
}

func (this unusedMigrationMethods) EnsureGeneratedDeviceGroup(models.Device, models.Device) error {
	this.t.Error("unexpected call of EnsureGeneratedDeviceGroup")
	return nil
}

func (this unusedMigrationMethods) SetContentVariableAspectIdsOnWrite(*models.DeviceType) {
	this.t.Error("unexpected call of SetContentVariableAspectIdsOnWrite")
}

func (this unusedMigrationMethods) GetDeviceGroupCriteria([]string) ([]models.DeviceGroupFilterCriteria, error, int) {
	this.t.Error("unexpected call of GetDeviceGroupCriteria")
	return nil, nil, 200
}

func (this unusedMigrationMethods) PublishFunction(models.Function) error {
	this.t.Error("unexpected call of PublishFunction")
	return nil
}

// createUser registers the cleanup first, so a partly failed creation is removed as well.
func createUser(t *testing.T, root *mongo.Client, user, password, role, db string) {
	t.Helper()
	admin := root.Database("admin")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = admin.RunCommand(ctx, bson.D{{Key: "dropUser", Value: user}}).Err()
		_ = root.Database(db).Drop(ctx)
	})
	cmd := bson.D{
		{Key: "createUser", Value: user},
		{Key: "pwd", Value: password},
		{Key: "roles", Value: bson.A{bson.D{{Key: "role", Value: role}, {Key: "db", Value: db}}}},
	}
	if err := admin.RunCommand(context.Background(), cmd).Err(); err != nil {
		t.Fatalf("create user: %v", err)
	}
}

func randomHex(t *testing.T) string {
	t.Helper()
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}
