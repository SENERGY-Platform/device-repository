/*
 * Copyright 2025 InfAI (CC SES)
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

package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunHealthcheck(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok\n"))
		}))
		defer srv.Close()
		out := &bytes.Buffer{}
		if code := runHealthcheck(srv.URL, out); code != 0 {
			t.Fatalf("exit code %d", code)
		}
		if out.String() != "ok\n" {
			t.Fatalf("unexpected output %q", out.String())
		}
	})
	t.Run("unhealthy prints the reason", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("devices stale\nhubs stale\n"))
		}))
		defer srv.Close()
		out := &bytes.Buffer{}
		if code := runHealthcheck(srv.URL, out); code != 1 {
			t.Fatalf("exit code %d", code)
		}
		if out.String() != "devices stale\nhubs stale\n" {
			t.Fatalf("unexpected output %q", out.String())
		}
	})
	t.Run("unreachable", func(t *testing.T) {
		srv := httptest.NewServer(http.NotFoundHandler())
		url := srv.URL
		srv.Close()
		out := &bytes.Buffer{}
		if code := runHealthcheck(url, out); code != 1 {
			t.Fatalf("exit code %d", code)
		}
		if out.Len() == 0 {
			t.Fatal("expected the error on stdout")
		}
	})
}
