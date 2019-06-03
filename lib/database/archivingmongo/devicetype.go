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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
	"strings"
)

const deviceTypeIdFieldName = "Id"
const deviceTypeNameFieldName = "Name"
const deviceTypeServicesFieldName = "Services"
const serviceIdFieldName = "Id"
const serviceInputFieldName = "Input"
const serviceOutputFieldName = "Output"
const assignmentTypeFieldName = "Type"
const valueTypeIdFieldName = "Id"

var deviceTypeIdKey string
var deviceTypeNameKey string
var deviceTypeServicesKey string
var serviceIdKey string
var serviceInputKey string
var serviceOutputKey string
var assignmentTypeKey string
var valueTypeIdKey string

var deviceTypeToServicePath string
var deviceTypeToInputPath string
var deviceTypeToOutputPath string

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

	deviceTypeServicesKey, err = getBsonFieldName(model.DeviceType{}, deviceTypeServicesFieldName)
	if err != nil {
		log.Fatal(err)
	}

	serviceIdKey, err = getBsonFieldName(model.Service{}, serviceIdFieldName)
	if err != nil {
		log.Fatal(err)
	}

	serviceInputKey, err = getBsonFieldName(model.Service{}, serviceInputFieldName)
	if err != nil {
		log.Fatal(err)
	}

	serviceOutputKey, err = getBsonFieldName(model.Service{}, serviceOutputFieldName)
	if err != nil {
		log.Fatal(err)
	}

	assignmentTypeKey, err = getBsonFieldName(model.TypeAssignment{}, assignmentTypeFieldName)
	if err != nil {
		log.Fatal(err)
	}

	valueTypeIdKey, err = getBsonFieldName(model.ValueType{}, valueTypeIdFieldName)
	if err != nil {
		log.Fatal(err)
	}

	deviceTypeToInputPath = strings.Join([]string{deviceTypeServicesKey, serviceInputKey, assignmentTypeKey, valueTypeIdKey}, ".")
	deviceTypeToOutputPath = strings.Join([]string{deviceTypeServicesKey, serviceOutputKey, assignmentTypeKey, valueTypeIdKey}, ".")
	deviceTypeToServicePath = strings.Join([]string{deviceTypeServicesKey, serviceIdKey}, ".")

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
		err = db.ensureIndex(collection, "devicetypeserviceindex", deviceTypeToServicePath, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicetypeinputvaluetypeindex", deviceTypeToInputPath, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicetypeoutputvaluetypeindex", deviceTypeToOutputPath, true, false)
		return err
	})
}

func (this *Mongo) deviceTypeCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceTypeCollection)
}

func (this *Mongo) GetDeviceType(ctx context.Context, id string) (deviceType model.DeviceType, exists bool, err error) {
	result := this.deviceTypeCollection().FindOne(ctx, bson.M{deviceTypeIdKey: id, "removed": bson.M{"$in": bson.A{nil, false}}})
	err = result.Err()
	if err != nil {
		return
	}
	err = result.Decode(&deviceType)
	if err == mongo.ErrNoDocuments {
		return deviceType, false, nil
	}
	return deviceType, true, err
}

func (this *Mongo) GetDeviceTypeWithService(ctx context.Context, id string) (deviceType model.DeviceType, exists bool, err error) {
	result := this.deviceTypeCollection().FindOne(ctx, bson.M{deviceTypeToServicePath: id, "removed": bson.M{"$in": bson.A{nil, false}}})
	err = result.Err()
	if err != nil {
		return
	}
	err = result.Decode(&deviceType)
	if err == mongo.ErrNoDocuments {
		return deviceType, false, nil
	}
	return deviceType, true, err
}

func (this *Mongo) ListDeviceTypes(ctx context.Context, listoptions listoptions.ListOptions) (result []model.DeviceType, err error) {
	opt := options.Find()
	if limit, ok := listoptions.GetLimit(); ok {
		opt.SetLimit(limit)
	}
	if offset, ok := listoptions.GetOffset(); ok {
		opt.SetSkip(offset)
	}
	if sort, ok := listoptions.Get("sort"); ok {
		sortstr, ok := sort.(string)
		if !ok {
			return result, errors.New("unable to interpret sort as string")
		}
		parts := strings.Split(sortstr, ".")
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
	}
	cursor, err := this.deviceTypeCollection().Find(ctx, bson.M{"removed": bson.M{"$in": bson.A{nil, false}}}, opt)
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
	_, err := this.deviceTypeCollection().UpdateOne(ctx, bson.M{deviceTypeIdKey: deviceType.Id}, bson.M{"$set": deviceType}, options.Update().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveDeviceType(ctx context.Context, id string) error {
	_, err := this.deviceTypeCollection().ReplaceOne(ctx, bson.M{deviceIdKey: id}, bson.M{deviceTypeIdKey: id, "removed": true}, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) ListDeviceTypesUsingValueType(ctx context.Context, id string, listoptions ...listoptions.ListOptions) (result []model.DeviceType, err error) {
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
	cursor, err := this.deviceTypeCollection().Find(ctx, bson.M{"$or": bson.A{bson.M{deviceTypeToInputPath: id}, bson.M{deviceTypeToOutputPath: id}}, "removed": bson.M{"$in": bson.A{nil, false}}}, opt)
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
