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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"regexp"
	"strings"
)

var HubBson = getBsonFieldObject[model.HubWithConnectionState]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoHubCollection)
		err = db.ensureIndex(collection, "hubidindex", HubBson.Id, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "hubnameindex", HubBson.Name, true, false) //to support faster sort
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "hubdeviceidindex", HubBson.DeviceIds[0], true, false)
		if err != nil {
			return err
		}
		err = db.removeIndex(collection, "hubdevicelocalidindex")
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) hubCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoHubCollection)
}

func (this *Mongo) GetHub(ctx context.Context, id string) (hub model.HubWithConnectionState, exists bool, err error) {
	result := this.hubCollection().FindOne(ctx, bson.M{HubBson.Id: id})
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

func (this *Mongo) SetHub(ctx context.Context, hub model.HubWithConnectionState) error {
	_, err := this.hubCollection().ReplaceOne(ctx, bson.M{HubBson.Id: hub.Id}, hub, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveHub(ctx context.Context, id string) error {
	_, err := this.hubCollection().DeleteOne(ctx, bson.M{HubBson.Id: id})
	return err
}

func (this *Mongo) GetHubsByDeviceId(ctx context.Context, id string) (hubs []model.HubWithConnectionState, err error) {
	cursor, err := this.hubCollection().Find(ctx, bson.M{HubBson.DeviceIds[0]: id})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(ctx) {
		hub := model.HubWithConnectionState{}
		err = cursor.Decode(&hub)
		if err != nil {
			return nil, err
		}
		hubs = append(hubs, hub)
	}
	err = cursor.Err()
	return hubs, err
}

func (this *Mongo) ListHubs(ctx context.Context, listOptions model.HubListOptions, withTotal bool) (result []model.HubWithConnectionState, total int64, err error) {
	opt := options.Find()
	if listOptions.Limit > 0 {
		opt.SetLimit(listOptions.Limit)
	}
	if listOptions.Offset > 0 {
		opt.SetSkip(listOptions.Offset)
	}

	if listOptions.SortBy == "" {
		listOptions.SortBy = HubBson.Name + ".asc"
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
		filter[HubBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		filter[HubBson.Name] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	if listOptions.ConnectionState != nil {
		filter[HubBson.ConnectionState] = listOptions.ConnectionState
	}

	if listOptions.LocalDeviceId != "" {
		filter[HubBson.DeviceLocalIds[0]] = listOptions.LocalDeviceId
	}
	if listOptions.OwnerId != "" {
		filter[HubBson.OwnerId] = listOptions.OwnerId
	}

	cursor, err := this.hubCollection().Find(ctx, filter, opt)
	if err != nil {
		return result, total, err
	}
	result, err, _ = readCursorResult[model.HubWithConnectionState](ctx, cursor)
	if err != nil {
		return result, total, err
	}
	if withTotal {
		total, err = this.hubCollection().CountDocuments(ctx, filter)
		if err != nil {
			return result, total, err
		}
	}
	return result, total, err
}

func (this *Mongo) SetHubConnectionState(ctx context.Context, id string, state models.ConnectionState) error {
	_, err := this.hubCollection().UpdateOne(ctx, bson.M{
		HubBson.Id: id,
	}, bson.M{
		"$set": bson.M{HubBson.ConnectionState: state},
	})
	return err
}
