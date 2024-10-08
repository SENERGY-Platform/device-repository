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

const aspectIdFieldName = "Id"

var aspectIdKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		aspectIdKey, err = getBsonFieldName(models.Aspect{}, aspectIdFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoAspectCollection)
		err = db.ensureIndex(collection, "aspectidindex", aspectIdKey, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) aspectCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoAspectCollection)
}

func (this *Mongo) GetAspect(ctx context.Context, id string) (aspect models.Aspect, exists bool, err error) {
	result := this.aspectCollection().FindOne(ctx, bson.M{aspectIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return aspect, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&aspect)
	if err == mongo.ErrNoDocuments {
		return aspect, false, nil
	}
	return aspect, true, err
}

func (this *Mongo) SetAspect(ctx context.Context, aspect models.Aspect) error {
	_, err := this.aspectCollection().ReplaceOne(ctx, bson.M{aspectIdKey: aspect.Id}, aspect, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveAspect(ctx context.Context, id string) error {
	_, err := this.aspectCollection().DeleteOne(ctx, bson.M{aspectIdKey: id})
	return err
}

func (this *Mongo) ListAllAspects(ctx context.Context) (result []models.Aspect, err error) {
	cursor, err := this.aspectCollection().Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{aspectIdKey, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.Aspect{}
	for cursor.Next(context.Background()) {
		aspect := models.Aspect{}
		err = cursor.Decode(&aspect)
		if err != nil {
			return nil, err
		}
		result = append(result, aspect)
	}
	err = cursor.Err()
	return
}

// returns all aspects used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
func (this *Mongo) ListAspectsWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) (result []models.Aspect, err error) {
	aspectIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.AspectId, bson.M{
		deviceTypeCriteriaIsControllingFunctionKey: false,
		DeviceTypeCriteriaBson.AspectId:            bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	var cursor *mongo.Cursor
	if ancestors || descendants {
		or := bson.A{
			bson.D{{aspectNodeIdKey, bson.M{"$in": aspectIds}}},
		}
		if ancestors {
			or = append(or, bson.D{{aspectNodeAncestorIdsKey, bson.M{"$in": aspectIds}}})
		}
		if descendants {
			or = append(or, bson.D{{aspectNodeDescendentIdsKey, bson.M{"$in": aspectIds}}})
		}
		rootIds, err := this.aspectNodeCollection().Distinct(ctx, aspectNodeRootIdKey, bson.D{{"$or", or}})
		if err != nil {
			return nil, err
		}
		cursor, err = this.aspectCollection().Find(ctx, bson.M{aspectIdKey: bson.M{"$in": rootIds}}, options.Find().SetSort(bson.D{{aspectIdKey, 1}}))
		if err != nil {
			return nil, err
		}
	} else {
		cursor, err = this.aspectCollection().Find(ctx, bson.M{aspectIdKey: bson.M{"$in": aspectIds}}, options.Find().SetSort(bson.D{{aspectIdKey, 1}}))
		if err != nil {
			return nil, err
		}
	}
	defer cursor.Close(context.Background())
	result = []models.Aspect{}
	for cursor.Next(context.Background()) {
		aspect := models.Aspect{}
		err = cursor.Decode(&aspect)
		if err != nil {
			return nil, err
		}
		result = append(result, aspect)
	}
	err = cursor.Err()
	return
}
