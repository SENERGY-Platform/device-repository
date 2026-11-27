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
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// timestampedWrite describes a Set*/Remove* method of a collection that backs a 304 answer (group A of the cache headers).
// seed is the Set method that creates the element the remove method is called with, empty for the Set methods themselves.
type timestampedWrite struct {
	seed        string
	collections func(c configuration.Config) []string
}

func collectionsOf(f func(c configuration.Config) string) func(c configuration.Config) []string {
	return func(c configuration.Config) []string { return []string{f(c)} }
}

var (
	aspectCollections      = collectionsOf(func(c configuration.Config) string { return c.MongoAspectCollection })
	aspectNodeCollections  = collectionsOf(GetAspectNodeCollectionName)
	aspectClassCollections = collectionsOf(func(c configuration.Config) string { return c.MongoAspectClassCollection })
	conceptCollections     = collectionsOf(func(c configuration.Config) string { return c.MongoConceptCollection })
	characteristicsColls   = collectionsOf(func(c configuration.Config) string { return c.MongoCharacteristicCollection })
	functionCollections    = collectionsOf(func(c configuration.Config) string { return c.MongoFunctionCollection })
	deviceClassCollections = collectionsOf(func(c configuration.Config) string { return c.MongoDeviceClassCollection })
	protocolCollections    = collectionsOf(func(c configuration.Config) string { return c.MongoProtocolCollection })
	deviceTypeCollections  = func(c configuration.Config) []string {
		return []string{c.MongoDeviceTypeCollection, GetDeviceTypeCriteriaCollectionName(c)}
	}
)

// timestampedWrites is the complete list of Set*/Remove* methods that write a collection of the fixed set
// observed by the cache headers: aspects, aspect nodes, aspect classes, concepts, characteristics, functions,
// device classes, device types, device type criteria and protocols.
var timestampedWrites = map[string]timestampedWrite{
	"SetAspect":                 {"", aspectCollections},
	"RemoveAspect":              {"SetAspect", aspectCollections},
	"RemoveAspects":             {"SetAspect", aspectCollections},
	"SetAspectNode":             {"", aspectNodeCollections},
	"RemoveAspectNodesByRootId": {"SetAspectNode", aspectNodeCollections},
	"SetAspectClass":            {"", aspectClassCollections},
	"RemoveAspectClass":         {"SetAspectClass", aspectClassCollections},
	"RemoveAspectClasses":       {"SetAspectClass", aspectClassCollections},
	"SetConcept":                {"", conceptCollections},
	"RemoveConcept":             {"SetConcept", conceptCollections},
	"RemoveConcepts":            {"SetConcept", conceptCollections},
	"SetCharacteristic":         {"", characteristicsColls},
	"RemoveCharacteristic":      {"SetCharacteristic", characteristicsColls},
	"RemoveCharacteristics":     {"SetCharacteristic", characteristicsColls},
	"SetFunction":               {"", functionCollections},
	"RemoveFunction":            {"SetFunction", functionCollections},
	"RemoveFunctions":           {"SetFunction", functionCollections},
	"SetDeviceClass":            {"", deviceClassCollections},
	"RemoveDeviceClass":         {"SetDeviceClass", deviceClassCollections},
	"RemoveDeviceClasses":       {"SetDeviceClass", deviceClassCollections},
	"SetDeviceType":             {"", deviceTypeCollections},
	"RemoveDeviceType":          {"SetDeviceType", deviceTypeCollections},
	"RemoveDeviceTypes":         {"SetDeviceType", deviceTypeCollections},
	"SetProtocol":               {"", protocolCollections},
	"RemoveProtocol":            {"SetProtocol", protocolCollections},
	"RemoveProtocols":           {"SetProtocol", protocolCollections},
}

// otherWrites lists the Set*/Remove* methods that are not part of the fixed set.
// A new Set*/Remove* method has to be added here or to timestampedWrites, otherwise TestLastUpdateTimestampOfFixedSet fails:
// the decision whether a 304 may be answered from the timestamp of its collection is made when the method is added.
var otherWrites = map[string]bool{
	"SetLastUpdateTimestamp":     true, // the writer of the timestamps itself
	"SetDevice":                  true,
	"RemoveDevice":               true,
	"RemoveDevices":              true,
	"SetDeviceConnectionState":   true,
	"SetDeviceConnectionStates":  true,
	"SetDefaultDeviceAttributes": true,
	"SetHub":                     true,
	"RemoveHub":                  true,
	"RemoveHubs":                 true,
	"SetHubConnectionState":      true,
	"SetHubConnectionStates":     true,
	"SetDeviceGroup":             true,
	"RemoveDeviceGroup":          true,
	"RemoveDeviceGroups":         true,
	"SetLocation":                true,
	"RemoveLocation":             true,
	"RemoveLocations":            true,
	"SetThing":                   true,
	"SetThingIfRevision":         true,
	"RemoveThing":                true,
	"RemoveThings":               true,
	"SetGraph":                   true,
	"SetGraphIfRevision":         true,
	"RemoveGraph":                true,
	"RemoveGraphs":               true,
}

