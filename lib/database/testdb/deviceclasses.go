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
	"golang.org/x/exp/maps"
)

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
	panic("not implemented") // TODO
}
func (db *DB) GetDeviceClass(_ context.Context, id string) (result models.DeviceClass, exists bool, err error) {
	return get(id, db.deviceClasses)
}
func (db *DB) DeviceClassIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	//TODO implement me
	panic("implement me")
}
