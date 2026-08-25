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
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	model2 "github.com/SENERGY-Platform/permissions-v2/pkg/model"
)

func TestPermSync(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.SyncLockDuration = "1ms"
	config.Debug = true
	config.DisableStrictValidationForTesting = true
	config.RestLogger()

	config, err = docker.NewEnv(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	db, err := database.New(config)
	if err != nil {
		t.Error(err)
		return
	}
	permClient := client.New(config.PermissionsV2Url)

	_, err, _ = permClient.SetTopic(client.InternalAdminToken, client.Topic{
		Id: config.DeviceTopic,
	})
	if err != nil {
		t.Error(err)
		return
	}
	_, err, _ = permClient.SetTopic(client.InternalAdminToken, client.Topic{
		Id: config.LocationTopic,
	})
	if err != nil {
		t.Error(err)
		return
	}
	_, err, _ = permClient.SetTopic(client.InternalAdminToken, client.Topic{
		Id: config.DeviceGroupTopic,
	})
	if err != nil {
		t.Error(err)
		return
	}
	_, err, _ = permClient.SetTopic(client.InternalAdminToken, client.Topic{
		Id: config.HubTopic,
	})
	if err != nil {
		t.Error(err)
		return
	}
	_, err, _ = permClient.SetTopic(client.InternalAdminToken, client.Topic{
		Id: config.GraphTopic,
	})
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("init state", func(t *testing.T) {

		getPermission := func(user string) client.ResourcePermissions {
			return client.ResourcePermissions{
				UserPermissions: map[string]model2.PermissionsMap{
					user: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupPermissions: map[string]model2.PermissionsMap{
					"admin": {Read: true, Write: true, Execute: true, Administrate: true},
				},
				RolePermissions: map[string]model2.PermissionsMap{
					"admin": {Read: true, Write: true, Execute: true, Administrate: true},
				},
			}
		}

		t.Run("create", func(t *testing.T) {
			err = db.SetDevice(ctx, model.DeviceWithConnectionState{
				Device: models.Device{
					Id:           "d1",
					LocalId:      "d1",
					Name:         "d1",
					DeviceTypeId: "dt1",
					OwnerId:      "u1",
				},
			}, func(old model.DeviceWithConnectionState, new model.DeviceWithConnectionState) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.DeviceTopic, new.Id, getPermission(new.OwnerId))
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
			err = db.SetDevice(ctx, model.DeviceWithConnectionState{
				Device: models.Device{
					Id:           "d2",
					LocalId:      "d2",
					Name:         "d2",
					DeviceTypeId: "dt1",
					OwnerId:      "u1",
				},
			}, func(old model.DeviceWithConnectionState, n model.DeviceWithConnectionState) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.DeviceTopic, n.Id, getPermission(n.OwnerId))
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
			err = db.SetDevice(ctx, model.DeviceWithConnectionState{
				Device: models.Device{
					Id:           "d3",
					LocalId:      "d3",
					Name:         "d3",
					DeviceTypeId: "dt1",
					OwnerId:      "u1",
				},
			}, func(old model.DeviceWithConnectionState, n model.DeviceWithConnectionState) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.DeviceTopic, n.Id, getPermission(n.OwnerId))
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = db.SetHub(ctx, model.HubWithConnectionState{
				Hub: models.Hub{
					Id:      "h1",
					Name:    "h1",
					OwnerId: "u1",
				},
			}, func(state model.HubWithConnectionState) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.HubTopic, state.Id, getPermission(state.OwnerId))
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = db.SetHub(ctx, model.HubWithConnectionState{
				Hub: models.Hub{
					Id:      "h2",
					Name:    "h2",
					OwnerId: "u2",
				},
			}, func(state model.HubWithConnectionState) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.HubTopic, state.Id, getPermission(state.OwnerId))
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
			err = db.SetHub(ctx, model.HubWithConnectionState{
				Hub: models.Hub{
					Id:      "h3",
					Name:    "h3",
					OwnerId: "u2",
				},
			}, func(state model.HubWithConnectionState) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.HubTopic, state.Id, getPermission(state.OwnerId))
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = db.SetDeviceGroup(ctx, models.DeviceGroup{
				Id:   "dg1",
				Name: "dg1",
			}, func(dg models.DeviceGroup, user string) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.DeviceGroupTopic, dg.Id, getPermission(user))
				return err
			}, "u1")
			if err != nil {
				t.Error(err)
				return
			}

			err = db.SetDeviceGroup(ctx, models.DeviceGroup{
				Id:   "dg2",
				Name: "dg2",
			}, func(dg models.DeviceGroup, user string) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.DeviceGroupTopic, dg.Id, getPermission(user))
				return err
			}, "u1")
			if err != nil {
				t.Error(err)
				return
			}
			err = db.SetDeviceGroup(ctx, models.DeviceGroup{
				Id:   "dg3",
				Name: "dg3",
			}, func(dg models.DeviceGroup, user string) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.DeviceGroupTopic, dg.Id, getPermission(user))
				return err
			}, "u1")
			if err != nil {
				t.Error(err)
				return
			}

			err = db.SetLocation(ctx, models.Location{
				Id:   "l1",
				Name: "l1",
			}, func(l models.Location, user string) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.LocationTopic, l.Id, getPermission(user))
				return err
			}, "u1")
			if err != nil {
				t.Error(err)
				return
			}
			err = db.SetLocation(ctx, models.Location{
				Id:   "l2",
				Name: "l2",
			}, func(l models.Location, user string) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.LocationTopic, l.Id, getPermission(user))
				return err
			}, "u1")
			if err != nil {
				t.Error(err)
				return
			}
			err = db.SetLocation(ctx, models.Location{
				Id:   "l3",
				Name: "l3",
			}, func(l models.Location, user string) error {
				_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.LocationTopic, l.Id, getPermission(user))
				return err
			}, "u1")
			if err != nil {
				t.Error(err)
				return
			}

			for _, id := range []string{"g1", "g2", "g3"} {
				err = db.SetGraph(ctx, models.Graph{
					Id:    id,
					Owner: "u1",
					Nodes: []models.Node{{Id: "1"}, {Id: "2"}},
					Edges: []models.Edge{{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 100}},
				}, func(g models.Graph) error {
					_, err, _ := permClient.SetPermission(client.InternalAdminToken, config.GraphTopic, g.Id, getPermission(g.Owner))
					return err
				})
				if err != nil {
					t.Error(err)
					return
				}
			}
		})

		t.Run("remove some from perm", func(t *testing.T) {
			err, _ := permClient.RemoveResource(client.InternalAdminToken, config.DeviceTopic, "d2")
			if err != nil {
				t.Error(err)
				return
			}
			err, _ = permClient.RemoveResource(client.InternalAdminToken, config.HubTopic, "h2")
			if err != nil {
				t.Error(err)
				return
			}
			err, _ = permClient.RemoveResource(client.InternalAdminToken, config.DeviceGroupTopic, "dg2")
			if err != nil {
				t.Error(err)
				return
			}
			err, _ = permClient.RemoveResource(client.InternalAdminToken, config.LocationTopic, "l2")
			if err != nil {
				t.Error(err)
				return
			}
			err, _ = permClient.RemoveResource(client.InternalAdminToken, config.GraphTopic, "g2")
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("remove some from db", func(t *testing.T) {
			err = db.RemoveDevice(ctx, "d3", func(state model.DeviceWithConnectionState) error { return nil })
			if err != nil {
				t.Error(err)
				return
			}
			err = db.RemoveHub(ctx, "h3", func(state model.HubWithConnectionState) error { return nil })
			if err != nil {
				t.Error(err)
				return
			}
			err = db.RemoveDeviceGroup(ctx, "dg3", func(state models.DeviceGroup) error { return nil })
			if err != nil {
				t.Error(err)
				return
			}
			err = db.RemoveLocation(ctx, "l3", func(state models.Location) error { return nil })
			if err != nil {
				t.Error(err)
				return
			}
			err = db.RemoveGraph(ctx, "g3", func(g models.Graph) error { return nil })
			if err != nil {
				t.Error(err)
				return
			}
		})

	})

	time.Sleep(1 * time.Second)
	t.Run("start lib to sync", func(t *testing.T) {
		err = lib.Start(ctx, wg, config)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check db", func(t *testing.T) {
		t.Run("devices", func(t *testing.T) {
			devices, total, err := db.ListDevices(ctx, model.DeviceListOptions{SortBy: "id.asc"}, true)
			if err != nil {
				t.Error(err)
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
				return
			}
			if len(devices) != 2 {
				t.Error("unexpected devices: ", len(devices))
				return
			}
			if devices[0].Id != "d1" {
				t.Error("unexpected device: ", devices[0].Id)
			}
			if devices[1].Id != "d2" {
				t.Error("unexpected device: ", devices[1].Id)
			}
		})

		t.Run("hubs", func(t *testing.T) {
			hubs, total, err := db.ListHubs(ctx, model.HubListOptions{SortBy: "id.asc"}, true)
			if err != nil {
				t.Error(err)
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
				return
			}
			if len(hubs) != 2 {
				t.Error("unexpected hubs: ", len(hubs))
				return
			}
			if hubs[0].Id != "h1" {
				t.Error("unexpected hub: ", hubs[0].Id)
			}
			if hubs[1].Id != "h2" {
				t.Error("unexpected hub: ", hubs[1].Id)
			}
		})

		t.Run("device-groups", func(t *testing.T) {
			dgs, total, err := db.ListDeviceGroups(ctx, model.DeviceGroupListOptions{SortBy: "id.asc"})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 4 {
				t.Error("unexpected total: ", total)
				return
			}
			if len(dgs) != 4 {
				t.Error("unexpected device-groups: ", len(dgs))
				return
			}
			if dgs[0].Id != "dg1" {
				t.Error("unexpected device-group: ", dgs[0].Id)
			}
			if dgs[1].Id != "dg2" {
				t.Error("unexpected device-group: ", dgs[1].Id)
			}
			if dgs[2].Id != "urn:infai:ses:device-group:d1" {
				t.Error("unexpected device-group: ", dgs[2].Id)
			}
			if dgs[3].Id != "urn:infai:ses:device-group:d2" {
				t.Error("unexpected device-group: ", dgs[3].Id)
			}
		})

		t.Run("locations", func(t *testing.T) {
			locations, total, err := db.ListLocations(ctx, model.LocationListOptions{SortBy: "id.asc"})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
				return
			}
			if len(locations) != 2 {
				t.Error("unexpected locations: ", len(locations))
				return
			}
			if locations[0].Id != "l1" {
				t.Error("unexpected location: ", locations[0].Id)
			}
			if locations[1].Id != "l2" {
				t.Error("unexpected location: ", locations[1].Id)
			}
		})

		t.Run("graphs", func(t *testing.T) {
			graphs, total, err := db.ListGraphs(ctx, model.GraphListOptions{SortBy: "id.asc"})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
				return
			}
			if len(graphs) != 2 {
				t.Error("unexpected graphs: ", len(graphs))
				return
			}
			if graphs[0].Id != "g1" {
				t.Error("unexpected graph: ", graphs[0].Id)
			}
			if graphs[1].Id != "g2" {
				t.Error("unexpected graph: ", graphs[1].Id)
			}
		})

	})

	t.Run("check perm", func(t *testing.T) {
		t.Run("devices", func(t *testing.T) {
			ids, err, _ := permClient.AdminListResourceIds(client.InternalAdminToken, config.DeviceTopic, client.ListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(ids) != 2 {
				t.Error("unexpected ids: ", len(ids))
				return
			}
			if !slices.Contains(ids, "d1") {
				t.Error("id not found:", "d1")
			}
			if !slices.Contains(ids, "d2") {
				t.Error("id not found:", "d2")
			}
		})
		t.Run("hubs", func(t *testing.T) {
			ids, err, _ := permClient.AdminListResourceIds(client.InternalAdminToken, config.HubTopic, client.ListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(ids) != 2 {
				t.Errorf("unexpected ids: %#v", ids)
				return
			}
			if !slices.Contains(ids, "h1") {
				t.Error("id not found:", "h1")
			}
			if !slices.Contains(ids, "h2") {
				t.Error("id not found:", "h2")
			}
		})
		t.Run("device-groups", func(t *testing.T) {
			ids, err, _ := permClient.AdminListResourceIds(client.InternalAdminToken, config.DeviceGroupTopic, client.ListOptions{})
			if err != nil {
				t.Error(err)
			}
			if len(ids) != 4 {
				t.Error("unexpected ids: ", len(ids))
				return
			}
			if !slices.Contains(ids, "dg1") {
				t.Error("id not found:", "dg1")
			}
			if !slices.Contains(ids, "dg2") {
				t.Error("id not found:", "dg2")
			}
		})
		t.Run("locations", func(t *testing.T) {
			ids, err, _ := permClient.AdminListResourceIds(client.InternalAdminToken, config.LocationTopic, client.ListOptions{})
			if err != nil {
				t.Error(err)
			}
			if len(ids) != 2 {
				t.Error("unexpected ids: ", len(ids))
				return
			}
			if !slices.Contains(ids, "l1") {
				t.Error("id not found:", "l1")
			}
			if !slices.Contains(ids, "l2") {
				t.Error("id not found:", "l2")
			}
		})
		t.Run("graphs", func(t *testing.T) {
			ids, err, _ := permClient.AdminListResourceIds(client.InternalAdminToken, config.GraphTopic, client.ListOptions{})
			if err != nil {
				t.Error(err)
			}
			if len(ids) != 2 {
				t.Error("unexpected ids: ", len(ids))
				return
			}
			if !slices.Contains(ids, "g1") {
				t.Error("id not found:", "g1")
			}
			if !slices.Contains(ids, "g2") {
				t.Error("id not found:", "g2")
			}
		})
	})
}
