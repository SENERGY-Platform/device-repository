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
	"slices"
	"strings"
)

func (db *DB) ListDeviceClasses(ctx context.Context, options model.DeviceClassListOptions) (result []models.DeviceClass, total int64, err error) {
	for _, dc := range db.deviceClasses {
		if (options.Search == "" || strings.Contains(strings.ToLower(dc.Name), strings.ToLower(options.Search))) &&
			(options.Ids == nil || slices.Contains(options.Ids, dc.Id)) {
			result = append(result, dc)
		}
	}
	limit := options.Limit
	offset := options.Offset
	if offset >= int64(len(result)) {
		return []models.DeviceClass{}, int64(len(result)), nil
	}
	return result[offset:min(len(result), int(offset+limit))], int64(len(result)), nil
}

func (db *DB) SetDeviceClass(_ context.Context, class models.DeviceClass) error {
	return set(class.Id, db.deviceClasses, class)
}
func (db *DB) RemoveDeviceClass(_ context.Context, id string) error {
	return del(id, db.deviceClasses)
}
func (db *DB) ListAllDeviceClasses(_ context.Context) ([]models.DeviceClass, error) {
	return maps.Values(db.deviceClasses), nil
}
func (db *DB) ListAllDeviceClassesUsedWithControllingFunctions(_ context.Context) ([]models.DeviceClass, error) {
	panic("not implemented")
}
func (db *DB) GetDeviceClass(_ context.Context, id string) (result models.DeviceClass, exists bool, err error) {
	return get(id, db.deviceClasses)
}
func (db *DB) DeviceClassIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	panic("implement me")
}
