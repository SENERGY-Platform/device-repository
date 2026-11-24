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

package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/database"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	permclient "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

// The bounds of these tests lie far in the past or far in the future, so the projection "now" does not depend on the clock.
func gTime(s string) *time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return &t
}

const (
	gJan2020   = "2020-01-01T00:00:00Z"
	gJun2020   = "2020-06-30T00:00:00Z"
	gDec2020   = "2020-12-01T00:00:00Z"
	gFarFuture = "2090-01-01T00:00:00Z"
)

// meterTestGraph is a meter m that feeds a with 90 and b with 10, both below root.
func meterTestGraph() models.Graph {
	return models.Graph{
		Attributes: []models.Attribute{{Key: "name", Value: "g"}},
		Nodes:      []models.Node{{Id: "root"}, {Id: "a"}, {Id: "b"}, {Id: "m"}},
		Edges: []models.Edge{
			{Id: "a>root", FromNodeId: "a", ToNodeId: "root", Weight: 100},
			{Id: "b>root", FromNodeId: "b", ToNodeId: "root", Weight: 100},
			{Id: "m>a", FromNodeId: "m", ToNodeId: "a", Weight: 90},
			{Id: "m>b", FromNodeId: "m", ToNodeId: "b", Weight: 10},
		},
	}
}

func edgeWeights(graph models.Graph) map[string]int {
	result := map[string]int{}
	for _, edge := range graph.Edges {
		result[edge.FromNodeId+">"+edge.ToNodeId] = edge.Weight
	}
	return result
}

func setMeterWeights(graph models.Graph, a int, b int) models.Graph {
	graph.Edges = slices.Clone(graph.Edges)
	for i, edge := range graph.Edges {
		if edge.Id == "m>a" || (edge.FromNodeId == "m" && edge.ToNodeId == "a") {
			graph.Edges[i].Weight = a
		}
		if edge.Id == "m>b" || (edge.FromNodeId == "m" && edge.ToNodeId == "b") {
			graph.Edges[i].Weight = b
		}
	}
	return graph
}

