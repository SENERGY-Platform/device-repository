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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"sort"
)

var aspectNodeIdFieldName, aspectNodeIdKey = "Id", ""
var aspectNodeRootIdFieldName, aspectNodeRootIdKey = "RootId", ""
var aspectNodeDescendentIdsFieldName, aspectNodeDescendentIdsKey = "DescendentIds", ""
var aspectNodeAncestorIdsFieldName, aspectNodeAncestorIdsKey = "AncestorIds", ""

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		aspectNodeIdKey, err = getBsonFieldName(model.AspectNode{}, aspectNodeIdFieldName)
		if err != nil {
			return err
		}
		aspectNodeRootIdKey, err = getBsonFieldName(model.AspectNode{}, aspectNodeRootIdFieldName)
		if err != nil {
			return err
		}
		aspectNodeDescendentIdsKey, err = getBsonFieldName(model.AspectNode{}, aspectNodeDescendentIdsFieldName)
		if err != nil {
			return err
		}
		aspectNodeAncestorIdsKey, err = getBsonFieldName(model.AspectNode{}, aspectNodeAncestorIdsFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(getAspectNodeCollectionName(db.config))
		err = db.ensureIndex(collection, "aspectNodeidindex", aspectNodeIdKey, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "aspectNoderootidindex", aspectNodeRootIdKey, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "aspectNodeDescendentIdsIndex", aspectNodeDescendentIdsKey, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "aspectNodeAncestorIdsIndex", aspectNodeAncestorIdsKey, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func getAspectNodeCollectionName(config config.Config) string {
	return config.MongoAspectCollection + "_node"
}

func (this *Mongo) aspectNodeCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(getAspectNodeCollectionName(this.config))
}

func (this *Mongo) GetAspectNode(ctx context.Context, id string) (aspectNode model.AspectNode, exists bool, err error) {
	result := this.aspectNodeCollection().FindOne(ctx, bson.M{aspectNodeIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return aspectNode, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&aspectNode)
	if err == mongo.ErrNoDocuments {
		return aspectNode, false, nil
	}
	sortSubIds(&aspectNode)
	return aspectNode, true, err
}

func (this *Mongo) AddAspectNode(ctx context.Context, aspectNode model.AspectNode) error {
	_, err := this.aspectNodeCollection().InsertOne(ctx, aspectNode)
	return err
}

func (this *Mongo) RemoveAspectNodesByRootId(ctx context.Context, id string) error {
	_, err := this.aspectNodeCollection().DeleteMany(ctx, bson.M{aspectNodeRootIdKey: id})
	return err
}

func (this *Mongo) ListAllAspectNodes(ctx context.Context) (result []model.AspectNode, err error) {
	cursor, err := this.aspectNodeCollection().Find(ctx, bson.D{}, options.Find().SetSort(bsonx.Doc{{aspectNodeIdKey, bsonx.Int32(1)}}))
	if err != nil {
		return nil, err
	}
	result = []model.AspectNode{}
	for cursor.Next(context.Background()) {
		aspectNode := model.AspectNode{}
		err = cursor.Decode(&aspectNode)
		if err != nil {
			return nil, err
		}
		sortSubIds(&aspectNode)
		result = append(result, aspectNode)
	}
	err = cursor.Err()
	return
}

//returns all aspects used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
func (this *Mongo) ListAspectNodesWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) (result []model.AspectNode, err error) {
	aspectNodeIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, deviceTypeCriteriaAspectIdKey, bson.M{
		deviceTypeCriteriaIsControllingFunctionKey: false,
		deviceTypeCriteriaAspectIdKey:              bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	or := bson.A{
		bson.D{{aspectNodeIdKey, bson.M{"$in": aspectNodeIds}}},
	}
	if ancestors {
		or = append(or, bson.D{{aspectNodeAncestorIdsKey, bson.M{"$in": aspectNodeIds}}})
	}
	if descendants {
		or = append(or, bson.D{{aspectNodeDescendentIdsKey, bson.M{"$in": aspectNodeIds}}})
	}
	cursor, err := this.aspectNodeCollection().Find(ctx, bson.D{{"$or", or}}, options.Find().SetSort(bsonx.Doc{{aspectNodeIdKey, bsonx.Int32(1)}}))
	if err != nil {
		return nil, err
	}
	result = []model.AspectNode{}
	for cursor.Next(context.Background()) {
		aspectNode := model.AspectNode{}
		err = cursor.Decode(&aspectNode)
		if err != nil {
			return nil, err
		}
		sortSubIds(&aspectNode)
		result = append(result, aspectNode)
	}
	err = cursor.Err()
	return
}

func sortSubIds(a *model.AspectNode) {
	sort.Strings(a.DescendentIds)
	sort.Strings(a.AncestorIds)
	sort.Strings(a.ChildIds)
}
