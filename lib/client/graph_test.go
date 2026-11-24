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
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func TestGraphRequests(t *testing.T) {
	var method, target string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, target = r.Method, r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/graphs/g/segments" {
			_, _ = w.Write([]byte(`[{"from":null,"until":"2025-03-01T00:00:00Z","graph":{"id":"g"}}]`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"g"}`))
	}))
	defer server.Close()
	c := NewClient(server.URL, Options{})
	ctx := context.Background()
	at := time.Date(2025, 6, 30, 12, 0, 0, 0, time.FixedZone("x", 2*3600))
	until := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		call           func() error
		expectedMethod string
		expectedTarget string
	}{
		{"read now", func() error { _, err, _ := c.ReadGraph(ctx, "t", "g"); return err }, "GET", "/graphs/g"},
		{"read at", func() error { _, err, _ := c.ReadGraphAt(ctx, "t", "g", at); return err }, "GET", "/graphs/g?at=2025-06-30T10%3A00%3A00Z"},
		{"read versions", func() error { _, err, _ := c.ReadGraphVersions(ctx, "t", "g"); return err }, "GET", "/graphs/g?versions=all"},
		{"segments without bounds", func() error { _, err, _ := c.ReadGraphSegments(ctx, "t", "g", nil, nil); return err }, "GET", "/graphs/g/segments"},
		{"segments", func() error { _, err, _ := c.ReadGraphSegments(ctx, "t", "g", &at, &until); return err }, "GET", "/graphs/g/segments?end=2025-09-01T00%3A00%3A00Z&start=2025-06-30T10%3A00%3A00Z"},
		{"set is a correction now", func() error { _, err, _ := c.SetGraph(ctx, "t", models.Graph{Id: "g"}); return err }, "PUT", "/graphs/g"},
		{"create", func() error { _, err, _ := c.SetGraph(ctx, "t", models.Graph{}); return err }, "POST", "/graphs"},
		{"move", func() error {
			_, err, _ := c.MoveGraphNode(ctx, "t", "g", "n/1", model.GraphNodeMove{At: &at, ToNodeId: "b"})
			return err
		}, "POST", "/graphs/g/nodes/n%2F1/move"},
		{"end", func() error { _, err, _ := c.EndGraphNode(ctx, "t", "g", "n", model.GraphNodeEnd{At: &at}); return err }, "POST", "/graphs/g/nodes/n/end"},
		{"swap", func() error {
			_, err, _ := c.SwapThingDeviceLink(ctx, "t", "th", "l", model.ThingDeviceLinkSwap{At: &at, DeviceId: "d"})
			return err
		}, "POST", "/things/th/device-links/l/swap"},
		{"edit", func() error {
			_, err, _ := c.EditGraph(ctx, "t", models.Graph{Id: "g"}, model.GraphEditOptions{At: &at, Until: &until, Mode: models.GraphEditModeChange})
			return err
		}, "PUT", "/graphs/g?at=2025-06-30T10%3A00%3A00Z&mode=change&until=2025-09-01T00%3A00%3A00Z"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err != nil {
				t.Fatal(err)
			}
			if method != test.expectedMethod || target != test.expectedTarget {
				t.Errorf("%v %v", method, target)
			}
		})
	}

	t.Run("segments decode an open side as nil", func(t *testing.T) {
		segments, err, _ := c.ReadGraphSegments(ctx, "t", "g", nil, nil)
		if err != nil || len(segments) != 1 || segments[0].From != nil || segments[0].Until == nil {
			t.Errorf("%#v %v", segments, err)
		}
	})
}
