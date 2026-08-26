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

package lib

import (
	"context"
	"slices"
	"testing"

	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/database/testdb"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

// TestSyncPermResourcesGraphs checks that graph permissions without a graph in the local db are removed
func TestSyncPermResourcesGraphs(t *testing.T) {
	ctx := context.Background()
	config := configuration.Config{
		DeviceTopic:      "devices",
		DeviceGroupTopic: "device-groups",
		HubTopic:         "hubs",
		LocationTopic:    "locations",
		GraphTopic:       "graphs",
	}

	permClient, err := client.NewTestClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, topic := range []string{config.DeviceTopic, config.DeviceGroupTopic, config.HubTopic, config.LocationTopic, config.GraphTopic} {
		if _, err, _ = permClient.SetTopic(client.InternalAdminToken, client.Topic{Id: topic}); err != nil {
			t.Fatal(err)
		}
	}

	db := testdb.NewTestDB(config)
	err = db.SetGraph(ctx, models.Graph{
		Id:    "g1",
		Owner: "u1",
		Nodes: []models.Node{{Id: "1"}, {Id: "2"}},
		Edges: []models.Edge{{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 100}},
	}, func(graph models.Graph) error { return nil })
	if err != nil {
		t.Fatal(err)
	}

	//g2 exists in permissions-v2 but not in the local db
	permissions := client.ResourcePermissions{
		UserPermissions: map[string]client.PermissionsMap{
			"u1": {Read: true, Write: true, Execute: true, Administrate: true},
		},
	}
	for _, id := range []string{"g1", "g2"} {
		if _, err, _ = permClient.SetPermission(client.InternalAdminToken, config.GraphTopic, id, permissions); err != nil {
			t.Fatal(err)
		}
	}

	err = SyncPermResources(ctx, config, permClient, db)
	if err != nil {
		t.Fatal(err)
	}

	ids, err, _ := permClient.AdminListResourceIds(client.InternalAdminToken, config.GraphTopic, client.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(ids, []string{"g1"}) {
		t.Errorf("unexpected graph permission resources: %#v", ids)
	}
}
