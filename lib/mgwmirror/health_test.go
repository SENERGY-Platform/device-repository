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
package mgwmirror

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/database"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
)

type testClock struct{ t time.Time }

func (this *testClock) now() time.Time { return this.t }

func newTestHealth(maxAge time.Duration) (*Health, *testClock) {
	clock := &testClock{t: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	return NewHealth([]string{"devices", "functions"}, maxAge, clock.now), clock
}

func TestHealthFreshStartIsHealthyWithinGrace(t *testing.T) {
	health, clock := newTestHealth(30 * time.Minute)
	clock.t = clock.t.Add(29 * time.Minute)
	if healthy, report := health.Check(); !healthy || report != "ok" {
		t.Errorf("fresh start inside the grace period: %v %q", healthy, report)
	}
	clock.t = clock.t.Add(2 * time.Minute)
	if healthy, _ := health.Check(); healthy {
		t.Error("never synced and older than max age must be unhealthy")
	}
}

func TestHealthStaleCollectionReportsError(t *testing.T) {
	health, clock := newTestHealth(30 * time.Minute)
	start := clock.t
	health.Failure("functions", errors.New("listing functions: 403 forbidden"))
	health.Success("devices")
	clock.t = clock.t.Add(31 * time.Minute)
	health.Success("devices")

	healthy, report := health.Check()
	expected := "functions: not synced since " + start.Format(time.RFC3339) + ": listing functions: 403 forbidden"
	if healthy || report != expected {
		t.Errorf("got %v %q, expected unhealthy %q", healthy, report, expected)
	}
}

func TestHealthStaleWithoutErrorSaysSo(t *testing.T) {
	health, clock := newTestHealth(time.Minute)
	clock.t = clock.t.Add(2 * time.Minute)
	_, report := health.Check()
	if !strings.Contains(report, "devices: not synced since 2026-01-01T12:00:00Z: no error recorded") {
		t.Errorf("unexpected report %q", report)
	}
}

func TestHealthSuccessRecoversAndClearsError(t *testing.T) {
	health, clock := newTestHealth(time.Minute)
	health.FailAll(errors.New("boom"))
	clock.t = clock.t.Add(2 * time.Minute)
	if healthy, _ := health.Check(); healthy {
		t.Fatal("expected unhealthy")
	}
	health.Success("devices")
	health.Success("functions")
	if healthy, report := health.Check(); !healthy || report != "ok" {
		t.Errorf("after success: %v %q", healthy, report)
	}
}

func TestHealthMaxAge(t *testing.T) {
	cases := []struct {
		name     string
		interval string
		override string
		expected time.Duration
	}{
		{"default interval", "10m", "", 30 * time.Minute},
		{"short interval", "1m", "", 30 * time.Minute},
		{"long interval", "20m", "", 60 * time.Minute},
		{"exactly 3 intervals equals the minimum", "10m", "", 30 * time.Minute},
		{"override wins", "20m", "5m", 5 * time.Minute},
		{"invalid interval falls back", "x", "", 30 * time.Minute},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			config := configuration.Config{MgwMirrorUpdateInterval: c.interval, MgwMirrorHealthMaxAge: c.override}
			if got := HealthMaxAge(config); got != c.expected {
				t.Errorf("got %v, expected %v", got, c.expected)
			}
		})
	}
}

// timestampDb answers only the one call pull() makes before it reaches the collections.
type timestampDb struct {
	database.Database
	timestamps []model.LastUpdateTimestamp
}

func (this timestampDb) GetLastUpdateTimestampsForUser(context.Context, string) ([]model.LastUpdateTimestamp, error) {
	return this.timestamps, nil
}

func healthTestConfig(sourceUrl string) configuration.Config {
	return configuration.Config{
		MgwMirrorSourceUrl: sourceUrl, MgwMirrorUserId: "user", MgwMirrorUpdateInterval: "10m", LogLevel: "error",
		MongoProtocolCollection: "protocols", MongoAspectClassCollection: "aspect_classes", MongoAspectCollection: "aspects",
		MongoCharacteristicCollection: "characteristics", MongoConceptCollection: "concepts", MongoDeviceClassCollection: "device_classes",
		MongoFunctionCollection: "functions", MongoDeviceTypeCollection: "device_types", MongoDeviceCollection: "devices",
		MongoDeviceGroupCollection: "device_groups", MongoHubCollection: "hubs", MongoLocationCollection: "locations",
	}
}

func TestPullSkipBecauseCurrentCountsAsSuccess(t *testing.T) {
	config := healthTestConfig("")
	var timestamps []model.LastUpdateTimestamp
	for _, collection := range mirroredCollections(config) {
		timestamps = append(timestamps, model.LastUpdateTimestamp{Collection: collection, UnixTimestamp: 100})
	}
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(timestamps)
	}))
	defer source.Close()
	config.MgwMirrorSourceUrl = source.URL

	health, clock := newTestHealth(30 * time.Minute)
	health.collections = map[string]*collectionHealth{}
	for _, collection := range mirroredCollections(config) {
		health.collections[collection] = &collectionHealth{lastSuccess: clock.t}
	}
	clock.t = clock.t.Add(time.Hour)
	if healthy, _ := health.Check(); healthy {
		t.Fatal("precondition: stale before the pull")
	}
	pull(config, timestampDb{timestamps: timestamps}, health, true, nil)
	if healthy, report := health.Check(); !healthy {
		t.Errorf("every collection is current in the source, got %q", report)
	}
}

func TestPullAbortBeforeCollectionsMarksAllCollections(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer source.Close()
	config := healthTestConfig(source.URL)

	health, clock := newTestHealth(30 * time.Minute)
	health.collections = map[string]*collectionHealth{}
	for _, collection := range mirroredCollections(config) {
		health.collections[collection] = &collectionHealth{lastSuccess: clock.t}
	}
	pull(config, nil, health, true, nil)
	clock.t = clock.t.Add(time.Hour)
	healthy, report := health.Check()
	if healthy {
		t.Fatal("expected unhealthy")
	}
	lines := strings.Split(report, "\n")
	if len(lines) != len(mirroredCollections(config)) {
		t.Errorf("expected a line per collection, got %v", lines)
	}
	for _, line := range lines {
		if !strings.Contains(line, "getting source last update timestamps") {
			t.Errorf("error missing in %q", line)
		}
	}
}
