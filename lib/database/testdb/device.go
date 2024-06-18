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

func (db *DB) GetDevice(_ context.Context, id string) (device models.Device, exists bool, err error) {
	return get(id, db.devices)
}

func (db *DB) SetDevice(_ context.Context, device models.Device) error {
	return set(device.Id, db.devices, device)
}

func (db *DB) RemoveDevice(_ context.Context, id string) error {
	return del(id, db.devices)

}

func (db *DB) GetDeviceByLocalId(_ context.Context, ownerId string, localId string) (device models.Device, exists bool, err error) {
	for i := range db.devices {
		if db.devices[i].LocalId == localId && (!db.config.LocalIdUniqueForOwner || db.devices[i].OwnerId == ownerId) {
			return db.devices[i], true, nil
		}
	}
	return models.Device{}, false, err
}
