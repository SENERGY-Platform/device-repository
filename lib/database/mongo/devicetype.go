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

const deviceTypeIdFieldName = "Id"
const deviceTypeNameFieldName = "Name"
const deviceTypeServiceFieldName = "Services"
const serviceIdFieldName = "Id"

var deviceTypeIdKey string
var deviceTypeNameKey string
var serviceIdKey string
var deviceTypeServicesKey string

var deviceTypeByServicePath string

func init() {
	var err error
	deviceTypeIdKey, err = getBsonFieldName(model.DeviceType{}, deviceTypeIdFieldName)
	if err != nil {
		log.Fatal(err)
	}
	deviceTypeNameKey, err = getBsonFieldName(model.DeviceType{}, deviceTypeNameFieldName)
	if err != nil {
		log.Fatal(err)
	}
	serviceIdKey, err = getBsonFieldName(model.Service{}, serviceIdFieldName)
	if err != nil {
		log.Fatal(err)
	}
	deviceTypeServicesKey, err = getBsonFieldName(model.DeviceType{}, deviceTypeServiceFieldName)
	if err != nil {
		log.Fatal(err)
	}

	deviceTypeByServicePath = deviceTypeServicesKey + "." + serviceIdKey

	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceTypeCollection)
		err = db.ensureIndex(collection, "devicetypeidindex", deviceTypeIdKey, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicetypenameindex", deviceTypeNameKey, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicetypeserviceindex", deviceTypeByServicePath, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceTypeCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceTypeCollection)
}

func (this *Mongo) GetDeviceType(ctx context.Context, id string) (deviceType model.DeviceType, exists bool, err error) {
	result := this.deviceTypeCollection().FindOne(ctx, bson.M{deviceTypeIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return deviceType, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&deviceType)
	if err == mongo.ErrNoDocuments {
		return deviceType, false, nil
	}
	return deviceType, true, err
}

func (this *Mongo) ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string, filterCriteria []model.FilterCriteria) (result []model.DeviceType, err error) {
	result = []model.DeviceType{}
	opt := options.Find()
	opt.SetLimit(limit)
	opt.SetSkip(offset)

	parts := strings.Split(sort, ".")
	sortby := deviceTypeIdKey
	switch parts[0] {
	case "id":
		sortby = deviceTypeIdKey
	case "name":
		sortby = deviceTypeNameKey
	default:
		sortby = deviceTypeIdKey
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bsonx.Doc{{sortby, bsonx.Int32(direction)}})

	filter := bson.M{}
	if len(filterCriteria) > 0 {
		deviceTypeIds, err := this.getDeviceTypeIdsByFilterCriteria(ctx, filterCriteria)
		if err != nil {
			return nil, err
		}
		filter = bson.M{deviceTypeIdKey: bson.M{"$in": deviceTypeIds}}
	}

	cursor, err := this.deviceTypeCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		deviceType := model.DeviceType{}
		err = cursor.Decode(&deviceType)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceType)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) SetDeviceType(ctx context.Context, deviceType model.DeviceType) error {
	_, err := this.deviceTypeCollection().ReplaceOne(ctx, bson.M{deviceTypeIdKey: deviceType.Id}, deviceType, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = this.setDeviceTypeCriteria(ctx, deviceType)
	if err != nil {
		return err
	}
	return err
}

func (this *Mongo) RemoveDeviceType(ctx context.Context, id string) error {
	_, err := this.deviceTypeCollection().DeleteOne(ctx, bson.M{deviceTypeIdKey: id})
	if err != nil {
		return err
	}
	err = this.removeDeviceTypeCriteriaByDeviceType(ctx, id)
	if err != nil {
		return err
	}
	return err
}

func (this *Mongo) GetDeviceTypesByServiceId(ctx context.Context, serviceId string) (result []model.DeviceType, err error) {
	opt := options.Find()
	opt.SetLimit(2)
	opt.SetSkip(0)

	cursor, err := this.deviceTypeCollection().Find(ctx, bson.M{deviceTypeByServicePath: serviceId}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		deviceType := model.DeviceType{}
		err = cursor.Decode(&deviceType)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceType)
	}
	err = cursor.Err()
	return
}

//all criteria must match
func (this *Mongo) getDeviceTypeIdsByFilterCriteria(ctx context.Context, criteria []model.FilterCriteria) (result []interface{}, err error) {
	for _, c := range criteria {
		result, err = this.filterDeviceTypeIdsByFilterCriteria(ctx, result, c)
		if err != nil {
			return result, err
		}
	}
	return
}

func (this *Mongo) filterDeviceTypeIdsByFilterCriteria(ctx context.Context, deviceTypeIds []interface{}, criteria model.FilterCriteria) (result []interface{}, err error) {
	result = []interface{}{}
	if deviceTypeIds != nil && len(deviceTypeIds) == 0 {
		return result, nil
	}
	filter := bson.M{}
	if deviceTypeIds != nil {
		filter = bson.M{
			deviceTypeCriteriaDeviceTypeIdKey: bson.M{"$in": deviceTypeIds},
		}
	}
	if criteria.DeviceClassId != "" {
		filter[deviceTypeCriteriaDeviceClassIdKey] = criteria.DeviceClassId
	}
	if criteria.FunctionId != "" {
		filter[deviceTypeCriteriaFunctionIdKey] = criteria.FunctionId
	}
	if criteria.AspectId != "" {
		node, exists, err := this.GetAspectNode(ctx, criteria.AspectId)
		if err != nil {
			return result, err
		}
		if exists {
			filter[deviceTypeCriteriaAspectIdKey] = bson.M{"$in": append(node.DescendentIds, node.Id)}
		} else {
			//return result, errors.New("unknown AspectId: "+criteria.AspectId)
			log.Println("WARNING: filterDeviceTypeIdsByFilterCriteria() aspect id not found as aspect-node", criteria.AspectId)
			filter[deviceTypeCriteriaAspectIdKey] = criteria.AspectId
		}
	}

	temp, err := this.deviceTypeCriteriaCollection().Distinct(ctx, deviceTypeCriteriaDeviceTypeIdKey, filter)
	if err != nil {
		return result, err
	}
	if temp != nil {
		result = temp
	}
	return
}
