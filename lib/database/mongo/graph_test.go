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
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoGraphValidity(t *testing.T) {
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
	noSync := func(models.Graph) error { return nil }

	from := time.Date(2026, 11, 1, 10, 0, 0, 123*int(time.Millisecond), time.UTC)
	until := time.Date(2026, 12, 1, 10, 0, 0, 456*int(time.Millisecond), time.UTC)
	graph := models.Graph{
		Id:         "g",
		Owner:      "u1",
		Attributes: []models.Attribute{{Key: "k", Value: "v", ValidFrom: &from}},
		Nodes: []models.Node{
			{Id: "root", Attributes: []models.Attribute{}},
			{Id: "a", ValidFrom: &from, ValidUntil: &until, Attributes: []models.Attribute{{Key: "k", Value: "v", ValidUntil: &until}}},
		},
		Edges: []models.Edge{{Id: "a>root", FromNodeId: "a", ToNodeId: "root", Weight: 100, ValidFrom: &from}},
	}
	err = m.SetGraph(ctx, graph, noSync)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("the validities survive a round trip in UTC", func(t *testing.T) {
		stored, exists, err := m.GetGraph(ctx, "g")
		if err != nil || !exists {
			t.Fatal(err, exists)
		}
		if !stored.Nodes[1].ValidFrom.Equal(from) || !stored.Nodes[1].ValidUntil.Equal(until) || !stored.Edges[0].ValidFrom.Equal(from) || stored.Edges[0].ValidUntil != nil || stored.Nodes[0].ValidFrom != nil {
			t.Errorf("%#v", stored)
		}
		if !stored.Attributes[0].ValidFrom.Equal(from) || stored.Attributes[0].ValidUntil != nil || !stored.Nodes[1].Attributes[0].ValidUntil.Equal(until) {
			t.Errorf("%#v", stored.Attributes)
		}
	})

	t.Run("unset bounds are not written, old and new documents look alike", func(t *testing.T) {
		raw := bson.M{}
		err := m.graphCollection().FindOne(ctx, bson.M{"id": "g"}).Decode(&raw)
		if err != nil {
			t.Fatal(err)
		}
		nodes := raw["nodes"].(bson.A)
		root := nodes[0].(bson.M)
		if _, ok := root["validfrom"]; ok {
			t.Errorf("%#v", root)
		}
		if _, ok := root["validuntil"]; ok {
			t.Errorf("%#v", root)
		}
		a := nodes[1].(bson.M)
		if _, ok := a["validfrom"]; !ok {
			t.Errorf("%#v", a)
		}
	})

	t.Run("a document written before the fields existed reads as unbounded", func(t *testing.T) {
		_, err := m.graphCollection().InsertOne(ctx, bson.M{"id": "old", "owner": "u1", "attributes": bson.A{bson.M{"key": "k", "value": "v", "origin": ""}}, "nodes": bson.A{bson.M{"id": "n", "attributes": bson.A{}, "resourcetype": "", "resourceid": ""}}, "edges": bson.A{}})
		if err != nil {
			t.Fatal(err)
		}
		stored, exists, err := m.GetGraph(ctx, "old")
		if err != nil || !exists {
			t.Fatal(err, exists)
		}
		if stored.Nodes[0].ValidFrom != nil || stored.Attributes[0].ValidUntil != nil {
			t.Errorf("%#v", stored)
		}
		if len(stored.At(time.Now()).Nodes) != 1 {
			t.Error("an old document is valid always")
		}
	})

	t.Run("a write that carries a stale revision is not applied", func(t *testing.T) {
		cas := models.Graph{Id: "cas", Owner: "u1", Nodes: []models.Node{{Id: "n"}}}
		swapped, err := m.SetGraphIfRevision(ctx, cas, 0, noSync)
		if err != nil || !swapped {
			t.Fatal(swapped, err)
		}
		_, r1, exists, err := m.GetGraphWithRevision(ctx, "cas")
		if err != nil || !exists || r1 == 0 {
			t.Fatal(r1, exists, err)
		}
		cas.Attributes = []models.Attribute{{Key: "k", Value: "first"}}
		swapped, err = m.SetGraphIfRevision(ctx, cas, r1, noSync)
		if err != nil || !swapped {
			t.Fatal(swapped, err)
		}
		cas.Attributes = []models.Attribute{{Key: "k", Value: "stale"}}
		swapped, err = m.SetGraphIfRevision(ctx, cas, r1, noSync)
		if err != nil || swapped {
			t.Fatal("the stale write was applied", swapped, err)
		}
		//creating what exists is a lost race, not an error
		swapped, err = m.SetGraphIfRevision(ctx, cas, 0, noSync)
		if err != nil || swapped {
			t.Fatal("the second creation was applied", swapped, err)
		}
		stored, r2, _, _ := m.GetGraphWithRevision(ctx, "cas")
		if stored.Attributes[0].Value != "first" || r2 <= r1 {
			t.Errorf("%#v %v %v", stored.Attributes, r1, r2)
		}
	})

	t.Run("a graph written before revisions existed has revision 0 and can be written", func(t *testing.T) {
		_, err := m.graphCollection().InsertOne(ctx, bson.M{"id": "norev", "owner": "u1", "attributes": bson.A{}, "nodes": bson.A{bson.M{"id": "n", "attributes": bson.A{}}}, "edges": bson.A{}})
		if err != nil {
			t.Fatal(err)
		}
		g, r, exists, err := m.GetGraphWithRevision(ctx, "norev")
		if err != nil || !exists || r != 0 {
			t.Fatal(r, exists, err)
		}
		swapped, err := m.SetGraphIfRevision(ctx, g, r, noSync)
		if err != nil || !swapped {
			t.Fatal(swapped, err)
		}
	})
}
