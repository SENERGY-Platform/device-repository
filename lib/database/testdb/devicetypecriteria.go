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
)

func (db *DB) GetDeviceTypeCriteriaByAspectIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
	panic("implement me")
}

func (db *DB) GetDeviceTypeCriteriaByFunctionIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
	panic("implement me")
}

func (db *DB) GetDeviceTypeCriteriaByDeviceClassIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
	panic("implement me")
}

func (db *DB) GetDeviceTypeCriteriaByCharacteristicIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
	panic("implement me")
}

func (db *DB) GetDeviceTypeCriteriaForDeviceTypeIdsAndFilterCriteria(ctx context.Context, deviceTypeIds []interface{}, criteria model.FilterCriteria, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
	panic("not implemented")
}
func (db *DB) GetDeviceTypeIdsByFilterCriteria(ctx context.Context, criteria []model.FilterCriteria, interactionsFilter []string, includeModified bool) (result []interface{}, err error) {
	panic("not implemented")
}

func (db *DB) GetConfigurableCandidates(_ context.Context, serviceId string) (result []model.DeviceTypeCriteria, err error) {
	panic("not implemented")
}

func (db *DB) GetDeviceTypeIdsByFilterCriteriaV2(ctx context.Context, criteria []model.FilterCriteria, includeModified bool) (result []interface{}, err error) {
	panic("implement me")
}
