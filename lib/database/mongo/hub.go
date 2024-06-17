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

package mongo

import (
	"context"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const hubIdFieldName = "Id"
const hubDeviceLocalIdFieldName = "DeviceLocalIds"
const hubOwnerIdFieldName = "OwnerId"

var hubIdKey string
var hubDeviceLocalIdKey string
var hubOwnerIdKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		hubIdKey, err = getBsonFieldName(models.Hub{}, hubIdFieldName)
		if err != nil {
			return err
		}
		hubDeviceLocalIdKey, err = getBsonFieldName(models.Hub{}, hubDeviceLocalIdFieldName)
		if err != nil {
			return err
		}
		hubOwnerIdKey, err = getBsonFieldName(models.Hub{}, hubOwnerIdFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoHubCollection)
		err = db.ensureIndex(collection, "hubidindex", hubIdKey, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "hubdevicelocalidindex", hubDeviceLocalIdKey, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) hubCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoHubCollection)
}

func (this *Mongo) GetHub(ctx context.Context, id string) (hub models.Hub, exists bool, err error) {
	result := this.hubCollection().FindOne(ctx, bson.M{hubIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return hub, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&hub)
	if err == mongo.ErrNoDocuments {
		return hub, false, nil
	}
	return hub, true, err
}

func (this *Mongo) SetHub(ctx context.Context, hub models.Hub) error {
	_, err := this.hubCollection().ReplaceOne(ctx, bson.M{hubIdKey: hub.Id}, hub, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveHub(ctx context.Context, id string) error {
	_, err := this.hubCollection().DeleteOne(ctx, bson.M{hubIdKey: id})
	return err
}

func (this *Mongo) GetHubsByDeviceLocalId(ctx context.Context, localId string) (hubs []models.Hub, err error) {
	cursor, err := this.hubCollection().Find(ctx, bson.M{hubDeviceLocalIdKey: localId})
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		hub := models.Hub{}
		err = cursor.Decode(&hub)
		if err != nil {
			return nil, err
		}
		hubs = append(hubs, hub)
	}
	err = cursor.Err()
	return hubs, err
}
