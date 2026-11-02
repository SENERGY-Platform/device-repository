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

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/signal"
)

type namedRead struct {
	name string
	read func(c Interface) error
}

// invalidationReads are reads of different shapes: single reads, reads by local id, derived single
// reads, lists and queries
func invalidationReads() []namedRead {
	ctx := context.Background()
	token := testToken("alice", "1", "user")
	cached := func(c Interface) Interface { return WithCache(c, time.Minute) }
	return []namedRead{
		{"device x", func(c Interface) error { _, err, _ := cached(c).ReadDevice(ctx, "x", token, model.READ); return err }},
		{"device y", func(c Interface) error { _, err, _ := cached(c).ReadDevice(ctx, "y", token, model.READ); return err }},
		{"device by local id", func(c Interface) error {
			_, err, _ := cached(c).ReadDeviceByLocalId(ctx, "alice", "local-x", token, model.READ)
			return err
		}},
		{"extended device x", func(c Interface) error {
			_, err, _ := cached(c).ReadExtendedDevice(ctx, "x", token, model.READ, false)
			return err
		}},
		{"device list", func(c Interface) error {
			_, err, _ := cached(c).ListDevices(ctx, token, model.DeviceListOptions{Limit: 10})
			return err
		}},
		{"device type t", func(c Interface) error { _, err, _ := cached(c).ReadDeviceType(ctx, "t", token); return err }},
		{"device type u", func(c Interface) error { _, err, _ := cached(c).ReadDeviceType(ctx, "u", token); return err }},
		{"service s", func(c Interface) error { _, err, _ := cached(c).GetService(ctx, "s"); return err }},
		{"function f", func(c Interface) error { _, err, _ := cached(c).GetFunction(ctx, "f"); return err }},
		{"function query", func(c Interface) error {
			_, _, err, _ := cached(c).ListFunctions(ctx, model.FunctionListOptions{Limit: 10})
			return err
		}},
		{"device class c", func(c Interface) error { _, err, _ := cached(c).GetDeviceClass(ctx, "c"); return err }},
		{"functions of device class c", func(c Interface) error {
			_, err, _ := cached(c).GetDeviceClassesFunctions(ctx, "c")
			return err
		}},
		{"aspect node a", func(c Interface) error { _, err, _ := cached(c).GetAspectNode(ctx, "a"); return err }},
		{"measuring functions of aspect node a", func(c Interface) error {
			_, err, _ := cached(c).GetAspectNodesMeasuringFunctions(ctx, "a", false, false)
			return err
		}},
		{"concept k", func(c Interface) error { _, err, _ := cached(c).GetConceptWithCharacteristics(ctx, "k"); return err }},
		{"characteristic h", func(c Interface) error { _, err, _ := cached(c).GetCharacteristic(ctx, "h"); return err }},
		{"hub h", func(c Interface) error { _, err, _ := cached(c).ReadHub(ctx, "h", token, model.READ); return err }},
	}
}

// refetched primes every read, invalidates and returns the reads that ask the server again
func refetched(t *testing.T, basePath string, invalidate func(c Interface)) []string {
	repo := &nullRepo{}
	server := httptest.NewServer(repo)
	defer server.Close()
	c := NewClient(server.URL+basePath, nil)
	requests := func() int {
		repo.mux.Lock()
		defer repo.mux.Unlock()
		return repo.requests
	}
	reads := invalidationReads()
	for _, r := range reads {
		if err := r.read(c); err != nil {
			t.Fatalf("%v: %v", r.name, err)
		}
	}
	invalidate(c)
	result := []string{}
	for _, r := range reads {
		before := requests()
		if err := r.read(c); err != nil {
			t.Fatalf("%v: %v", r.name, err)
		}
		if requests() != before {
			result = append(result, r.name)
		}
	}
	return result
}

