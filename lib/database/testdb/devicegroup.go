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
	"golang.org/x/exp/maps"
	"time"
)

func (db *DB) SetDeviceGroup(ctx context.Context, deviceGroup models.DeviceGroup, syncHandler func(dg models.DeviceGroup, user string) error, user string) error {
	return set(deviceGroup.Id, db.deviceGroups, deviceGroup, func(group models.DeviceGroup) error {
		return syncHandler(group, user)
	})
}

func (db *DB) GetDeviceGroupSyncUser(ctx context.Context, deviceGroupId string) (syncUser string, exists bool, err error) {
	return "", false, nil
}

func (db *DB) RemoveDeviceGroup(ctx context.Context, id string, syncDeleteHandler func(models.DeviceGroup) error) error {
	return del(id, db.deviceGroups, syncDeleteHandler)
}

func (db *DB) RetryDeviceGroupSync(lockduration time.Duration, syncDeleteHandler func(models.DeviceGroup) error, syncHandler func(dg models.DeviceGroup, user string) error) error {
	return nil
}

func (db *DB) GetDeviceGroup(_ context.Context, id string) (deviceGroup models.DeviceGroup, exists bool, err error) {
	return get(id, db.deviceGroups)
}

func (db *DB) ListDeviceGroups(_ context.Context, options model.DeviceGroupListOptions) (result []models.DeviceGroup, total int64, err error) {
	return maps.Values(db.deviceGroups), int64(len(db.deviceGroups)), nil
}
