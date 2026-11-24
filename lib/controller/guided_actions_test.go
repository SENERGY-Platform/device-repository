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
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/SENERGY-Platform/device-repository/v3/lib/database"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	permclient "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

func TestSwapThingDeviceLink(t *testing.T) {
	var ctrl *Controller
	var permClient permclient.Client
	ctx := context.Background()
	//every case starts with a fresh database: a device is linked to one thing at a time
	setup := func(t *testing.T) {
		t.Helper()
		ctrl, permClient, _ = newThingTestController(t)
		for _, id := range []string{"d1", "d2", "d3"} {
			thingTestDevice(t, ctrl, permClient, id, "")
		}
	}
	energy := func(from float64) []models.ThingDeviceLinkReading {
		return []models.ThingDeviceLinkReading{{ServiceId: "s1", Column: "metrics.energy", FromValue: from}}
	}
	removal := func(value float64) []model.ThingDeviceLinkEndReading {
		return []model.ThingDeviceLinkEndReading{{ServiceId: "s1", Column: "metrics.energy", UntilValue: value}}
	}
	newMeter := func(t *testing.T, name string, device string) (models.Thing, models.ThingDeviceLink) {
		t.Helper()
		meter := thingTestMeter(t, ctrl, name)
		meter, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, meter.Id, model.ThingDeviceLinkCreate{DeviceId: device, From: thingTestTime(1), Readings: energy(100)})
		expectCode(t, err, code, http.StatusOK)
		return meter, meter.DeviceLinks[0]
	}
	stored := func(t *testing.T, id string) models.Thing {
		t.Helper()
		thing, err, code := ctrl.ReadThing(ctx, graphTestUserJwt, id)
		expectCode(t, err, code, http.StatusOK)
		return thing
	}

	t.Run("the same device with a counter reset: the old link ends with its reading, the new one starts with the new", func(t *testing.T) {
		setup(t)
		meter, link := newMeter(t, "reset", "d1")
		result, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, link.Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), UntilValues: removal(900), DeviceId: "d1", Readings: energy(0)})
		expectCode(t, err, code, http.StatusOK)
		if !reflect.DeepEqual(result, stored(t, meter.Id)) {
			t.Errorf("the answer is not what is stored")
		}
		if len(result.DeviceLinks) != 2 {
			t.Fatalf("%#v", result.DeviceLinks)
		}
		old, next := result.DeviceLinks[0], result.DeviceLinks[1]
		if old.Id != link.Id || old.Until == nil || !old.Until.Equal(*thingTestTime(5)) || old.Readings[0].UntilValue == nil || *old.Readings[0].UntilValue != 900 || old.Readings[0].FromValue != 100 {
			t.Errorf("%#v", old)
		}
		if next.DeviceId != "d1" || !next.From.Equal(*thingTestTime(5)) || next.Until != nil || next.Readings[0].FromValue != 0 || next.Readings[0].UntilValue != nil {
			t.Errorf("%#v", next)
		}
	})

	t.Run("a new device and a new factor", func(t *testing.T) {
		setup(t)
		meter, link := newMeter(t, "new device", "d1")
		factor := 40.0
		result, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, link.Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d2", Readings: energy(7), Factor: &factor})
		expectCode(t, err, code, http.StatusOK)
		next := result.DeviceLinks[1]
		if next.DeviceId != "d2" || next.Factor == nil || *next.Factor != 40 || result.DeviceLinks[0].Until == nil {
			t.Errorf("%#v", result.DeviceLinks)
		}
		//the old device is free from the instant on
		other := thingTestMeter(t, ctrl, "takes over d1")
		_, err, code = ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, other.Id, model.ThingDeviceLinkCreate{DeviceId: "d1", From: thingTestTime(5)})
		expectCode(t, err, code, http.StatusOK)
	})

	t.Run("a swap in the past is allowed, at now is the default", func(t *testing.T) {
		setup(t)
		meter, link := newMeter(t, "backdated", "d1")
		result, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, link.Id, model.ThingDeviceLinkSwap{At: thingTestTime(2), DeviceId: "d2"})
		expectCode(t, err, code, http.StatusOK)
		if !result.DeviceLinks[0].Until.Equal(*thingTestTime(2)) || !result.DeviceLinks[1].From.Equal(*thingTestTime(2)) {
			t.Errorf("%#v", result.DeviceLinks)
		}
		next := result.DeviceLinks[1]
		result, err, code = ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, next.Id, model.ThingDeviceLinkSwap{DeviceId: "d3"})
		expectCode(t, err, code, http.StatusOK)
		if result.DeviceLinks[1].Until == nil || !result.DeviceLinks[2].From.Equal(*result.DeviceLinks[1].Until) {
			t.Errorf("%#v", result.DeviceLinks)
		}
	})

	t.Run("nothing is written when one half is refused", func(t *testing.T) {
		setup(t)
		meter, link := newMeter(t, "atomic", "d1")
		//d2 is linked to another thing from hour 8: a swap at 5 would overlap it
		blocker := thingTestMeter(t, ctrl, "blocker")
		_, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, blocker.Id, model.ThingDeviceLinkCreate{DeviceId: "d2", From: thingTestTime(8)})
		expectCode(t, err, code, http.StatusOK)

		refused := map[string]struct {
			swap model.ThingDeviceLinkSwap
			code int
		}{
			"overlaps a later link of the device": {model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d2"}, http.StatusBadRequest},
			"unknown reading at removal":          {model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3", UntilValues: []model.ThingDeviceLinkEndReading{{ServiceId: "s1", Column: "other", UntilValue: 1}}}, http.StatusBadRequest},
			"duplicate reading at removal":        {model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3", UntilValues: append(removal(1), removal(2)...)}, http.StatusBadRequest},
			"invalid new readings":                {model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3", Readings: []models.ThingDeviceLinkReading{{ServiceId: "s1", Column: "c"}, {ServiceId: "s1", Column: "c"}}}, http.StatusBadRequest},
			"an until_value in the new readings":  {model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3", Readings: []models.ThingDeviceLinkReading{{ServiceId: "s1", Column: "c", UntilValue: new(float64)}}}, http.StatusBadRequest},
			"factor zero":                         {model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3", Factor: new(float64)}, http.StatusBadRequest},
			"unknown device":                      {model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "nope"}, http.StatusBadRequest},
			"no device":                           {model.ThingDeviceLinkSwap{At: thingTestTime(5)}, http.StatusBadRequest},
			"at is the start of the ended link":   {model.ThingDeviceLinkSwap{At: thingTestTime(1), DeviceId: "d3"}, http.StatusBadRequest},
			"at is before the start of the link":  {model.ThingDeviceLinkSwap{At: thingTestTime(0), DeviceId: "d3"}, http.StatusBadRequest},
		}
		for name, test := range refused {
			t.Run(name, func(t *testing.T) {
				_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, link.Id, test.swap)
				expectCode(t, err, code, test.code)
				after := stored(t, meter.Id)
				if len(after.DeviceLinks) != 1 || after.DeviceLinks[0].Until != nil || after.DeviceLinks[0].Readings[0].UntilValue != nil {
					t.Errorf("the refused swap changed the thing: %#v", after.DeviceLinks)
				}
			})
		}
		//the message names the instant
		_, err, _ = ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, link.Id, model.ThingDeviceLinkSwap{At: thingTestTime(1), DeviceId: "d3"})
		if !strings.Contains(err.Error(), "2026-11-01T01:00:00Z") {
			t.Errorf("%v", err)
		}
	})

	t.Run("an ended or unknown link can not be swapped", func(t *testing.T) {
		setup(t)
		meter, link := newMeter(t, "ended", "d1")
		_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, link.Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d1"})
		expectCode(t, err, code, http.StatusOK)
		_, err, code = ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, link.Id, model.ThingDeviceLinkSwap{At: thingTestTime(6), DeviceId: "d1"})
		expectCode(t, err, code, http.StatusBadRequest)
		_, err, code = ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, "unknown", model.ThingDeviceLinkSwap{At: thingTestTime(6), DeviceId: "d1"})
		expectCode(t, err, code, http.StatusNotFound)
	})

	t.Run("a graph with the thing and the new device is refused, and nothing is written", func(t *testing.T) {
		setup(t)
		meter, link := newMeter(t, "graph rule", "d1")
		_, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, thingTestGraph(meter.Id, "d3"))
		expectCode(t, err, code, http.StatusOK)
		_, err, code = ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, link.Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3"})
		expectCode(t, err, code, http.StatusBadRequest)
		if after := stored(t, meter.Id); len(after.DeviceLinks) != 1 || after.DeviceLinks[0].Until != nil {
			t.Errorf("%#v", after.DeviceLinks)
		}
	})

	t.Run("permissions: write on the thing and read on the device", func(t *testing.T) {
		setup(t)
		meter, link := newMeter(t, "permissions", "d1")
		swap := model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d2"}
		_, err, code := ctrl.SwapThingDeviceLink(ctx, thingTestOtherUserJwt, meter.Id, link.Id, swap)
		expectCode(t, err, code, http.StatusForbidden)

		//write on the thing, but the device d2 is not readable for the other user
		_, err, _ = permClient.SetPermission(permclient.InternalAdminToken, "things", meter.Id, permclient.ResourcePermissions{UserPermissions: map[string]permclient.PermissionsMap{
			graphTestUserId:      {Read: true, Write: true, Execute: true, Administrate: true},
			thingTestOtherUserId: {Read: true, Write: true},
		}})
		if err != nil {
			t.Fatal(err)
		}
		_, err, code = ctrl.SwapThingDeviceLink(ctx, thingTestOtherUserJwt, meter.Id, link.Id, swap)
		expectCode(t, err, code, http.StatusForbidden)
		if after := stored(t, meter.Id); len(after.DeviceLinks) != 1 || after.DeviceLinks[0].Until != nil {
			t.Errorf("%#v", after.DeviceLinks)
		}

		thingTestDevice(t, ctrl, permClient, "d4", thingTestOtherUserId)
		_, err, code = ctrl.SwapThingDeviceLink(ctx, thingTestOtherUserJwt, meter.Id, link.Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d4"})
		expectCode(t, err, code, http.StatusOK)
	})

	t.Run("the open link of a device is found by device_ids, ended ones too", func(t *testing.T) {
		setup(t)
		meter, link := newMeter(t, "lookup", "d1")
		_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, link.Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d2"})
		expectCode(t, err, code, http.StatusOK)
		things, _, err, code := ctrl.ListThings(ctx, graphTestUserJwt, model.ThingListOptions{DeviceIds: []string{"d2"}})
		expectCode(t, err, code, http.StatusOK)
		found := false
		for _, thing := range things {
			if thing.Id == meter.Id {
				found = true
			}
		}
		if !found {
			t.Errorf("%#v", things)
		}
	})
}