// callWrite calls a Set*/Remove* method with an element of the id "x": the context is passed through,
// structs are zero values with the id "x", sync handlers do nothing.
func callWrite(ctx context.Context, m *Mongo, name string) error {
	method := reflect.ValueOf(m).MethodByName(name)
	args := []reflect.Value{}
	for i := 0; i < method.Type().NumIn(); i++ {
		argType := method.Type().In(i)
		switch {
		case argType == reflect.TypeOf((*context.Context)(nil)).Elem():
			args = append(args, reflect.ValueOf(ctx))
		case argType.Kind() == reflect.Func:
			args = append(args, reflect.MakeFunc(argType, func([]reflect.Value) []reflect.Value {
				results := []reflect.Value{}
				for j := 0; j < argType.NumOut(); j++ {
					results = append(results, reflect.Zero(argType.Out(j)))
				}
				return results
			}))
		case argType.Kind() == reflect.String:
			args = append(args, reflect.ValueOf("x").Convert(argType))
		case argType.Kind() == reflect.Slice && argType.Elem().Kind() == reflect.String:
			args = append(args, reflect.ValueOf([]string{"x"}).Convert(argType))
		case argType.Kind() == reflect.Struct:
			element := reflect.New(argType).Elem()
			if id := element.FieldByName("Id"); id.IsValid() && id.Kind() == reflect.String {
				id.SetString("x")
			}
			args = append(args, element)
		default:
			args = append(args, reflect.Zero(argType))
		}
	}
	results := method.Call(args)
	last := results[len(results)-1]
	if last.IsNil() {
		return nil
	}
	return last.Interface().(error)
}

