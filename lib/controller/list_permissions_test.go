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
	"errors"
	"net/http"
	"testing"

	"github.com/SENERGY-Platform/device-repository/v3/lib/database"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

// TestListsWithoutReadableIdsAreEmpty checks that a non-admin caller who may read nothing gets nothing, also when the
// list is asked for without ids. The permission lookup then answers nil, and nil reaching the database means "no id filter".
func TestListsWithoutReadableIdsAreEmpty(t *testing.T) {
	ctrl, permClient, _ := newThingTestController(t)
	ctrl.db = &idFilterCheckingDB{Database: ctrl.db}
	ctx := context.Background()
	other := thingTestOtherUserJwt //a non-admin user without any rights

	thingTestDevice(t, ctrl, permClient, "d1", "")
	thing := thingTestCreate(t, ctrl, "thing")
	graph, err, code := ctrl.SetGraph(ctx, graphTestUserJwt, models.Graph{
		Attributes: []models.Attribute{{Key: "name", Value: "g"}},
		Nodes:      []models.Node{{Id: "1"}, {Id: "2"}},
		Edges:      []models.Edge{{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 100}},
	})
	expectCode(t, err, code, http.StatusOK)
	//one element in every collection that the other user may not read
	if err = ctrl.db.SetHub(ctx, model.HubWithConnectionState{Hub: models.Hub{Id: "h1", Name: "h", OwnerId: graphTestUserId}}, func(model.HubWithConnectionState) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err = ctrl.db.SetDeviceGroup(ctx, models.DeviceGroup{Id: "g1", Name: "g"}, func(models.DeviceGroup, string) error { return nil }, graphTestUserId); err != nil {
		t.Fatal(err)
	}
	if err = ctrl.db.SetLocation(ctx, models.Location{Id: "l1", Name: "l"}, func(models.Location, string) error { return nil }, graphTestUserId); err != nil {
		t.Fatal(err)
	}
	for name, list := range map[string]func(ids []string) (int, error){
		"things": func(ids []string) (int, error) {
			r, _, err, _ := ctrl.ListThings(ctx, other, model.ThingListOptions{Ids: ids})
			return len(r), err
		},
		"graphs": func(ids []string) (int, error) {
			r, _, err, _ := ctrl.ListGraphs(ctx, other, model.GraphListOptions{Ids: ids})
			return len(r), err
		},
		"devices": func(ids []string) (int, error) {
			r, err, _ := ctrl.ListDevices(ctx, other, model.DeviceListOptions{Ids: ids, Permission: models.Read})
			return len(r), err
		},
		"extended devices": func(ids []string) (int, error) {
			r, _, err, _ := ctrl.ListExtendedDevices(ctx, other, model.ExtendedDeviceListOptions{Ids: ids, Permission: models.Read})
			return len(r), err
		},
		"device-groups": func(ids []string) (int, error) {
			r, _, err, _ := ctrl.ListDeviceGroups(ctx, other, model.DeviceGroupListOptions{Ids: ids})
			return len(r), err
		},
		"hubs": func(ids []string) (int, error) {
			r, err, _ := ctrl.ListHubs(ctx, other, model.HubListOptions{Ids: ids})
			return len(r), err
		},
		"extended hubs": func(ids []string) (int, error) {
			r, _, err, _ := ctrl.ListExtendedHubs(ctx, other, model.HubListOptions{Ids: ids})
			return len(r), err
		},
		"locations": func(ids []string) (int, error) {
			r, _, err, _ := ctrl.ListLocations(ctx, other, model.LocationListOptions{Ids: ids})
			return len(r), err
		},
	} {
		t.Run(name, func(t *testing.T) {
			for label, ids := range map[string][]string{
				"explicit ids, none readable": {thing.Id, graph.Id, "d1", "h1", "g1", "l1"},
				"explicit ids, none exist":    {"nope", "nope2"},
				"empty explicit ids":          {},
				"no ids, nothing readable":    nil,
			} {
				n, err := list(ids)
				if err != nil || n != 0 {
					t.Errorf("%v: %v elements, err %v", label, n, err)
				}
			}
		})
	}
}

// idFilterCheckingDB stands for mongo: a query without id filter (Ids == nil) answers everything, a query with an empty
// id filter answers nothing. testdb ignores the ids of some lists, so the test database alone can not show the difference.
// The test calls only with a non-admin token, so every list that reaches the database must carry an id filter.
type idFilterCheckingDB struct {
	database.Database
}

var errUnfilteredQuery = errors.New("query without id filter for a non-admin caller")

func (this *idFilterCheckingDB) ListThings(ctx context.Context, o model.ThingListOptions) ([]models.Thing, int64, error) {
	if o.Ids == nil {
		return nil, 0, errUnfilteredQuery
	}
	if len(o.Ids) == 0 {
		return nil, 0, nil
	}
	return this.Database.ListThings(ctx, o)
}

func (this *idFilterCheckingDB) ListGraphs(ctx context.Context, o model.GraphListOptions) ([]models.Graph, int64, error) {
	if o.Ids == nil {
		return nil, 0, errUnfilteredQuery
	}
	if len(o.Ids) == 0 {
		return nil, 0, nil
	}
	return this.Database.ListGraphs(ctx, o)
}

func (this *idFilterCheckingDB) ListDevices(ctx context.Context, o model.DeviceListOptions, withTotal bool) ([]model.DeviceWithConnectionState, int64, error) {
	if o.Ids == nil {
		return nil, 0, errUnfilteredQuery
	}
	if len(o.Ids) == 0 {
		return nil, 0, nil
	}
	return this.Database.ListDevices(ctx, o, withTotal)
}

func (this *idFilterCheckingDB) ListDeviceGroups(ctx context.Context, o model.DeviceGroupListOptions) ([]models.DeviceGroup, int64, error) {
	if o.Ids == nil {
		return nil, 0, errUnfilteredQuery
	}
	if len(o.Ids) == 0 {
		return nil, 0, nil
	}
	return this.Database.ListDeviceGroups(ctx, o)
}

func (this *idFilterCheckingDB) ListHubs(ctx context.Context, o model.HubListOptions, withTotal bool) ([]model.HubWithConnectionState, int64, error) {
	if o.Ids == nil {
		return nil, 0, errUnfilteredQuery
	}
	if len(o.Ids) == 0 {
		return nil, 0, nil
	}
	return this.Database.ListHubs(ctx, o, withTotal)
}

func (this *idFilterCheckingDB) ListLocations(ctx context.Context, o model.LocationListOptions) ([]models.Location, int64, error) {
	if o.Ids == nil {
		return nil, 0, errUnfilteredQuery
	}
	if len(o.Ids) == 0 {
		return nil, 0, nil
	}
	return this.Database.ListLocations(ctx, o)
}
