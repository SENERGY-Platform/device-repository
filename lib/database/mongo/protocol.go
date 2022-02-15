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
	"strings"
)

const protocolIdFieldName = "Id"
const protocolNameFieldName = "Name"

var protocolIdKey string
var protocolNameKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		protocolIdKey, err = getBsonFieldName(model.Protocol{}, protocolIdFieldName)
		if err != nil {
			return err
		}
		protocolNameKey, err = getBsonFieldName(model.Protocol{}, protocolNameFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoProtocolCollection)
		err = db.ensureIndex(collection, "protocolidindex", protocolIdKey, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "protocolnameindex", protocolNameKey, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) protocolCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoProtocolCollection)
}

func (this *Mongo) GetProtocol(ctx context.Context, id string) (protocol model.Protocol, exists bool, err error) {
	result := this.protocolCollection().FindOne(ctx, bson.M{protocolIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return protocol, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&protocol)
	if err == mongo.ErrNoDocuments {
		return protocol, false, nil
	}
	return protocol, true, err
}

func (this *Mongo) ListProtocols(ctx context.Context, limit int64, offset int64, sort string) (result []model.Protocol, err error) {
	opt := options.Find()
	opt.SetLimit(limit)
	opt.SetSkip(offset)

	parts := strings.Split(sort, ".")
	sortby := protocolIdKey
	switch parts[0] {
	case "id":
		sortby = protocolIdKey
	case "name":
		sortby = protocolNameKey
	default:
		sortby = protocolIdKey
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bsonx.Doc{{sortby, bsonx.Int32(direction)}})

	cursor, err := this.protocolCollection().Find(ctx, bson.M{}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		protocol := model.Protocol{}
		err = cursor.Decode(&protocol)
		if err != nil {
			return nil, err
		}
		result = append(result, protocol)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) SetProtocol(ctx context.Context, protocol model.Protocol) error {
	_, err := this.protocolCollection().ReplaceOne(ctx, bson.M{protocolIdKey: protocol.Id}, protocol, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveProtocol(ctx context.Context, id string) error {
	_, err := this.protocolCollection().DeleteOne(ctx, bson.M{protocolIdKey: id})
	return err
}