func TestLastUpdateTimestampOfFixedSet(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := configuration.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	port, _, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	conf.MongoUrl = "mongodb://localhost:" + port
	m, err := New(conf)
	if err != nil {
		t.Fatal(err)
	}
	timeout := func() context.Context {
		c, cancelTimeout := context.WithTimeout(ctx, 10*time.Second)
		t.Cleanup(cancelTimeout)
		return c
	}

	t.Run("every write method is classified", func(t *testing.T) {
		methodType := reflect.TypeOf(m)
		for i := 0; i < methodType.NumMethod(); i++ {
			name := methodType.Method(i).Name
			if !strings.HasPrefix(name, "Set") && !strings.HasPrefix(name, "Remove") {
				continue
			}
			_, fixed := timestampedWrites[name]
			if !fixed && !otherWrites[name] {
				t.Errorf("%v is not classified: add it to timestampedWrites if it writes a collection of the fixed set "+
					"(aspects, aspect nodes, aspect classes, concepts, characteristics, functions, device classes, device types, "+
					"device type criteria, protocols), otherwise to otherWrites", name)
			}
		}
		for name := range timestampedWrites {
			if _, ok := methodType.MethodByName(name); !ok {
				t.Errorf("timestampedWrites lists %v, which does not exist", name)
			}
		}
	})

	names := []string{}
	for name := range timestampedWrites {
		names = append(names, name)
	}
	sort.Strings(names)

	// clearTimestamps removes all timestamps, so that a collection has one afterwards only if the method under test wrote it
	clearTimestamps := func(t *testing.T) {
		t.Helper()
		_, err := m.lastUpdateTimestampsCollection().DeleteMany(timeout(), bson.M{})
		if err != nil {
			t.Fatal(err)
		}
	}
	hasTimestamp := func(t *testing.T, collection string) bool {
		t.Helper()
		err := m.lastUpdateTimestampsCollection().FindOne(timeout(), bson.M{
			LastUpdateTimestampsBson.Collection: collection,
			LastUpdateTimestampsBson.UserId:     "",
		}).Err()
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false
		}
		if err != nil {
			t.Fatal(err)
		}
		return true
	}

	t.Run("every write method moves the timestamp of its collections", func(t *testing.T) {
		for _, name := range names {
			write := timestampedWrites[name]
			t.Run(name, func(t *testing.T) {
				if write.seed != "" {
					if err := callWrite(timeout(), m, write.seed); err != nil {
						t.Fatal(err)
					}
				}
				clearTimestamps(t)
				if err := callWrite(timeout(), m, name); err != nil {
					t.Fatal(err)
				}
				for _, collection := range write.collections(m.config) {
					if !hasTimestamp(t, collection) {
						t.Errorf("%v does not set the last update timestamp of %v", name, collection)
					}
				}
			})
		}
	})

	t.Run("a failing timestamp write is returned by every write method", func(t *testing.T) {
		// a view cannot be written to: every write to the timestamp collection fails, no hook in production code needed
		view := m.config.MongoLastUpdateTimestampsCollection + "_view"
		err := m.client.Database(m.config.MongoDatabase).CreateView(timeout(), view, m.config.MongoLastUpdateTimestampsCollection, mongo.Pipeline{})
		if err != nil {
			t.Fatal(err)
		}
		healthy := m.config
		m.config.MongoLastUpdateTimestampsCollection = view
		defer func() { m.config = healthy }()

		for _, name := range names {
			write := timestampedWrites[name]
			t.Run(name, func(t *testing.T) {
				if write.seed != "" {
					// the element is stored before the timestamp fails
					_ = callWrite(timeout(), m, write.seed)
				}
				err := callWrite(timeout(), m, name)
				if err == nil {
					t.Fatalf("%v returned no error although the last update timestamp could not be written", name)
				}
				if !errors.Is(err, ErrLastUpdateTimestamp) {
					t.Errorf("%v returned an error that is not the timestamp error: %v", name, err)
				}
			})
		}

		// collections outside the fixed set keep ignoring the error
		t.Run("SetThing stays unchanged", func(t *testing.T) {
			if err := callWrite(timeout(), m, "SetThing"); err != nil {
				t.Errorf("SetThing returned %v", err)
			}
		})
	})

	// A Set or Remove whose timestamp could not be written leaves the element marked; a repeated Remove does not find it
	// any more, so the Retry*Sync loop has to write the timestamp.
	t.Run("the retry loop writes the timestamp that a failed write left out", func(t *testing.T) {
		view := m.config.MongoLastUpdateTimestampsCollection + "_retry_view"
		err := m.client.Database(m.config.MongoDatabase).CreateView(timeout(), view, m.config.MongoLastUpdateTimestampsCollection, mongo.Pipeline{})
		if err != nil {
			t.Fatal(err)
		}
		healthy := m.config
		broken := m.config
		broken.MongoLastUpdateTimestampsCollection = view
		defer func() { m.config = healthy }()

		retries := []struct{ retry, set, remove string }{
			{"RetryAspectSync", "SetAspect", "RemoveAspect"},
			{"RetryAspectClassSync", "SetAspectClass", "RemoveAspectClass"},
			{"RetryConceptSync", "SetConcept", "RemoveConcept"},
			{"RetryCharacteristicSync", "SetCharacteristic", "RemoveCharacteristic"},
			{"RetryFunctionSync", "SetFunction", "RemoveFunction"},
			{"RetryDeviceClassSync", "SetDeviceClass", "RemoveDeviceClass"},
			{"RetryProtocolSync", "SetProtocol", "RemoveProtocol"},
			{"RetryDeviceTypeSync", "SetDeviceType", "RemoveDeviceType"},
		}
		for _, r := range retries {
			for _, remove := range []bool{false, true} {
				name := r.retry + "/set"
				write := r.set
				pending := SyncTodoBson
				if remove {
					name = r.retry + "/remove"
					write = r.remove
					pending = SyncDeleteBson
				}
				t.Run(name, func(t *testing.T) {
					collections := timestampedWrites[r.set].collections(healthy)
					clearTimestamps(t)
					if remove {
						if err := callWrite(timeout(), m, r.set); err != nil {
							t.Fatal(err)
						}
					}
					m.config = broken
					err := callWrite(timeout(), m, write)
					m.config = healthy
					if err == nil {
						t.Fatalf("%v returned no error although the last update timestamp could not be written", write)
					}
					// the retry fetches only jobs that are older than the lock duration, in whole seconds
					time.Sleep(2100 * time.Millisecond)
					clearTimestamps(t)

					retry := reflect.ValueOf(m).MethodByName(r.retry)
					args := []reflect.Value{reflect.ValueOf(time.Second)}
					for i := 1; i < retry.Type().NumIn(); i++ {
						argType := retry.Type().In(i)
						args = append(args, reflect.MakeFunc(argType, func([]reflect.Value) []reflect.Value {
							return []reflect.Value{reflect.Zero(argType.Out(0))}
						}))
					}
					if result := retry.Call(args)[0]; !result.IsNil() {
						t.Fatal(result.Interface())
					}

					for _, collection := range collections {
						if !hasTimestamp(t, collection) {
							t.Errorf("%v does not set the last update timestamp of %v", r.retry, collection)
						}
					}
					count, err := m.client.Database(m.config.MongoDatabase).Collection(collections[0]).CountDocuments(timeout(), bson.M{pending: true})
					if err != nil {
						t.Fatal(err)
					}
					if count != 0 {
						t.Errorf("%v left %v documents with %v", r.retry, count, pending)
					}
					if remove {
						count, err = m.client.Database(m.config.MongoDatabase).Collection(collections[0]).CountDocuments(timeout(), bson.M{})
						if err != nil {
							t.Fatal(err)
						}
						if count != 0 {
							t.Errorf("%v did not delete the element", r.retry)
						}
					}
				})
			}
		}
	})
}
