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
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
)

const defaultHealthMaxAge = 30 * time.Minute

// HealthMaxAge is how long a collection may go without a successful sync before the mirror is unhealthy:
// mgw_mirror_health_max_age if set, else 30 minutes but at least three update intervals.
func HealthMaxAge(config configuration.Config) time.Duration {
	if config.MgwMirrorHealthMaxAge != "" {
		if d, err := time.ParseDuration(config.MgwMirrorHealthMaxAge); err == nil {
			return d
		}
	}
	result := defaultHealthMaxAge
	if interval, err := time.ParseDuration(config.MgwMirrorUpdateInterval); err == nil && 3*interval > result {
		result = 3 * interval
	}
	return result
}

func mirroredCollections(config configuration.Config) []string {
	return []string{
		config.MongoProtocolCollection,
		config.MongoAspectClassCollection,
		config.MongoAspectCollection,
		config.MongoCharacteristicCollection,
		config.MongoConceptCollection,
		config.MongoDeviceClassCollection,
		config.MongoFunctionCollection,
		config.MongoDeviceTypeCollection,
		config.MongoDeviceCollection,
		config.MongoDeviceGroupCollection,
		config.MongoHubCollection,
		config.MongoLocationCollection,
	}
}

// Health tracks per collection when the mirror was last known to be current.
// pull() writes while the health endpoint reads.
type Health struct {
	now    func() time.Time
	maxAge time.Duration

	mux         sync.Mutex
	collections map[string]*collectionHealth
}

type collectionHealth struct {
	lastSuccess time.Time
	lastError   string
	lastErrorAt time.Time
}

// NewHealth starts every collection with the current time as its last success,
// so a fresh process has one max age to complete its first pull.
func NewHealth(collections []string, maxAge time.Duration, now func() time.Time) *Health {
	result := &Health{now: now, maxAge: maxAge, collections: map[string]*collectionHealth{}}
	start := now()
	for _, collection := range collections {
		result.collections[collection] = &collectionHealth{lastSuccess: start}
	}
	return result
}

// oneLine keeps the report at one line per collection; client errors carry the response body with its newline.
func oneLine(err error) string {
	return strings.Join(strings.Fields(err.Error()), " ")
}

// Success records that the collection is current: it was pulled completely or the source has nothing newer.
func (this *Health) Success(collection string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	if c, ok := this.collections[collection]; ok {
		c.lastSuccess = this.now()
		c.lastError = ""
		c.lastErrorAt = time.Time{}
	}
}

// Failure records why the collection could not be synced; the last success stays as it was.
func (this *Health) Failure(collection string, err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	if c, ok := this.collections[collection]; ok {
		c.lastError = oneLine(err)
		c.lastErrorAt = this.now()
	}
}

// FailAll records the error of a pull that aborted before it reached the collections.
func (this *Health) FailAll(err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	now := this.now()
	for _, c := range this.collections {
		c.lastError = oneLine(err)
		c.lastErrorAt = now
	}
}

// Check reports healthy if every collection was synced within max age;
// otherwise report has one line per stale collection.
func (this *Health) Check() (healthy bool, report string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	now := this.now()
	lines := []string{}
	for name, c := range this.collections {
		if now.Sub(c.lastSuccess) <= this.maxAge {
			continue
		}
		reason := c.lastError
		if reason == "" {
			reason = "no error recorded"
		}
		lines = append(lines, fmt.Sprintf("%v: not synced since %v: %v", name, c.lastSuccess.UTC().Format(time.RFC3339), reason))
	}
	if len(lines) == 0 {
		return true, "ok"
	}
	sort.Strings(lines)
	return false, strings.Join(lines, "\n")
}
