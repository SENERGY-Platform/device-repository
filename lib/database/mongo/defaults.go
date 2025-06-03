/*
 * Copyright 2025 InfAI (CC SES)
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
	"errors"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DefaultDeviceAttributes struct {
	UserId     string             `json:"user_id" bson:"user_id"`
	Attributes []models.Attribute `json:"attributes" bson:"attributes"`
}

var DefaultDeviceAttributesBson = getBsonFieldObject[DefaultDeviceAttributes]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDefaultDeviceAttributesCollection)
		err = db.ensureIndex(collection, "defaultdeviceattributesuseridindex", DefaultDeviceAttributesBson.UserId, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) defaultDeviceAttributesCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDefaultDeviceAttributesCollection)
}

func (this *Mongo) GetDefaultDeviceAttributes(ctx context.Context, userId string) (attributes []models.Attribute, err error) {
	result := this.defaultDeviceAttributesCollection().FindOne(ctx, bson.M{DefaultDeviceAttributesBson.UserId: userId})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return []models.Attribute{}, nil
	}
	if err != nil {
		return attributes, err
	}
	entry := DefaultDeviceAttributes{}
	err = result.Decode(&entry)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return []models.Attribute{}, nil
	}
	return entry.Attributes, err
}

func (this *Mongo) SetDefaultDeviceAttributes(ctx context.Context, userId string, attributes []models.Attribute) (err error) {
	_, err = this.defaultDeviceAttributesCollection().ReplaceOne(ctx, bson.M{DefaultDeviceAttributesBson.UserId: userId}, DefaultDeviceAttributes{
		UserId:     userId,
		Attributes: attributes,
	}, options.Replace().SetUpsert(true))
	return err
}
