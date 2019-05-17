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
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strings"
)

const deviceTypeIdFieldName = "Id"
const deviceTypeServicesFieldName = "Services"
const serviceInputFieldName = "Input"
const serviceOutputFieldName = "Output"
const assignmentTypeFieldName = "Type"
const valueTypeIdFieldName = "Id"

var deviceTypeIdKey string
var deviceTypeServicesKey string
var serviceInputKey string
var serviceOutputKey string
var assignmentTypeKey string
var valueTypeIdKey string

var deviceTypeToInputPath string
var deviceTypeToOutputPath string

func init() {
	var err error
	deviceTypeIdKey, err = getBsonFieldName(model.DeviceType{}, deviceTypeIdFieldName)
	if err != nil {
		log.Fatal(err)
	}

	deviceTypeServicesKey, err = getBsonFieldName(model.DeviceType{}, deviceTypeServicesFieldName)
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

	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceTypeCollection)
		err = db.ensureIndex(collection, "devicetypeidindex", deviceTypeIdKey, true, true)
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
	result := this.deviceTypeCollection().FindOne(ctx, bson.D{{deviceTypeIdKey, id}})
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

func (this *Mongo) SetDeviceType(ctx context.Context, deviceType model.DeviceType) error {
	_, err := this.deviceTypeCollection().ReplaceOne(ctx, bson.M{deviceTypeIdKey: deviceType.Id}, deviceType, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveDeviceType(ctx context.Context, id string) error {
	_, err := this.deviceTypeCollection().DeleteOne(ctx, bson.M{deviceTypeIdKey: id})
	return err
}

func (this *Mongo) ListDeviceTypesUsingValueType(ctx context.Context, id string) (result []model.DeviceType, err error) {
	cursor, err := this.deviceTypeCollection().Find(ctx, bson.M{"$or": bson.A{bson.M{deviceTypeToInputPath: id}, bson.M{deviceTypeToOutputPath: id}}})
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
