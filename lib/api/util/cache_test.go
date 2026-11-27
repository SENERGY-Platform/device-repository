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
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
)

func TestParseIfNoneMatch(t *testing.T) {
	etag := `"abc"`
	for name, tc := range map[string]struct {
		headers []string
		any     bool
		match   bool
	}{
		"none":            {},
		"exact":           {headers: []string{etag}, match: true},
		"weak":            {headers: []string{`W/"abc"`}, match: true},
		"list":            {headers: []string{`"x", "abc"`}, match: true},
		"list no space":   {headers: []string{`"x","abc"`}, match: true},
		"two headers":     {headers: []string{`"x"`, `"abc"`}, match: true},
		"star":            {headers: []string{`*`}, any: true, match: true},
		"other":           {headers: []string{`"abcd"`}},
		"unquoted":        {headers: []string{`abc`}},
		"unterminated":    {headers: []string{`"abc`}},
		"comma in tag":    {headers: []string{`"a,b", "abc"`}, match: true},
		"garbage":         {headers: []string{`,,W/, x`}},
		"garbage then ok": {headers: []string{`x, "abc"`}, match: true},
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for _, h := range tc.headers {
				req.Header.Add("If-None-Match", h)
			}
			result := newIfNoneMatch(req)
			if result.any != tc.any {
				t.Errorf("any: got %v", result.any)
			}
			if result.matches(etag) != tc.match {
				t.Errorf("match: got %v", result.matches(etag))
			}
		})
	}
}

func TestCacheMiddleware(t *testing.T) {
	var handlerCalls atomic.Int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalls.Add(1)
		if r.URL.Query().Get("fail") != "" {
			http.Error(w, "fail", http.StatusBadRequest)
			return
		}
		w.Header().Set("X-Total-Count", "3")
		_, _ = w.Write([]byte("body of " + r.URL.Path))
	})
	stamps := []model.LastUpdateTimestamp{{Collection: "b", UnixTimestamp: 2}, {Collection: "a", UnixTimestamp: 1}}
	var stampErr error
	timestamps := func(ctx context.Context) ([]model.LastUpdateTimestamp, error) { return stamps, stampErr }
	cache := NewCache(handler, timestamps)

	do := func(method string, target string, inm string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, target, nil)
		if inm != "" {
			req.Header.Set("If-None-Match", inm)
		}
		rec := httptest.NewRecorder()
		cache.ServeHTTP(rec, req)
		return rec
	}

	t.Run("timestamp group answers 304 without the handler", func(t *testing.T) {
		first := do(http.MethodGet, "/v2/aspects", "")
		etag := first.Header().Get("ETag")
		if first.Code != 200 || etag == "" || first.Header().Get("Cache-Control") != "private, no-cache" {
			t.Fatalf("%v %v", first.Code, first.Header())
		}
		calls := handlerCalls.Load()
		second := do(http.MethodGet, "/v2/aspects", etag)
		if second.Code != http.StatusNotModified || second.Body.Len() != 0 || handlerCalls.Load() != calls {
			t.Fatalf("%v %q calls %v", second.Code, second.Body.String(), handlerCalls.Load()-calls)
		}
		stamps[0].UnixTimestamp = 3
		third := do(http.MethodGet, "/v2/aspects", etag)
		if third.Code != 200 || third.Header().Get("ETag") == etag {
			t.Fatalf("%v %v", third.Code, third.Header())
		}
	})

	t.Run("the order of the timestamps does not matter, the url does", func(t *testing.T) {
		a := do(http.MethodGet, "/v2/aspects", "").Header().Get("ETag")
		stamps[0], stamps[1] = stamps[1], stamps[0]
		b := do(http.MethodGet, "/v2/aspects", "").Header().Get("ETag")
		if a != b {
			t.Errorf("%v %v", a, b)
		}
		c := do(http.MethodGet, "/v2/aspects?limit=1", "").Header().Get("ETag")
		if a == c {
			t.Errorf("expected another ETag for another url")
		}
	})

	t.Run("errors carry no cache headers", func(t *testing.T) {
		etag := do(http.MethodGet, "/v2/aspects", "").Header().Get("ETag")
		resp := do(http.MethodGet, "/v2/aspects?fail=1", etag)
		if resp.Code != 400 || resp.Header().Get("ETag") != "" || resp.Header().Get("Cache-Control") != "" {
			t.Errorf("%v %v", resp.Code, resp.Header())
		}
	})

	t.Run("response group hashes body and total", func(t *testing.T) {
		first := do(http.MethodGet, "/devices", "")
		etag := first.Header().Get("ETag")
		if first.Code != 200 || etag == "" {
			t.Fatalf("%v %v", first.Code, first.Header())
		}
		calls := handlerCalls.Load()
		second := do(http.MethodGet, "/devices", etag)
		if second.Code != http.StatusNotModified || second.Body.Len() != 0 {
			t.Fatalf("%v", second.Code)
		}
		if handlerCalls.Load() != calls+1 {
			t.Errorf("the handler must run for group B")
		}
		if do(http.MethodGet, "/devices/x", etag).Code != 200 {
			t.Errorf("another path must not match")
		}
	})

	t.Run("fails safe", func(t *testing.T) {
		for _, tc := range []struct{ method, target string }{
			{http.MethodHead, "/hubs/x"},
			{http.MethodHead, "/v2/aspects"},
			{http.MethodPost, "/devices"},
			{http.MethodGet, "/last-update-timestamps"},
			{http.MethodGet, "/export"},
			{http.MethodGet, "/health"},
			{http.MethodGet, "/hubs/x/unknown"},
		} {
			resp := do(tc.method, tc.target, `*`)
			if resp.Code != 200 || resp.Header().Get("ETag") != "" || resp.Header().Get("Cache-Control") != "" {
				t.Errorf("%v %v: %v %v", tc.method, tc.target, resp.Code, resp.Header())
			}
		}
	})

	t.Run("timestamp source failing or missing passes group A through", func(t *testing.T) {
		stampErr = errors.New("db down")
		resp := do(http.MethodGet, "/v2/aspects", "")
		if resp.Code != 200 || resp.Header().Get("ETag") != "" {
			t.Errorf("%v %v", resp.Code, resp.Header())
		}
		stampErr = nil
		without := NewCache(handler, nil)
		rec := httptest.NewRecorder()
		without.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v2/aspects", nil))
		if rec.Code != 200 || rec.Header().Get("ETag") != "" {
			t.Errorf("%v %v", rec.Code, rec.Header())
		}
		rec = httptest.NewRecorder()
		without.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/devices", nil))
		if rec.Header().Get("ETag") == "" {
			t.Errorf("group B does not need the timestamps")
		}
	})
}
