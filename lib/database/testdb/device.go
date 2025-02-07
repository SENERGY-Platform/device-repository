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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"regexp"
	"slices"
	"strings"
	"time"
)

func (db *DB) GetDevice(_ context.Context, id string) (device model.DeviceWithConnectionState, exists bool, err error) {
	return get(id, db.devices)
}

func (db *DB) RetryDeviceSync(lockduration time.Duration, syncDeleteHandler func(model.DeviceWithConnectionState) error, syncHandler func(model.DeviceWithConnectionState) error) error {
	return nil
}

func (db *DB) SetDevice(_ context.Context, device model.DeviceWithConnectionState, syncHandler func(model.DeviceWithConnectionState) error) error {
	return set(device.Id, db.devices, device, syncHandler)
}

func (db *DB) RemoveDevice(_ context.Context, id string, syncDeleteHandler func(model.DeviceWithConnectionState) error) error {
	return del(id, db.devices, syncDeleteHandler)
}

func (db *DB) GetDeviceByLocalId(_ context.Context, ownerId string, localId string) (device model.DeviceWithConnectionState, exists bool, err error) {
	for i := range db.devices {
		if db.devices[i].LocalId == localId && (!db.config.LocalIdUniqueForOwner || db.devices[i].OwnerId == ownerId) {
			return db.devices[i], true, nil
		}
	}
	return model.DeviceWithConnectionState{}, false, err
}

func (db *DB) DeviceLocalIdsToIds(ctx context.Context, owner string, localIds []string) (ids []string, err error) {
	ids = []string{}
	for _, lid := range localIds {
		device, exists, err := db.GetDeviceByLocalId(ctx, owner, lid)
		if err != nil {
			return nil, err
		}
		if exists {
			ids = append(ids, device.Id)
		}
	}
	return ids, nil
}

func (db *DB) ListDevices(ctx context.Context, options model.DeviceListOptions, withTotal bool) (devices []model.DeviceWithConnectionState, total int64, err error) {
	devices = []model.DeviceWithConnectionState{}
	var r *regexp.Regexp
	if options.Search != "" {
		r, err = regexp.Compile("(?i)" + regexp.QuoteMeta(options.Search))
		if err != nil {
			return nil, total, err
		}

	}
	for _, device := range db.devices {
		if options.Ids != nil && !slices.Contains(options.Ids, device.Id) {
			continue
		}
		if options.ConnectionState != nil && *options.ConnectionState != device.ConnectionState {
			continue
		}
		if options.Search != "" && r != nil {
			if !r.MatchString(device.Name) {
				continue
			}
		}
		devices = append(devices, device)
	}
	if options.SortBy == "" {
		options.SortBy = "name.asc"
	}
	sortby := options.SortBy
	sortby = strings.TrimSuffix(sortby, ".asc")
	sortby = strings.TrimSuffix(sortby, ".desc")

	direction := 1
	if strings.HasSuffix(options.SortBy, ".desc") {
		direction = -1
	}
	slices.SortFunc(devices, func(a, b model.DeviceWithConnectionState) int {
		afield := a.Name
		bfield := b.Name
		if sortby == "id" {
			afield = a.Id
			bfield = b.Id
		}
		return strings.Compare(afield, bfield) * direction
	})

	total = int64(len(devices))
	if options.Limit > 0 || options.Offset > 0 {
		if options.Offset >= int64(len(devices)) {
			return []model.DeviceWithConnectionState{}, total, nil
		}
		if (options.Limit + options.Offset) >= int64(len(devices)) {
			return devices[options.Offset:], total, nil
		}
		return devices[options.Offset : options.Limit+options.Offset], total, nil
	}

	return devices, total, nil
}

func (db *DB) SetDeviceConnectionState(ctx context.Context, id string, state models.ConnectionState) error {
	device, ok := db.devices[id]
	if !ok {
		return nil
	}
	device.ConnectionState = state
	return db.SetDevice(ctx, device, nil)
}