// moveTestGraph: M is a unit below A, X and Y are its children; A and B are below root.
func moveTestGraph() models.Graph {
	return models.Graph{
		Attributes: []models.Attribute{{Key: "name", Value: "g"}},
		Nodes:      []models.Node{{Id: "root"}, {Id: "A"}, {Id: "B"}, {Id: "M"}, {Id: "X"}, {Id: "Y"}},
		Edges: []models.Edge{
			{Id: "A>root", FromNodeId: "A", ToNodeId: "root", Weight: 100},
			{Id: "B>root", FromNodeId: "B", ToNodeId: "root", Weight: 100},
			{Id: "M>A", FromNodeId: "M", ToNodeId: "A", Weight: 100},
			{Id: "X>M", FromNodeId: "X", ToNodeId: "M", Weight: 100},
			{Id: "Y>M", FromNodeId: "Y", ToNodeId: "M", Weight: 100},
		},
	}
}

func parentOf(graph models.Graph, node string) []string {
	result := []string{}
	for _, edge := range graph.Edges {
		if edge.FromNodeId == node {
			result = append(result, edge.ToNodeId)
		}
	}
	return result
}

func TestMoveGraphNode(t *testing.T) {
	ctx := context.Background()
	setup := func(t *testing.T) (*Controller, models.Graph) {
		ctrl, _, _ := newGraphTestController(t)
		created, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, moveTestGraph())
		expectCode(t, err, code, http.StatusOK)
		return ctrl, created
	}
	readAt := func(t *testing.T, ctrl *Controller, id string, at string) models.Graph {
		t.Helper()
		g, err, code := ctrl.ReadGraphAt(ctx, graphTestUserJwt, id, *gTime(at))
		expectCode(t, err, code, http.StatusOK)
		return g
	}

	t.Run("open move: the node is below B from the instant on, the children stay below it, the history stays", func(t *testing.T) {
		ctrl, g := setup(t)
		result, err, code := ctrl.MoveGraphNode(ctx, graphTestUserJwt, g.Id, "M", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "B"})
		expectCode(t, err, code, http.StatusOK)
		if !reflect.DeepEqual(parentOf(result, "M"), []string{"B"}) {
			t.Errorf("the answer is the graph now: %#v", result.Edges)
		}
		before := readAt(t, ctrl, g.Id, "2020-03-01T00:00:00Z")
		if !reflect.DeepEqual(parentOf(before, "M"), []string{"A"}) {
			t.Errorf("history: %#v", before.Edges)
		}
		onTheDay := readAt(t, ctrl, g.Id, gJun2020)
		if !reflect.DeepEqual(parentOf(onTheDay, "M"), []string{"B"}) || parentOf(onTheDay, "X")[0] != "M" || parentOf(onTheDay, "Y")[0] != "M" || len(onTheDay.Nodes) != 6 {
			t.Errorf("%#v", onTheDay.Edges)
		}
		w := edgeWeights(onTheDay)
		if w["M>B"] != 100 {
			t.Errorf("%#v", w)
		}
	})

	t.Run("bounded move: A, B, A", func(t *testing.T) {
		ctrl, g := setup(t)
		_, err, code := ctrl.MoveGraphNode(ctx, graphTestUserJwt, g.Id, "M", model.GraphNodeMove{At: gTime(gJun2020), Until: gTime(gDec2020), ToNodeId: "B"})
		expectCode(t, err, code, http.StatusOK)
		for at, want := range map[string]string{"2020-03-01T00:00:00Z": "A", gJun2020: "B", "2020-10-01T00:00:00Z": "B", gDec2020: "A", "2021-03-01T00:00:00Z": "A"} {
			if got := parentOf(readAt(t, ctrl, g.Id, at), "M"); !reflect.DeepEqual(got, []string{want}) {
				t.Errorf("at %v: %v, want %v", at, got, want)
			}
		}
		now, _, _ := ctrl.ReadGraph(ctx, graphTestUserJwt, g.Id)
		if !reflect.DeepEqual(parentOf(now, "M"), []string{"A"}) {
			t.Errorf("%#v", now.Edges)
		}
	})

	t.Run("refused moves change nothing", func(t *testing.T) {
		ctrl, g := setup(t)
		for name, test := range map[string]struct {
			node string
			move model.GraphNodeMove
		}{
			"unknown target":      {"M", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "nope"}},
			"unknown node":        {"nope", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "B"}},
			"no target":           {"M", model.GraphNodeMove{At: gTime(gJun2020)}},
			"below itself":        {"M", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "M"}},
			"below its own child": {"M", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "X"}},
			"already below":       {"M", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "A"}},
			"the end node":        {"root", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "B"}},
			"until not after at":  {"M", model.GraphNodeMove{At: gTime(gJun2020), Until: gTime(gJun2020), ToNodeId: "B"}},
		} {
			t.Run(name, func(t *testing.T) {
				_, err, code := ctrl.MoveGraphNode(ctx, graphTestUserJwt, g.Id, test.node, test.move)
				expectCode(t, err, code, http.StatusBadRequest)
			})
		}
		stored, _, _ := ctrl.ReadGraphVersions(ctx, graphTestUserJwt, g.Id)
		if !reflect.DeepEqual(stored.Edges, g.Edges) {
			t.Errorf("%#v", stored.Edges)
		}
	})

	t.Run("a node that is not valid at the instant can not be moved, the message names it", func(t *testing.T) {
		ctrl, g := setup(t)
		_, err, code := ctrl.EndGraphNode(ctx, graphTestUserJwt, g.Id, "B", model.GraphNodeEnd{At: gTime(gJun2020)})
		expectCode(t, err, code, http.StatusOK)
		_, err, code = ctrl.MoveGraphNode(ctx, graphTestUserJwt, g.Id, "M", model.GraphNodeMove{At: gTime(gDec2020), ToNodeId: "B"})
		expectCode(t, err, code, http.StatusBadRequest)
		if !strings.Contains(err.Error(), "2020-12-01T00:00:00Z") {
			t.Errorf("%v", err)
		}
	})

	t.Run("permissions: write on the graph", func(t *testing.T) {
		ctrl, g := setup(t)
		_, err, code := ctrl.MoveGraphNode(ctx, otherUserJwt(), g.Id, "M", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "B"})
		expectCode(t, err, code, http.StatusForbidden)
		_, err, code = ctrl.EndGraphNode(ctx, otherUserJwt(), g.Id, "M", model.GraphNodeEnd{At: gTime(gJun2020)})
		expectCode(t, err, code, http.StatusForbidden)
		_, err, code = ctrl.MoveGraphNode(ctx, permclient.InternalAdminToken, g.Id, "M", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "B"})
		expectCode(t, err, code, http.StatusOK)
	})

	t.Run("a lost race is applied again to what the winner wrote, a writer that always loses gets 409", func(t *testing.T) {
		ctrl, g := setup(t)
		db := &racingDB{Database: ctrl.db}
		ctrl.db = db
		db.beforeWrite = func() {
			current, _, _ := ctrl.ReadGraph(ctx, graphTestUserJwt, g.Id)
			current.Attributes = []models.Attribute{{Key: "name", Value: "renamed"}}
			_, err, code := ctrl.EditGraph(ctx, graphTestUserJwt, current, model.GraphEditOptions{})
			expectCode(t, err, code, http.StatusOK)
		}
		_, err, code := ctrl.MoveGraphNode(ctx, graphTestUserJwt, g.Id, "M", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "B"})
		expectCode(t, err, code, http.StatusOK)
		if db.conflicts != 1 {
			t.Errorf("%v", db.conflicts)
		}
		now, _, _ := ctrl.ReadGraph(ctx, graphTestUserJwt, g.Id)
		if !reflect.DeepEqual(parentOf(now, "M"), []string{"B"}) || now.Attributes[0].Value != "renamed" {
			t.Errorf("%#v", now)
		}
		always := &alwaysLosing{Database: ctrl.db}
		ctrl.db = always
		_, err, code = ctrl.EndGraphNode(ctx, graphTestUserJwt, g.Id, "X", model.GraphNodeEnd{At: gTime(gDec2020)})
		if code != http.StatusConflict || err == nil || always.calls != graphWriteAttempts {
			t.Errorf("%v %v %v", code, err, always.calls)
		}
	})
}