func TestGraphEditThroughController(t *testing.T) {
	ctrl, _, p := newGraphTestController(t)
	ctx := context.Background()

	created, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, meterTestGraph())
	expectCode(t, err, code, http.StatusOK)

	versions := func(t *testing.T) models.Graph {
		t.Helper()
		g, err, code := ctrl.ReadGraphVersions(ctx, graphTestUserJwt, created.Id)
		expectCode(t, err, code, http.StatusOK)
		return g
	}

	t.Run("a change of the weights from a past date is two versions per edge, the answer is the graph now", func(t *testing.T) {
		current := setMeterWeights(created, 70, 30)
		result, err, code := ctrl.EditGraph(ctx, graphTestUserJwt, current, model.GraphEditOptions{At: gTime(gJun2020), Mode: models.GraphEditModeChange})
		expectCode(t, err, code, http.StatusOK)
		if w := edgeWeights(result); w["m>a"] != 70 || w["m>b"] != 30 || len(result.Edges) != 4 {
			t.Errorf("the answer is not the graph now: %#v", result.Edges)
		}
		stored := versions(t)
		if len(stored.Edges) != 6 {
			t.Fatalf("%#v", stored.Edges)
		}
		before, err, code := ctrl.ReadGraphAt(ctx, graphTestUserJwt, created.Id, *gTime("2020-03-01T00:00:00Z"))
		expectCode(t, err, code, http.StatusOK)
		if w := edgeWeights(before); w["m>a"] != 90 || w["m>b"] != 10 {
			t.Errorf("%#v", before.Edges)
		}
		onTheDay, err, code := ctrl.ReadGraphAt(ctx, graphTestUserJwt, created.Id, *gTime(gJun2020))
		expectCode(t, err, code, http.StatusOK)
		if w := edgeWeights(onTheDay); w["m>a"] != 70 || w["m>b"] != 30 {
			t.Errorf("the instant the new version starts belongs to it: %#v", onTheDay.Edges)
		}
		justBefore, err, code := ctrl.ReadGraphAt(ctx, graphTestUserJwt, created.Id, gJun2020Minus(time.Millisecond))
		expectCode(t, err, code, http.StatusOK)
		if w := edgeWeights(justBefore); w["m>a"] != 90 {
			t.Errorf("%#v", justBefore.Edges)
		}
	})

	t.Run("a PUT without at keeps the ended versions: read, modify, write must not lose history", func(t *testing.T) {
		now, err, code := ctrl.ReadGraph(ctx, graphTestUserJwt, created.Id)
		expectCode(t, err, code, http.StatusOK)
		now.Attributes = []models.Attribute{{Key: "name", Value: "renamed"}}
		for i := range now.Nodes {
			now.Nodes[i].Attributes = nil
		}
		result, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, now)
		expectCode(t, err, code, http.StatusOK)
		if len(result.Edges) != 4 || result.Attributes[0].Value != "renamed" {
			t.Errorf("%#v", result)
		}
		stored := versions(t)
		ended := 0
		for _, edge := range stored.Edges {
			if edge.ValidUntil != nil {
				ended++
			}
		}
		if ended != 2 || len(stored.Edges) != 6 {
			t.Errorf("the history was lost: %#v", stored.Edges)
		}
		last := p.published[len(p.published)-1]
		if len(last.Edges) != 6 {
			t.Errorf("the stored document is published, with all versions: %#v", last.Edges)
		}
	})

	t.Run("an edit that leaves a gap names the first invalid instant", func(t *testing.T) {
		current, _, _ := ctrl.ReadGraph(ctx, graphTestUserJwt, created.Id)
		current = setMeterWeights(current, 60, 30) // sums to 90 from the day on
		_, err, code := ctrl.EditGraph(ctx, graphTestUserJwt, current, model.GraphEditOptions{At: gTime(gDec2020), Mode: models.GraphEditModeChange})
		expectCode(t, err, code, http.StatusBadRequest)
		if !strings.Contains(err.Error(), "invalid at 2020-12-01T00:00:00Z") || !strings.Contains(err.Error(), "got 90") {
			t.Error(err)
		}
		if len(versions(t).Edges) != 6 {
			t.Error("a refused edit must not write")
		}
	})

	t.Run("segments and the empty projection", func(t *testing.T) {
		segments, err, code := ctrl.ReadGraphSegments(ctx, graphTestUserJwt, created.Id, nil, nil)
		expectCode(t, err, code, http.StatusOK)
		if len(segments) != 2 || segments[0].From != nil || segments[1].Until != nil || !segments[0].Until.Equal(*gTime(gJun2020)) {
			t.Fatalf("%#v", segments)
		}
		_, err, code = ctrl.ReadGraphSegments(ctx, graphTestUserJwt, created.Id, gTime(gDec2020), gTime(gJun2020))
		expectCode(t, err, code, http.StatusBadRequest)

		//a graph whose nodes all start later is there, without content
		later := meterTestGraph()
		for i := range later.Nodes {
			later.Nodes[i].ValidFrom = gTime(gDec2020)
		}
		for i := range later.Edges {
			later.Edges[i].ValidFrom = gTime(gDec2020)
		}
		later, err, code = ctrl.SetGraph(ctx, graphTestUserJwt, later)
		expectCode(t, err, code, http.StatusOK)
		before, err, code := ctrl.ReadGraphAt(ctx, graphTestUserJwt, later.Id, *gTime(gJan2020))
		expectCode(t, err, code, http.StatusOK)
		if before.Id != later.Id || before.Owner != graphTestUserId || len(before.Nodes) != 0 || len(before.Edges) != 0 || len(before.Attributes) != 1 {
			t.Errorf("%#v", before)
		}
		_, err, code = ctrl.ReadGraphAt(ctx, graphTestUserJwt, "unknown", *gTime(gJan2020))
		expectCode(t, err, code, http.StatusForbidden)
	})

	t.Run("the instants are cut to milliseconds in UTC", func(t *testing.T) {
		g, _, _ := ctrl.SetGraph(ctx, graphTestUserJwt, meterTestGraph())
		current := setMeterWeights(g, 80, 20)
		_, err, code := ctrl.EditGraph(ctx, graphTestUserJwt, current, model.GraphEditOptions{At: gTime("2020-03-04T05:06:07.123456789+02:00"), Mode: models.GraphEditModeChange})
		expectCode(t, err, code, http.StatusOK)
		stored, _, _ := ctrl.ReadGraphVersions(ctx, graphTestUserJwt, g.Id)
		found := false
		for _, edge := range stored.Edges {
			if edge.ValidUntil != nil {
				found = true
				if !edge.ValidUntil.Equal(*gTime("2020-03-04T03:06:07.123Z")) || edge.ValidUntil.Location() != time.UTC {
					t.Errorf("%v", edge.ValidUntil)
				}
			}
		}
		if !found {
			t.Error("no version ended")
		}
	})

	t.Run("wrong options are a 400", func(t *testing.T) {
		current, _, _ := ctrl.ReadGraph(ctx, graphTestUserJwt, created.Id)
		for name, options := range map[string]model.GraphEditOptions{
			"unknown mode":             {Mode: "replace"},
			"until for a correction":   {Until: gTime(gFarFuture)},
			"until before at":          {At: gTime(gDec2020), Until: gTime(gJun2020), Mode: models.GraphEditModeChange},
			"until equal to at":        {At: gTime(gDec2020), Until: gTime(gDec2020), Mode: models.GraphEditModeChange},
			"until for explicit corr.": {Until: gTime(gFarFuture), Mode: models.GraphEditModeCorrection},
		} {
			_, err, code := ctrl.EditGraph(ctx, graphTestUserJwt, current, options)
			if code != http.StatusBadRequest || err == nil {
				t.Errorf("%v: %v %v", name, code, err)
			}
		}
	})

	t.Run("somebody without write access can not edit", func(t *testing.T) {
		current, _, _ := ctrl.ReadGraph(ctx, graphTestUserJwt, created.Id)
		_, err, code := ctrl.EditGraph(ctx, otherUserJwt(), current, model.GraphEditOptions{Mode: models.GraphEditModeChange})
		if code != http.StatusForbidden || err == nil {
			t.Errorf("%v %v", code, err)
		}
		_, err, code = ctrl.ReadGraphVersions(ctx, otherUserJwt(), created.Id)
		if code != http.StatusForbidden || err == nil {
			t.Errorf("%v %v", code, err)
		}
	})
}

