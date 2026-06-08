/*
 * Copyright 2026 InfAI (CC SES)
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
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var LastUpdateTimestampsBson = getBsonFieldObject[model.LastUpdateTimestamp]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoLastUpdateTimestampsCollection)
		err = db.ensureIndex(collection, "lastupdate_collection_index", LastUpdateTimestampsBson.Collection, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "lastupdate_userid_index", LastUpdateTimestampsBson.UserId, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) lastUpdateTimestampsCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoLastUpdateTimestampsCollection)
}

func (this *Mongo) GetLastUpdateTimestampsForUser(ctx context.Context, userId string) (result []model.LastUpdateTimestamp, err error) {
	c, err := this.lastUpdateTimestampsCollection().Find(ctx, bson.M{"$or": []bson.M{{LastUpdateTimestampsBson.UserId: ""}, {LastUpdateTimestampsBson.UserId: userId}}})
	if err != nil {
		return nil, err
	}
	err = c.All(ctx, &result)
	return result, err
}

func (this *Mongo) SetLastUpdateTimestamp(ctx context.Context, collection string, userId string) (err error) {
	timestamp := time.Now().Unix()
	filter := bson.M{LastUpdateTimestampsBson.Collection: collection, LastUpdateTimestampsBson.UserId: userId}
	_, err = this.lastUpdateTimestampsCollection().ReplaceOne(ctx, filter, model.LastUpdateTimestamp{
		UnixTimestamp: timestamp,
		Collection:    collection,
		UserId:        userId,
	}, options.Replace().SetUpsert(true))
	if err != nil {
		this.config.GetLogger().Error("unable to update last update timestamp", "collection", collection, "userid", userId, "error", err)
		return err
	}
	return err
}
