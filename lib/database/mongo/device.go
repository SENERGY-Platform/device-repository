/*
 * Copyright 2019 InfAI (CC SES)
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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const deviceIdFieldName = "Id"
const deviceLocalIdFieldName = "LocalId"

var deviceIdKey string
var deviceLocalIdKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		deviceIdKey, err = getBsonFieldName(models.Device{}, deviceIdFieldName)
		if err != nil {
			return err
		}
		deviceLocalIdKey, err = getBsonFieldName(models.Device{}, deviceLocalIdFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceCollection)
		err = db.ensureIndex(collection, "deviceidindex", deviceIdKey, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicelocalidindex", deviceLocalIdKey, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceCollection)
}

func (this *Mongo) GetDevice(ctx context.Context, id string) (device models.Device, exists bool, err error) {
	result := this.deviceCollection().FindOne(ctx, bson.M{deviceIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return device, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&device)
	if err == mongo.ErrNoDocuments {
		return device, false, nil
	}
	return device, true, err
}

func (this *Mongo) SetDevice(ctx context.Context, device models.Device) error {
	_, err := this.deviceCollection().ReplaceOne(ctx, bson.M{deviceIdKey: device.Id}, device, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveDevice(ctx context.Context, id string) error {
	_, err := this.deviceCollection().DeleteOne(ctx, bson.M{deviceIdKey: id})
	return err
}

func (this *Mongo) GetDeviceByLocalId(ctx context.Context, localId string) (device models.Device, exists bool, err error) {
	result := this.deviceCollection().FindOne(ctx, bson.M{deviceLocalIdKey: localId})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return device, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&device)
	if err == mongo.ErrNoDocuments {
		return device, false, nil
	}
	return device, true, err
}
