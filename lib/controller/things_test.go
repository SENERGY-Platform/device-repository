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
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/controller/publisher"
	"github.com/SENERGY-Platform/device-repository/v3/lib/database/testdb"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

const thingTestOtherUserId = "otherUser"

// thingTestOtherUserJwt is a token of a non-admin user without any permissions. The signature is not validated by the controller.
var thingTestOtherUserJwt = func() string {
	enc := base64.RawURLEncoding
	header, _ := json.Marshal(map[string]any{"alg": "HS256", "typ": "JWT"})
	claims, _ := json.Marshal(map[string]any{
		"sub":          thingTestOtherUserId,
		"typ":          "Bearer",
		"roles":        []string{"user"},
		"realm_access": map[string]any{"roles": []string{"user"}},
	})
	return "Bearer " + enc.EncodeToString(header) + "." + enc.EncodeToString(claims) + ".c2ln"
}()

// thingPublisherMock records thing and graph commands; everything else is void
type thingPublisherMock struct {
	publisher.Void
	things        []models.Thing
	deletedThings []string
	graphs        []models.Graph
}

func (this *thingPublisherMock) PublishThing(thing models.Thing) error {
	this.things = append(this.things, thing)
	return nil
}

func (this *thingPublisherMock) PublishThingDelete(id string) error {
	this.deletedThings = append(this.deletedThings, id)
	return nil
}

func (this *thingPublisherMock) PublishGraph(graph models.Graph) error {
	this.graphs = append(this.graphs, graph)
	return nil
}

