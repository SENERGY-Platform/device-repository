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

package tests

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib"
	"github.com/SENERGY-Platform/device-repository/v3/lib/client"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

// TestThings runs against the started service and fails loudly if the environment does not come up.
func TestThings(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const userid = "testOwner"
	const userjwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	config.SyncLockDuration = time.Second.String()
	config.Debug = true

	config, err = docker.NewEnv(ctx, wg, config)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)

	c := client.NewClient("http://localhost:"+config.ServerPort, client.Options{})
	bg := context.Background()

	protocol, err, _ := c.SetProtocol(bg, client.InternalAdminToken, models.Protocol{
		Name:             "p1",
		Handler:          "p1",
		ProtocolSegments: []models.ProtocolSegment{{Id: "ps1", Name: "ps1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	deviceType, err, _ := c.SetDeviceType(bg, client.InternalAdminToken, models.DeviceType{
		Name: "dt1",
		Services: []models.Service{{
			LocalId:     "s1",
			Name:        "s1",
			Interaction: models.REQUEST,
			ProtocolId:  protocol.Id,
		}},
	}, client.DeviceTypeUpdateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	deviceIds := []string{}
	for _, name := range []string{"d0", "d1", "d2"} {
		d, err, _ := c.CreateDevice(bg, userjwt, models.Device{LocalId: name, Name: name, DeviceTypeId: deviceType.Id})
		if err != nil {
			t.Fatal(err)
		}
		deviceIds = append(deviceIds, d.Id)
	}

	//all periods lie in the past, so ending a link "now" is after its from
	base := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -30)
	day := func(n int) time.Time {
		return base.AddDate(0, 0, n)
	}
	expectCode := func(t *testing.T, err error, code int, expected int) {
		t.Helper()
		if err == nil {
			t.Fatalf("expected error with code %v, got none", expected)
		}
		if code != expected {
			t.Fatalf("expected code %v, got %v: %v", expected, code, err)
		}
	}

	var thing models.Thing
	t.Run("create thing", func(t *testing.T) {
		thing, err, _ = c.SetThing(bg, userjwt, models.Thing{
			Name:       "oven",
			Attributes: []models.Attribute{{Key: "power_w", Value: "3500", Origin: "test"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if thing.Id == "" || thing.OwnerId != userid || thing.Name != "oven" {
			t.Fatalf("unexpected thing: %#v", thing)
		}
		if len(thing.DeviceLinks) != 0 {
			t.Fatalf("unexpected links: %#v", thing.DeviceLinks)
		}
	})

	t.Run("read thing", func(t *testing.T) {
		result, err, _ := c.ReadThing(bg, userjwt, thing.Id)
		if err != nil {
			t.Fatal(err)
		}
		if result.Id != thing.Id || result.Name != "oven" || len(result.Attributes) != 1 {
			t.Fatalf("unexpected thing: %#v", result)
		}
	})

	t.Run("create with preset id is rejected", func(t *testing.T) {
		//client.SetThing uses PUT for a preset id, so POST is sent directly
		req, err := http.NewRequestWithContext(bg, http.MethodPost, "http://localhost:"+config.ServerPort+"/things", strings.NewReader(`{"id":"preset","name":"x"}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", userjwt)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("unexpected status %v", resp.StatusCode)
		}
	})

	var link models.ThingDeviceLink
	t.Run("add link", func(t *testing.T) {
		from, until := day(1), day(10)
		result, err, _ := c.AddThingDeviceLink(bg, userjwt, thing.Id, client.ThingDeviceLinkCreate{DeviceId: deviceIds[0], From: &from, Until: &until})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.DeviceLinks) != 1 {
			t.Fatalf("unexpected links: %#v", result.DeviceLinks)
		}
		link = result.DeviceLinks[0]
		if link.Id == "" || link.DeviceId != deviceIds[0] || !link.From.Equal(from) || link.Until == nil || !link.Until.Equal(until) {
			t.Fatalf("unexpected link: %#v", link)
		}
	})

	t.Run("put keeps links", func(t *testing.T) {
		changed := thing
		changed.Name = "oven kitchen"
		changed.DeviceLinks = nil
		result, err, _ := c.SetThing(bg, userjwt, changed)
		if err != nil {
			t.Fatal(err)
		}
		if result.Name != "oven kitchen" || len(result.DeviceLinks) != 1 || result.DeviceLinks[0].Id != link.Id {
			t.Fatalf("unexpected thing: %#v", result)
		}
		read, err, _ := c.ReadThing(bg, userjwt, thing.Id)
		if err != nil {
			t.Fatal(err)
		}
		if read.Name != "oven kitchen" || len(read.DeviceLinks) != 1 || read.DeviceLinks[0].Id != link.Id {
			t.Fatalf("unexpected thing: %#v", read)
		}
		thing = read
	})

	t.Run("overlapping link is rejected", func(t *testing.T) {
		from := day(5)
		_, err, code := c.AddThingDeviceLink(bg, userjwt, thing.Id, client.ThingDeviceLinkCreate{DeviceId: deviceIds[0], From: &from})
		expectCode(t, err, code, http.StatusBadRequest)
	})

	t.Run("touching link is accepted", func(t *testing.T) {
		from := day(10)
		result, err, _ := c.AddThingDeviceLink(bg, userjwt, thing.Id, client.ThingDeviceLinkCreate{DeviceId: deviceIds[0], From: &from})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.DeviceLinks) != 2 {
			t.Fatalf("unexpected links: %#v", result.DeviceLinks)
		}
	})

	t.Run("end link with empty body", func(t *testing.T) {
		//the open link of device 0 ends now; an empty body must not be a 400
		var open models.ThingDeviceLink
		current, err, _ := c.ReadThing(bg, userjwt, thing.Id)
		if err != nil {
			t.Fatal(err)
		}
		for _, l := range current.DeviceLinks {
			if l.Until == nil {
				open = l
			}
		}
		if open.Id == "" {
			t.Fatalf("no open link: %#v", current.DeviceLinks)
		}
		before := time.Now().Add(-time.Minute)
		req, err := http.NewRequestWithContext(bg, http.MethodPost, "http://localhost:"+config.ServerPort+"/things/"+thing.Id+"/device-links/"+open.Id+"/end", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", userjwt)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status %v", resp.StatusCode)
		}
		after, err, _ := c.ReadThing(bg, userjwt, thing.Id)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, l := range after.DeviceLinks {
			if l.Id == open.Id {
				found = true
				if l.Until == nil || l.Until.Before(before) {
					t.Fatalf("unexpected until: %#v", l)
				}
			}
		}
		if !found {
			t.Fatalf("link vanished: %#v", after.DeviceLinks)
		}
		thing = after

		t.Run("ending again is rejected", func(t *testing.T) {
			_, err, code := c.EndThingDeviceLink(bg, userjwt, thing.Id, open.Id, client.ThingDeviceLinkEnd{})
			expectCode(t, err, code, http.StatusBadRequest)
		})
	})

	t.Run("list things by device ids", func(t *testing.T) {
		list, total, err, _ := c.ListThings(bg, userjwt, client.ThingListOptions{DeviceIds: []string{deviceIds[0]}})
		if err != nil {
			t.Fatal(err)
		}
		if total != 1 || len(list) != 1 || list[0].Id != thing.Id {
			t.Fatalf("unexpected list: total=%v %#v", total, list)
		}
		list, total, err, _ = c.ListThings(bg, userjwt, client.ThingListOptions{DeviceIds: []string{deviceIds[2]}})
		if err != nil {
			t.Fatal(err)
		}
		if total != 0 || len(list) != 0 {
			t.Fatalf("unexpected list: total=%v %#v", total, list)
		}
	})

	t.Run("graph with thing and linked device is rejected", func(t *testing.T) {
		//an ended link counts as well
		_, err, code := c.SetGraph(bg, userjwt, models.Graph{
			Nodes: []models.Node{
				{Id: "1"},
				{Id: "2", ResourceType: models.GraphResourceTypeThing, ResourceId: thing.Id},
				{Id: "3", ResourceType: client.GraphResourceTypeDevice, ResourceId: deviceIds[0]},
				{Id: "4"},
			},
			Edges: []models.Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 50},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 50},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 100},
				{Id: "3->4", FromNodeId: "3", ToNodeId: "4", Weight: 100},
			},
		})
		expectCode(t, err, code, http.StatusBadRequest)
	})

	var graph models.Graph
	t.Run("graph with thing and unlinked device", func(t *testing.T) {
		graph, err, _ = c.SetGraph(bg, userjwt, models.Graph{
			Nodes: []models.Node{
				{Id: "1"},
				{Id: "2", ResourceType: models.GraphResourceTypeThing, ResourceId: thing.Id},
				{Id: "3", ResourceType: client.GraphResourceTypeDevice, ResourceId: deviceIds[1]},
				{Id: "4"},
			},
			Edges: []models.Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 50},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 50},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 100},
				{Id: "3->4", FromNodeId: "3", ToNodeId: "4", Weight: 100},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("linking a device of the graph is rejected", func(t *testing.T) {
		_, err, code := c.AddThingDeviceLink(bg, userjwt, thing.Id, client.ThingDeviceLinkCreate{DeviceId: deviceIds[1]})
		expectCode(t, err, code, http.StatusBadRequest)
	})

	t.Run("delete thing removes it from graphs", func(t *testing.T) {
		err, _ = c.DeleteThing(bg, userjwt, thing.Id)
		if err != nil {
			t.Fatal(err)
		}
		//the permissions go with the thing, so the former owner gets 403 like for a deleted graph
		_, err, code := c.ReadThing(bg, userjwt, thing.Id)
		expectCode(t, err, code, http.StatusForbidden)
		//an admin lists without a permission filter, so an empty answer shows the document is gone
		remaining, _, err, code := c.ListThings(bg, client.InternalAdminToken, client.ThingListOptions{Search: thing.Id, Limit: 10})
		expectCode(t, err, code, http.StatusOK)
		if len(remaining) != 0 {
			t.Fatalf("thing still stored: %#v", remaining)
		}

		//the graph is updated by the sync handler
		deadline := time.Now().Add(30 * time.Second)
		for {
			g, err, _ := c.ReadGraph(bg, userjwt, graph.Id)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, n := range g.Nodes {
				if n.ResourceType == models.GraphResourceTypeThing && n.ResourceId == thing.Id {
					found = true
				}
			}
			if !found {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("thing node still in graph: %#v", g.Nodes)
			}
			time.Sleep(500 * time.Millisecond)
		}
	})
}
