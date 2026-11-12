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
	"sync"

	"net/url"
	"strings"

	"github.com/SENERGY-Platform/service-commons/pkg/signal"
)

// CacheResource names a kind of resource whose cached responses InvalidateCache removes.
type CacheResource string

const (
	CacheDevices         CacheResource = "devices"
	CacheDeviceTypes     CacheResource = "device-types"
	CacheDeviceGroups    CacheResource = "device-groups"
	CacheProtocols       CacheResource = "protocols"
	CacheHubs            CacheResource = "hubs"
	CacheLocations       CacheResource = "locations"
	CacheConcepts        CacheResource = "concepts"
	CacheCharacteristics CacheResource = "characteristics"
	CacheFunctions       CacheResource = "functions"
	CacheAspects         CacheResource = "aspects"
	CacheAspectClasses   CacheResource = "aspect-classes"
	CacheDeviceClasses   CacheResource = "device-classes"
	CacheGraphs          CacheResource = "graphs"
	CacheThings          CacheResource = "things"
)

// cacheDependencies says, per resource, under which paths a single read returns it, and which
// single reads of other paths are derived from it. A derived pattern matches the leading segments
// of a path, "*" matching any one segment.
var cacheDependencies = map[CacheResource]struct {
	own     []string
	derived []string
}{
	CacheDevices:      {own: []string{"devices", "extended-devices"}},
	CacheDeviceTypes:  {own: []string{"device-types"}, derived: []string{"services", "extended-devices", "device-classes/*/functions", "device-classes/*/controlling-functions", "aspect-nodes/*/measuring-functions", "aspect-nodes/*/controlling-functions"}},
	CacheDeviceGroups: {own: []string{"device-groups"}},
	CacheProtocols:    {own: []string{"protocols"}},
	CacheHubs:         {own: []string{"hubs", "extended-hubs"}},
	CacheLocations:    {own: []string{"locations", "extended-locations"}},
	CacheConcepts:     {own: []string{"concepts"}},
	//a concept is read with its characteristics
	CacheCharacteristics: {own: []string{"characteristics"}, derived: []string{"concepts"}},
	CacheFunctions:       {own: []string{"functions"}, derived: []string{"device-classes/*/functions", "device-classes/*/controlling-functions", "aspect-nodes/*/measuring-functions", "aspect-nodes/*/controlling-functions"}},
	//an aspect node carries ancestors and descendants, so a change anywhere in a hierarchy changes all its nodes
	CacheAspects:       {own: []string{"aspects"}, derived: []string{"aspect-nodes"}},
	CacheAspectClasses: {own: []string{"aspect-classes"}, derived: []string{"aspects", "aspect-nodes"}},
	CacheDeviceClasses: {own: []string{"device-classes"}},
	CacheGraphs:        {own: []string{"graphs"}},
	CacheThings:        {own: []string{"things"}},
}

// singleReadCollections are the first path segments under which a path with a further segment
// reads one resource; every other cached response is a list or a query
var singleReadCollections = func() map[string]bool {
	result := map[string]bool{}
	for _, dependency := range cacheDependencies {
		for _, collection := range dependency.own {
			result[collection] = true
		}
	}
	result["services"] = true
	result["aspect-nodes"] = true
	return result
}()

// cacheTarget is the path of a cached request relative to the base url of the client, and its query
type cacheTarget struct {
	segments []string
	query    url.Values
}

func newCacheTarget(basePath string, requestUrl *url.URL) cacheTarget {
	path := strings.Trim(strings.TrimPrefix(requestUrl.Path, basePath), "/")
	segments := []string{}
	if path != "" {
		segments = strings.Split(path, "/")
	}
	return cacheTarget{segments: segments, query: requestUrl.Query()}
}

func (this cacheTarget) isSingleRead() bool {
	return len(this.segments) >= 2 && singleReadCollections[this.segments[0]]
}

func (this cacheTarget) matches(pattern string) bool {
	parts := strings.Split(pattern, "/")
	if len(parts) > len(this.segments) {
		return false
	}
	for i, part := range parts {
		if part != "*" && part != this.segments[i] {
			return false
		}
	}
	return true
}

