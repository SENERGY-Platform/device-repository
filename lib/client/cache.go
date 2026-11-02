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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

// WithCache returns a copy of c whose read methods (Read*, Get*, List*) answer from a cache
// that is shared by all copies of the same client. Writes, validations and GetLastUpdateTimestamps
// are not cached. Implementations other than the one returned by NewClient are returned unchanged.
//
// A cached response is used while it is younger than the ttl it was stored with and younger than the
// ttl of the reading copy, so a caller that needs fresher data than another caller stored gets it.
// Errors are not cached.
//
// Responses are cached per request and per user id and roles of its token. The token is parsed
// without validation: a refreshed token of the same user reuses the cached responses, and an
// expired or revoked token may still be answered from the cache until the ttl ends.
// A token that can not be parsed bypasses the cache.
func WithCache(c Interface, ttl time.Duration) Interface {
	client, ok := c.(*Client)
	if !ok || client.cache == nil {
		return c
	}
	result := *client
	result.cacheTtl = ttl
	return &result
}

// doCached is do for read requests, which are cached if the client has a cache ttl
func doCached[T any](c *Client, req *http.Request) (result T, err error, code int) {
	resp, err, code := c.fetchCached(req)
	if err != nil {
		return result, err, code
	}
	err = json.Unmarshal(resp.body, &result)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, 0 //like do, which leaves the code unset on success
}

// doCachedWithTotal is doWithTotalInResult for read requests, which are cached if the client has a cache ttl
func doCachedWithTotal[T any](c *Client, req *http.Request) (result T, total int64, err error, code int) {
	resp, err, code := c.fetchCached(req)
	if err != nil {
		return result, total, err, code
	}
	total, err = strconv.ParseInt(resp.total, 10, 64)
	if err != nil {
		return result, total, fmt.Errorf("unable to read X-Total-Count header %w", err), http.StatusInternalServerError
	}
	err = json.Unmarshal(resp.body, &result)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, 0 //like doWithTotalInResult, which leaves the code unset on success
}

// cachedResponse is only read after it is stored, so readers may share it
type cachedResponse struct {
	body  []byte
	total string
}

func (c *Client) fetchCached(req *http.Request) (resp cachedResponse, err error, code int) {
	if c.optionalAuthTokenForApiGatewayRequest != nil && req.Header.Get("Authorization") == "" {
		token, err := c.optionalAuthTokenForApiGatewayRequest()
		if err != nil {
			return resp, err, http.StatusInternalServerError
		}
		req.Header.Set("Authorization", token)
	}
	if c.cacheTtl <= 0 || c.cache == nil {
		return c.fetch(req)
	}
	key, ok := requestCacheKey(req)
	if !ok {
		return c.fetch(req)
	}
	if resp, hit := c.cache.get(key, c.cacheTtl); hit {
		return resp, nil, 0
	}
	generation := c.cache.currentGeneration()
	resp, err, code = c.fetch(req)
	if err == nil {
		c.cache.set(key, newCacheTarget(c.basePath, req.URL), resp, c.cacheTtl, generation)
	}
	return resp, err, code
}

func (c *Client) fetch(req *http.Request) (result cachedResponse, err error, code int) {
	resp, err := c.send(req)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 299 {
		return result, fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(body)), resp.StatusCode
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return cachedResponse{body: body, total: resp.Header.Get("X-Total-Count")}, nil, 0
}

// requestCacheKey returns false if the token can not be parsed or the body can not be read twice
func requestCacheKey(req *http.Request) (key string, ok bool) {
	user := ""
	if token := req.Header.Get("Authorization"); token != "" {
		parsed, err := jwt.Parse(token)
		if err != nil {
			return "", false
		}
		roles := slices.Clone(parsed.GetRoles())
		slices.Sort(roles)
		user = parsed.GetUserId() + "|" + strings.Join(roles, ",")
	}
	body := ""
	if req.Body != nil && req.Body != http.NoBody {
		if req.GetBody == nil {
			return "", false
		}
		reader, err := req.GetBody()
		if err != nil {
			return "", false
		}
		temp, err := io.ReadAll(reader)
		if err != nil {
			return "", false
		}
		body = string(temp)
	}
	return req.Method + " " + req.URL.String() + "\x00" + user + "\x00" + body, true
}

const cacheCleanupInterval = time.Minute

type cacheStore struct {
	mux         sync.Mutex
	entries     map[string]cacheEntry
	lastCleanup time.Time
	now         func() time.Time
	//counts invalidations, so that a response requested before one is not stored after it
	generation uint64
}

type cacheEntry struct {
	response cachedResponse
	target   cacheTarget
	stored   time.Time
	expires  time.Time
}

func newCacheStore() *cacheStore {
	return &cacheStore{entries: map[string]cacheEntry{}, lastCleanup: time.Now(), now: time.Now}
}

func (this *cacheStore) get(key string, ttl time.Duration) (response cachedResponse, ok bool) {
	this.mux.Lock()
	defer this.mux.Unlock()
	entry, ok := this.entries[key]
	if !ok {
		return response, false
	}
	now := this.now()
	if !now.Before(entry.expires) || now.Sub(entry.stored) >= ttl {
		return response, false
	}
	return entry.response, true
}

func (this *cacheStore) currentGeneration() uint64 {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.generation
}

// set stores the response unless the cache was invalidated since generation was read before the request
func (this *cacheStore) set(key string, target cacheTarget, response cachedResponse, ttl time.Duration, generation uint64) {
	this.mux.Lock()
	defer this.mux.Unlock()
	if generation != this.generation {
		return
	}
	now := this.now()
	this.entries[key] = cacheEntry{response: response, target: target, stored: now, expires: now.Add(ttl)}
	if now.Sub(this.lastCleanup) >= cacheCleanupInterval {
		this.lastCleanup = now
		for k, entry := range this.entries {
			if !now.Before(entry.expires) {
				delete(this.entries, k)
			}
		}
	}
}
