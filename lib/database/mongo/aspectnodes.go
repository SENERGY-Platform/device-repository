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
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sort"
)

var aspectNodeIdFieldName, aspectNodeIdKey = "Id", ""
var aspectNodeRootIdFieldName, aspectNodeRootIdKey = "RootId", ""
var aspectNodeDescendentIdsFieldName, aspectNodeDescendentIdsKey = "DescendentIds", ""
var aspectNodeAncestorIdsFieldName, aspectNodeAncestorIdsKey = "AncestorIds", ""

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		aspectNodeIdKey, err = getBsonFieldName(models.AspectNode{}, aspectNodeIdFieldName)
		if err != nil {
			return err
		}
		aspectNodeRootIdKey, err = getBsonFieldName(models.AspectNode{}, aspectNodeRootIdFieldName)
		if err != nil {
			return err
		}
		aspectNodeDescendentIdsKey, err = getBsonFieldName(models.AspectNode{}, aspectNodeDescendentIdsFieldName)
		if err != nil {
			return err
		}
		aspectNodeAncestorIdsKey, err = getBsonFieldName(models.AspectNode{}, aspectNodeAncestorIdsFieldName)
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

func (this *Mongo) GetAspectNode(ctx context.Context, id string) (aspectNode models.AspectNode, exists bool, err error) {
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

func (this *Mongo) SetAspectNode(ctx context.Context, aspectNode models.AspectNode) error {
	//_, err := this.aspectNodeCollection().InsertOne(ctx, aspectNode)
	_, err := this.aspectNodeCollection().ReplaceOne(ctx, bson.M{aspectNodeIdKey: aspectNode.Id}, aspectNode, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveAspectNodesByRootId(ctx context.Context, id string) error {
	_, err := this.aspectNodeCollection().DeleteMany(ctx, bson.M{aspectNodeRootIdKey: id})
	return err
}

func (this *Mongo) ListAllAspectNodes(ctx context.Context) (result []models.AspectNode, err error) {
	cursor, err := this.aspectNodeCollection().Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{aspectNodeIdKey, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.AspectNode{}
	for cursor.Next(context.Background()) {
		aspectNode := models.AspectNode{}
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

// returns all aspects used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
func (this *Mongo) ListAspectNodesWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) (result []models.AspectNode, err error) {
	aspectNodeIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.AspectId, bson.M{
		deviceTypeCriteriaIsControllingFunctionKey: false,
		DeviceTypeCriteriaBson.AspectId:            bson.M{"$exists": true, "$ne": ""},
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
	cursor, err := this.aspectNodeCollection().Find(ctx, bson.D{{"$or", or}}, options.Find().SetSort(bson.D{{aspectNodeIdKey, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.AspectNode{}
	for cursor.Next(context.Background()) {
		aspectNode := models.AspectNode{}
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

func sortSubIds(a *models.AspectNode) {
	sort.Strings(a.DescendentIds)
	sort.Strings(a.AncestorIds)
	sort.Strings(a.ChildIds)
}

func (this *Mongo) ListAspectNodesByIdList(ctx context.Context, ids []string) (result []models.AspectNode, err error) {
	cursor, err := this.aspectNodeCollection().Find(ctx, bson.M{aspectNodeIdKey: bson.M{"$in": ids}}, options.Find().SetSort(bson.D{{aspectNodeIdKey, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.AspectNode{}
	for cursor.Next(context.Background()) {
		aspectNode := models.AspectNode{}
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
