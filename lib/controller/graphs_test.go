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
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller/publisher"
	"github.com/SENERGY-Platform/device-repository/lib/database/testdb"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

const graphTestUserId = "testOwner"
const graphTestUserJwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`

// graphPublisherMock records the graph commands; everything else is void
type graphPublisherMock struct {
	publisher.Void
	published []models.Graph
	deleted   []string
}

func (this *graphPublisherMock) PublishGraph(graph models.Graph) error {
	this.published = append(this.published, graph)
	return nil
}

func (this *graphPublisherMock) PublishGraphDelete(id string) error {
	this.deleted = append(this.deleted, id)
	return nil
}

func newGraphTestController(t *testing.T) (ctrl *Controller, permClient client.Client, p *graphPublisherMock) {
	t.Helper()
	conf := configuration.Config{
		DeviceTopic:           "devices",
		DeviceGroupTopic:      "device-groups",
		HubTopic:              "hubs",
		LocationTopic:         "locations",
		GraphTopic:            "graphs",
		InitPermissionsTopics: true,
	}
	permClient, err := client.NewTestClient(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	p = &graphPublisherMock{}
	ctrl, err = New(conf, testdb.NewTestDB(conf), p, permClient)
	if err != nil {
		t.Fatal(err)
	}
	return ctrl, permClient, p
}

func TestGraphTopicRegistration(t *testing.T) {
	_, permClient, _ := newGraphTestController(t)
	topic, err, _ := permClient.GetTopic(client.InternalAdminToken, "graphs")
	if err != nil {
		t.Fatal(err)
	}
	if topic.PublishToKafkaTopic != "graphs" {
		t.Error("unexpected publish_to_kafka_topic: ", topic.PublishToKafkaTopic)
	}
}

func TestGraphPublishing(t *testing.T) {
	ctrl, _, p := newGraphTestController(t)

	graph := models.Graph{
		Attributes: []models.Attribute{{Key: "name", Value: "g1"}},
		Nodes:      []models.Node{{Id: "1"}, {Id: "2"}},
		Edges:      []models.Edge{{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 100}},
	}

	graph, err, _ := ctrl.SetGraph(graphTestUserJwt, graph)
	if err != nil {
		t.Fatal(err)
	}
	if graph.Owner != graphTestUserId {
		t.Fatal("unexpected owner: ", graph.Owner)
	}
	if !reflect.DeepEqual(p.published, []models.Graph{graph}) {
		t.Errorf("create was not published:\n%#v\n%#v", p.published, []models.Graph{graph})
	}

	graph.Attributes = []models.Attribute{{Key: "name", Value: "g1-renamed"}}
	graph, err, _ = ctrl.SetGraph(graphTestUserJwt, graph)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.published) != 2 || !reflect.DeepEqual(p.published[1], graph) {
		t.Errorf("update was not published:\n%#v\n%#v", p.published, graph)
	}

	err, _ = ctrl.DeleteGraph(client.InternalAdminToken, graph.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p.deleted, []string{graph.Id}) {
		t.Errorf("delete was not published: %#v", p.deleted)
	}
}
