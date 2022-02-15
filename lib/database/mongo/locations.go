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
)

const locationIdFieldName = "Id"

var locationIdKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		locationIdKey, err = getBsonFieldName(model.Location{}, locationIdFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoLocationCollection)
		err = db.ensureIndex(collection, "locationidindex", locationIdKey, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) locationCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoLocationCollection)
}

func (this *Mongo) GetLocation(ctx context.Context, id string) (location model.Location, exists bool, err error) {
	result := this.locationCollection().FindOne(ctx, bson.M{locationIdKey: id})
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

func (this *Mongo) SetLocation(ctx context.Context, location model.Location) error {
	_, err := this.locationCollection().ReplaceOne(ctx, bson.M{locationIdKey: location.Id}, location, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveLocation(ctx context.Context, id string) error {
	_, err := this.locationCollection().DeleteOne(ctx, bson.M{locationIdKey: id})
	return err
}
