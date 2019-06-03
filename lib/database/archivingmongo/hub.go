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
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

const hubIdFieldName = "Id"

var hubIdKey string

func init() {
	var err error
	hubIdKey, err = getBsonFieldName(model.Hub{}, hubIdFieldName)
	if err != nil {
		log.Fatal(err)
	}
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoHubCollection)
		err = db.ensureIndex(collection, "hubidindex", hubIdKey, true, true)
		return err
	})
}

func (this *Mongo) hubCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoHubCollection)
}

func (this *Mongo) GetHub(ctx context.Context, id string) (hub model.Hub, exists bool, err error) {
	result := this.hubCollection().FindOne(ctx, bson.M{hubIdKey: id, "removed": bson.M{"$in": bson.A{nil, false}}})
	err = result.Err()
	if err != nil {
		return
	}
	err = result.Decode(&hub)
	if err == mongo.ErrNoDocuments {
		return hub, false, nil
	}
	return hub, true, err
}

func (this *Mongo) SetHub(ctx context.Context, hub model.Hub) error {
	_, err := this.hubCollection().UpdateOne(ctx, bson.M{hubIdKey: hub.Id}, bson.M{"$set": hub}, options.Update().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveHub(ctx context.Context, id string) error {
	_, err := this.hubCollection().ReplaceOne(ctx, bson.M{hubIdKey: id}, bson.M{hubIdKey: id, "removed": true}, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) ListHubs(ctx context.Context, listoptions listoptions.ListOptions) (result []model.Hub, err error) {
	opt := options.Find()
	if limit, ok := listoptions.GetLimit(); ok {
		opt.SetLimit(limit)
	}
	if offset, ok := listoptions.GetOffset(); ok {
		opt.SetSkip(offset)
	}
	cursor, err := this.hubCollection().Find(ctx, bson.M{"removed": bson.M{"$in": bson.A{nil, false}}}, opt)
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		hub := model.Hub{}
		err = cursor.Decode(&hub)
		if err != nil {
			return nil, err
		}
		result = append(result, hub)
	}
	err = cursor.Err()
	return
}
