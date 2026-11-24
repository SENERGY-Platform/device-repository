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
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestRemoveMany(t *testing.T) {
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
	set := func(t *testing.T, id string, owner string) {
		t.Helper()
		if err := m.SetThing(timeout(), models.Thing{Id: id, Name: id, OwnerId: owner}, noSync); err != nil {
			t.Fatal(err)
		}
	}
	// stored reports whether the document exists at all and whether it is marked as deleted
	stored := func(t *testing.T, id string) (exists bool, markedDeleted bool) {
		t.Helper()
		var doc struct {
			SyncDelete bool `bson:"sync_delete"`
		}
		err := m.thingCollection().FindOne(timeout(), bson.M{ThingBson.Id: id}).Decode(&doc)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return false, false
			}
			t.Fatal(err)
		}
		return true, doc.SyncDelete
	}
	timestampOwners := func(t *testing.T, owners ...string) (result map[string]int64) {
		t.Helper()
		result = map[string]int64{}
		for _, owner := range owners {
			list, err := m.GetLastUpdateTimestampsForUser(timeout(), owner)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range list {
				if e.Collection == conf.MongoThingCollection {
					result[owner] = e.UnixTimestamp
				}
			}
		}
		return result
	}

	t.Run("known and unknown ids with duplicates", func(t *testing.T) {
		set(t, "t1", "u1")
		set(t, "t2", "u1")
		set(t, "t3", "u2")
		set(t, "keep", "u1")

		calls := 0
		var received []string
		err := m.RemoveThings(timeout(), []string{"t1", "unknown", "t2", "t3", "t1"}, func(things []models.Thing) error {
			calls++
			for _, thing := range things {
				received = append(received, thing.Id)
				//the elements are marked as deleted while the handler runs, like in the single variant
				exists, marked := stored(t, thing.Id)
				if !exists || !marked {
					t.Errorf("%v: exists=%v marked=%v", thing.Id, exists, marked)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Errorf("handler called %v times", calls)
		}
		slices.Sort(received)
		if !slices.Equal(received, []string{"t1", "t2", "t3"}) {
			t.Errorf("unexpected handler input %v", received)
		}
		for _, id := range []string{"t1", "t2", "t3"} {
			if exists, _ := stored(t, id); exists {
				t.Errorf("%v still stored", id)
			}
		}
		if exists, marked := stored(t, "keep"); !exists || marked {
			t.Errorf("keep: exists=%v marked=%v", exists, marked)
		}
		timestamps := timestampOwners(t, "u1", "u2")
		if timestamps["u1"] == 0 || timestamps["u2"] == 0 {
			t.Errorf("missing last update timestamps: %v", timestamps)
		}
	})

	t.Run("only unknown ids do not call the handler", func(t *testing.T) {
		err := m.RemoveThings(timeout(), []string{"unknown1", "unknown2"}, func([]models.Thing) error {
			t.Error("handler must not be called")
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		err = m.RemoveThings(timeout(), []string{}, func([]models.Thing) error {
			t.Error("handler must not be called")
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("failing handler leaves the documents marked", func(t *testing.T) {
		set(t, "f1", "u1")
		set(t, "f2", "u1")
		set(t, "f3", "u1")
		err := m.RemoveThings(timeout(), []string{"f1", "f2"}, func([]models.Thing) error {
			return errors.New("handler failed")
		})
		if err != nil {
			t.Fatalf("a failing sync handler is retried later and must not be returned: %v", err)
		}
		for _, id := range []string{"f1", "f2"} {
			if exists, marked := stored(t, id); !exists || !marked {
				t.Errorf("%v: exists=%v marked=%v", id, exists, marked)
			}
		}
		if exists, marked := stored(t, "f3"); !exists || marked {
			t.Errorf("f3: exists=%v marked=%v", exists, marked)
		}

		t.Run("marked documents are ignored by another batch", func(t *testing.T) {
			err := m.RemoveThings(timeout(), []string{"f1", "f3"}, func(things []models.Thing) error {
				if len(things) != 1 || things[0].Id != "f3" {
					t.Errorf("unexpected handler input %#v", things)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			//f1 must keep its mark: its sync handler did not succeed yet
			if exists, marked := stored(t, "f1"); !exists || !marked {
				t.Errorf("f1: exists=%v marked=%v", exists, marked)
			}
			if exists, _ := stored(t, "f3"); exists {
				t.Error("f3 still stored")
			}
		})

		t.Run("retry loop picks the marked documents up", func(t *testing.T) {
			//sync jobs are selected by whole seconds and only if they are older than the lock duration,
			//see docs/running-the-container-tests.md
			var retried []string
			docker.WaitFor(t, 10*time.Second, func() error {
				err := m.RetryThingSync(time.Second, func(thing models.Thing) error {
					retried = append(retried, thing.Id)
					return nil
				}, noSync)
				if err != nil {
					t.Fatal(err)
				}
				if len(retried) < 2 {
					return errors.New("marked documents not retried yet")
				}
				return nil
			})
			slices.Sort(retried)
			if !slices.Equal(retried, []string{"f1", "f2"}) {
				t.Errorf("unexpected retried documents %v", retried)
			}
		})
	})
}