func TestInvalidateCache(t *testing.T) {
	lists := []string{"device list", "function query"}
	cases := []struct {
		name     string
		resource CacheResource
		id       string
		want     []string
	}{
		{"device", CacheDevices, "x", []string{"device x", "device by local id", "extended device x"}},
		{"every device", CacheDevices, "", []string{"device x", "device y", "device by local id", "extended device x"}},
		{"device type", CacheDeviceTypes, "t", []string{"extended device x", "device type t", "service s", "functions of device class c", "measuring functions of aspect node a"}},
		{"function", CacheFunctions, "f", []string{"function f", "functions of device class c", "measuring functions of aspect node a"}},
		{"aspect", CacheAspects, "z", []string{"aspect node a", "measuring functions of aspect node a"}},
		{"characteristic", CacheCharacteristics, "h", []string{"concept k", "characteristic h"}},
		{"hub", CacheHubs, "h", []string{"hub h"}},
		{"device class", CacheDeviceClasses, "c", []string{"device class c", "functions of device class c"}},
	}
	for _, c := range cases {
		for _, basePath := range []string{"", "/device-repository"} {
			got := refetched(t, basePath, func(client Interface) { InvalidateCache(client, c.resource, c.id) })
			want := append(slices.Clone(c.want), lists...)
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("%v (base path %q): refetched\n%v\nwant\n%v", c.name, basePath, strings.Join(got, ", "), strings.Join(want, ", "))
			}
		}
	}
}

func TestResetCache(t *testing.T) {
	got := refetched(t, "", ResetCache)
	if len(got) != len(invalidationReads()) {
		t.Errorf("expected every read to ask the server again, got %v", got)
	}
}

func TestInvalidateCacheOfOtherImplementations(t *testing.T) {
	type other struct {
		Interface
	}
	InvalidateCache(other{}, CacheDevices, "x")
	ResetCache(other{})
}

// a response requested before an invalidation is not stored after it, it may predate the change
func TestInvalidationDuringARequest(t *testing.T) {
	requested := make(chan struct{})
	release := make(chan struct{})
	mux := sync.Mutex{}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.Lock()
		requests++
		first := requests == 1
		mux.Unlock()
		if first {
			close(requested)
			<-release
		}
		w.Write([]byte(`{"id":"f"}`))
	}))
	defer server.Close()
	c := NewClient(server.URL, nil)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		defer close(done)
		WithCache(c, time.Minute).GetFunction(ctx, "f")
	}()
	<-requested
	InvalidateCache(c, CacheFunctions, "f")
	close(release)
	<-done

	WithCache(c, time.Minute).GetFunction(ctx, "f")
	mux.Lock()
	defer mux.Unlock()
	if requests != 2 {
		t.Errorf("a response requested before the invalidation was served from the cache, %v requests", requests)
	}
}

// publishOn sends a signal and waits until every subscriber handled it
func publishOn(t *testing.T, broker *signal.Broker, sig signal.Signal, value string) {
	done := make(chan struct{})
	id := broker.Sub("", sig, func(_ string, wg *sync.WaitGroup) {
		go func() {
			wg.Wait()
			close(done)
		}()
	})
	defer broker.Unsub(id)
	broker.Pub(sig, value)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("signal not handled")
	}
}

func TestInvalidateOnSignals(t *testing.T) {
	cases := []struct {
		name   string
		signal signal.Signal
		value  string
		want   []string
	}{
		{"device", signal.Known.DeviceCacheInvalidation, "x", []string{"device x", "device by local id", "extended device x", "device list", "function query"}},
		{"function", signal.Known.FunctionCacheInvalidation, "f", []string{"function f", "functions of device class c", "measuring functions of aspect node a", "device list", "function query"}},
		{"all", signal.Known.CacheInvalidationAll, "", nil},
	}
	for _, c := range cases {
		broker := &signal.Broker{}
		got := refetched(t, "", func(client Interface) {
			InvalidateOnSignals(client, broker)
			publishOn(t, broker, c.signal, c.value)
		})
		want := c.want
		if want == nil {
			for _, r := range invalidationReads() {
				want = append(want, r.name)
			}
		}
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Errorf("%v: refetched\n%v\nwant\n%v", c.name, strings.Join(got, ", "), strings.Join(want, ", "))
		}
	}

	t.Run("unsubscribe", func(t *testing.T) {
		broker := &signal.Broker{}
		got := refetched(t, "", func(client Interface) {
			InvalidateOnSignals(client, broker)()
			publishOn(t, broker, signal.Known.CacheInvalidationAll, "")
		})
		if len(got) != 0 {
			t.Errorf("signals after unsubscribing must not clear the cache, refetched %v", got)
		}
	})
}