func gJun2020Minus(d time.Duration) time.Time {
	return gTime(gJun2020).Add(-d)
}

func TestGraphBoundedChangeThroughController(t *testing.T) {
	ctrl, _, _ := newGraphTestController(t)
	ctx := context.Background()
	g := models.Graph{
		Nodes: []models.Node{{Id: "root"}, {Id: "pa"}, {Id: "pb"}, {Id: "x", ResourceType: "device", ResourceId: "dx"}},
		Edges: []models.Edge{
			{Id: "pa>root", FromNodeId: "pa", ToNodeId: "root", Weight: 100},
			{Id: "pb>root", FromNodeId: "pb", ToNodeId: "root", Weight: 100},
			{Id: "x>pa", FromNodeId: "x", ToNodeId: "pa", Weight: 100},
		},
	}
	g, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, g)
	expectCode(t, err, code, http.StatusOK)

	current := g
	current.Edges = slices.Clone(g.Edges)
	current.Edges[2] = models.Edge{Id: "x>pb", FromNodeId: "x", ToNodeId: "pb", Weight: 100}
	_, err, code = ctrl.EditGraph(ctx, graphTestUserJwt, current, model.GraphEditOptions{At: gTime("2020-01-01T00:00:00Z"), Until: gTime("2020-04-01T00:00:00Z"), Mode: models.GraphEditModeChange})
	expectCode(t, err, code, http.StatusOK)

	segments, err, code := ctrl.ReadGraphSegments(ctx, graphTestUserJwt, g.Id, nil, nil)
	expectCode(t, err, code, http.StatusOK)
	parents := []string{}
	for _, s := range segments {
		for _, e := range s.Graph.Edges {
			if e.FromNodeId == "x" {
				parents = append(parents, e.ToNodeId)
			}
		}
	}
	if !slices.Equal(parents, []string{"pa", "pb", "pa"}) {
		t.Errorf("%v", parents)
	}
}

