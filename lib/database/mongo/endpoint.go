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
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const endpointIdFieldName = "Id"
const endpointDeviceFieldName = "Device"
const endpointServiceFieldName = "Service"
const endpointEndpointFieldName = "Endpoint"
const endpointProtocolFieldName = "ProtocolHandler"

var endpointIdKey string
var endpointDeviceKey string
var endpointServiceKey string
var endpointEndpointKey string
var endpointProtocolKey string

func init() {
	var err error
	endpointIdKey, err = getBsonFieldName(model.Endpoint{}, endpointIdFieldName)
	if err != nil {
		log.Fatal(err)
	}

	endpointDeviceKey, err = getBsonFieldName(model.Endpoint{}, endpointDeviceFieldName)
	if err != nil {
		log.Fatal(err)
	}

	endpointServiceKey, err = getBsonFieldName(model.Endpoint{}, endpointServiceFieldName)
	if err != nil {
		log.Fatal(err)
	}

	endpointEndpointKey, err = getBsonFieldName(model.Endpoint{}, endpointEndpointFieldName)
	if err != nil {
		log.Fatal(err)
	}

	endpointProtocolKey, err = getBsonFieldName(model.Endpoint{}, endpointProtocolFieldName)
	if err != nil {
		log.Fatal(err)
	}

	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoEndpointCollection)
		err = db.ensureIndex(collection, "endpointidindex", endpointIdKey, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "endpointdeviceindex", endpointDeviceKey, true, false)
		if err != nil {
			return err
		}

		err = db.ensureCompoundIndex(collection, "endpointoutindex", true, false, endpointDeviceKey, endpointServiceKey)
		if err != nil {
			return err
		}

		err = db.ensureCompoundIndex(collection, "endpointinindex", true, false, endpointEndpointKey, endpointProtocolKey)
		return err
	})
}

func (this *Mongo) endpointCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoEndpointCollection)
}

func (this *Mongo) ListEndpoints(ctx context.Context, listoptions ...listoptions.ListOptions) (result []model.Endpoint, err error) {
	opt := options.Find()
	filter := bson.M{}
	if len(listoptions) > 0 {
		if limit, ok := listoptions[0].GetLimit(); ok {
			opt.SetLimit(limit)
		}
		if offset, ok := listoptions[0].GetOffset(); ok {
			opt.SetSkip(offset)
		}
		if deviceId, ok := listoptions[0].Get("device"); ok {
			filter[endpointDeviceKey] = deviceId
		}
		if serviceId, ok := listoptions[0].Get("service"); ok {
			filter[endpointServiceKey] = serviceId
		}
		if endpoint, ok := listoptions[0].Get("endpoint"); ok {
			filter[endpointEndpointKey] = endpoint
		}
		if protocol, ok := listoptions[0].Get("protocol"); ok {
			filter[endpointProtocolKey] = protocol
		}
		err = listoptions[0].EvalStrict()
		if err != nil {
			return result, err
		}
	}

	cursor, err := this.endpointCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		device := model.Endpoint{}
		err = cursor.Decode(&device)
		if err != nil {
			return nil, err
		}
		result = append(result, device)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) RemoveEndpoint(ctx context.Context, id string) error {
	_, err := this.endpointCollection().DeleteOne(ctx, bson.M{endpointIdKey: id})
	return err
}

func (this *Mongo) SetEndpoint(ctx context.Context, endpoint model.Endpoint) error {
	_, err := this.endpointCollection().ReplaceOne(ctx, bson.M{endpointIdKey: endpoint.Id}, endpoint, options.Replace().SetUpsert(true))
	return err
}
