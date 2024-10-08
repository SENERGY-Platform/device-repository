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
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const functionIdFieldName = "Id"
const functionRdfTypeFieldName = "RdfType"
const functionConceptFieldName = "ConceptId"

var functionIdKey string
var functionRdfTypeKey string
var functionConceptKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		functionIdKey, err = getBsonFieldName(models.Function{}, functionIdFieldName)
		if err != nil {
			return err
		}
		functionRdfTypeKey, err = getBsonFieldName(models.Function{}, functionRdfTypeFieldName)
		if err != nil {
			return err
		}
		functionConceptKey, err = getBsonFieldName(models.Function{}, functionConceptFieldName)
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
		err = db.ensureIndex(collection, "functionconceptindex", functionConceptKey, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) functionCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoFunctionCollection)
}

func (this *Mongo) GetFunction(ctx context.Context, id string) (function models.Function, exists bool, err error) {
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

func (this *Mongo) SetFunction(ctx context.Context, function models.Function) error {
	_, err := this.functionCollection().ReplaceOne(ctx, bson.M{functionIdKey: function.Id}, function, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveFunction(ctx context.Context, id string) error {
	_, err := this.functionCollection().DeleteOne(ctx, bson.M{functionIdKey: id})
	return err
}

func (this *Mongo) ListAllFunctionsByType(ctx context.Context, rdfType string) (result []models.Function, err error) {
	cursor, err := this.functionCollection().Find(ctx, bson.M{functionRdfTypeKey: rdfType}, options.Find().SetSort(bson.D{{functionIdKey, 1}}))
	if err != nil {
		return nil, err
	}
	result = []models.Function{}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		function := models.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

// returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
func (this *Mongo) ListAllMeasuringFunctionsByAspect(ctx context.Context, aspect string, ancestors bool, descendants bool) (result []models.Function, err error) {
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
	functionIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.FunctionId, bson.M{
		deviceTypeCriteriaIsControllingFunctionKey: false,
		DeviceTypeCriteriaBson.AspectId:            aspectFilter,
		DeviceTypeCriteriaBson.FunctionId:          bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.functionCollection().Find(ctx, bson.M{functionIdKey: bson.M{"$in": functionIds}}, options.Find().SetSort(bson.D{{functionIdKey, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.Function{}
	for cursor.Next(context.Background()) {
		function := models.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListAllFunctionsByDeviceClass(ctx context.Context, class string) (result []models.Function, err error) {
	functionIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.FunctionId, bson.M{
		DeviceTypeCriteriaBson.DeviceClassId: class,
		DeviceTypeCriteriaBson.FunctionId:    bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.functionCollection().Find(ctx, bson.M{functionIdKey: bson.M{"$in": functionIds}}, options.Find().SetSort(bson.D{{functionIdKey, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.Function{}
	for cursor.Next(context.Background()) {
		function := models.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListAllControllingFunctionsByDeviceClass(ctx context.Context, class string) (result []models.Function, err error) {
	functionIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.FunctionId, bson.M{
		DeviceTypeCriteriaBson.DeviceClassId:       class,
		deviceTypeCriteriaIsControllingFunctionKey: true,
		DeviceTypeCriteriaBson.FunctionId:          bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.functionCollection().Find(ctx, bson.M{functionIdKey: bson.M{"$in": functionIds}}, options.Find().SetSort(bson.D{{functionIdKey, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.Function{}
	for cursor.Next(context.Background()) {
		function := models.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ConceptIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	filter := bson.M{
		functionConceptKey: id,
	}
	temp := this.functionCollection().FindOne(ctx, filter)
	err = temp.Err()
	if err == mongo.ErrNoDocuments {
		return false, nil, nil
	}
	if err != nil {
		return result, nil, err
	}
	function := models.Function{}
	_ = temp.Decode(&function)
	return true, []string{function.Id}, nil
}
