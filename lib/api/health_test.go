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

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
)

type healthStub struct {
	Controller
	healthy bool
	report  string
}

func (this healthStub) MirrorHealth() (bool, string) { return this.healthy, this.report }

func getHealth(config configuration.Config, ctrl Controller) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	GetRouter(config, ctrl).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	return response
}

func TestHealthPlatformIsAlwaysOk(t *testing.T) {
	response := getHealth(configuration.Config{LogLevel: "error", OtelEndpoint: configuration.OtelEndpointOff}, healthStub{healthy: false, report: "stale"})
	if response.Code != http.StatusOK || response.Body.String() != "ok\n" {
		t.Errorf("got %v %q", response.Code, response.Body.String())
	}
}

func TestHealthMirrorReportsStaleCollections(t *testing.T) {
	config := configuration.Config{LogLevel: "error", AsMgwMirror: true, OtelEndpoint: configuration.OtelEndpointOff}
	report := "functions: not synced since 2026-01-01T12:00:00Z: unexpected statuscode 403"

	response := getHealth(config, healthStub{healthy: false, report: report})
	if response.Code != http.StatusServiceUnavailable || response.Body.String() != report+"\n" {
		t.Errorf("stale mirror: got %v %q", response.Code, response.Body.String())
	}

	response = getHealth(config, healthStub{healthy: true, report: "ok"})
	if response.Code != http.StatusOK || response.Body.String() != "ok\n" {
		t.Errorf("current mirror: got %v %q", response.Code, response.Body.String())
	}
}
