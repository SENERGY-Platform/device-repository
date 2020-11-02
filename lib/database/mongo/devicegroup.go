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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
	"strings"
)

const deviceGroupIdFieldName = "Id"
const deviceGroupNameFieldName = "Name"

var deviceGroupIdKey string
var deviceGroupNameKey string

func init() {
	var err error
	deviceGroupIdKey, err = getBsonFieldName(model.DeviceGroup{}, deviceGroupIdFieldName)
	if err != nil {
		log.Fatal(err)
	}
	deviceGroupNameKey, err = getBsonFieldName(model.DeviceGroup{}, deviceGroupNameFieldName)
	if err != nil {
		log.Fatal(err)
	}
	serviceIdKey, err = getBsonFieldName(model.Service{}, serviceIdFieldName)
	if err != nil {
		log.Fatal(err)
	}

	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceGroupCollection)
		err = db.ensureIndex(collection, "deviceGroupidindex", deviceGroupIdKey, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceGroupCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceGroupCollection)
}

func (this *Mongo) GetDeviceGroup(ctx context.Context, id string) (deviceGroup model.DeviceGroup, exists bool, err error) {
	result := this.deviceGroupCollection().FindOne(ctx, bson.M{deviceGroupIdKey: id})
	err = result.Err()
	if err != nil {
		return
	}
	err = result.Decode(&deviceGroup)
	if err == mongo.ErrNoDocuments {
		return deviceGroup, false, nil
	}
	return deviceGroup, true, err
}

func (this *Mongo) ListDeviceGroups(ctx context.Context, limit int64, offset int64, sort string) (result []model.DeviceGroup, err error) {
	opt := options.Find()
	opt.SetLimit(limit)
	opt.SetSkip(offset)

	parts := strings.Split(sort, ".")
	sortby := deviceGroupIdKey
	switch parts[0] {
	case "id":
		sortby = deviceGroupIdKey
	case "name":
		sortby = deviceGroupNameKey
	default:
		sortby = deviceGroupIdKey
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bsonx.Doc{{sortby, bsonx.Int32(direction)}})

	cursor, err := this.deviceGroupCollection().Find(ctx, bson.M{}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		deviceGroup := model.DeviceGroup{}
		err = cursor.Decode(&deviceGroup)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceGroup)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) SetDeviceGroup(ctx context.Context, deviceGroup model.DeviceGroup) error {
	_, err := this.deviceGroupCollection().ReplaceOne(ctx, bson.M{deviceGroupIdKey: deviceGroup.Id}, deviceGroup, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveDeviceGroup(ctx context.Context, id string) error {
	_, err := this.deviceGroupCollection().DeleteOne(ctx, bson.M{deviceGroupIdKey: id})
	return err
}
