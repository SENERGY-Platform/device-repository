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

package testdb

import (
	"context"
	"maps"
	"slices"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func (db *DB) SetThing(ctx context.Context, thing models.Thing, syncHandler func(models.Thing) error) error {
	return set(thing.Id, db.things, thing, func(thing models.Thing) error {
		return syncHandler(thing)
	})
}

func (db *DB) RemoveThing(ctx context.Context, id string, syncDeleteHandler func(models.Thing) error) error {
	return del(id, db.things, syncDeleteHandler)
}

func (db *DB) GetThing(ctx context.Context, id string) (thing models.Thing, exists bool, err error) {
	return get(id, db.things)
}

func (db *DB) ListThings(ctx context.Context, listOptions model.ThingListOptions) (result []models.Thing, total int64, err error) {
	all := iterToSlice(maps.Values(db.things))
	result = []models.Thing{}
	for _, thing := range all {
		if listOptions.Ids != nil && !slices.Contains(listOptions.Ids, thing.Id) {
			continue
		}
		if len(listOptions.DeviceIds) > 0 && !slices.ContainsFunc(thing.DeviceLinks, func(link models.ThingDeviceLink) bool {
			return slices.Contains(listOptions.DeviceIds, link.DeviceId)
		}) {
			continue
		}
		result = append(result, thing)
	}
	return result, int64(len(result)), nil
}

func (db *DB) RetryThingSync(lockduration time.Duration, syncDeleteHandler func(models.Thing) error, syncHandler func(models.Thing) error) error {
	return nil
}

func (db *DB) DesyncUnknownThings(ctx context.Context, knownThings []string) (err error) {
	return nil
}
