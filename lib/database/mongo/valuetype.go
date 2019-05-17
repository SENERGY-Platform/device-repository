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
	"strings"
)

const valueTypeFieldFieldName = "Fields"
const fieldTypeTypeFieldName = "Type"

var valueTypeFieldKey string
var fieldTypeTypeKey string

var valueTypeToValueTypePath string

func init() {
	var err error
	valueTypeFieldKey, err = getBsonFieldName(model.ValueType{}, valueTypeFieldFieldName)
	if err != nil {
		log.Fatal(err)
	}
	fieldTypeTypeKey, err = getBsonFieldName(model.FieldType{}, fieldTypeTypeFieldName)
	if err != nil {
		log.Fatal(err)
	}

	valueTypeToValueTypePath = strings.Join([]string{valueTypeFieldKey, fieldTypeTypeKey, valueTypeIdKey}, ".")

	// valueTypeIdKey and valueTypeIdFieldName are defined in devicetype.go
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoValueTypeCollection)
		err = db.ensureIndex(collection, "valuetypeidindex", valueTypeIdKey, true, true)
		return err
	})
}

func (this *Mongo) valueTypeCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoValueTypeCollection)
}

func (this *Mongo) GetValueType(ctx context.Context, id string) (vt model.ValueType, exists bool, err error) {
	result := this.valueTypeCollection().FindOne(ctx, bson.D{{valueTypeIdKey, id}})
	err = result.Err()
	if err != nil {
		return
	}
	err = result.Decode(&vt)
	if err == mongo.ErrNoDocuments {
		return vt, false, nil
	}
	return vt, true, err
}

func (this *Mongo) SetValueType(ctx context.Context, valueType model.ValueType) error {
	_, err := this.valueTypeCollection().ReplaceOne(ctx, bson.M{valueTypeIdKey: valueType.Id}, valueType, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveValueType(ctx context.Context, id string) error {
	_, err := this.valueTypeCollection().DeleteOne(ctx, bson.M{valueTypeIdKey: id})
	return err
}

func (this *Mongo) ListValueTypesUsingValueType(ctx context.Context, id string, listoptions ...listoptions.ListOptions) (result []model.ValueType, err error) {
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
	cursor, err := this.valueTypeCollection().Find(ctx, bson.M{valueTypeToValueTypePath: id}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		vt := model.ValueType{}
		err = cursor.Decode(&vt)
		if err != nil {
			return nil, err
		}
		result = append(result, vt)
	}
	err = cursor.Err()
	return
}
