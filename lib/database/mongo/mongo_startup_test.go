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
	"errors"
	"net"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v2/lib/configuration"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
)

const replicaSetURL = "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0&readPreference=primary"

func TestClientOptions_AuthWhenUserGiven(t *testing.T) {
	opts := clientOptions(configuration.Config{
		MongoUrl:        replicaSetURL,
		MongoUser:       "device-repository",
		MongoPassword:   "s3cr3t",
		MongoAuthSource: "admin",
		MongoDatabase:   "device_repository",
	})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	want := &options.Credential{Username: "device-repository", Password: "s3cr3t", AuthSource: "admin"}
	if !reflect.DeepEqual(opts.Auth, want) {
		t.Errorf("auth = %+v, want %+v", opts.Auth, want)
	}
}

func TestClientOptions_NoAuthWhenUserEmpty(t *testing.T) {
	// A password without a user must not switch auth on.
	opts := clientOptions(configuration.Config{
		MongoUrl:        "mongodb://localhost:27017",
		MongoPassword:   "s3cr3t",
		MongoAuthSource: "admin",
		MongoDatabase:   "device_repository",
	})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	if opts.Auth != nil {
		t.Errorf("auth = %+v, want nil", opts.Auth)
	}
}

func TestClientOptions_ConfiguredCredentialsReplaceURICredentials(t *testing.T) {
	opts := clientOptions(configuration.Config{
		MongoUrl:        "mongodb://old:oldpw@localhost:27017/?authSource=other&authMechanism=SCRAM-SHA-1",
		MongoUser:       "device-repository",
		MongoPassword:   "newpw",
		MongoAuthSource: "admin",
		MongoDatabase:   "device_repository",
	})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	want := &options.Credential{Username: "device-repository", Password: "newpw", AuthSource: "admin"}
	if !reflect.DeepEqual(opts.Auth, want) {
		t.Errorf("auth = %+v, want %+v", opts.Auth, want)
	}
}

func TestClientOptions_URIPassedUnchanged(t *testing.T) {
	opts := clientOptions(configuration.Config{MongoUrl: replicaSetURL, MongoDatabase: "device_repository"})
	if err := opts.Validate(); err != nil {
		t.Fatal(err)
	}
	if got := opts.GetURI(); got != replicaSetURL {
		t.Errorf("uri = %q, want %q", got, replicaSetURL)
	}
	if want := []string{"mongo-0.mongo:27017", "mongo-1.mongo:27017"}; !reflect.DeepEqual(opts.Hosts, want) {
		t.Errorf("hosts = %v, want %v", opts.Hosts, want)
	}
	if opts.ReplicaSet == nil || *opts.ReplicaSet != "rs0" {
		t.Errorf("replica set = %v, want rs0", opts.ReplicaSet)
	}
}

func TestClientOptions_KeepsMajorityReadConcern(t *testing.T) {
	opts := clientOptions(configuration.Config{MongoUrl: "mongodb://localhost:27017/?readConcernLevel=local", MongoDatabase: "device_repository"})
	if opts.ReadConcern == nil || opts.ReadConcern.Level != readconcern.Majority().Level {
		t.Errorf("read concern = %+v, want majority", opts.ReadConcern)
	}
}

