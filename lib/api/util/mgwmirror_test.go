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

package util

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
)

type pullFunc func() error

func (f pullFunc) MirrorUpdate() error {
	return f()
}

func TestMirrorMiddlewareNotFoundPull(t *testing.T) {
	config := configuration.Config{MgwMirrorUserId: "user", MgwMirrorMissPullTimeout: "200ms"}

	//answers 404 until pulled is set
	newHandler := func(pulled *atomic.Bool) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !pulled.Load() {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			w.Header().Set("X-Test", "found")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("found"))
		})
	}

	t.Run("pull finds the entry", func(t *testing.T) {
		pulled := &atomic.Bool{}
		m := NewMirrorMiddleware(newHandler(pulled), config, pullFunc(func() error {
			pulled.Store(true)
			return nil
		}))
		resp := httptest.NewRecorder()
		m.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/devices/d1", nil))
		if resp.Code != http.StatusOK || resp.Body.String() != "found" || resp.Header().Get("X-Test") != "found" {
			t.Error(resp.Code, resp.Body.String(), resp.Header())
		}
	})

	t.Run("pull without the entry", func(t *testing.T) {
		pulls := atomic.Int64{}
		m := NewMirrorMiddleware(newHandler(&atomic.Bool{}), config, pullFunc(func() error {
			pulls.Add(1)
			return nil
		}))
		resp := httptest.NewRecorder()
		m.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/devices/d1", nil))
		if resp.Code != http.StatusNotFound || pulls.Load() != 1 {
			t.Error(resp.Code, pulls.Load())
		}
	})

	t.Run("failed pull", func(t *testing.T) {
		m := NewMirrorMiddleware(newHandler(&atomic.Bool{}), config, pullFunc(func() error {
			return errors.New("test")
		}))
		resp := httptest.NewRecorder()
		m.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/devices/d1", nil))
		if resp.Code != http.StatusNotFound {
			t.Error(resp.Code)
		}
	})

	t.Run("pull exceeds timeout", func(t *testing.T) {
		release := make(chan struct{})
		defer close(release)
		m := NewMirrorMiddleware(newHandler(&atomic.Bool{}), config, pullFunc(func() error {
			<-release
			return nil
		}))
		start := time.Now()
		resp := httptest.NewRecorder()
		m.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/devices/d1", nil))
		if resp.Code != http.StatusNotFound {
			t.Error(resp.Code)
		}
		if d := time.Since(start); d > time.Second {
			t.Error("timeout not applied", d)
		}
	})

	t.Run("found entry does not pull", func(t *testing.T) {
		pulled := &atomic.Bool{}
		pulled.Store(true)
		pulls := atomic.Int64{}
		m := NewMirrorMiddleware(newHandler(pulled), config, pullFunc(func() error {
			pulls.Add(1)
			return nil
		}))
		resp := httptest.NewRecorder()
		m.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/devices/d1", nil))
		if resp.Code != http.StatusOK || pulls.Load() != 0 {
			t.Error(resp.Code, pulls.Load())
		}
	})

	t.Run("query post does not pull", func(t *testing.T) {
		pulls := atomic.Int64{}
		m := NewMirrorMiddleware(newHandler(&atomic.Bool{}), config, pullFunc(func() error {
			pulls.Add(1)
			return nil
		}))
		resp := httptest.NewRecorder()
		m.ServeHTTP(resp, httptest.NewRequest(http.MethodPost, "/query/devices", nil))
		if resp.Code != http.StatusNotFound || pulls.Load() != 0 {
			t.Error(resp.Code, pulls.Load())
		}
	})
}
