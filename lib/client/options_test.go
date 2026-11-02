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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/models/go/models"
)

func TestOptionsTimeoutBoundsARequest(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	defer close(release)

	done := make(chan error, 1)
	go func() {
		_, err, _ := NewClient(server.URL, Options{Timeout: 50 * time.Millisecond}).GetFunction(context.Background(), "f")
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Error("a request past the timeout has to fail")
		}
	case <-time.After(5 * time.Second):
		t.Error("the request was not bounded by the timeout")
	}
}

type recordingTransport struct {
	mux      sync.Mutex
	requests []*http.Request
}

func (this *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	this.mux.Lock()
	this.requests = append(this.requests, req)
	this.mux.Unlock()
	return http.DefaultTransport.RoundTrip(req)
}

func TestOptionsHttpClientSendsTheRequestsAndKeepsItsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(models.Function{Id: "f"})
	}))
	defer server.Close()
	transport := &recordingTransport{}
	given := &http.Client{Transport: transport, Timeout: time.Hour}

	if _, err, _ := NewClient(server.URL, Options{HttpClient: given, Timeout: time.Minute}).GetFunction(context.Background(), "f"); err != nil {
		t.Fatal(err)
	}
	if len(transport.requests) != 1 {
		t.Errorf("the request has to go through the given client, got %v requests", len(transport.requests))
	}
	if given.Timeout != time.Hour {
		t.Errorf("the timeout is set on a copy, the given client changed to %v", given.Timeout)
	}
}

func TestOptionsTokenForTheApiGatewayIsSentWithoutOwnToken(t *testing.T) {
	var mux sync.Mutex
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.Lock()
		authorization = r.Header.Get("Authorization")
		mux.Unlock()
		_ = json.NewEncoder(w).Encode(models.Function{Id: "f"})
	}))
	defer server.Close()
	token := func() (string, error) { return "Bearer gateway", nil }

	if _, err, _ := NewClient(server.URL, Options{OptionalAuthTokenForApiGatewayRequest: token}).GetFunction(context.Background(), "f"); err != nil {
		t.Fatal(err)
	}
	mux.Lock()
	defer mux.Unlock()
	if authorization != "Bearer gateway" {
		t.Errorf("Authorization %q", authorization)
	}
}
