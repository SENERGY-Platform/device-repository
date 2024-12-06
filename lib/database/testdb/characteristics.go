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

func (db *DB) ListCharacteristics(ctx context.Context, options model.CharacteristicListOptions) (result []models.Characteristic, total int64, err error) {
	for _, characteristic := range db.characteristics {
		if (options.Search == "" || strings.Contains(strings.ToLower(characteristic.Name), strings.ToLower(options.Search))) &&
			(options.Ids == nil || slices.Contains(options.Ids, characteristic.Id)) {
			result = append(result, characteristic)
		}
	}
	limit := options.Limit
	offset := options.Offset
	if offset >= int64(len(result)) {
		return []models.Characteristic{}, int64(len(result)), nil
	}
	return result[offset:min(len(result), int(offset+limit))], int64(len(result)), nil
}

func (db *DB) SetCharacteristic(_ context.Context, characteristic models.Characteristic) error {
	return set(characteristic.Id, db.characteristics, characteristic)
}
func (db *DB) RemoveCharacteristic(_ context.Context, id string) error {
	return del(id, db.characteristics)
}
func (db *DB) GetCharacteristic(_ context.Context, id string) (result models.Characteristic, exists bool, err error) {
	return get(id, db.characteristics)
}
func (db *DB) ListAllCharacteristics(_ context.Context) ([]models.Characteristic, error) {
	return maps.Values(db.characteristics), nil
}
func (db *DB) CharacteristicIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	panic("implement me")
}
func (db *DB) CharacteristicIsUsedWithConceptInDeviceType(ctx context.Context, characteristicId string, conceptId string) (result bool, where []string, err error) {
	panic("implement me")
}
