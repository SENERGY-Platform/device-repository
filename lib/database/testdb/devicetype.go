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
	"golang.org/x/exp/slices"
	"strings"
)

func (db *DB) GetDeviceType(_ context.Context, id string) (deviceType models.DeviceType, exists bool, err error) {
	return get(id, db.deviceTypes)
}
func (db *DB) SetDeviceType(_ context.Context, deviceType models.DeviceType) error {
	return set(deviceType.Id, db.deviceTypes, deviceType)

}
func (db *DB) RemoveDeviceType(_ context.Context, id string) error {
	return del(id, db.deviceTypes)

}
func (db *DB) ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string, includeModified bool) (result []models.DeviceType, err error) {
	// TODO filtering
	// sort can be id or name with .asc or .desc
	deviceTypes := maps.Values(db.deviceTypes)
	if offset >= int64(len(deviceTypes)) {
		return []models.DeviceType{}, nil
	}

	parts := strings.Split(sort, ".")
	desc := parts[1] == "desc"
	switch parts[0] {
	case "name":
		slices.SortFunc(deviceTypes, func(a, b models.DeviceType) int {
			if desc {
				return strings.Compare(a.Name, b.Name)
			}
			return strings.Compare(a.Name, b.Name) * -1
		})
	default:
	case "id":
		slices.SortFunc(deviceTypes, func(a, b models.DeviceType) int {
			if desc {
				return strings.Compare(a.Id, b.Id)
			}
			return strings.Compare(a.Id, b.Id) * -1
		})
	}
	if offset >= int64(len(deviceTypes)) {
		return []models.DeviceType{}, nil
	}

	return deviceTypes[offset:min(len(deviceTypes), int(offset+limit))], nil
}

func (db *DB) ListDeviceTypesV2(ctx context.Context, limit int64, offset int64, sort string, filter []model.FilterCriteria, includeModified bool) (result []models.DeviceType, err error) {
	panic("implement me")
}

func (db *DB) GetDeviceTypesByServiceId(_ context.Context, serviceId string) (result []models.DeviceType, err error) {
	for i := range db.deviceTypes {
		for j := range db.deviceTypes[i].Services {
			if db.deviceTypes[i].Services[j].Id == serviceId {
				result = append(result, db.deviceTypes[i])
				break
			}
		}
	}
	return
}
