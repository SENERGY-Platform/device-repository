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

const conceptIdFieldName = "Id"

var conceptIdKey string

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		conceptIdKey, err = getBsonFieldName(model.Concept{}, conceptIdFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoConceptCollection)
		err = db.ensureIndex(collection, "conceptidindex", conceptIdKey, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) conceptCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoConceptCollection)
}

func (this *Mongo) SetConcept(ctx context.Context, concept model.Concept) error {
	_, err := this.conceptCollection().ReplaceOne(ctx, bson.M{conceptIdKey: concept.Id}, concept, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveConcept(ctx context.Context, id string) error {
	_, err := this.conceptCollection().DeleteOne(ctx, bson.M{conceptIdKey: id})
	return err
}

func (this *Mongo) GetConceptWithCharacteristics(ctx context.Context, id string) (concept model.ConceptWithCharacteristics, exists bool, err error) {
	temp, exists, err := this.GetConceptWithoutCharacteristics(ctx, id)
	if err != nil {
		return concept, exists, err
	}
	if !exists {
		return concept, exists, err
	}
	concept.Id = temp.Id
	concept.Name = temp.Name
	concept.BaseCharacteristicId = temp.BaseCharacteristicId
	concept.Characteristics, err = this.getCharacteristicsByIds(ctx, temp.CharacteristicIds)
	return concept, exists, err
}

func (this *Mongo) GetConceptWithoutCharacteristics(ctx context.Context, id string) (concept model.Concept, exists bool, err error) {
	result := this.conceptCollection().FindOne(ctx, bson.M{conceptIdKey: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return concept, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&concept)
	if err == mongo.ErrNoDocuments {
		return concept, false, nil
	}
	return concept, true, err
}