// otherUserJwt is a token of a user that owns nothing: the subject is not graphTestUserId.
func otherUserJwt() string {
	header := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	claims, _ := json.Marshal(map[string]any{"sub": "someoneElse", "realm_access": map[string]any{"roles": []string{"user"}}})
	return "Bearer " + header + "." + base64.RawURLEncoding.EncodeToString(claims) + ".sig"
}

func TestGraphThingRuleIsTimeAware(t *testing.T) {
	ctrl, permClient, _ := newThingTestController(t)
	ctx := context.Background()
	thingTestDevice(t, ctrl, permClient, "dev", "")
	thing := thingTestCreate(t, ctrl, "meter")
	// the device measures the thing from 2020-06-01 on
	thing, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "dev", From: gTime("2020-06-01T00:00:00Z")})
	expectCode(t, err, code, http.StatusOK)

	build := func(deviceFrom, deviceUntil, thingFrom, thingUntil string) models.Graph {
		ptr := func(s string) *time.Time {
			if s == "" {
				return nil
			}
			return gTime(s)
		}
		return models.Graph{
			Attributes: []models.Attribute{{Key: "name", Value: "g"}},
			Nodes: []models.Node{
				{Id: "end"},
				{Id: "d", ResourceType: models.GraphResourceTypeDevice, ResourceId: "dev", ValidFrom: ptr(deviceFrom), ValidUntil: ptr(deviceUntil)},
				{Id: "t", ResourceType: models.GraphResourceTypeThing, ResourceId: thing.Id, ValidFrom: ptr(thingFrom), ValidUntil: ptr(thingUntil)},
			},
			Edges: []models.Edge{
				{Id: "d>end", FromNodeId: "d", ToNodeId: "end", Weight: 100, ValidFrom: ptr(deviceFrom), ValidUntil: ptr(deviceUntil)},
				{Id: "t>end", FromNodeId: "t", ToNodeId: "end", Weight: 100, ValidFrom: ptr(thingFrom), ValidUntil: ptr(thingUntil)},
			},
		}
	}

	tests := []struct {
		name  string
		graph models.Graph
		code  int
	}{
		{"both unbounded, as before", build("", "", "", ""), http.StatusBadRequest},
		{"the device node ends where the thing node starts, inside the link", build("", "2020-09-01T00:00:00Z", "2020-09-01T00:00:00Z", ""), http.StatusOK},
		{"the device node ends a day after the thing node starts", build("", "2020-09-02T00:00:00Z", "2020-09-01T00:00:00Z", ""), http.StatusBadRequest},
		{"the thing node ends where the device node starts", build("2020-09-01T00:00:00Z", "", "", "2020-09-01T00:00:00Z"), http.StatusOK},
		{"they overlap before the link starts", build("", "2020-05-01T00:00:00Z", "2020-04-01T00:00:00Z", "2020-05-01T00:00:00Z"), http.StatusOK},
		{"they overlap during the link", build("", "2020-07-01T00:00:00Z", "2020-06-15T00:00:00Z", ""), http.StatusBadRequest},
		{"the thing node starts in the far future: unbounded device node meets the open link", build("", "", gFarFuture, ""), http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, test.graph)
			expectCode(t, err, code, test.code)
		})
	}

	t.Run("links outside the common validity of both nodes are accepted", func(t *testing.T) {
		thingTestDevice(t, ctrl, permClient, "dev2", "")
		g := build("", "", "", "")
		g.Nodes[1].ResourceId = "dev2"
		g.Nodes[1].ValidUntil = gTime("2020-09-01T00:00:00Z")
		g.Edges[0].ValidUntil = gTime("2020-09-01T00:00:00Z")
		g.Nodes[2].ValidFrom = gTime("2020-09-01T00:00:00Z")
		g.Edges[1].ValidFrom = gTime("2020-09-01T00:00:00Z")
		_, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, g)
		expectCode(t, err, code, http.StatusOK)

		// the thing node starts 2020-09-01: a link of dev2 that ends before, or starts after the device node, is fine
		_, err, code = ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "dev2", From: gTime("2020-01-01T00:00:00Z"), Until: gTime("2020-09-01T00:00:00Z")})
		expectCode(t, err, code, http.StatusOK)
		thingTestDevice(t, ctrl, permClient, "dev3", "")
		g3 := build("", "", "", "")
		g3.Nodes[1].ResourceId = "dev3"
		g3.Nodes[1].ValidUntil = gTime("2020-09-01T00:00:00Z")
		g3.Edges[0].ValidUntil = gTime("2020-09-01T00:00:00Z")
		g3.Nodes[2].ValidFrom = gTime("2020-09-01T00:00:00Z")
		g3.Edges[1].ValidFrom = gTime("2020-09-01T00:00:00Z")
		g3.Nodes[2].ResourceId = thing.Id
		_, err, code = ctrl.SetGraph(ctx, graphTestUserJwt, g3)
		expectCode(t, err, code, http.StatusOK)
		_, err, code = ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "dev3", From: gTime("2020-08-01T00:00:00Z"), Until: gTime("2020-08-15T00:00:00Z")})
		expectCode(t, err, code, http.StatusOK)
		// both nodes are valid together only before 2020-09-01 for the device and from then on for the thing: never together
		_, err, code = ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "dev3", From: gTime("2020-10-01T00:00:00Z")})
		expectCode(t, err, code, http.StatusOK)
	})

	t.Run("a link is refused where the graph holds both at once", func(t *testing.T) {
		thingTestDevice(t, ctrl, permClient, "dev4", "")
		g := build("", "", "", "")
		g.Nodes[1].ResourceId = "dev4"
		g.Nodes[1].ValidUntil = gTime("2020-09-15T00:00:00Z")
		g.Edges[0].ValidUntil = gTime("2020-09-15T00:00:00Z")
		g.Nodes[2].ValidFrom = gTime("2020-09-01T00:00:00Z")
		g.Edges[1].ValidFrom = gTime("2020-09-01T00:00:00Z")
		_, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, g)
		expectCode(t, err, code, http.StatusOK) // no link yet
		// a link that holds before the overlap of the nodes is fine, one that reaches into it is not
		_, err, code = ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "dev4", From: gTime("2020-01-01T00:00:00Z"), Until: gTime("2020-09-01T00:00:00Z")})
		expectCode(t, err, code, http.StatusOK)
		_, err, code = ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "dev4", From: gTime("2020-09-10T00:00:00Z")})
		expectCode(t, err, code, http.StatusBadRequest)
	})
}

