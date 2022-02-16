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

const characteristicIdFieldName = "Id"

var characteristicIdKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		characteristicIdKey, err = getBsonFieldName(model.Characteristic{}, characteristicIdFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoCharacteristicCollection)
		err = db.ensureIndex(collection, "characteristicidindex", characteristicIdKey, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) characteristicCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoCharacteristicCollection)
}

func (this *Mongo) GetCharacteristic(ctx context.Context, id string) (characteristic model.Characteristic, exists bool, err error) {
	result := this.characteristicCollection().FindOne(ctx, bson.M{characteristicIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return characteristic, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&characteristic)
	if err == mongo.ErrNoDocuments {
		return characteristic, false, nil
	}
	return characteristic, true, err
}

func (this *Mongo) SetCharacteristic(ctx context.Context, characteristic model.Characteristic) error {
	_, err := this.characteristicCollection().ReplaceOne(ctx, bson.M{characteristicIdKey: characteristic.Id}, characteristic, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveCharacteristic(ctx context.Context, id string) error {
	_, err := this.characteristicCollection().DeleteOne(ctx, bson.M{characteristicIdKey: id})
	return err
}

func (this *Mongo) ListAllCharacteristics(ctx context.Context) (result []model.Characteristic, err error) {
	cursor, err := this.characteristicCollection().Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		characteristic := model.Characteristic{}
		err = cursor.Decode(&characteristic)
		if err != nil {
			return nil, err
		}
		result = append(result, characteristic)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) getCharacteristicsByIds(ctx context.Context, ids []string) (result []model.Characteristic, err error) {
	if len(ids) == 0 {
		return []model.Characteristic{}, nil
	}
	cursor, err := this.characteristicCollection().Find(ctx, bson.M{characteristicIdKey: bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		characteristic := model.Characteristic{}
		err = cursor.Decode(&characteristic)
		if err != nil {
			return nil, err
		}
		result = append(result, characteristic)
	}
	err = cursor.Err()
	return
}