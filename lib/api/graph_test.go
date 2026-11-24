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
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

// graphStub records what the endpoints ask the controller for
type graphStub struct {
	Controller
	calls   []string
	at      *time.Time
	start   *time.Time
	end     *time.Time
	options model.GraphEditOptions
}

func (this *graphStub) ReadGraph(ctx context.Context, token string, id string) (models.Graph, error, int) {
	this.calls = append(this.calls, "ReadGraph")
	return models.Graph{Id: id}, nil, http.StatusOK
}

func (this *graphStub) ReadGraphAt(ctx context.Context, token string, id string, at time.Time) (models.Graph, error, int) {
	this.calls = append(this.calls, "ReadGraphAt")
	this.at = &at
	return models.Graph{Id: id}, nil, http.StatusOK
}

func (this *graphStub) ReadGraphVersions(ctx context.Context, token string, id string) (models.Graph, error, int) {
	this.calls = append(this.calls, "ReadGraphVersions")
	return models.Graph{Id: id}, nil, http.StatusOK
}

func (this *graphStub) ReadGraphSegments(ctx context.Context, token string, id string, start *time.Time, end *time.Time) ([]models.GraphSegment, error, int) {
	this.calls = append(this.calls, "ReadGraphSegments")
	this.start, this.end = start, end
	return []models.GraphSegment{{From: start, Until: end, Graph: models.Graph{Id: id}}}, nil, http.StatusOK
}

func (this *graphStub) EditGraph(ctx context.Context, token string, graph models.Graph, options model.GraphEditOptions) (models.Graph, error, int) {
	this.calls = append(this.calls, "EditGraph")
	this.options = options
	return graph, nil, http.StatusOK
}

func graphRequest(stub *graphStub, method string, target string, body string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	config := configuration.Config{LogLevel: "error", OtelEndpoint: configuration.OtelEndpointOff}
	GetRouterWithoutMiddleware(config, stub).ServeHTTP(response, httptest.NewRequest(method, target, strings.NewReader(body)))
	return response
}

func TestGraphGetQuery(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		code     int
		expected string
	}{
		{"without query it is the graph now", "/graphs/g", http.StatusOK, "ReadGraph"},
		{"at", "/graphs/g?at=2025-06-30T12:00:00Z", http.StatusOK, "ReadGraphAt"},
		{"at with an offset", "/graphs/g?at=2025-06-30T12:00:00%2B02:00", http.StatusOK, "ReadGraphAt"},
		{"versions", "/graphs/g?versions=all", http.StatusOK, "ReadGraphVersions"},
		{"at and versions", "/graphs/g?versions=all&at=2025-06-30T12:00:00Z", http.StatusBadRequest, ""},
		{"unknown versions value", "/graphs/g?versions=some", http.StatusBadRequest, ""},
		{"at that is no time", "/graphs/g?at=yesterday", http.StatusBadRequest, ""},
		{"at without time zone", "/graphs/g?at=2025-06-30T12:00:00", http.StatusBadRequest, ""},
		{"empty at", "/graphs/g?at=", http.StatusBadRequest, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &graphStub{}
			response := graphRequest(stub, http.MethodGet, test.target, "")
			if response.Code != test.code {
				t.Fatalf("%v %v", response.Code, response.Body.String())
			}
			if test.expected == "" {
				if len(stub.calls) != 0 {
					t.Errorf("%v", stub.calls)
				}
				return
			}
			if len(stub.calls) != 1 || stub.calls[0] != test.expected {
				t.Errorf("%v", stub.calls)
			}
		})
	}

	t.Run("the instant reaches the controller", func(t *testing.T) {
		stub := &graphStub{}
		graphRequest(stub, http.MethodGet, "/graphs/g?at=2025-06-30T12:00:00%2B02:00", "")
		if stub.at == nil || !stub.at.Equal(time.Date(2025, 6, 30, 10, 0, 0, 0, time.UTC)) {
			t.Errorf("%v", stub.at)
		}
	})
}

func TestGraphSegmentsQuery(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		code         int
		start, end   bool
		expectedBody string
	}{
		{"both", "/graphs/g/segments?start=2025-01-01T00:00:00Z&end=2025-12-31T00:00:00Z", http.StatusOK, true, true, `"from":"2025-01-01T00:00:00Z","until":"2025-12-31T00:00:00Z"`},
		{"without start", "/graphs/g/segments?end=2025-12-31T00:00:00Z", http.StatusOK, false, true, `"from":null,"until":"2025-12-31T00:00:00Z"`},
		{"without end", "/graphs/g/segments?start=2025-01-01T00:00:00Z", http.StatusOK, true, false, `"from":"2025-01-01T00:00:00Z","until":null`},
		{"without both", "/graphs/g/segments", http.StatusOK, false, false, `"from":null,"until":null`},
		{"start that is no time", "/graphs/g/segments?start=x", http.StatusBadRequest, false, false, ""},
		{"end that is no time", "/graphs/g/segments?end=x", http.StatusBadRequest, false, false, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &graphStub{}
			response := graphRequest(stub, http.MethodGet, test.target, "")
			if response.Code != test.code {
				t.Fatalf("%v %v", response.Code, response.Body.String())
			}
			if test.code != http.StatusOK {
				return
			}
			if (stub.start != nil) != test.start || (stub.end != nil) != test.end {
				t.Errorf("%v %v", stub.start, stub.end)
			}
			if !strings.Contains(response.Body.String(), test.expectedBody) {
				t.Errorf("%v", response.Body.String())
			}
		})
	}
}

