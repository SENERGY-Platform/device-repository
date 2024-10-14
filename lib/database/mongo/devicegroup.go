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
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"regexp"
	"strings"
)

var DeviceGroupBson = getBsonFieldObject[models.DeviceGroup]()

var deviceGroupCriteriaShortKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		deviceGroupCriteriaShortKey, err = getBsonFieldName(models.DeviceGroup{}, "CriteriaShort")
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceGroupCollection)
		err = db.ensureIndex(collection, "deviceGroupidindex", DeviceGroupBson.Id, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceGroupCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceGroupCollection)
}

func (this *Mongo) GetDeviceGroup(ctx context.Context, id string) (deviceGroup models.DeviceGroup, exists bool, err error) {
	result := this.deviceGroupCollection().FindOne(ctx, bson.M{DeviceGroupBson.Id: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return deviceGroup, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&deviceGroup)
	if err == mongo.ErrNoDocuments {
		return deviceGroup, false, nil
	}
	return deviceGroup, true, err
}

func (this *Mongo) ListDeviceGroups(ctx context.Context, listOptions model.DeviceGroupListOptions) (result []models.DeviceGroup, total int64, err error) {
	opt := options.Find()
	opt.SetLimit(listOptions.Limit)
	opt.SetSkip(listOptions.Offset)

	parts := strings.Split(listOptions.SortBy, ".")
	sortby := DeviceGroupBson.Id
	switch parts[0] {
	case "id":
		sortby = DeviceGroupBson.Id
	case "name":
		sortby = DeviceGroupBson.Name
	default:
		sortby = DeviceGroupBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{}
	if listOptions.Ids != nil {
		filter[DeviceGroupBson.Id] = bson.M{"$in": listOptions.Ids}
	}

	if listOptions.IgnoreGenerated {
		filter[DeviceGroupBson.Id] = bson.M{"$in": listOptions.Ids}
	}

	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter[DeviceGroupBson.Name] = bson.M{"$regex": escapedSearch, "$options": "i"}
	}

	if listOptions.Criteria != nil {
		criteriaFilter := []bson.M{}
		for _, c := range listOptions.Criteria {
			criteriaFilter = append(criteriaFilter, bson.M{deviceGroupCriteriaShortKey: models.DeviceGroupFilterCriteria{
				Interaction:   c.Interaction,
				FunctionId:    c.FunctionId,
				AspectId:      c.AspectId,
				DeviceClassId: c.DeviceClassId,
			}.Short()})
		}
		filter["$and"] = criteriaFilter
	}

	cursor, err := this.deviceGroupCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.deviceCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) SetDeviceGroup(ctx context.Context, deviceGroup models.DeviceGroup) error {
	_, err := this.deviceGroupCollection().ReplaceOne(ctx, bson.M{DeviceGroupBson.Id: deviceGroup.Id}, deviceGroup, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveDeviceGroup(ctx context.Context, id string) error {
	_, err := this.deviceGroupCollection().DeleteOne(ctx, bson.M{DeviceGroupBson.Id: id})
	return err
}
