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

const valueTypeFieldFieldName = "Fields"
const fieldTypeTypeFieldName = "Type"
const valueTypeNameFieldName = "Name"

var valueTypeFieldKey string
var fieldTypeTypeKey string
var valueTypeNameKey string

var valueTypeToValueTypePath string

func init() {
	var err error
	valueTypeNameKey, err = getBsonFieldName(model.ValueType{}, valueTypeNameFieldName)
	if err != nil {
		log.Fatal(err)
	}
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
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "valuetypenameindex", valueTypeNameKey, true, false)
		return err
	})
}

func (this *Mongo) valueTypeCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoValueTypeCollection)
}

func (this *Mongo) GetValueType(ctx context.Context, id string) (vt model.ValueType, exists bool, err error) {
	result := this.valueTypeCollection().FindOne(ctx, bson.M{valueTypeIdKey: id, "removed": bson.M{"$in": bson.A{nil, false}}})
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

func (this *Mongo) ListValueTypes(ctx context.Context, listoptions listoptions.ListOptions) (result []model.ValueType, err error) {
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
		sortby := valueTypeIdKey
		switch parts[0] {
		case "id":
			sortby = valueTypeIdKey
		case "name":
			sortby = valueTypeNameKey
		default:
			sortby = valueTypeIdKey
		}
		direction := int32(1)
		if len(parts) > 1 && parts[1] == "desc" {
			direction = int32(-1)
		}
		opt.SetSort(bsonx.Doc{{sortby, bsonx.Int32(direction)}})
	}
	cursor, err := this.valueTypeCollection().Find(ctx, bson.M{"removed": bson.M{"$in": bson.A{nil, false}}}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		valueType := model.ValueType{}
		err = cursor.Decode(&valueType)
		if err != nil {
			return nil, err
		}
		result = append(result, valueType)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) SetValueType(ctx context.Context, valueType model.ValueType) error {
	_, err := this.valueTypeCollection().UpdateOne(ctx, bson.M{valueTypeIdKey: valueType.Id}, bson.M{"$set": valueType}, options.Update().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveValueType(ctx context.Context, id string) error {
	_, err := this.valueTypeCollection().ReplaceOne(ctx, bson.M{valueTypeIdKey: id}, bson.M{valueTypeIdKey: id, "removed": true}, options.Replace().SetUpsert(true))
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
	cursor, err := this.valueTypeCollection().Find(ctx, bson.M{valueTypeToValueTypePath: id, "removed": bson.M{"$in": bson.A{nil, false}}}, opt)
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