func TestGraphPutQuery(t *testing.T) {
	body := `{"id":"g","owner":"o"}`
	tests := []struct {
		name     string
		target   string
		code     int
		expected model.GraphEditOptions
	}{
		{"without query it is a correction now", "/graphs/g", http.StatusOK, model.GraphEditOptions{}},
		{"at and mode", "/graphs/g?at=2025-06-30T00:00:00Z&mode=change", http.StatusOK, model.GraphEditOptions{At: ptrTime(2025, 6, 30), Mode: "change"}},
		{"bounded", "/graphs/g?at=2025-06-30T00:00:00Z&until=2025-09-01T00:00:00Z&mode=change", http.StatusOK, model.GraphEditOptions{At: ptrTime(2025, 6, 30), Until: ptrTime(2025, 9, 1), Mode: "change"}},
		{"at that is no time", "/graphs/g?at=x", http.StatusBadRequest, model.GraphEditOptions{}},
		{"until that is no time", "/graphs/g?until=x&mode=change", http.StatusBadRequest, model.GraphEditOptions{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &graphStub{}
			response := graphRequest(stub, http.MethodPut, test.target, body)
			if response.Code != test.code {
				t.Fatalf("%v %v", response.Code, response.Body.String())
			}
			if test.code != http.StatusOK {
				if len(stub.calls) != 0 {
					t.Errorf("%v", stub.calls)
				}
				return
			}
			o := stub.options
			if o.Mode != test.expected.Mode || !sameTime(o.At, test.expected.At) || !sameTime(o.Until, test.expected.Until) {
				t.Errorf("%#v", o)
			}
		})
	}
}

func ptrTime(year int, month time.Month, day int) *time.Time {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &t
}

func sameTime(a *time.Time, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

type actionStub struct {
	Controller
	move  model.GraphNodeMove
	end   model.GraphNodeEnd
	swap  model.ThingDeviceLinkSwap
	paths []string
}

func (this *actionStub) MoveGraphNode(ctx context.Context, token string, graphId string, nodeId string, move model.GraphNodeMove) (models.Graph, error, int) {
	this.paths = append(this.paths, graphId+"/"+nodeId)
	this.move = move
	return models.Graph{Id: graphId}, nil, http.StatusOK
}

func (this *actionStub) EndGraphNode(ctx context.Context, token string, graphId string, nodeId string, end model.GraphNodeEnd) (models.Graph, error, int) {
	this.paths = append(this.paths, graphId+"/"+nodeId)
	this.end = end
	return models.Graph{Id: graphId}, nil, http.StatusOK
}

func (this *actionStub) SwapThingDeviceLink(ctx context.Context, token string, thingId string, linkId string, swap model.ThingDeviceLinkSwap) (models.Thing, error, int) {
	this.paths = append(this.paths, thingId+"/"+linkId)
	this.swap = swap
	return models.Thing{Id: thingId}, nil, http.StatusOK
}

func TestGuidedActionEndpoints(t *testing.T) {
	run := func(stub *actionStub, target string, body string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		config := configuration.Config{LogLevel: "error", OtelEndpoint: configuration.OtelEndpointOff}
		GetRouterWithoutMiddleware(config, stub).ServeHTTP(response, httptest.NewRequest(http.MethodPost, target, strings.NewReader(body)))
		return response
	}
	t.Run("move", func(t *testing.T) {
		stub := &actionStub{}
		response := run(stub, "/graphs/g/nodes/n/move", `{"at":"2025-06-30T00:00:00Z","to_node_id":"b","until":"2025-09-01T00:00:00Z"}`)
		if response.Code != http.StatusOK || stub.paths[0] != "g/n" || stub.move.ToNodeId != "b" || !sameTime(stub.move.At, ptrTime(2025, 6, 30)) || !sameTime(stub.move.Until, ptrTime(2025, 9, 1)) {
			t.Errorf("%v %v %#v", response.Code, response.Body.String(), stub.move)
		}
		if response := run(&actionStub{}, "/graphs/g/nodes/n/move", `{"at":"yesterday"}`); response.Code != http.StatusBadRequest {
			t.Errorf("%v", response.Code)
		}
	})
	t.Run("end, also with an empty body", func(t *testing.T) {
		stub := &actionStub{}
		response := run(stub, "/graphs/g/nodes/n/end", `{"at":"2025-06-30T00:00:00Z"}`)
		if response.Code != http.StatusOK || !sameTime(stub.end.At, ptrTime(2025, 6, 30)) {
			t.Errorf("%v %v", response.Code, response.Body.String())
		}
		stub = &actionStub{}
		if response := run(stub, "/graphs/g/nodes/n/end", ``); response.Code != http.StatusOK || stub.end.At != nil {
			t.Errorf("%v", response.Code)
		}
	})
	t.Run("swap", func(t *testing.T) {
		stub := &actionStub{}
		body := `{"at":"2025-06-30T00:00:00Z","until_values":[{"service_id":"s","column":"c","until_value":5}],"device_id":"d2","readings":[{"service_id":"s","column":"c","from_value":1}],"factor":40}`
		response := run(stub, "/things/t/device-links/l/swap", body)
		if response.Code != http.StatusOK || stub.paths[0] != "t/l" || stub.swap.DeviceId != "d2" || len(stub.swap.UntilValues) != 1 || stub.swap.UntilValues[0].UntilValue != 5 || len(stub.swap.Readings) != 1 || stub.swap.Factor == nil || *stub.swap.Factor != 40 {
			t.Errorf("%v %v %#v", response.Code, response.Body.String(), stub.swap)
		}
		if response := run(&actionStub{}, "/things/t/device-links/l/swap", `not json`); response.Code != http.StatusBadRequest {
			t.Errorf("%v", response.Code)
		}
	})
}