func dev(id string, from *time.Time, until *time.Time) models.Node {
	return models.Node{Id: id, ResourceType: models.GraphResourceTypeDevice, ResourceId: "dev-" + id, ValidFrom: from, ValidUntil: until}
}

func edgeV(id, from, to string, weight int, validFrom *time.Time, validUntil *time.Time) models.Edge {
	return models.Edge{Id: id, FromNodeId: from, ToNodeId: to, Weight: weight, ValidFrom: validFrom, ValidUntil: validUntil}
}

// storeGraph writes a graph with history straight into the database, as a former write left it
func storeGraph(t *testing.T, ctrl *Controller, graph models.Graph) models.Graph {
	t.Helper()
	graph.Id = "urn:infai:ses:graph:history"
	graph.Owner = graphTestUserId
	if err := graph.Valid(); err != nil {
		t.Fatal(err)
	}
	if err := ctrl.db.SetGraph(context.Background(), graph, func(models.Graph) error { return nil }); err != nil {
		t.Fatal(err)
	}
	return graph
}

func TestDeleteResourcesFromGraphWithHistory(t *testing.T) {
	past := gTime(gJan2020)
	T := gTime("2020-06-01T00:00:00Z")
	root := models.Node{Id: "P", ResourceType: "location", ResourceId: "l"}

	t.Run("a device with history in the middle: its edges end now, the rerouted one starts now", func(t *testing.T) {
		ctrl, _, _ := newGraphTestController(t)
		// X -> M -> P directly until T, X -> M -> Q -> P since
		storeGraph(t, ctrl, models.Graph{
			Nodes: []models.Node{root, {Id: "Q", ResourceType: "location", ResourceId: "l2"}, dev("M", nil, nil), dev("X", nil, nil)},
			Edges: []models.Edge{edgeV("e0", "X", "M", 100, nil, nil), edgeV("e1", "M", "P", 100, nil, T), edgeV("e2", "M", "Q", 100, T, nil), edgeV("e3", "Q", "P", 100, nil, nil)},
		})
		before := time.Now()
		err := ctrl.removeResourcesFromGraphs(context.Background(), models.GraphResourceTypeDevice, []string{"dev-M"})
		if err != nil {
			t.Fatal(err)
		}
		now, _, _ := ctrl.ReadGraph(context.Background(), graphTestUserJwt, "urn:infai:ses:graph:history")
		ids := []string{}
		for _, n := range now.Nodes {
			ids = append(ids, n.Id)
		}
		slices.Sort(ids)
		if !slices.Equal(ids, []string{"P", "Q", "X"}) || len(now.Edges) != 2 {
			t.Errorf("%#v", now)
		}
		if w := edgeWeights(now); w["X>Q"] != 100 || w["Q>P"] != 100 {
			t.Errorf("%#v", now.Edges)
		}
		old, err, code := ctrl.ReadGraphAt(context.Background(), graphTestUserJwt, "urn:infai:ses:graph:history", *gTime("2020-03-01T00:00:00Z"))
		expectCode(t, err, code, http.StatusOK)
		if w := edgeWeights(old); w["M>P"] != 100 || w["X>M"] != 100 || len(old.Nodes) != 4 {
			t.Errorf("the history is gone: %#v", old)
		}
		stored, _, _ := ctrl.ReadGraphVersions(ctx0(), graphTestUserJwt, "urn:infai:ses:graph:history")
		for _, e := range stored.Edges {
			if e.FromNodeId == "X" && e.ToNodeId == "Q" && (e.ValidFrom == nil || e.ValidFrom.Before(before.Add(-time.Second))) {
				t.Errorf("the rerouted edge starts now: %#v", e)
			}
		}
	})

	t.Run("an unrelated device is deleted although the graph has weights with history", func(t *testing.T) {
		ctrl, _, _ := newGraphTestController(t)
		storeGraph(t, ctrl, models.Graph{
			Nodes: []models.Node{root, {Id: "P2", ResourceType: "device", ResourceId: "dev-P2"}, dev("M", nil, nil), dev("Z", nil, nil), dev("W", nil, nil)},
			Edges: []models.Edge{
				edgeV("e1", "M", "P", 100, nil, T), edgeV("e2", "M", "P", 50, T, nil), edgeV("e3", "M", "P2", 50, T, nil),
				edgeV("e4", "Z", "P", 100, nil, nil), edgeV("e5", "P2", "P", 100, nil, nil), edgeV("e6", "W", "Z", 100, nil, nil),
			},
		})
		err := ctrl.removeResourcesFromGraphs(context.Background(), models.GraphResourceTypeDevice, []string{"dev-Z"})
		if err != nil {
			t.Fatal(err)
		}
		now, _, _ := ctrl.ReadGraph(context.Background(), graphTestUserJwt, "urn:infai:ses:graph:history")
		if w := edgeWeights(now); w["W>P"] != 100 || w["M>P"] != 50 || w["M>P2"] != 50 || len(now.Nodes) != 4 {
			t.Errorf("%#v", now)
		}
		old, _, _ := ctrl.ReadGraphAt(context.Background(), graphTestUserJwt, "urn:infai:ses:graph:history", *past)
		if w := edgeWeights(old); w["M>P"] != 100 || w["Z>P"] != 100 || w["W>Z"] != 100 {
			t.Errorf("%#v", old)
		}
	})

	t.Run("a graph without validities loses the node as before", func(t *testing.T) {
		ctrl, _, _ := newGraphTestController(t)
		g := storeGraph(t, ctrl, models.Graph{
			Nodes: []models.Node{root, dev("M", nil, nil)},
			Edges: []models.Edge{edgeV("e1", "M", "P", 100, nil, nil)},
		})
		g.Nodes = []models.Node{root, dev("N", nil, nil)}
		g.Edges = []models.Edge{edgeV("e1", "N", "P", 100, nil, nil)}
		storeGraph(t, ctrl, g)
		err := ctrl.removeResourcesFromGraphs(context.Background(), models.GraphResourceTypeDevice, []string{"dev-N"})
		if err != nil {
			t.Fatal(err)
		}
		stored, _, _ := ctrl.ReadGraphVersions(ctx0(), graphTestUserJwt, g.Id)
		if len(stored.Nodes) != 1 || len(stored.Edges) != 0 {
			t.Errorf("%#v", stored)
		}
	})
}

