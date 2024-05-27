/*
 * Copyright 2024 InfAI (CC SES)
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

package mongo

import (
	"context"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (this *Mongo) RunStartupMigrations() error {
	return this.runDeviceOwnerMigration()
}

func (this *Mongo) runDeviceOwnerMigration() error {
	cursor, err := this.deviceCollection().Find(context.Background(), bson.M{
		"$or": bson.A{
			bson.M{deviceOwnerIdKey: bson.M{"$exists": false}},
			bson.M{deviceOwnerIdKey: ""},
		},
	})
	if err != nil {
		return err
	}
	deviceKind, err := this.getInternalKind(this.config.DeviceTopic)
	if err != nil {
		return err
	}
	for cursor.Next(context.Background()) {
		element := models.Device{}
		err = cursor.Decode(&element)
		if err != nil {
			return err
		}
		if element.OwnerId != "" {
			panic("owner must be empty because we searched for devices without owner")
		}
		rights, err := this.getRights(deviceKind, element.Id)
		if err != nil {
			return err
		}
		if len(rights.AdminUsers) > 0 {
			element.OwnerId = rights.AdminUsers[0]
		}
		ctx, _ := getTimeoutContext()
		err = this.SetDevice(ctx, element)
		if err != nil {
			return err
		}
	}
	return cursor.Err()
}