func newThingTestController(t *testing.T) (ctrl *Controller, permClient client.Client, p *thingPublisherMock) {
	t.Helper()
	conf := configuration.Config{
		DeviceTopic:           "devices",
		DeviceGroupTopic:      "device-groups",
		HubTopic:              "hubs",
		LocationTopic:         "locations",
		GraphTopic:            "graphs",
		ThingTopic:            "things",
		InitPermissionsTopics: true,
	}
	permClient, err := client.NewTestClient(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	p = &thingPublisherMock{}
	ctrl, err = New(conf, testdb.NewTestDB(conf), p, permClient)
	if err != nil {
		t.Fatal(err)
	}
	return ctrl, permClient, p
}

// thingTestDevice creates a device owned by graphTestUserId. If readableBy is not empty, that user may read it too.
func thingTestDevice(t *testing.T, ctrl *Controller, permClient client.Client, id string, readableBy string) {
	t.Helper()
	err := ctrl.db.SetDevice(context.Background(), model.DeviceWithConnectionState{Device: models.Device{Id: id, LocalId: id, Name: id, OwnerId: graphTestUserId}}, func(old, new model.DeviceWithConnectionState) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	permissions := client.ResourcePermissions{UserPermissions: map[string]client.PermissionsMap{
		graphTestUserId: {Read: true, Write: true, Execute: true, Administrate: true},
	}}
	if readableBy != "" {
		permissions.UserPermissions[readableBy] = client.PermissionsMap{Read: true}
	}
	if _, err, _ = permClient.SetPermission(client.InternalAdminToken, "devices", id, permissions); err != nil {
		t.Fatal(err)
	}
}

func thingTestCreate(t *testing.T, ctrl *Controller, name string) models.Thing {
	t.Helper()
	thing, err, _ := ctrl.SetThing(context.Background(), graphTestUserJwt, models.Thing{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return thing
}

func thingTestTime(hour int) *time.Time {
	result := time.Date(2026, 11, 1, hour, 0, 0, 0, time.UTC)
	return &result
}

func expectCode(t *testing.T, err error, code int, expected int) {
	t.Helper()
	if code != expected {
		t.Fatalf("expected code %v, got %v (err=%v)", expected, code, err)
	}
	if expected >= 400 && err == nil {
		t.Fatalf("expected an error for code %v", expected)
	}
	if expected < 400 && err != nil {
		t.Fatalf("unexpected error for code %v: %v", expected, err)
	}
}

func TestThingTopicRegistration(t *testing.T) {
	_, permClient, _ := newThingTestController(t)
	topic, err, _ := permClient.GetTopic(client.InternalAdminToken, "things")
	if err != nil {
		t.Fatal(err)
	}
	if topic.PublishToKafkaTopic != "things" {
		t.Error("unexpected publish_to_kafka_topic: ", topic.PublishToKafkaTopic)
	}
}

func TestThingPublishing(t *testing.T) {
	ctrl, _, p := newThingTestController(t)
	thing := thingTestCreate(t, ctrl, "oven")
	if !reflect.DeepEqual(p.things, []models.Thing{thing}) {
		t.Errorf("create was not published:\n%#v\n%#v", p.things, thing)
	}
	err, code := ctrl.DeleteThing(context.Background(), graphTestUserJwt, thing.Id)
	expectCode(t, err, code, http.StatusOK)
	if !reflect.DeepEqual(p.deletedThings, []string{thing.Id}) {
		t.Errorf("delete was not published: %#v", p.deletedThings)
	}
}

func TestThingCreate(t *testing.T) {
	ctrl, _, _ := newThingTestController(t)
	thing, err, code := ctrl.SetThing(context.Background(), graphTestUserJwt, models.Thing{
		Name:             "oven",
		ContentVariables: []models.ContentVariable{{Name: "power"}},
		DeviceLinks:      []models.ThingDeviceLink{{Id: "x", DeviceId: "d", From: *thingTestTime(1)}},
	})
	expectCode(t, err, code, http.StatusOK)
	if !strings.HasPrefix(thing.Id, models.THING_PREFIX) {
		t.Error("unexpected id", thing.Id)
	}
	if thing.OwnerId != graphTestUserId {
		t.Error("unexpected owner", thing.OwnerId)
	}
	if thing.ContentVariables[0].Id == "" {
		t.Error("content variable has no id")
	}
	if len(thing.DeviceLinks) != 0 {
		t.Error("device links of a new thing must be dropped", thing.DeviceLinks)
	}
	read, err, code := ctrl.ReadThing(context.Background(), graphTestUserJwt, thing.Id)
	expectCode(t, err, code, http.StatusOK)
	if read.Id != thing.Id || read.Name != "oven" {
		t.Error("unexpected thing", read)
	}

	_, err, code = ctrl.SetThing(context.Background(), graphTestUserJwt, models.Thing{})
	expectCode(t, err, code, http.StatusBadRequest)

	//a non-admin may not create a thing for another user
	foreign, err, code := ctrl.SetThing(context.Background(), graphTestUserJwt, models.Thing{Name: "fridge", OwnerId: thingTestOtherUserId})
	expectCode(t, err, code, http.StatusOK)
	if foreign.OwnerId != graphTestUserId {
		t.Error("unexpected owner", foreign.OwnerId)
	}
	_, err, code = ctrl.ReadThing(context.Background(), graphTestUserJwt, foreign.Id)
	expectCode(t, err, code, http.StatusOK)

	//empty lists are answered as [] instead of null
	empty, err, code := ctrl.SetThing(context.Background(), graphTestUserJwt, models.Thing{Name: "lamp"})
	expectCode(t, err, code, http.StatusOK)
	encoded, err := json.Marshal(empty)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"attributes":[]`, `"content_variables":[]`, `"device_links":[]`} {
		if !strings.Contains(string(encoded), field) {
			t.Errorf("expected %v in %v", field, string(encoded))
		}
	}
}

func TestThingForeignUser(t *testing.T) {
	ctrl, permClient, _ := newThingTestController(t)
	thing := thingTestCreate(t, ctrl, "oven")
	thingTestDevice(t, ctrl, permClient, "d1", thingTestOtherUserId)
	ctx := context.Background()

	_, err, code := ctrl.ReadThing(ctx, thingTestOtherUserJwt, thing.Id)
	expectCode(t, err, code, http.StatusForbidden)

	thing.Name = "hacked"
	_, err, code = ctrl.SetThing(ctx, thingTestOtherUserJwt, thing)
	expectCode(t, err, code, http.StatusForbidden)

	_, err, code = ctrl.AddThingDeviceLink(ctx, thingTestOtherUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "d1"})
	expectCode(t, err, code, http.StatusForbidden)

	_, err, code = ctrl.EndThingDeviceLink(ctx, thingTestOtherUserJwt, thing.Id, "any", model.ThingDeviceLinkEnd{})
	expectCode(t, err, code, http.StatusForbidden)

	err, code = ctrl.DeleteThing(ctx, thingTestOtherUserJwt, thing.Id)
	expectCode(t, err, code, http.StatusForbidden)

	list, total, err, code := ctrl.ListThings(ctx, thingTestOtherUserJwt, model.ThingListOptions{})
	expectCode(t, err, code, http.StatusOK)
	if len(list) != 0 || total != 0 {
		t.Error("foreign user must not see the thing", list)
	}

	//admin may do everything
	read, err, code := ctrl.ReadThing(ctx, client.InternalAdminToken, thing.Id)
	expectCode(t, err, code, http.StatusOK)
	read.Name = "renamed by admin"
	_, err, code = ctrl.SetThing(ctx, client.InternalAdminToken, read)
	expectCode(t, err, code, http.StatusOK)
	updated, err, code := ctrl.AddThingDeviceLink(ctx, client.InternalAdminToken, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "d1", From: thingTestTime(1)})
	expectCode(t, err, code, http.StatusOK)
	_, err, code = ctrl.EndThingDeviceLink(ctx, client.InternalAdminToken, thing.Id, updated.DeviceLinks[0].Id, model.ThingDeviceLinkEnd{Until: thingTestTime(23)})
	expectCode(t, err, code, http.StatusOK)
	list, _, err, code = ctrl.ListThings(ctx, client.InternalAdminToken, model.ThingListOptions{})
	expectCode(t, err, code, http.StatusOK)
	if len(list) != 1 {
		t.Error("admin must see the thing", list)
	}
	err, code = ctrl.DeleteThing(ctx, client.InternalAdminToken, thing.Id)
	expectCode(t, err, code, http.StatusOK)
}

func TestThingMissing(t *testing.T) {
	ctrl, _, _ := newThingTestController(t)
	ctx := context.Background()
	_, err, code := ctrl.AddThingDeviceLink(ctx, client.InternalAdminToken, "urn:infai:ses:thing:missing", model.ThingDeviceLinkCreate{DeviceId: "d1"})
	expectCode(t, err, code, http.StatusNotFound)
	_, err, code = ctrl.EndThingDeviceLink(ctx, client.InternalAdminToken, "urn:infai:ses:thing:missing", "l", model.ThingDeviceLinkEnd{})
	expectCode(t, err, code, http.StatusNotFound)
	_, err, code = ctrl.ReadThing(ctx, client.InternalAdminToken, "urn:infai:ses:thing:missing")
	expectCode(t, err, code, http.StatusNotFound)
}

func TestThingPutKeepsDeviceLinks(t *testing.T) {
	ctrl, permClient, _ := newThingTestController(t)
	thingTestDevice(t, ctrl, permClient, "d1", "")
	ctx := context.Background()
	thing := thingTestCreate(t, ctrl, "oven")
	thing, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "d1", From: thingTestTime(1)})
	expectCode(t, err, code, http.StatusOK)
	stored := slices.Clone(thing.DeviceLinks)

	changed := thing
	changed.Name = "oven 2"
	changed.DeviceLinks = []models.ThingDeviceLink{{Id: "other", DeviceId: "d2", From: *thingTestTime(5)}}
	result, err, code := ctrl.SetThing(ctx, graphTestUserJwt, changed)
	expectCode(t, err, code, http.StatusOK)
	if result.Name != "oven 2" || !reflect.DeepEqual(result.DeviceLinks, stored) {
		t.Errorf("unexpected result %#v", result)
	}
	read, err, code := ctrl.ReadThing(ctx, graphTestUserJwt, thing.Id)
	expectCode(t, err, code, http.StatusOK)
	if !reflect.DeepEqual(read.DeviceLinks, stored) {
		t.Errorf("stored links changed: %#v", read.DeviceLinks)
	}
}

func TestThingOwnerChange(t *testing.T) {
	ctrl, _, _ := newThingTestController(t)
	ctx := context.Background()
	thing := thingTestCreate(t, ctrl, "oven")
	thing.OwnerId = thingTestOtherUserId
	//new owner has no admin rights on the thing
	_, err, code := ctrl.SetThing(ctx, graphTestUserJwt, thing)
	expectCode(t, err, code, http.StatusBadRequest)
}

func TestThingListSort(t *testing.T) {
	ctrl, _, _ := newThingTestController(t)
	thingTestCreate(t, ctrl, "b")
	_, _, err, code := ctrl.ListThings(context.Background(), graphTestUserJwt, model.ThingListOptions{SortBy: "name.asc"})
	expectCode(t, err, code, http.StatusOK)
}

func TestThingAddLink(t *testing.T) {
	ctrl, permClient, _ := newThingTestController(t)
	ctx := context.Background()
	thingTestDevice(t, ctrl, permClient, "d1", "")
	thingTestDevice(t, ctrl, permClient, "d2", "")
	thingTestDevice(t, ctrl, permClient, "foreign", "")
	thing := thingTestCreate(t, ctrl, "oven")
	other := thingTestCreate(t, ctrl, "fridge")
	add := func(thingId string, deviceId string, from *time.Time, until *time.Time) (models.Thing, error, int) {
		return ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thingId, model.ThingDeviceLinkCreate{DeviceId: deviceId, From: from, Until: until})
	}

	t.Run("unknown device", func(t *testing.T) {
		_, err, code := add(thing.Id, "unknown", nil, nil)
		expectCode(t, err, code, http.StatusBadRequest)
	})
	t.Run("missing device id", func(t *testing.T) {
		_, err, code := add(thing.Id, "", nil, nil)
		expectCode(t, err, code, http.StatusBadRequest)
	})
	t.Run("no read access to device", func(t *testing.T) {
		//otherUser may write a thing of their own but may not read the device
		own, err, code := ctrl.SetThing(ctx, thingTestOtherUserJwt, models.Thing{Name: "own"})
		expectCode(t, err, code, http.StatusOK)
		_, err, code = ctrl.AddThingDeviceLink(ctx, thingTestOtherUserJwt, own.Id, model.ThingDeviceLinkCreate{DeviceId: "foreign"})
		expectCode(t, err, code, http.StatusForbidden)
	})
	t.Run("until not after from", func(t *testing.T) {
		_, err, code := add(thing.Id, "d1", thingTestTime(5), thingTestTime(5))
		expectCode(t, err, code, http.StatusBadRequest)
		_, err, code = add(thing.Id, "d1", thingTestTime(5), thingTestTime(4))
		expectCode(t, err, code, http.StatusBadRequest)
	})
	t.Run("defaults from to now and normalizes", func(t *testing.T) {
		before := time.Now().Add(-time.Second)
		result, err, code := add(other.Id, "d2", nil, nil)
		expectCode(t, err, code, http.StatusOK)
		link := result.DeviceLinks[0]
		if link.Id == "" || link.Until != nil || link.From.Before(before) || link.From.Location() != time.UTC || link.From.Nanosecond()%int(time.Millisecond) != 0 {
			t.Errorf("unexpected link %#v", link)
		}
		later := link.From.Add(time.Hour)
		_, err, code = ctrl.EndThingDeviceLink(ctx, graphTestUserJwt, other.Id, link.Id, model.ThingDeviceLinkEnd{Until: &later})
		expectCode(t, err, code, http.StatusOK)
	})

	//[1,5) d1 in thing
	result, err, code := add(thing.Id, "d1", thingTestTime(1), thingTestTime(5))
	expectCode(t, err, code, http.StatusOK)
	if len(result.DeviceLinks) != 1 {
		t.Fatal(result.DeviceLinks)
	}

	t.Run("overlap in same thing", func(t *testing.T) {
		_, err, code := add(thing.Id, "d1", thingTestTime(4), thingTestTime(8))
		expectCode(t, err, code, http.StatusBadRequest)
	})
	t.Run("overlap in other thing", func(t *testing.T) {
		_, err, code := add(other.Id, "d1", thingTestTime(2), thingTestTime(3))
		expectCode(t, err, code, http.StatusBadRequest)
	})
	t.Run("touching ranges are fine", func(t *testing.T) {
		_, err, code := add(thing.Id, "d1", thingTestTime(5), thingTestTime(7))
		expectCode(t, err, code, http.StatusOK)
		_, err, code = add(other.Id, "d1", thingTestTime(0), thingTestTime(1))
		expectCode(t, err, code, http.StatusOK)
	})
	t.Run("other device same range", func(t *testing.T) {
		_, err, code := add(thing.Id, "d2", thingTestTime(1), thingTestTime(5))
		expectCode(t, err, code, http.StatusOK)
	})
	t.Run("open link blocks later link of same device", func(t *testing.T) {
		_, err, code := add(thing.Id, "d1", thingTestTime(9), nil)
		expectCode(t, err, code, http.StatusOK)
		_, err, code = add(other.Id, "d1", thingTestTime(12), thingTestTime(13))
		expectCode(t, err, code, http.StatusBadRequest)
	})
}

func TestThingEndLink(t *testing.T) {
	ctrl, permClient, _ := newThingTestController(t)
	ctx := context.Background()
	thingTestDevice(t, ctrl, permClient, "d1", "")
	thing := thingTestCreate(t, ctrl, "oven")
	thing, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "d1", From: thingTestTime(1)})
	expectCode(t, err, code, http.StatusOK)
	linkId := thing.DeviceLinks[0].Id

	_, err, code = ctrl.EndThingDeviceLink(ctx, graphTestUserJwt, thing.Id, "unknown", model.ThingDeviceLinkEnd{})
	expectCode(t, err, code, http.StatusNotFound)

	_, err, code = ctrl.EndThingDeviceLink(ctx, graphTestUserJwt, thing.Id, linkId, model.ThingDeviceLinkEnd{Until: thingTestTime(0)})
	expectCode(t, err, code, http.StatusBadRequest)
	_, err, code = ctrl.EndThingDeviceLink(ctx, graphTestUserJwt, thing.Id, linkId, model.ThingDeviceLinkEnd{Until: thingTestTime(1)})
	expectCode(t, err, code, http.StatusBadRequest)

	result, err, code := ctrl.EndThingDeviceLink(ctx, graphTestUserJwt, thing.Id, linkId, model.ThingDeviceLinkEnd{})
	expectCode(t, err, code, http.StatusOK)
	if result.DeviceLinks[0].Until == nil || result.DeviceLinks[0].Until.Before(*thingTestTime(1)) {
		t.Errorf("until not set: %#v", result.DeviceLinks[0])
	}
	read, err, code := ctrl.ReadThing(ctx, graphTestUserJwt, thing.Id)
	expectCode(t, err, code, http.StatusOK)
	if read.DeviceLinks[0].Until == nil {
		t.Error("end was not stored")
	}

	_, err, code = ctrl.EndThingDeviceLink(ctx, graphTestUserJwt, thing.Id, linkId, model.ThingDeviceLinkEnd{})
	expectCode(t, err, code, http.StatusBadRequest)

	//a link that has not started yet ends one millisecond after its start, an explicit until before its start is rejected
	future := normalizeThingTime(time.Now().Add(48 * time.Hour))
	thing, err, code = ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "d1", From: &future})
	expectCode(t, err, code, http.StatusOK)
	futureLinkId := thing.DeviceLinks[1].Id
	past := future.Add(-time.Hour)
	_, err, code = ctrl.EndThingDeviceLink(ctx, graphTestUserJwt, thing.Id, futureLinkId, model.ThingDeviceLinkEnd{Until: &past})
	expectCode(t, err, code, http.StatusBadRequest)
	result, err, code = ctrl.EndThingDeviceLink(ctx, graphTestUserJwt, thing.Id, futureLinkId, model.ThingDeviceLinkEnd{})
	expectCode(t, err, code, http.StatusOK)
	if until := result.DeviceLinks[1].Until; until == nil || !until.Equal(future.Add(time.Millisecond)) {
		t.Errorf("unexpected until of future link: %#v", result.DeviceLinks[1])
	}
}

// thingTestGraph builds a valid graph: neither thing nor device may be an end node, so all of them lead to a plain node.
func thingTestGraph(thingId string, deviceIds ...string) models.Graph {
	graph := models.Graph{
		Attributes: []models.Attribute{{Key: "name", Value: "g"}},
		Nodes: []models.Node{
			{Id: "t", ResourceType: models.GraphResourceTypeThing, ResourceId: thingId},
			{Id: "end"},
		},
		Edges: []models.Edge{{Id: "t->end", FromNodeId: "t", ToNodeId: "end", Weight: 100}},
	}
	for _, id := range deviceIds {
		graph.Nodes = append(graph.Nodes, models.Node{Id: id, ResourceType: models.GraphResourceTypeDevice, ResourceId: id})
		graph.Edges = append(graph.Edges, models.Edge{Id: id + "->end", FromNodeId: id, ToNodeId: "end", Weight: 100})
	}
	return graph
}

func TestThingGraphRule(t *testing.T) {
	ctrl, permClient, _ := newThingTestController(t)
	ctx := context.Background()
	for _, id := range []string{"d-open", "d-ended", "d-free"} {
		thingTestDevice(t, ctrl, permClient, id, "")
	}
	thing := thingTestCreate(t, ctrl, "oven")
	thing, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "d-open", From: thingTestTime(1)})
	expectCode(t, err, code, http.StatusOK)
	thing, err, code = ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "d-ended", From: thingTestTime(1), Until: thingTestTime(2)})
	expectCode(t, err, code, http.StatusOK)

	t.Run("unknown thing", func(t *testing.T) {
		_, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, thingTestGraph("urn:infai:ses:thing:unknown"))
		expectCode(t, err, code, http.StatusBadRequest)
	})
	t.Run("linked open device", func(t *testing.T) {
		_, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, thingTestGraph(thing.Id, "d-open"))
		expectCode(t, err, code, http.StatusBadRequest)
	})
	t.Run("linked ended device", func(t *testing.T) {
		_, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, thingTestGraph(thing.Id, "d-ended"))
		expectCode(t, err, code, http.StatusBadRequest)
	})
	t.Run("unlinked device", func(t *testing.T) {
		_, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, thingTestGraph(thing.Id, "d-free"))
		expectCode(t, err, code, http.StatusOK)
	})
	t.Run("link to device already in graph with thing", func(t *testing.T) {
		//the graph of the previous subtest contains thing and d-free
		_, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "d-free", From: thingTestTime(3)})
		expectCode(t, err, code, http.StatusBadRequest)
	})
}

func TestThingDeleteRemovesGraphNode(t *testing.T) {
	ctrl, permClient, p := newThingTestController(t)
	ctx := context.Background()
	thingTestDevice(t, ctrl, permClient, "d1", "")
	thing := thingTestCreate(t, ctrl, "oven")
	graph, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, thingTestGraph(thing.Id, "d1"))
	expectCode(t, err, code, http.StatusOK)
	//a second graph without the thing must stay untouched
	untouched, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, models.Graph{Nodes: []models.Node{{Id: "1"}}})
	expectCode(t, err, code, http.StatusOK)
	published := len(p.graphs)

	err, code = ctrl.DeleteThing(ctx, graphTestUserJwt, thing.Id)
	expectCode(t, err, code, http.StatusOK)

	read, err, code := ctrl.ReadGraph(ctx, graphTestUserJwt, graph.Id)
	expectCode(t, err, code, http.StatusOK)
	for _, node := range read.Nodes {
		if node.ResourceType == models.GraphResourceTypeThing {
			t.Error("thing node still in graph", read.Nodes)
		}
	}
	if len(read.Nodes) != 2 {
		t.Error("unexpected nodes", read.Nodes)
	}
	if err = read.Valid(); err != nil {
		t.Error(err)
	}
	if len(p.graphs) != published+1 {
		t.Errorf("only the graph with the thing must be republished, %v -> %v", published, len(p.graphs))
	}
	_, err, code = ctrl.ReadGraph(ctx, graphTestUserJwt, untouched.Id)
	expectCode(t, err, code, http.StatusOK)
	_, err, code = ctrl.ReadThing(ctx, client.InternalAdminToken, thing.Id)
	expectCode(t, err, code, http.StatusNotFound)
}

func TestThingDeviceDeleteEndsOpenLinks(t *testing.T) {
	ctrl, permClient, _ := newThingTestController(t)
	ctx := context.Background()
	thingTestDevice(t, ctrl, permClient, "d1", "")
	thingTestDevice(t, ctrl, permClient, "d2", "")
	thing := thingTestCreate(t, ctrl, "oven")
	other := thingTestCreate(t, ctrl, "fridge")
	add := func(thingId string, deviceId string, from *time.Time, until *time.Time) models.Thing {
		result, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thingId, model.ThingDeviceLinkCreate{DeviceId: deviceId, From: from, Until: until})
		expectCode(t, err, code, http.StatusOK)
		return result
	}
	add(thing.Id, "d1", thingTestTime(1), thingTestTime(2)) //ended
	add(thing.Id, "d1", thingTestTime(3), nil)              //open
	add(thing.Id, "d2", thingTestTime(1), nil)              //open, other device
	add(other.Id, "d1", thingTestTime(0), thingTestTime(1)) //ended, other thing

	err, code := ctrl.DeleteDevice(ctx, client.InternalAdminToken, "d1")
	expectCode(t, err, code, http.StatusOK)

	read, err, code := ctrl.ReadThing(ctx, graphTestUserJwt, thing.Id)
	expectCode(t, err, code, http.StatusOK)
	if len(read.DeviceLinks) != 3 {
		t.Fatal(read.DeviceLinks)
	}
	for _, link := range read.DeviceLinks {
		switch {
		case link.DeviceId == "d1" && link.From.Equal(*thingTestTime(1)):
			if link.Until == nil || !link.Until.Equal(*thingTestTime(2)) {
				t.Errorf("ended link changed: %#v", link)
			}
		case link.DeviceId == "d1" && link.From.Equal(*thingTestTime(3)):
			if link.Until == nil || !link.Until.After(link.From) {
				t.Errorf("open link was not ended: %#v", link)
			}
		case link.DeviceId == "d2":
			if link.Until != nil {
				t.Errorf("link of other device changed: %#v", link)
			}
		default:
			t.Errorf("unexpected link %#v", link)
		}
	}
	otherRead, err, code := ctrl.ReadThing(ctx, graphTestUserJwt, other.Id)
	expectCode(t, err, code, http.StatusOK)
	if len(otherRead.DeviceLinks) != 1 || otherRead.DeviceLinks[0].Until == nil || !otherRead.DeviceLinks[0].Until.Equal(*thingTestTime(1)) {
		t.Errorf("ended link of other thing changed: %#v", otherRead.DeviceLinks)
	}

	//running the handler again changes nothing
	before := read.DeviceLinks
	if err = ctrl.endOpenThingLinksOfDevice(ctx, "d1"); err != nil {
		t.Fatal(err)
	}
	again, _, _ := ctrl.ReadThing(ctx, graphTestUserJwt, thing.Id)
	if !reflect.DeepEqual(before, again.DeviceLinks) {
		t.Errorf("second run changed links:\n%#v\n%#v", before, again.DeviceLinks)
	}
}

func TestThingDeviceDeleteEndsFutureLink(t *testing.T) {
	ctrl, permClient, _ := newThingTestController(t)
	ctx := context.Background()
	thingTestDevice(t, ctrl, permClient, "d1", "")
	thing := thingTestCreate(t, ctrl, "oven")
	future := normalizeThingTime(time.Now().Add(48 * time.Hour))
	_, err, code := ctrl.AddThingDeviceLink(ctx, graphTestUserJwt, thing.Id, model.ThingDeviceLinkCreate{DeviceId: "d1", From: &future})
	expectCode(t, err, code, http.StatusOK)

	err, code = ctrl.DeleteDevice(ctx, client.InternalAdminToken, "d1")
	expectCode(t, err, code, http.StatusOK)

	read, err, code := ctrl.ReadThing(ctx, graphTestUserJwt, thing.Id)
	expectCode(t, err, code, http.StatusOK)
	if len(read.DeviceLinks) != 1 {
		t.Fatal(read.DeviceLinks)
	}
	if until := read.DeviceLinks[0].Until; until == nil || !until.Equal(future.Add(time.Millisecond)) {
		t.Errorf("unexpected until of future link: %#v", read.DeviceLinks[0])
	}
}

func TestThingListEmptyIsNotNull(t *testing.T) {
	ctrl, _, _ := newThingTestController(t)
	result, _, err, code := ctrl.ListThings(context.Background(), graphTestUserJwt, model.ThingListOptions{})
	expectCode(t, err, code, http.StatusOK)
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != "[]" {
		t.Errorf("expected [], got %v", string(encoded))
	}
}