func ctx0() context.Context { return context.Background() }

// racingDB lets a second writer win once between the read and the write of the first
type racingDB struct {
	database.Database
	beforeWrite func()
	conflicts   int
}

func (this *racingDB) SetGraphIfRevision(ctx context.Context, graph models.Graph, revision int64, syncHandler func(models.Graph) error) (bool, error) {
	if this.beforeWrite != nil {
		race := this.beforeWrite
		this.beforeWrite = nil
		race()
	}
	swapped, err := this.Database.SetGraphIfRevision(ctx, graph, revision, syncHandler)
	if !swapped {
		this.conflicts++
	}
	return swapped, err
}

func TestGraphWritesAreCompareAndSet(t *testing.T) {
	ctrl, _, _ := newGraphTestController(t)
	ctx := context.Background()
	created, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, meterTestGraph())
	expectCode(t, err, code, http.StatusOK)
	db := &racingDB{Database: ctrl.db}
	ctrl.db = db

	t.Run("two edits interleaved: the loser is applied again to what the winner wrote, and the winner is not lost", func(t *testing.T) {
		//writer 1 changes the weights from 06-30 on, writer 2 renames the graph in between
		db.conflicts = 0
		db.beforeWrite = func() {
			current, _, _ := ctrl.ReadGraph(ctx, graphTestUserJwt, created.Id)
			current.Attributes = []models.Attribute{{Key: "name", Value: "renamed"}}
			_, err, code := ctrl.EditGraph(ctx, graphTestUserJwt, current, model.GraphEditOptions{})
			expectCode(t, err, code, http.StatusOK)
		}
		current := setMeterWeights(created, 70, 30)
		_, err, code := ctrl.EditGraph(ctx, graphTestUserJwt, current, model.GraphEditOptions{At: gTime(gJun2020), Mode: models.GraphEditModeChange})
		expectCode(t, err, code, http.StatusOK)
		if db.conflicts != 1 {
			t.Errorf("expected one lost race, got %v", db.conflicts)
		}
		stored, _, _ := ctrl.ReadGraphVersions(ctx, graphTestUserJwt, created.Id)
		//writer 1's body still carries the old name, so from its instant on it wins; the rename of writer 2 holds before
		if len(stored.Edges) != 6 || len(stored.Attributes) != 2 || stored.Attributes[0].Value != "renamed" || stored.Attributes[1].Value != "g" {
			t.Errorf("%#v", stored)
		}
		now, _, _ := ctrl.ReadGraph(ctx, graphTestUserJwt, created.Id)
		if w := edgeWeights(now); w["m>a"] != 70 {
			t.Errorf("%#v", now.Edges)
		}
	})

	t.Run("a writer that always loses gets a 409", func(t *testing.T) {
		always := &alwaysLosing{Database: ctrl.db}
		ctrl.db = always
		defer func() { ctrl.db = db }()
		current, _, _ := ctrl.ReadGraph(ctx, graphTestUserJwt, created.Id)
		_, err, code := ctrl.EditGraph(ctx, graphTestUserJwt, current, model.GraphEditOptions{})
		if code != http.StatusConflict || err == nil {
			t.Errorf("%v %v", code, err)
		}
		if always.calls != graphWriteAttempts {
			t.Errorf("%v", always.calls)
		}
	})

	t.Run("two creations with the same id: the second one is an edit of the first", func(t *testing.T) {
		g := meterTestGraph()
		g.Id = "urn:infai:ses:graph:new"
		db.beforeWrite = func() {
			_, err, code := ctrl.SetGraph(ctx, permclient.InternalAdminToken, g)
			expectCode(t, err, code, http.StatusOK)
		}
		_, err, code := ctrl.SetGraph(ctx, permclient.InternalAdminToken, g)
		expectCode(t, err, code, http.StatusOK)
	})
}

