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

package mongo

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const deviceClassIdFieldName = "Id"

var deviceClassIdKey string

func init() {
	var err error
	deviceClassIdKey, err = getBsonFieldName(model.DeviceClass{}, deviceClassIdFieldName)
	if err != nil {
		log.Fatal(err)
	}

	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceClassCollection)
		err = db.ensureIndex(collection, "deviceclassidindex", deviceClassIdKey, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceClassCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceClassCollection)
}

func (this *Mongo) GetDeviceClass(ctx context.Context, id string) (deviceClass model.DeviceClass, exists bool, err error) {
	result := this.deviceClassCollection().FindOne(ctx, bson.M{deviceClassIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return deviceClass, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&deviceClass)
	if err == mongo.ErrNoDocuments {
		return deviceClass, false, nil
	}
	return deviceClass, true, err
}

func (this *Mongo) SetDeviceClass(ctx context.Context, deviceClass model.DeviceClass) error {
	_, err := this.deviceClassCollection().ReplaceOne(ctx, bson.M{deviceClassIdKey: deviceClass.Id}, deviceClass, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveDeviceClass(ctx context.Context, id string) error {
	_, err := this.deviceClassCollection().DeleteOne(ctx, bson.M{deviceClassIdKey: id})
	return err
}

func (this *Mongo) ListAllDeviceClasses(ctx context.Context) (result []model.DeviceClass, err error) {
	cursor, err := this.deviceClassCollection().Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		deviceClass := model.DeviceClass{}
		err = cursor.Decode(&deviceClass)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceClass)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListAllDeviceClassesUsedWithControllingFunctions(ctx context.Context) (result []model.DeviceClass, err error) {
	deviceClassIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, deviceTypeCriteriaDeviceClassIdKey, bson.M{
		deviceTypeCriteriaIsControllingFunctionKey: true,
		deviceTypeCriteriaDeviceClassIdKey:         bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.deviceClassCollection().Find(ctx, bson.M{deviceClassIdKey: bson.M{"$in": deviceClassIds}})
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		deviceClass := model.DeviceClass{}
		err = cursor.Decode(&deviceClass)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceClass)
	}
	err = cursor.Err()
	return
}