func TestEndGraphNode(t *testing.T) {
	ctx := context.Background()
	ctrl, permClient, _ := newThingTestController(t)
	thingTestDevice(t, ctrl, permClient, "d1", "")
	meter := thingTestMeter(t, ctrl, "meter")
	meter, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, meter.Id, model.ThingDeviceLinkCreate{DeviceId: "d1", From: gTime(gJan2020)})
	expectCode(t, err, code, http.StatusOK)

	graph := moveTestGraph()
	graph.Nodes = append(graph.Nodes, models.Node{Id: "T", ResourceType: models.GraphResourceTypeThing, ResourceId: meter.Id})
	graph.Edges = append(graph.Edges, models.Edge{Id: "T>M", FromNodeId: "T", ToNodeId: "M", Weight: 100})
	g, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, graph)
	expectCode(t, err, code, http.StatusOK)
	readAt := func(at string) models.Graph {
		r, err, code := ctrl.ReadGraphAt(ctx, graphTestUserJwt, g.Id, *gTime(at))
		expectCode(t, err, code, http.StatusOK)
		return r
	}

	t.Run("the unit ends at the instant, its children are rerouted to its parent from there, the history is intact", func(t *testing.T) {
		result, err, code := ctrl.EndGraphNode(ctx, graphTestUserJwt, g.Id, "M", model.GraphNodeEnd{At: gTime(gJun2020)})
		expectCode(t, err, code, http.StatusOK)
		for _, child := range []string{"X", "Y", "T"} {
			if !reflect.DeepEqual(parentOf(result, child), []string{"A"}) {
				t.Errorf("%v: %#v", child, result.Edges)
			}
		}
		if slicesContainsNode(result, "M") {
			t.Errorf("M is still valid now")
		}
		before := readAt("2020-03-01T00:00:00Z")
		if !slicesContainsNode(before, "M") || !reflect.DeepEqual(parentOf(before, "X"), []string{"M"}) || !reflect.DeepEqual(parentOf(before, "M"), []string{"A"}) || len(before.Nodes) != 7 {
			t.Errorf("history: %#v", before)
		}
		onTheDay := readAt(gJun2020)
		if slicesContainsNode(onTheDay, "M") || !reflect.DeepEqual(parentOf(onTheDay, "Y"), []string{"A"}) {
			t.Errorf("%#v", onTheDay)
		}
	})

	t.Run("the open link of the meter thing is not ended", func(t *testing.T) {
		stored, err, code := ctrl.ReadThing(ctx, graphTestUserJwt, meter.Id)
		expectCode(t, err, code, http.StatusOK)
		if len(stored.DeviceLinks) != 1 || stored.DeviceLinks[0].Until != nil {
			t.Errorf("%#v", stored.DeviceLinks)
		}
		//ending the meter's own node leaves the link as it is, too
		_, err, code = ctrl.EndGraphNode(ctx, graphTestUserJwt, g.Id, "T", model.GraphNodeEnd{At: gTime(gDec2020)})
		expectCode(t, err, code, http.StatusOK)
		stored, _, _ = ctrl.ReadThing(ctx, graphTestUserJwt, meter.Id)
		if stored.DeviceLinks[0].Until != nil {
			t.Errorf("%#v", stored.DeviceLinks)
		}
	})

	t.Run("a node that is gone at the instant, or the end node with children, is refused", func(t *testing.T) {
		_, err, code := ctrl.EndGraphNode(ctx, graphTestUserJwt, g.Id, "M", model.GraphNodeEnd{At: gTime(gDec2020)})
		expectCode(t, err, code, http.StatusBadRequest)
		_, err, code = ctrl.EndGraphNode(ctx, graphTestUserJwt, g.Id, "root", model.GraphNodeEnd{At: gTime(gDec2020)})
		expectCode(t, err, code, http.StatusBadRequest)
	})
}