type alwaysLosing struct {
	database.Database
	calls int
}

func (this *alwaysLosing) SetGraphIfRevision(ctx context.Context, graph models.Graph, revision int64, syncHandler func(models.Graph) error) (bool, error) {
	this.calls++
	return false, nil
}

func TestSetThingKeepsKindWithoutKind(t *testing.T) {
	ctrl, _, _ := newThingTestController(t)
	ctx := context.Background()
	meter, err, code := ctrl.SetThing(ctx, graphTestUserJwt, models.Thing{Name: "m", Kind: models.ThingKindMeter})
	expectCode(t, err, code, http.StatusOK)
	renamed, err, code := ctrl.SetThing(ctx, graphTestUserJwt, models.Thing{Id: meter.Id, Name: "m2"})
	expectCode(t, err, code, http.StatusOK)
	if renamed.Kind != models.ThingKindMeter {
		t.Errorf("%#v", renamed)
	}
	consumer, err, code := ctrl.SetThing(ctx, graphTestUserJwt, models.Thing{Id: meter.Id, Name: "m3", Kind: models.ThingKindConsumer})
	expectCode(t, err, code, http.StatusOK)
	if consumer.Kind != models.ThingKindConsumer {
		t.Errorf("%#v", consumer)
	}
	plain, err, code := ctrl.SetThing(ctx, graphTestUserJwt, models.Thing{Name: "plain"})
	expectCode(t, err, code, http.StatusOK)
	again, err, code := ctrl.SetThing(ctx, graphTestUserJwt, models.Thing{Id: plain.Id, Name: "plain2"})
	expectCode(t, err, code, http.StatusOK)
	if again.Kind != "" {
		t.Errorf("a thing without kind stays a consumer without one: %#v", again)
	}
}

