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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

const functionIdFieldName = "Id"
const functionRdfTypeFieldName = "RdfType"

var functionIdKey string
var functionRdfTypeKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		functionIdKey, err = getBsonFieldName(model.Function{}, functionIdFieldName)
		if err != nil {
			return err
		}
		functionRdfTypeKey, err = getBsonFieldName(model.Function{}, functionRdfTypeFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoFunctionCollection)
		err = db.ensureIndex(collection, "functionidindex", functionIdKey, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "functionrdftypeindex", functionRdfTypeKey, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) functionCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoFunctionCollection)
}

func (this *Mongo) GetFunction(ctx context.Context, id string) (function model.Function, exists bool, err error) {
	result := this.functionCollection().FindOne(ctx, bson.M{functionIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return function, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&function)
	if err == mongo.ErrNoDocuments {
		return function, false, nil
	}
	return function, true, err
}

func (this *Mongo) SetFunction(ctx context.Context, function model.Function) error {
	_, err := this.functionCollection().ReplaceOne(ctx, bson.M{functionIdKey: function.Id}, function, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveFunction(ctx context.Context, id string) error {
	_, err := this.functionCollection().DeleteOne(ctx, bson.M{functionIdKey: id})
	return err
}

func (this *Mongo) ListAllFunctionsByType(ctx context.Context, rdfType string) (result []model.Function, err error) {
	cursor, err := this.functionCollection().Find(ctx, bson.M{functionRdfTypeKey: rdfType}, options.Find().SetSort(bsonx.Doc{{functionIdKey, bsonx.Int32(1)}}))
	if err != nil {
		return nil, err
	}
	result = []model.Function{}
	for cursor.Next(context.Background()) {
		function := model.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

//returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
func (this *Mongo) ListAllMeasuringFunctionsByAspect(ctx context.Context, aspect string, ancestors bool, descendants bool) (result []model.Function, err error) {
	var aspectFilter interface{}
	if ancestors || descendants {
		relatedIds := []string{aspect}
		node, exists, err := this.GetAspectNode(ctx, aspect)
		if err != nil {
			return nil, err
		}
		if exists {
			if ancestors {
				relatedIds = append(relatedIds, node.AncestorIds...)
			}
			if descendants {
				relatedIds = append(relatedIds, node.DescendentIds...)
			}
			aspectFilter = bson.M{"$in": relatedIds}
		}
	} else {
		aspectFilter = aspect
	}
	functionIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, deviceTypeCriteriaFunctionIdKey, bson.M{
		deviceTypeCriteriaIsControllingFunctionKey: false,
		deviceTypeCriteriaAspectIdKey:              aspectFilter,
		deviceTypeCriteriaFunctionIdKey:            bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.functionCollection().Find(ctx, bson.M{functionIdKey: bson.M{"$in": functionIds}}, options.Find().SetSort(bsonx.Doc{{functionIdKey, bsonx.Int32(1)}}))
	if err != nil {
		return nil, err
	}
	result = []model.Function{}
	for cursor.Next(context.Background()) {
		function := model.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListAllFunctionsByDeviceClass(ctx context.Context, class string) (result []model.Function, err error) {
	functionIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, deviceTypeCriteriaFunctionIdKey, bson.M{
		deviceTypeCriteriaDeviceClassIdKey: class,
		deviceTypeCriteriaFunctionIdKey:    bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.functionCollection().Find(ctx, bson.M{functionIdKey: bson.M{"$in": functionIds}}, options.Find().SetSort(bsonx.Doc{{functionIdKey, bsonx.Int32(1)}}))
	if err != nil {
		return nil, err
	}
	result = []model.Function{}
	for cursor.Next(context.Background()) {
		function := model.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListAllControllingFunctionsByDeviceClass(ctx context.Context, class string) (result []model.Function, err error) {
	functionIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, deviceTypeCriteriaFunctionIdKey, bson.M{
		deviceTypeCriteriaDeviceClassIdKey:         class,
		deviceTypeCriteriaIsControllingFunctionKey: true,
		deviceTypeCriteriaFunctionIdKey:            bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.functionCollection().Find(ctx, bson.M{functionIdKey: bson.M{"$in": functionIds}}, options.Find().SetSort(bsonx.Doc{{functionIdKey, bsonx.Int32(1)}}))
	if err != nil {
		return nil, err
	}
	result = []model.Function{}
	for cursor.Next(context.Background()) {
		function := model.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}
