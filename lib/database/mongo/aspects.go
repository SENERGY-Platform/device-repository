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
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"regexp"
	"strings"
)

var AspectBson = getBsonFieldObject[models.Aspect]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoAspectCollection)
		err = db.ensureIndex(collection, "aspectidindex", AspectBson.Id, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) aspectCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoAspectCollection)
}

func (this *Mongo) ListAspects(ctx context.Context, listOptions model.AspectListOptions) (result []models.Aspect, total int64, err error) {
	opt := options.Find()
	opt.SetLimit(listOptions.Limit)
	opt.SetSkip(listOptions.Offset)

	parts := strings.Split(listOptions.SortBy, ".")
	sortby := AspectBson.Id
	switch parts[0] {
	case "id":
		sortby = AspectBson.Id
	case "name":
		sortby = AspectBson.Name
	default:
		sortby = AspectBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{}
	if listOptions.Ids != nil {
		filter[AspectBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter[AspectBson.Name] = bson.M{"$regex": escapedSearch, "$options": "i"}
	}

	cursor, err := this.aspectCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.aspectCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) GetAspect(ctx context.Context, id string) (aspect models.Aspect, exists bool, err error) {
	result := this.aspectCollection().FindOne(ctx, bson.M{AspectBson.Id: id})
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
	_, err := this.aspectCollection().ReplaceOne(ctx, bson.M{AspectBson.Id: aspect.Id}, aspect, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveAspect(ctx context.Context, id string) error {
	_, err := this.aspectCollection().DeleteOne(ctx, bson.M{AspectBson.Id: id})
	return err
}

func (this *Mongo) ListAllAspects(ctx context.Context) (result []models.Aspect, err error) {
	cursor, err := this.aspectCollection().Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{AspectBson.Id, 1}}))
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
			bson.D{{AspectNodeBson.Id, bson.M{"$in": aspectIds}}},
		}
		if ancestors {
			or = append(or, bson.D{{aspectNodeAncestorIdsKey, bson.M{"$in": aspectIds}}})
		}
		if descendants {
			or = append(or, bson.D{{aspectNodeDescendentIdsKey, bson.M{"$in": aspectIds}}})
		}
		if len(or) == 0 {
			return []models.Aspect{}, nil
		}
		rootIds, err := this.aspectNodeCollection().Distinct(ctx, AspectNodeBson.RootId, bson.D{{"$or", or}})
		if err != nil {
			return nil, err
		}
		cursor, err = this.aspectCollection().Find(ctx, bson.M{AspectBson.Id: bson.M{"$in": rootIds}}, options.Find().SetSort(bson.D{{AspectBson.Id, 1}}))
		if err != nil {
			return nil, err
		}
	} else {
		cursor, err = this.aspectCollection().Find(ctx, bson.M{AspectBson.Id: bson.M{"$in": aspectIds}}, options.Find().SetSort(bson.D{{AspectBson.Id, 1}}))
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