func slicesContainsNode(graph models.Graph, id string) bool {
	for _, node := range graph.Nodes {
		if node.Id == id {
			return true
		}
	}
	return false
}

// thingRacingDB lets a second writer win once between the read and the write of the first
type thingRacingDB struct {
	database.Database
	beforeWrite func()
	conflicts   int
	calls       int
}

func (this *thingRacingDB) SetThingIfRevision(ctx context.Context, thing models.Thing, revision int64, syncHandler func(models.Thing) error) (bool, error) {
	this.calls++
	if this.beforeWrite != nil {
		race := this.beforeWrite
		this.beforeWrite = nil
		race()
	}
	swapped, err := this.Database.SetThingIfRevision(ctx, thing, revision, syncHandler)
	if !swapped {
		this.conflicts++
	}
	return swapped, err
}

func TestThingWritesAreCompareAndSet(t *testing.T) {
	ctx := context.Background()
	setup := func(t *testing.T) (*Controller, *thingRacingDB, models.Thing) {
		ctrl, permClient, _ := newThingTestController(t)
		for _, id := range []string{"d1", "d2", "d3", "d4"} {
			thingTestDevice(t, ctrl, permClient, id, "")
		}
		meter := thingTestMeter(t, ctrl, "meter")
		for _, id := range []string{"d1", "d2"} {
			var err error
			var code int
			meter, err, code = ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, meter.Id, model.ThingDeviceLinkCreate{DeviceId: id, From: thingTestTime(1)})
			expectCode(t, err, code, http.StatusOK)
		}
		db := &thingRacingDB{Database: ctrl.db}
		ctrl.db = db
		return ctrl, db, meter
	}
	devices := func(thing models.Thing) (result []string) {
		for _, link := range thing.DeviceLinks {
			if link.Until == nil {
				result = append(result, link.DeviceId)
			}
		}
		slices.Sort(result)
		return result
	}

	t.Run("two interleaved swaps of different links: neither loses the other's link", func(t *testing.T) {
		ctrl, db, meter := setup(t)
		db.beforeWrite = func() {
			_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, meter.DeviceLinks[1].Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d4"})
			expectCode(t, err, code, http.StatusOK)
		}
		_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, meter.DeviceLinks[0].Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3"})
		expectCode(t, err, code, http.StatusOK)
		if db.conflicts != 1 {
			t.Errorf("expected one lost race, got %v", db.conflicts)
		}
		stored, _, _ := ctrl.ReadThing(ctx, graphTestUserJwt, meter.Id)
		if len(stored.DeviceLinks) != 4 || !slices.Equal(devices(stored), []string{"d3", "d4"}) {
			t.Errorf("%#v", stored.DeviceLinks)
		}
	})

	t.Run("two swaps of the same link: the loser is refused, the winner's link stays", func(t *testing.T) {
		ctrl, db, meter := setup(t)
		db.beforeWrite = func() {
			_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, meter.DeviceLinks[0].Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3"})
			expectCode(t, err, code, http.StatusOK)
		}
		_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, meter.DeviceLinks[0].Id, model.ThingDeviceLinkSwap{At: thingTestTime(6), DeviceId: "d4"})
		expectCode(t, err, code, http.StatusBadRequest)
		stored, _, _ := ctrl.ReadThing(ctx, graphTestUserJwt, meter.Id)
		if !slices.Equal(devices(stored), []string{"d2", "d3"}) {
			t.Errorf("%#v", stored.DeviceLinks)
		}
	})

	t.Run("a swap racing a PUT /things keeps both", func(t *testing.T) {
		ctrl, db, meter := setup(t)
		db.beforeWrite = func() {
			renamed := meter
			renamed.Name = "renamed"
			_, err, code := ctrl.SetThing(ctx, graphTestUserJwt, renamed)
			expectCode(t, err, code, http.StatusOK)
		}
		_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, meter.DeviceLinks[0].Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3"})
		expectCode(t, err, code, http.StatusOK)
		stored, _, _ := ctrl.ReadThing(ctx, graphTestUserJwt, meter.Id)
		if stored.Name != "renamed" || !slices.Equal(devices(stored), []string{"d2", "d3"}) {
			t.Errorf("%v %#v", stored.Name, stored.DeviceLinks)
		}
	})

	t.Run("a PUT /things racing a swap keeps both, and the links of the body are still ignored", func(t *testing.T) {
		ctrl, db, meter := setup(t)
		db.beforeWrite = func() {
			_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, meter.DeviceLinks[0].Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3"})
			expectCode(t, err, code, http.StatusOK)
		}
		renamed := meter //carries the two links as they were read before the swap
		renamed.Name = "renamed"
		_, err, code := ctrl.SetThing(ctx, graphTestUserJwt, renamed)
		expectCode(t, err, code, http.StatusOK)
		stored, _, _ := ctrl.ReadThing(ctx, graphTestUserJwt, meter.Id)
		if stored.Name != "renamed" || len(stored.DeviceLinks) != 3 || !slices.Equal(devices(stored), []string{"d2", "d3"}) {
			t.Errorf("%v %#v", stored.Name, stored.DeviceLinks)
		}
	})

	t.Run("link add and link end race too", func(t *testing.T) {
		ctrl, db, meter := setup(t)
		db.beforeWrite = func() {
			_, err, code := ctrl.EndThingDeviceLink(ctx, graphTestUserJwt, meter.Id, meter.DeviceLinks[1].Id, model.ThingDeviceLinkEnd{Until: thingTestTime(4)})
			expectCode(t, err, code, http.StatusOK)
		}
		_, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, meter.Id, model.ThingDeviceLinkCreate{DeviceId: "d3", From: thingTestTime(2)})
		expectCode(t, err, code, http.StatusOK)
		stored, _, _ := ctrl.ReadThing(ctx, graphTestUserJwt, meter.Id)
		if len(stored.DeviceLinks) != 3 || !slices.Equal(devices(stored), []string{"d1", "d3"}) {
			t.Errorf("%#v", stored.DeviceLinks)
		}
	})

	t.Run("a writer that always loses gets a 409", func(t *testing.T) {
		ctrl, _, meter := setup(t)
		always := &alwaysLosingThings{Database: ctrl.db}
		ctrl.db = always
		_, err, code := ctrl.SwapThingDeviceLink(ctx, graphTestUserJwt, meter.Id, meter.DeviceLinks[0].Id, model.ThingDeviceLinkSwap{At: thingTestTime(5), DeviceId: "d3"})
		if code != http.StatusConflict || err == nil || always.calls != thingWriteAttempts {
			t.Errorf("%v %v %v", code, err, always.calls)
		}
		renamed := meter
		renamed.Name = "x"
		always.calls = 0
		_, err, code = ctrl.SetThing(ctx, graphTestUserJwt, renamed)
		if code != http.StatusConflict || err == nil || always.calls != thingWriteAttempts {
			t.Errorf("%v %v %v", code, err, always.calls)
		}
	})
}

