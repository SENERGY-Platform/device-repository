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

package archivingmongo

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const deviceIdFieldName = "Id"
const deviceUrlFieldName = "Url"
const deviceDeviceTypeFieldName = "DeviceType"
const deviceHubFieldName = "Gateway"

var deviceIdKey string
var deviceUrlKey string
var deviceDeviceTypeKey string
var deviceHubKey string

func init() {
	var err error
	deviceIdKey, err = getBsonFieldName(model.DeviceInstance{}, deviceIdFieldName)
	if err != nil {
		log.Fatal(err)
	}
	deviceUrlKey, err = getBsonFieldName(model.DeviceInstance{}, deviceUrlFieldName)
	if err != nil {
		log.Fatal(err)
	}
	deviceDeviceTypeKey, err = getBsonFieldName(model.DeviceInstance{}, deviceDeviceTypeFieldName)
	if err != nil {
		log.Fatal(err)
	}
	deviceHubKey, err = getBsonFieldName(model.DeviceInstance{}, deviceHubFieldName)
	if err != nil {
		log.Fatal(err)
	}
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceCollection)
		err = db.ensureIndex(collection, "deviceidindex", deviceIdKey, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceurlindex", deviceUrlKey, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicedevicetypeindex", deviceDeviceTypeKey, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicehubindex", deviceHubKey, true, false)
		return err
	})
}

func (this *Mongo) deviceCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceCollection)
}

func (this *Mongo) GetDevice(ctx context.Context, id string) (device model.DeviceInstance, exists bool, err error) {
	result := this.deviceCollection().FindOne(ctx, bson.M{deviceIdKey: id, "removed": bson.M{"$in": bson.A{nil, false}}})
	err = result.Err()
	if err != nil {
		return
	}
	err = result.Decode(&device)
	if err == mongo.ErrNoDocuments {
		return device, false, nil
	}
	return device, true, err
}

func (this *Mongo) SetDevice(ctx context.Context, device model.DeviceInstance) error {
	_, err := this.deviceCollection().UpdateOne(ctx, bson.M{deviceIdKey: device.Id}, bson.M{"$set": device}, options.Update().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveDevice(ctx context.Context, id string) error {
	_, err := this.deviceCollection().ReplaceOne(ctx, bson.M{deviceIdKey: id}, bson.M{deviceIdKey: id, "removed": true}, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) ListDevicesOfDeviceType(ctx context.Context, deviceTypeId string, listoptions ...listoptions.ListOptions) (result []model.DeviceInstance, err error) {
	opt := options.Find()
	if len(listoptions) > 0 {
		if limit, ok := listoptions[0].GetLimit(); ok {
			opt.SetLimit(limit)
		}
		if offset, ok := listoptions[0].GetOffset(); ok {
			opt.SetSkip(offset)
		}
		err = listoptions[0].EvalStrict()
		if err != nil {
			return result, err
		}
	}
	cursor, err := this.deviceCollection().Find(ctx, bson.M{deviceDeviceTypeKey: deviceTypeId, "removed": bson.M{"$in": bson.A{nil, false}}}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		device := model.DeviceInstance{}
		err = cursor.Decode(&device)
		if err != nil {
			return nil, err
		}
		result = append(result, device)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListDevicesWithHub(ctx context.Context, id string, listoptions ...listoptions.ListOptions) (result []model.DeviceInstance, err error) {
	opt := options.Find()
	if len(listoptions) > 0 {
		if limit, ok := listoptions[0].GetLimit(); ok {
			opt.SetLimit(limit)
		}
		if offset, ok := listoptions[0].GetOffset(); ok {
			opt.SetSkip(offset)
		}
		err = listoptions[0].EvalStrict()
		if err != nil {
			return result, err
		}
	}
	cursor, err := this.deviceCollection().Find(ctx, bson.M{deviceHubKey: id, "removed": bson.M{"$in": bson.A{nil, false}}}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		device := model.DeviceInstance{}
		err = cursor.Decode(&device)
		if err != nil {
			return nil, err
		}
		result = append(result, device)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) GetDeviceByUri(ctx context.Context, uri string) (device model.DeviceInstance, exists bool, err error) {
	result := this.deviceCollection().FindOne(ctx, bson.M{deviceUrlKey: uri, "removed": bson.M{"$in": bson.A{nil, false}}})
	err = result.Err()
	if err != nil {
		return
	}
	err = result.Decode(&device)
	if err == mongo.ErrNoDocuments {
		return device, false, nil
	}
	return device, true, err
}