func TestGraphProviderAddsADeviceToAGraphWithHistory(t *testing.T) {
	T := gTime("2020-06-01T00:00:00Z")
	for name, until := range map[string]*time.Time{"B is open": nil, "B ends in the far future": gTime(gFarFuture)} {
		t.Run(name, func(t *testing.T) {
			ctrl, _, _ := newGraphTestController(t)
			ctx := context.Background()
			stored := storeGraph(t, ctrl, models.Graph{
				Nodes: []models.Node{{Id: "P", ResourceType: "location", ResourceId: "l"}, dev("A", nil, T), dev("B", T, until)},
				Edges: []models.Edge{edgeV("e1", "A", "P", 100, nil, T), edgeV("e2", "B", "P", 100, T, until)},
			})
			body, err, code := ctrl.ReadGraph(ctx, permclient.InternalAdminToken, stored.Id)
			expectCode(t, err, code, http.StatusOK)
			body.Nodes = append(body.Nodes, dev("D", nil, nil))
			body.Edges = append(body.Edges, edgeV("e3", "D", "B", 100, nil, nil))
			_, err, code = ctrl.SetGraph(ctx, permclient.InternalAdminToken, body)
			expectCode(t, err, code, http.StatusOK)
			versions, _, _ := ctrl.ReadGraphVersions(ctx, permclient.InternalAdminToken, stored.Id)
			found := false
			for _, e := range versions.Edges {
				if e.FromNodeId == "D" {
					found = true
					if e.ValidFrom == nil || !e.ValidFrom.Equal(*T) {
						t.Errorf("the edge starts where B starts: %#v", e)
					}
				}
			}
			if !found {
				t.Error("edge not stored")
			}
		})
	}
}