type alwaysLosingThings struct {
	database.Database
	calls int
}

func (this *alwaysLosingThings) SetThingIfRevision(ctx context.Context, thing models.Thing, revision int64, syncHandler func(models.Thing) error) (bool, error) {
	this.calls++
	return false, nil
}

func TestGuidedActionsBeforePlannedChanges(t *testing.T) {
	ctx := context.Background()
	//M is below A until 12-01 and below C from then on, X is below M
	setup := func(t *testing.T) (*Controller, models.Graph) {
		ctrl, _, _ := newGraphTestController(t)
		g := storeGraph(t, ctrl, models.Graph{
			Nodes: []models.Node{{Id: "root"}, {Id: "A"}, {Id: "C"}, {Id: "M"}, {Id: "X"}},
			Edges: []models.Edge{
				edgeV("A>root", "A", "root", 100, nil, nil), edgeV("C>root", "C", "root", 100, nil, nil), edgeV("X>M", "X", "M", 100, nil, nil),
				edgeV("M>A", "M", "A", 100, nil, gTime(gDec2020)), edgeV("M>C", "M", "C", 100, gTime(gDec2020), nil),
			},
		})
		return ctrl, g
	}
	parentAt := func(t *testing.T, ctrl *Controller, id string, at string, node string) []string {
		r, err, code := ctrl.ReadGraphAt(ctx, permclient.InternalAdminToken, id, *gTime(at))
		expectCode(t, err, code, http.StatusOK)
		return parentOf(r, node)
	}

	t.Run("a move without until across a planned version is refused and names it", func(t *testing.T) {
		ctrl, g := setup(t)
		_, err, code := ctrl.MoveGraphNode(ctx, permclient.InternalAdminToken, g.Id, "M", model.GraphNodeMove{At: gTime(gJun2020), ToNodeId: "root"})
		expectCode(t, err, code, http.StatusBadRequest)
		if !strings.Contains(err.Error(), "node M has changes from 2020-12-01T00:00:00Z") || !strings.Contains(err.Error(), "until") {
			t.Errorf("%v", err)
		}
		//a bound that still reaches the planned version is refused the same way
		_, err, code = ctrl.MoveGraphNode(ctx, permclient.InternalAdminToken, g.Id, "M", model.GraphNodeMove{At: gTime(gJun2020), Until: gTime("2021-03-01T00:00:00Z"), ToNodeId: "root"})
		expectCode(t, err, code, http.StatusBadRequest)
		stored, _, _ := ctrl.ReadGraphVersions(ctx, permclient.InternalAdminToken, g.Id)
		if len(stored.Edges) != 5 {
			t.Errorf("a refused move wrote: %#v", stored.Edges)
		}
	})

	t.Run("a move bounded before the planned version succeeds and leaves it alone", func(t *testing.T) {
		ctrl, g := setup(t)
		_, err, code := ctrl.MoveGraphNode(ctx, permclient.InternalAdminToken, g.Id, "M", model.GraphNodeMove{At: gTime(gJun2020), Until: gTime("2020-10-01T00:00:00Z"), ToNodeId: "root"})
		expectCode(t, err, code, http.StatusOK)
		for at, want := range map[string]string{"2020-03-01T00:00:00Z": "A", "2020-07-01T00:00:00Z": "root", "2020-11-01T00:00:00Z": "A", "2021-03-01T00:00:00Z": "C"} {
			if got := parentAt(t, ctrl, g.Id, at, "M"); !reflect.DeepEqual(got, []string{want}) {
				t.Errorf("at %v: %v, want %v", at, got, want)
			}
		}
		//a bound exactly at the planned version is fine, too
		ctrl, g = setup(t)
		_, err, code = ctrl.MoveGraphNode(ctx, permclient.InternalAdminToken, g.Id, "M", model.GraphNodeMove{At: gTime(gJun2020), Until: gTime(gDec2020), ToNodeId: "root"})
		expectCode(t, err, code, http.StatusOK)
		if got := parentAt(t, ctrl, g.Id, "2021-03-01T00:00:00Z", "M"); !reflect.DeepEqual(got, []string{"C"}) {
			t.Errorf("%v", got)
		}
	})

	t.Run("a move after the planned version works as usual", func(t *testing.T) {
		ctrl, g := setup(t)
		_, err, code := ctrl.MoveGraphNode(ctx, permclient.InternalAdminToken, g.Id, "M", model.GraphNodeMove{At: gTime("2021-03-01T00:00:00Z"), ToNodeId: "A"})
		expectCode(t, err, code, http.StatusOK)
	})

	t.Run("ending a node before a planned version of its edges is refused and names it", func(t *testing.T) {
		ctrl, g := setup(t)
		_, err, code := ctrl.EndGraphNode(ctx, permclient.InternalAdminToken, g.Id, "M", model.GraphNodeEnd{At: gTime(gJun2020)})
		expectCode(t, err, code, http.StatusBadRequest)
		if !strings.Contains(err.Error(), "node M has changes from 2020-12-01T00:00:00Z") {
			t.Errorf("%v", err)
		}
		//the parent's edge counts too: X leads to M, and a planned version of an edge to M changes where X is rerouted
		_, err, code = ctrl.EndGraphNode(ctx, permclient.InternalAdminToken, g.Id, "C", model.GraphNodeEnd{At: gTime(gJun2020)})
		expectCode(t, err, code, http.StatusBadRequest)
		_, err, code = ctrl.EndGraphNode(ctx, permclient.InternalAdminToken, g.Id, "M", model.GraphNodeEnd{At: gTime("2021-03-01T00:00:00Z")})
		expectCode(t, err, code, http.StatusOK)
	})
}
