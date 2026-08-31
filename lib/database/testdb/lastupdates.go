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
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/model"
)

func (db *DB) GetLastUpdateTimestampsForUser(ctx context.Context, userId string) (result []model.LastUpdateTimestamp, err error) {
	ts := time.Now().Unix()
	return []model.LastUpdateTimestamp{
		{
			Collection:    db.config.MongoDeviceCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoHubCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoCharacteristicCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoProtocolCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoLocationCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoConceptCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoAspectCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoAspectClassCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoDeviceClassCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoDeviceGroupCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoDeviceTypeCollection,
			UnixTimestamp: ts,
		},
		{
			Collection:    db.config.MongoFunctionCollection,
			UnixTimestamp: ts,
		},
	}, nil
}
