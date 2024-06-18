/*
 * Copyright 2022 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/models/go/models"
)

func (db *DB) GetHub(_ context.Context, id string) (hub models.Hub, exists bool, err error) {
	return get(id, db.hubs)
}
func (db *DB) SetHub(_ context.Context, hub models.Hub) error {
	return set(hub.Id, db.hubs, hub)
}
func (db *DB) RemoveHub(_ context.Context, id string) error {
	return del(id, db.hubs)
}
func (db *DB) GetHubsByDeviceId(_ context.Context, id string) (hubs []models.Hub, err error) {
	for i := range db.hubs {
		for j := range db.hubs[i].DeviceIds {
			if db.hubs[i].DeviceLocalIds[j] == id {
				hubs = append(hubs, db.hubs[i])
				break
			}
		}
	}
	return
}
