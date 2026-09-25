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

package configuration

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type mongoFields struct {
	Url, User, Password, AuthSource, Database, DeviceCollection string
}

func mongoOf(c Config) mongoFields {
	return mongoFields{c.MongoUrl, c.MongoUser, c.MongoPassword, c.MongoAuthSource, c.MongoDatabase, c.MongoDeviceCollection}
}

// clearMongoEnv empties the variables for this test; the loader ignores empty values.
func clearMongoEnv(t *testing.T) {
	for _, k := range []string{"MONGO_URL", "MONGO_USER", "MONGO_PASSWORD", "MONGO_AUTH_SOURCE", "MONGO_DATABASE", "MONGO_TABLE", "MONGO_DEVICE_COLLECTION", "MGW_CERT_MANAGER_URL"} {
		t.Setenv(k, "")
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func load(t *testing.T, location string) (cfg Config, stdout string) {
	t.Helper()
	stdout = captureStdout(t, func() {
		var err error
		if cfg, err = Load(location); err != nil {
			t.Fatal(err)
		}
	})
	return cfg, stdout
}

func TestLoad_RepoConfigMongoDefaults(t *testing.T) {
	clearMongoEnv(t)
	cfg, _ := load(t, "../../config.json")
	want := mongoFields{Url: "mongodb://localhost:27017", AuthSource: "admin", Database: "device_repository", DeviceCollection: "device"}
	if got := mongoOf(cfg); got != want {
		t.Errorf("mongo = %+v, want %+v", got, want)
	}
}

func TestLoad_MongoDefaultsWhenFileOmitsThem(t *testing.T) {
	clearMongoEnv(t)
	cfg, _ := load(t, writeConfig(t, `{}`))
	want := mongoFields{Url: "mongodb://localhost:27017", AuthSource: "admin", Database: "device_repository"}
	if got := mongoOf(cfg); got != want {
		t.Errorf("mongo = %+v, want %+v", got, want)
	}
}

func TestLoad_MongoConfigFile(t *testing.T) {
	clearMongoEnv(t)
	cfg, _ := load(t, writeConfig(t, `{"mongo_url": "mongodb://file:27017", "mongo_user": "u", "mongo_password": "s3cr3t", "mongo_auth_source": "a", "mongo_database": "d", "mongo_device_collection": "c"}`))
	want := mongoFields{Url: "mongodb://file:27017", User: "u", Password: "s3cr3t", AuthSource: "a", Database: "d", DeviceCollection: "c"}
	if got := mongoOf(cfg); got != want {
		t.Errorf("mongo = %+v, want %+v", got, want)
	}
}

func TestLoad_MongoEnvNames(t *testing.T) {
	clearMongoEnv(t)
	t.Setenv("MONGO_URL", "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0")
	t.Setenv("MONGO_USER", "device-repository")
	t.Setenv("MONGO_PASSWORD", "p@ss:w/rd")
	t.Setenv("MONGO_AUTH_SOURCE", "users")
	t.Setenv("MONGO_DATABASE", "device_repository_test")
	cfg, _ := load(t, "../../config.json")
	want := mongoFields{
		Url:              "mongodb://mongo-0.mongo:27017,mongo-1.mongo:27017/?replicaSet=rs0",
		User:             "device-repository",
		Password:         "p@ss:w/rd",
		AuthSource:       "users",
		Database:         "device_repository_test",
		DeviceCollection: "device",
	}
	if got := mongoOf(cfg); got != want {
		t.Errorf("mongo = %+v, want %+v", got, want)
	}
}

// MONGO_TABLE was replaced by MONGO_DATABASE without a fallback.
func TestLoad_MongoTableNoLongerRead(t *testing.T) {
	clearMongoEnv(t)
	t.Setenv("MONGO_TABLE", "other")
	cfg, _ := load(t, writeConfig(t, `{"mongo_table": "other"}`))
	if cfg.MongoDatabase != "device_repository" {
		t.Errorf("database = %q, want device_repository", cfg.MongoDatabase)
	}
}

func TestSecretFields(t *testing.T) {
	secrets := []string{}
	typ := reflect.TypeOf(Config{})
	for i := 0; i < typ.NumField(); i++ {
		if isSecret(typ.Field(i)) {
			secrets = append(secrets, typ.Field(i).Name)
		}
	}
	if want := []string{"MongoPassword"}; !reflect.DeepEqual(secrets, want) {
		t.Errorf("secret fields = %v, want %v", secrets, want)
	}
}

func TestLoad_EnvPrintMasksMongoPassword(t *testing.T) {
	clearMongoEnv(t)
	t.Setenv("MONGO_USER", "device-repository")
	t.Setenv("MONGO_PASSWORD", "s3cr3t-pw")
	cfg, out := load(t, "../../config.json")
	if strings.Contains(out, "s3cr3t-pw") {
		t.Errorf("printed environment leaks the password: %s", out)
	}
	if !strings.Contains(out, "MONGO_PASSWORD  =  ***") {
		t.Errorf("expected the password variable to be reported as masked, got: %s", out)
	}
	if !strings.Contains(out, "MONGO_USER  =  device-repository") {
		t.Errorf("expected the applied variables to be printed, got: %s", out)
	}
	if cfg.MongoPassword != "s3cr3t-pw" {
		t.Errorf("password = %q, want the value of MONGO_PASSWORD", cfg.MongoPassword)
	}
}

func TestConfigFormattingMasksMongoPassword(t *testing.T) {
	clearMongoEnv(t)
	t.Setenv("MONGO_PASSWORD", "s3cr3t-pw")
	cfg, _ := load(t, "../../config.json")
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	bp, err := json.Marshal(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	outputs := map[string]string{
		"json":         string(b),
		"json pointer": string(bp),
		"%v":           fmt.Sprintf("%v", cfg),
		"%+v":          fmt.Sprintf("%+v", cfg),
		"%#v":          fmt.Sprintf("%#v", cfg),
		"%s":           fmt.Sprintf("%s", cfg),
		"%v pointer":   fmt.Sprintf("%v", &cfg),
		"%+v pointer":  fmt.Sprintf("%+v", &cfg),
		"%#v pointer":  fmt.Sprintf("%#v", &cfg),
		"%v in slice":  fmt.Sprintf("%v", []Config{cfg}),
		"%+v in map":   fmt.Sprintf("%+v", map[string]Config{"c": cfg}),
	}
	for name, s := range outputs {
		if strings.Contains(s, "s3cr3t-pw") {
			t.Errorf("%s leaks the password: %s", name, s)
		}
		if !strings.Contains(s, "device_repository") {
			t.Errorf("%s lost the other fields: %s", name, s)
		}
	}
	if !strings.Contains(string(b), `"mongo_password":"***"`) {
		t.Errorf("json does not show the password as masked: %s", b)
	}
	if cfg.MongoPassword != "s3cr3t-pw" {
		t.Errorf("masking changed the loaded password to %q", cfg.MongoPassword)
	}
}

func TestConfigFormattingKeepsEmptyPasswordEmpty(t *testing.T) {
	b, err := json.Marshal(Config{MongoDatabase: "device_repository"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"mongo_password":""`) {
		t.Errorf("an unset password should stay visibly unset: %s", b)
	}
}

func TestFieldNameToEnvName(t *testing.T) {
	want := map[string]string{
		"MongoUrl":        "MONGO_URL",
		"MongoUser":       "MONGO_USER",
		"MongoPassword":   "MONGO_PASSWORD",
		"MongoAuthSource": "MONGO_AUTH_SOURCE",
		"MongoDatabase":   "MONGO_DATABASE",
	}
	for field, env := range want {
		if got := fieldNameToEnvName(field); got != env {
			t.Errorf("%s -> %s, want %s", field, got, env)
		}
	}
}

// captureStdout also catches the logger Load creates, since it writes to os.Stdout.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	defer func() { os.Stdout = orig }()
	f()
	_ = w.Close()
	return <-done
}