// affectedBy decides whether a change of the resource with id may change the cached response
func (this cacheTarget) affectedBy(resource CacheResource, id string) bool {
	dependency, known := cacheDependencies[resource]
	if !known || !this.isSingleRead() {
		return true
	}
	for _, collection := range dependency.own {
		if this.segments[0] != collection {
			continue
		}
		//a read by local id names another id than the one invalidated
		if id == "" || this.segments[1] == id || this.query.Get("as") == "local_id" {
			return true
		}
	}
	for _, pattern := range dependency.derived {
		if this.matches(pattern) {
			return true
		}
	}
	return false
}

// InvalidateCache removes the responses cached through WithCache that a change of the resource with
// id may have made stale, for every user and every copy of c: the single reads of that resource,
// reads by local id of its kind, every list and query, and the single reads of other resources derived
// from it, such as the services of a device type. An empty id removes every single read of the kind.
// Implementations other than the one returned by NewClient are left untouched.
func InvalidateCache(c Interface, resource CacheResource, id string) {
	client, ok := c.(*Client)
	if !ok || client.cache == nil {
		return
	}
	client.cache.invalidate(func(target cacheTarget) bool {
		return target.affectedBy(resource, id)
	})
}

// ResetCache removes every response cached through WithCache for every copy of c.
// Implementations other than the one returned by NewClient are left untouched.
func ResetCache(c Interface) {
	client, ok := c.(*Client)
	if !ok || client.cache == nil {
		return
	}
	client.cache.invalidate(func(cacheTarget) bool { return true })
}

// signalResources maps the cache invalidation signals of service-commons to the resources they name.
// signal.Known.GenericCacheInvalidation carries a free key without a resource and is left out.
var signalResources = map[signal.Signal]CacheResource{
	signal.Known.DeviceCacheInvalidation:         CacheDevices,
	signal.Known.DeviceTypeCacheInvalidation:     CacheDeviceTypes,
	signal.Known.DeviceGroupInvalidation:         CacheDeviceGroups,
	signal.Known.ProtocolInvalidation:            CacheProtocols,
	signal.Known.HubCacheInvalidation:            CacheHubs,
	signal.Known.LocationInvalidation:            CacheLocations,
	signal.Known.ConceptCacheInvalidation:        CacheConcepts,
	signal.Known.CharacteristicCacheInvalidation: CacheCharacteristics,
	signal.Known.FunctionCacheInvalidation:       CacheFunctions,
	signal.Known.AspectCacheInvalidation:         CacheAspects,
	signal.Known.DeviceClassCacheInvalidation:    CacheDeviceClasses,
}

// InvalidateOnSignals clears the cache of c on the cache invalidation signals of service-commons on
// broker (signal.DefaultBroker if nil): signal.Known.CacheInvalidationAll resets it, a signal of a
// resource invalidates that resource. The signals come from the kafka consumers of
// service-commons/pkg/cache/invalidator, which the service starts once for all its caches.
// For implementations other than the one returned by NewClient nothing is subscribed.
func InvalidateOnSignals(c Interface, broker *signal.Broker) (unsubscribe func()) {
	if _, ok := c.(*Client); !ok {
		return func() {}
	}
	if broker == nil {
		broker = signal.DefaultBroker
	}
	ids := []string{broker.Sub("", signal.Known.CacheInvalidationAll, func(string, *sync.WaitGroup) { ResetCache(c) })}
	for sig, resource := range signalResources {
		ids = append(ids, broker.Sub("", sig, func(id string, _ *sync.WaitGroup) { InvalidateCache(c, resource, id) }))
	}
	return func() {
		for _, id := range ids {
			broker.Unsub(id)
		}
	}
}

func (this *cacheStore) invalidate(affected func(cacheTarget) bool) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.generation++
	for key, entry := range this.entries {
		if affected(entry.target) {
			delete(this.entries, key)
		}
	}
}
