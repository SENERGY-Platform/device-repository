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
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"regexp"
	"strings"
)

var DeviceClassBson = getBsonFieldObject[models.DeviceClass]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceClassCollection)
		err = db.ensureIndex(collection, "deviceclassidindex", DeviceClassBson.Id, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceClassCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceClassCollection)
}

func (this *Mongo) ListDeviceClasses(ctx context.Context, listOptions model.DeviceClassListOptions) (result []models.DeviceClass, total int64, err error) {
	opt := options.Find()
	opt.SetLimit(listOptions.Limit)
	opt.SetSkip(listOptions.Offset)

	parts := strings.Split(listOptions.SortBy, ".")
	sortby := DeviceClassBson.Id
	switch parts[0] {
	case "id":
		sortby = DeviceClassBson.Id
	case "name":
		sortby = DeviceClassBson.Name
	default:
		sortby = DeviceClassBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{}
	if listOptions.Ids != nil {
		filter[DeviceClassBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter[DeviceClassBson.Name] = bson.M{"$regex": escapedSearch, "$options": "i"}
	}

	cursor, err := this.deviceClassCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.deviceClassCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) GetDeviceClass(ctx context.Context, id string) (deviceClass models.DeviceClass, exists bool, err error) {
	result := this.deviceClassCollection().FindOne(ctx, bson.M{DeviceClassBson.Id: id})
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

func (this *Mongo) SetDeviceClass(ctx context.Context, deviceClass models.DeviceClass) error {
	_, err := this.deviceClassCollection().ReplaceOne(ctx, bson.M{DeviceClassBson.Id: deviceClass.Id}, deviceClass, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveDeviceClass(ctx context.Context, id string) error {
	_, err := this.deviceClassCollection().DeleteOne(ctx, bson.M{DeviceClassBson.Id: id})
	return err
}

func (this *Mongo) ListAllDeviceClasses(ctx context.Context) (result []models.DeviceClass, err error) {
	cursor, err := this.deviceClassCollection().Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		deviceClass := models.DeviceClass{}
		err = cursor.Decode(&deviceClass)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceClass)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListAllDeviceClassesUsedWithControllingFunctions(ctx context.Context) (result []models.DeviceClass, err error) {
	deviceClassIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.DeviceClassId, bson.M{
		deviceTypeCriteriaIsControllingFunctionKey: true,
		DeviceTypeCriteriaBson.DeviceClassId:       bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.deviceClassCollection().Find(ctx, bson.M{DeviceClassBson.Id: bson.M{"$in": deviceClassIds}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		deviceClass := models.DeviceClass{}
		err = cursor.Decode(&deviceClass)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceClass)
	}
	err = cursor.Err()
	return
}