func TestClientOptions_NoSchemeAdded(t *testing.T) {
	opts := clientOptions(configuration.Config{MongoUrl: "localhost:27017", MongoDatabase: "device_repository"})
	if err := opts.Validate(); err == nil {
		t.Fatal("expected an error for a url without scheme")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     configuration.Config
		wantErr error
	}{
		{"no auth", configuration.Config{MongoDatabase: "device_repository"}, nil},
		{"user and password", configuration.Config{MongoDatabase: "device_repository", MongoUser: "u", MongoPassword: "p"}, nil},
		{"password without user", configuration.Config{MongoDatabase: "device_repository", MongoPassword: "p"}, nil},
		{"user without password", configuration.Config{MongoDatabase: "device_repository", MongoUser: "u"}, errMissingPassword},
		{"empty database", configuration.Config{}, errEmptyDatabase},
		{"empty database with credentials", configuration.Config{MongoUser: "u", MongoPassword: "p"}, errEmptyDatabase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateConfig(tt.cfg); !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCollectionsUseConfiguredDatabase(t *testing.T) {
	// mongo.Connect does not contact the server, so no running instance is needed.
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	db := &Mongo{config: configuration.Config{MongoDatabase: "custom_db", MongoDeviceCollection: "device", MongoMigrationLockCollection: "migration_lock"}, client: client}
	for _, coll := range []*mongo.Collection{db.deviceCollection(), db.migrationLockCollection()} {
		if coll.Database().Name() != "custom_db" {
			t.Errorf("collection %s is in database %s, want custom_db", coll.Name(), coll.Database().Name())
		}
	}
}

// unreachableURL points at a port that was just free, so only the startup check can fail.
func unreachableURL(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	if err = l.Close(); err != nil {
		t.Fatal(err)
	}
	return "mongodb://" + addr + "/?directConnection=true"
}

// The startup check would fail as well, so these check for the specific validation error.
func TestNew_RejectsBeforeConnecting(t *testing.T) {
	cases := map[string]struct {
		cfg  configuration.Config
		want error
	}{
		"empty database":        {configuration.Config{MongoUser: "device-repository", MongoPassword: "s3cr3t"}, errEmptyDatabase},
		"user without password": {configuration.Config{MongoUser: "device-repository", MongoDatabase: "device_repository"}, errMissingPassword},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			c.cfg.MongoUrl = unreachableURL(t)
			begin := time.Now()
			db, err := New(c.cfg)
			if !errors.Is(err, c.want) {
				t.Fatalf("err = %v, want %v", err, c.want)
			}
			if db != nil {
				t.Error("expected no db on failure")
			}
			if strings.Contains(err.Error(), "s3cr3t") {
				t.Errorf("error leaks the password: %v", err)
			}
			if elapsed := time.Since(begin); elapsed > time.Second {
				t.Errorf("validation took %v, it must not wait for a server", elapsed)
			}
		})
	}
}

// poolCounter counts connection pools; Disconnect closes every pool Connect created.
type poolCounter struct{ created, closed atomic.Int32 }

func (p *poolCounter) monitor() *event.PoolMonitor {
	return &event.PoolMonitor{Event: func(e *event.PoolEvent) {
		switch e.Type {
		case event.PoolCreated:
			p.created.Add(1)
		case event.PoolClosedEvent:
			p.closed.Add(1)
		}
	}}
}

func (p *poolCounter) assertAllClosed(t *testing.T) {
	t.Helper()
	created, closed := p.created.Load(), p.closed.Load()
	if created == 0 || closed != created {
		t.Errorf("%d of %d connection pools closed, the client was left connected", closed, created)
	}
}

func TestStart_StartupCheckFailsWithoutServer(t *testing.T) {
	const password = "pw-must-not-appear-7f3a"
	conf := configuration.Config{
		MongoUrl:        unreachableURL(t),
		MongoUser:       "device-repository",
		MongoPassword:   password,
		MongoAuthSource: "admin",
		MongoDatabase:   "device_repository",
	}
	pools := &poolCounter{}
	begin := time.Now()
	db, err := start(conf, clientOptions(conf).SetPoolMonitor(pools.monitor()), 500*time.Millisecond)
	if err == nil {
		t.Fatal("expected an error when the server is unreachable")
	}
	if db != nil {
		t.Error("expected no db on failure")
	}
	if !strings.HasPrefix(err.Error(), "mongo startup check failed: ") {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), password) {
		t.Error("error text contains the password")
	}
	if elapsed := time.Since(begin); elapsed > 5*time.Second {
		t.Errorf("start took %v, the timeout was not applied", elapsed)
	}
	pools.assertAllClosed(t)
}

func TestStart_StartupCheckRunsBeforeCollectionHooks(t *testing.T) {
	called := false
	orig := CreateCollections
	CreateCollections = []func(db *Mongo) error{func(db *Mongo) error { called = true; return nil }}
	t.Cleanup(func() { CreateCollections = orig })

	conf := configuration.Config{MongoUrl: unreachableURL(t), MongoDatabase: "device_repository"}
	if _, err := start(conf, clientOptions(conf), 500*time.Millisecond); err == nil {
		t.Fatal("expected an error when the server is unreachable")
	}
	if called {
		t.Error("a collection hook ran although the startup check failed")
	}
}
