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

package mongo

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

func TestMongoThings(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := configuration.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	port, _, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Fatal(err)
	}
	conf.MongoUrl = "mongodb://localhost:" + port
	m, err := New(conf)
	if err != nil {
		t.Fatal(err)
	}
	noSync := func(models.Thing) error { return nil }
	timeout := func() context.Context {
		c, cancelTimeout := context.WithTimeout(ctx, 5*time.Second)
		t.Cleanup(cancelTimeout)
		return c
	}

	from := time.Date(2026, 11, 1, 10, 0, 0, 123*int(time.Millisecond), time.UTC)
	until := time.Date(2026, 11, 2, 10, 0, 0, 456*int(time.Millisecond), time.UTC)
	things := []models.Thing{
		{
			Id: "urn:infai:ses:thing:1", Name: "Kombidaempfer", OwnerId: "u1",
			Attributes:       []models.Attribute{{Key: "power_w", Value: "3500", Origin: "test"}},
			ContentVariables: []models.ContentVariable{{Id: "cv1", Name: "power"}},
			DeviceLinks: []models.ThingDeviceLink{
				{Id: "l1", DeviceId: "d-open", From: from, Until: nil},
				{Id: "l2", DeviceId: "d-ended", From: from, Until: &until},
			},
		},
		{Id: "urn:infai:ses:thing:2", Name: "Backofen", OwnerId: "u1", DeviceLinks: []models.ThingDeviceLink{{Id: "l3", DeviceId: "d-other", From: from}}},
		{Id: "urn:infai:ses:thing:3", Name: "Kuehlschrank", OwnerId: "u1"},
	}
	for _, thing := range things {
		if err = m.SetThing(timeout(), thing, noSync); err != nil {
			t.Fatal(err)
		}
	}

	ids := func(list []models.Thing) (result []string) {
		for _, thing := range list {
			result = append(result, thing.Id)
		}
		return result
	}

	t.Run("round trip", func(t *testing.T) {
		got, exists, err := m.GetThing(timeout(), things[0].Id)
		if err != nil || !exists {
			t.Fatal(err, exists)
		}
		if got.Name != things[0].Name || got.OwnerId != "u1" || len(got.DeviceLinks) != 2 || len(got.Attributes) != 1 || got.ContentVariables[0].Id != "cv1" {
			t.Fatalf("%#v", got)
		}
		open, ended := got.DeviceLinks[0], got.DeviceLinks[1]
		if open.Until != nil {
			t.Errorf("open link has until %v", open.Until)
		}
		if !open.From.Equal(from) || ended.Until == nil || !ended.Until.Equal(until) {
			t.Errorf("times changed: %#v", got.DeviceLinks)
		}
		_, exists, err = m.GetThing(timeout(), "unknown")
		if err != nil || exists {
			t.Error(err, exists)
		}
	})

	t.Run("replace keeps one document", func(t *testing.T) {
		changed := things[2]
		changed.Name = "Kuehlschrank 2"
		if err := m.SetThing(timeout(), changed, noSync); err != nil {
			t.Fatal(err)
		}
		list, total, err := m.ListThings(timeout(), model.ThingListOptions{})
		if err != nil || total != 3 || len(list) != 3 {
			t.Fatal(err, total, len(list))
		}
		things[2] = changed
	})

	t.Run("list filters", func(t *testing.T) {
		cases := []struct {
			name     string
			options  model.ThingListOptions
			expected []string
		}{
			{"all", model.ThingListOptions{}, []string{things[0].Id, things[1].Id, things[2].Id}},
			{"open link device", model.ThingListOptions{DeviceIds: []string{"d-open"}}, []string{things[0].Id}},
			{"ended link device", model.ThingListOptions{DeviceIds: []string{"d-ended"}}, []string{things[0].Id}},
			{"two devices", model.ThingListOptions{DeviceIds: []string{"d-open", "d-other"}}, []string{things[0].Id, things[1].Id}},
			{"unknown device", model.ThingListOptions{DeviceIds: []string{"nope"}}, nil},
			{"empty device ids add no condition", model.ThingListOptions{DeviceIds: []string{}}, []string{things[0].Id, things[1].Id, things[2].Id}},
			{"ids", model.ThingListOptions{Ids: []string{things[1].Id, things[2].Id}}, []string{things[1].Id, things[2].Id}},
			{"empty ids", model.ThingListOptions{Ids: []string{}}, nil},
			{"search name", model.ThingListOptions{Search: "backof"}, []string{things[1].Id}},
			{"search id", model.ThingListOptions{Search: "thing:3"}, []string{things[2].Id}},
			{"search is literal", model.ThingListOptions{Search: ".*"}, nil},
			{"limit and offset", model.ThingListOptions{Limit: 1, Offset: 1}, []string{things[1].Id}},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				list, _, err := m.ListThings(timeout(), c.options)
				if err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(ids(list), c.expected) {
					t.Errorf("expected %v, got %v", c.expected, ids(list))
				}
			})
		}
	})

	t.Run("total ignores limit", func(t *testing.T) {
		list, total, err := m.ListThings(timeout(), model.ThingListOptions{Limit: 1})
		if err != nil || len(list) != 1 || total != 3 {
			t.Fatal(err, len(list), total)
		}
	})

	t.Run("sort", func(t *testing.T) {
		list, _, err := m.ListThings(timeout(), model.ThingListOptions{SortBy: "name.desc"})
		if err != nil {
			t.Fatal(err)
		}
		//Kuehlschrank 2, Kombidaempfer, Backofen
		if !slices.Equal(ids(list), []string{things[2].Id, things[0].Id, things[1].Id}) {
			t.Errorf("unexpected order %v", ids(list))
		}
		list, _, err = m.ListThings(timeout(), model.ThingListOptions{SortBy: "id.desc"})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(ids(list), []string{things[2].Id, things[1].Id, things[0].Id}) {
			t.Errorf("unexpected order %v", ids(list))
		}
		_, _, err = m.ListThings(timeout(), model.ThingListOptions{SortBy: "owner_id.asc"})
		if !errors.Is(err, model.ErrUnsupportedSort) {
			t.Errorf("expected ErrUnsupportedSort, got %v", err)
		}
	})

	t.Run("graphs by thing", func(t *testing.T) {
		noGraphSync := func(models.Graph) error { return nil }
		node := func(id string, resourceType models.GraphResourceType, resourceId string) models.Node {
			return models.Node{Id: id, ResourceType: resourceType, ResourceId: resourceId}
		}
		graphs := []models.Graph{
			{Id: "g1", Owner: "u1", Nodes: []models.Node{node("a", models.GraphResourceTypeThing, things[0].Id), node("b", models.GraphResourceTypeDevice, "d1"), node("c", "", "")}},
			{Id: "g2", Owner: "u1", Nodes: []models.Node{node("a", models.GraphResourceTypeThing, things[1].Id), node("b", models.GraphResourceTypeDevice, "d2")}},
			//a device node with the id of a thing must not match a thing filter
			{Id: "g3", Owner: "u1", Nodes: []models.Node{node("a", models.GraphResourceTypeDevice, things[0].Id)}},
		}
		for _, graph := range graphs {
			if err := m.SetGraph(timeout(), graph, noGraphSync); err != nil {
				t.Fatal(err)
			}
		}
		graphIds := func(options model.GraphListOptions) []string {
			list, _, err := m.ListGraphs(timeout(), options)
			if err != nil {
				t.Fatal(err)
			}
			result := []string{}
			for _, graph := range list {
				result = append(result, graph.Id)
			}
			slices.Sort(result)
			return result
		}
		if got := graphIds(model.GraphListOptions{ThingIds: []string{things[0].Id}}); !slices.Equal(got, []string{"g1"}) {
			t.Errorf("thing 1: %v", got)
		}
		if got := graphIds(model.GraphListOptions{ThingIds: []string{things[0].Id, things[1].Id}}); !slices.Equal(got, []string{"g1", "g2"}) {
			t.Errorf("thing 1+2: %v", got)
		}
		if got := graphIds(model.GraphListOptions{ThingIds: []string{things[0].Id}, DeviceIds: []string{"d1"}}); !slices.Equal(got, []string{"g1"}) {
			t.Errorf("thing 1 and d1: %v", got)
		}
		if got := graphIds(model.GraphListOptions{ThingIds: []string{things[0].Id}, DeviceIds: []string{"d2"}}); len(got) != 0 {
			t.Errorf("thing 1 and d2: %v", got)
		}
		if got := graphIds(model.GraphListOptions{ThingIds: []string{"nope"}}); len(got) != 0 {
			t.Errorf("unknown thing: %v", got)
		}
		if got := graphIds(model.GraphListOptions{ThingIds: []string{}}); !slices.Equal(got, []string{"g1", "g2", "g3"}) {
			t.Errorf("empty thing ids: %v", got)
		}
	})

	t.Run("remove", func(t *testing.T) {
		deleted := []string{}
		err := m.RemoveThing(timeout(), things[0].Id, func(thing models.Thing) error {
			deleted = append(deleted, thing.Id)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(deleted, []string{things[0].Id}) {
			t.Errorf("sync delete handler: %v", deleted)
		}
		_, exists, err := m.GetThing(timeout(), things[0].Id)
		if err != nil || exists {
			t.Error(err, exists)
		}
		list, _, err := m.ListThings(timeout(), model.ThingListOptions{DeviceIds: []string{"d-open"}})
		if err != nil || len(list) != 0 {
			t.Error(err, list)
		}
		//removing an unknown thing is not an error
		if err = m.RemoveThing(timeout(), "unknown", func(models.Thing) error { return nil }); err != nil {
			t.Error(err)
		}
	})
}
