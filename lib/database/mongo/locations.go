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

var LocationBson = getBsonFieldObject[models.Location]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoLocationCollection)
		err = db.ensureIndex(collection, "locationidindex", LocationBson.Id, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) locationCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoLocationCollection)
}

func (this *Mongo) GetLocation(ctx context.Context, id string) (location models.Location, exists bool, err error) {
	result := this.locationCollection().FindOne(ctx, bson.M{LocationBson.Id: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return location, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&location)
	if err == mongo.ErrNoDocuments {
		return location, false, nil
	}
	return location, true, err
}

func (this *Mongo) SetLocation(ctx context.Context, location models.Location) error {
	_, err := this.locationCollection().ReplaceOne(ctx, bson.M{LocationBson.Id: location.Id}, location, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveLocation(ctx context.Context, id string) error {
	_, err := this.locationCollection().DeleteOne(ctx, bson.M{LocationBson.Id: id})
	return err
}

func (this *Mongo) ListLocations(ctx context.Context, listOptions model.LocationListOptions) (result []models.Location, total int64, err error) {
	opt := options.Find()
	if listOptions.Limit > 0 {
		opt.SetLimit(listOptions.Limit)
	}
	if listOptions.Offset > 0 {
		opt.SetSkip(listOptions.Offset)
	}

	if listOptions.SortBy == "" {
		listOptions.SortBy = LocationBson.Name + ".asc"
	}

	sortby := listOptions.SortBy
	sortby = strings.TrimSuffix(sortby, ".asc")
	sortby = strings.TrimSuffix(sortby, ".desc")

	direction := int32(1)
	if strings.HasSuffix(listOptions.SortBy, ".desc") {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{}
	if listOptions.Ids != nil {
		filter[LocationBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter["$or"] = []interface{}{
			bson.M{LocationBson.Name: bson.M{"$regex": escapedSearch, "$options": "i"}},
			bson.M{LocationBson.Description: bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	}

	cursor, err := this.locationCollection().Find(ctx, filter, opt)
	if err != nil {
		return result, total, err
	}
	result, err, _ = readCursorResult[models.Location](ctx, cursor)
	if err != nil {
		return result, total, err
	}
	total, err = this.locationCollection().CountDocuments(ctx, filter)
	if err != nil {
		return result, total, err
	}
	return result, total, err
}
