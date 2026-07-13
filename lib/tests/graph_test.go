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
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

func TestGraphUpdateOnDeviceDelete(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const userid = "testOwner"
	const userjwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.SyncLockDuration = time.Second.String()
	config.Debug = true

	config, err = docker.NewEnv(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	c := client.NewClient("http://localhost:"+config.ServerPort, nil)

	var protocol models.Protocol

	t.Run("create protocols", func(t *testing.T) {
		protocol, err, _ = c.SetProtocol(client.InternalAdminToken, models.Protocol{
			Name:    "p1",
			Handler: "p1",
			ProtocolSegments: []models.ProtocolSegment{{
				Id:   "ps1",
				Name: "ps1",
			}},
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	var deviceType models.DeviceType
	t.Run("create device-types", func(t *testing.T) {
		deviceType, err, _ = c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Name: "dt1",
			Services: []models.Service{
				{
					LocalId:     "s1",
					Name:        "s1",
					Interaction: models.REQUEST,
					ProtocolId:  protocol.Id,
				},
			},
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
	})

	deviceIds := []string{}
	t.Run("create devices", func(t *testing.T) {
		for i := range 10 {
			id := "d" + strconv.Itoa(i)
			d, err, _ := c.CreateDevice(userjwt, models.Device{
				LocalId:      id,
				Name:         id,
				DeviceTypeId: deviceType.Id,
			})
			if err != nil {
				t.Error(err)
				return
			}
			deviceIds = append(deviceIds, d.Id)
		}
	})

	var g1 models.Graph
	var g2 models.Graph
	var g3 models.Graph
	t.Run("create graphs", func(t *testing.T) {
		g1, err, _ = c.SetGraph(userjwt, models.Graph{
			Attributes: []models.Attribute{{Key: "name", Value: "g1"}},
			Nodes: []models.Node{
				{Id: "1"},
				{Id: "2", ResourceType: client.GraphResourceTypeDevice, ResourceId: deviceIds[0]},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []models.Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 80},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 20},
				{Id: "2->3", FromNodeId: "2", ToNodeId: "3", Weight: 50},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 50},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		if g1.Owner != userid {
			t.Error("unexpected owner: ", g1.Owner)
		}
		g2, err, _ = c.SetGraph(userjwt, models.Graph{
			Attributes: []models.Attribute{{Key: "name", Value: "g2"}},
			Nodes: []models.Node{
				{Id: "1"},
				{Id: "2", ResourceType: client.GraphResourceTypeDevice, ResourceId: deviceIds[0]},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []models.Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 80},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 20},
				{Id: "2->3", FromNodeId: "2", ToNodeId: "3", Weight: 50},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 50},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		if g2.Owner != userid {
			t.Error("unexpected owner: ", g2.Owner)
		}
		g3, err, _ = c.SetGraph(userjwt, models.Graph{
			Attributes: []models.Attribute{{Key: "name", Value: "g3"}},
			Nodes: []models.Node{
				{Id: "1"},
				{Id: "2", ResourceType: client.GraphResourceTypeDevice, ResourceId: deviceIds[1]},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []models.Edge{
				{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 80},
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 20},
				{Id: "2->3", FromNodeId: "2", ToNodeId: "3", Weight: 50},
				{Id: "2->4", FromNodeId: "2", ToNodeId: "4", Weight: 50},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		if g3.Owner != userid {
			t.Error("unexpected owner: ", g3.Owner)
		}
	})

	t.Run("list graphs", func(t *testing.T) {
		t.Run("no params", func(t *testing.T) {
			list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 3 {
				t.Error("unexpected total: ", total)
			}
			if len(list) != 3 {
				t.Error("unexpected list: ", list)
			}
			for _, g := range list {
				if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
					t.Error("unexpected attributes: ", g.Attributes)
					return
				}
			}
			slices.SortFunc(list, func(a, b models.Graph) int {
				return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
			})
			if !reflect.DeepEqual(list, []models.Graph{g1, g2, g3}) {
				t.Error("unexpected list: ", list)
			}
		})
		t.Run("limit offset", func(t *testing.T) {
			all, _, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{SortBy: "id.asc"})
			if err != nil {
				t.Error(err)
				return
			}
			if len(all) != 3 {
				t.Error("unexpected list: ", all)
				return
			}

			limited, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{SortBy: "id.asc", Limit: 1, Offset: 1})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 3 {
				t.Error("unexpected total: ", total)
			}
			if len(limited) != 1 {
				t.Error("unexpected list: ", limited)
				return
			}
			if !reflect.DeepEqual(limited, []models.Graph{all[1]}) {
				t.Error("unexpected list: ", limited)
				return
			}
		})
		t.Run("by ids", func(t *testing.T) {
			list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{Ids: []string{g2.Id, g3.Id}})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
			}
			if len(list) != 2 {
				t.Error("unexpected list: ", list)
			}
			for _, g := range list {
				if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
					t.Error("unexpected attributes: ", g.Attributes)
					return
				}
			}
			slices.SortFunc(list, func(a, b models.Graph) int {
				return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
			})
			if !reflect.DeepEqual(list, []models.Graph{g2, g3}) {
				t.Error("unexpected list: ", list)
			}
		})
		t.Run("by device ids 0", func(t *testing.T) {
			list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{DeviceIds: []string{deviceIds[0]}})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
			}
			if len(list) != 2 {
				t.Error("unexpected list: ", list)
			}
			for _, g := range list {
				if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
					t.Error("unexpected attributes: ", g.Attributes)
					return
				}
			}
			slices.SortFunc(list, func(a, b models.Graph) int {
				return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
			})
			if !reflect.DeepEqual(list, []models.Graph{g1, g2}) {
				t.Error("unexpected list: ", list)
			}
		})
		t.Run("by device ids 0,1", func(t *testing.T) {
			list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{DeviceIds: []string{deviceIds[0], deviceIds[1]}})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 3 {
				t.Error("unexpected total: ", total)
			}
			if len(list) != 3 {
				t.Error("unexpected list: ", list)
			}
			for _, g := range list {
				if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
					t.Error("unexpected attributes: ", g.Attributes)
					return
				}
			}
			slices.SortFunc(list, func(a, b models.Graph) int {
				return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
			})
			if !reflect.DeepEqual(list, []models.Graph{g1, g2, g3}) {
				t.Error("unexpected list: ", list)
			}
		})
		t.Run("by device ids 0,2", func(t *testing.T) {
			list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{DeviceIds: []string{deviceIds[0], deviceIds[2]}})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
			}
			if len(list) != 2 {
				t.Error("unexpected list: ", list)
			}
			for _, g := range list {
				if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
					t.Error("unexpected attributes: ", g.Attributes)
					return
				}
			}
			slices.SortFunc(list, func(a, b models.Graph) int {
				return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
			})
			if !reflect.DeepEqual(list, []models.Graph{g1, g2}) {
				t.Error("unexpected list: ", list)
			}
		})
		t.Run("by device ids 2", func(t *testing.T) {
			//device is not assigned to any graph
			list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{DeviceIds: []string{deviceIds[2]}})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 0 {
				t.Error("unexpected total: ", total)
			}
			if len(list) != 0 {
				t.Error("unexpected list: ", list)
			}
		})
		t.Run("by attr name", func(t *testing.T) {
			list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{Attributes: []models.Attribute{{Key: "name"}}})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 3 {
				t.Error("unexpected total: ", total)
			}
			if len(list) != 3 {
				t.Error("unexpected list: ", list)
			}
			for _, g := range list {
				if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
					t.Error("unexpected attributes: ", g.Attributes)
					return
				}
			}
			slices.SortFunc(list, func(a, b models.Graph) int {
				return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
			})
			if !reflect.DeepEqual(list, []models.Graph{g1, g2, g3}) {
				t.Error("unexpected list: ", list)
			}
		})
		t.Run("by attr name:g2", func(t *testing.T) {
			list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{Attributes: []models.Attribute{{Key: "name", Value: "g2"}}})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 1 {
				t.Error("unexpected total: ", total)
			}
			if len(list) != 1 {
				t.Error("unexpected list: ", list)
			}
			for _, g := range list {
				if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
					t.Error("unexpected attributes: ", g.Attributes)
					return
				}
			}
			slices.SortFunc(list, func(a, b models.Graph) int {
				return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
			})
			if !reflect.DeepEqual(list, []models.Graph{g2}) {
				t.Error("unexpected list: ", list)
			}
		})
		t.Run("by attr search g2", func(t *testing.T) {
			list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{Search: "g2"})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 1 {
				t.Error("unexpected total: ", total)
			}
			if len(list) != 1 {
				t.Error("unexpected list: ", list)
			}
			for _, g := range list {
				if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
					t.Error("unexpected attributes: ", g.Attributes)
					return
				}
			}
			slices.SortFunc(list, func(a, b models.Graph) int {
				return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
			})
			if !reflect.DeepEqual(list, []models.Graph{g2}) {
				t.Error("unexpected list: ", list)
			}
		})
	})

	t.Run("get graph", func(t *testing.T) {
		result, err, _ := c.ReadGraph(userjwt, g1.Id)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(result, g1) {
			t.Error("unexpected result: ", result)
		}
	})

	temp := models.GraphIdProvider
	defer func() {
		models.GraphIdProvider = temp
	}()
	idProviderValue := 0
	models.GraphIdProvider = func() string {
		idProviderValue++
		return fmt.Sprintf("id-%v", idProviderValue)
	}

	t.Run("delete device", func(t *testing.T) {
		err, _ = c.DeleteDevice(userjwt, deviceIds[0])
		if err != nil {
			t.Error(err)
			return
		}
		err, _ = c.DeleteDevice(userjwt, deviceIds[2]) //device is not assigned to any graph
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check graphs after device delete", func(t *testing.T) {
		g1 = models.Graph{
			Id:         g1.Id,
			Owner:      userid,
			Attributes: []models.Attribute{{Key: "name", Value: "g1"}},
			Nodes: []models.Node{
				{Id: "1"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []models.Edge{
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 60, Attributes: []models.Attribute{{Key: models.GraphEdgeAttrSystemChanged, Value: "true"}}},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
				{Id: "id-2", FromNodeId: "1", ToNodeId: "4", Weight: 40, Attributes: []models.Attribute{{Key: models.GraphEdgeAttrSystemChanged, Value: "true"}}},
			},
		}
		g2 = models.Graph{
			Id:         g2.Id,
			Owner:      userid,
			Attributes: []models.Attribute{{Key: "name", Value: "g2"}},
			Nodes: []models.Node{
				{Id: "1"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []models.Edge{
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 60, Attributes: []models.Attribute{{Key: models.GraphEdgeAttrSystemChanged, Value: "true"}}},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
				{Id: "id-4", FromNodeId: "1", ToNodeId: "4", Weight: 40, Attributes: []models.Attribute{{Key: models.GraphEdgeAttrSystemChanged, Value: "true"}}},
			},
		}
		g1v2 := models.Graph{
			Id:         g1.Id,
			Owner:      userid,
			Attributes: []models.Attribute{{Key: "name", Value: "g1"}},
			Nodes: []models.Node{
				{Id: "1"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []models.Edge{
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 60, Attributes: []models.Attribute{{Key: models.GraphEdgeAttrSystemChanged, Value: "true"}}},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
				{Id: "id-4", FromNodeId: "1", ToNodeId: "4", Weight: 40, Attributes: []models.Attribute{{Key: models.GraphEdgeAttrSystemChanged, Value: "true"}}},
			},
		}
		g2v2 := models.Graph{
			Id:         g2.Id,
			Owner:      userid,
			Attributes: []models.Attribute{{Key: "name", Value: "g2"}},
			Nodes: []models.Node{
				{Id: "1"},
				{Id: "3"},
				{Id: "4"},
				{Id: "5"},
			},
			Edges: []models.Edge{
				{Id: "1->3", FromNodeId: "1", ToNodeId: "3", Weight: 60, Attributes: []models.Attribute{{Key: models.GraphEdgeAttrSystemChanged, Value: "true"}}},
				{Id: "3->5", FromNodeId: "3", ToNodeId: "5", Weight: 100},
				{Id: "4->5", FromNodeId: "4", ToNodeId: "5", Weight: 100},
				{Id: "id-2", FromNodeId: "1", ToNodeId: "4", Weight: 40, Attributes: []models.Attribute{{Key: models.GraphEdgeAttrSystemChanged, Value: "true"}}},
			},
		}
		list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 3 {
			t.Error("unexpected total: ", total)
		}
		if len(list) != 3 {
			t.Error("unexpected list: ", list)
		}
		for _, g := range list {
			if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
				t.Error("unexpected attributes: ", g.Attributes)
				return
			}
		}
		slices.SortFunc(list, func(a, b models.Graph) int {
			return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
		})
		if !reflect.DeepEqual(list, []models.Graph{g1, g2, g3}) {
			//generated id order is not deterministic --> try the other way around
			g1 = g1v2
			g2 = g2v2
			if !reflect.DeepEqual(list, []models.Graph{g1, g2, g3}) {
				t.Errorf("unexpected list: \na=%#v\ne=%#v\n", list, []models.Graph{g1, g2, g3})
			}
		}
	})

	t.Run("delete graph", func(t *testing.T) {
		err, _ = c.DeleteGraph(userjwt, g2.Id)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check graphs after graph delete", func(t *testing.T) {
		list, total, err, _ := c.ListGraphs(userjwt, client.GraphListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 2 {
			t.Error("unexpected total: ", total)
		}
		if len(list) != 2 {
			t.Error("unexpected list: ", list)
		}
		for _, g := range list {
			if len(g.Attributes) != 1 || g.Attributes[0].Key != "name" {
				t.Error("unexpected attributes: ", g.Attributes)
				return
			}
		}
		slices.SortFunc(list, func(a, b models.Graph) int {
			return strings.Compare(a.Attributes[0].Value, b.Attributes[0].Value)
		})
		if !reflect.DeepEqual(list, []models.Graph{g1, g3}) {
			t.Errorf("unexpected list: \na=%#v\ne=%#v\n", list, []models.Graph{g1, g3})
		}
	})
}
