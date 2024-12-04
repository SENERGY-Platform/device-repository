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
	"slices"
	"strings"
)

var ConceptBson = getBsonFieldObject[models.Concept]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoConceptCollection)
		err = db.ensureIndex(collection, "conceptidindex", ConceptBson.Id, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "conceptcharacteristicsindex", ConceptBson.CharacteristicIds[0], true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) conceptCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoConceptCollection)
}

func (this *Mongo) SetConcept(ctx context.Context, concept models.Concept) error {
	_, err := this.conceptCollection().ReplaceOne(ctx, bson.M{ConceptBson.Id: concept.Id}, concept, options.Replace().SetUpsert(true))
	return err
}

func (this *Mongo) RemoveConcept(ctx context.Context, id string) error {
	_, err := this.conceptCollection().DeleteOne(ctx, bson.M{ConceptBson.Id: id})
	return err
}

func (this *Mongo) ListConceptsWithCharacteristics(ctx context.Context, listOptions model.ConceptListOptions) (result []models.ConceptWithCharacteristics, total int64, err error) {
	var temp []models.Concept
	temp, total, err = this.ListConcepts(ctx, listOptions)
	if err != nil {
		return result, total, err
	}
	characteristicIds := []string{}
	for _, concept := range temp {
		for _, characteristicId := range concept.CharacteristicIds {
			if !slices.Contains(characteristicIds, characteristicId) {
				characteristicIds = append(characteristicIds, characteristicId)
			}
		}
	}
	characteristics, err := this.getCharacteristicsByIds(ctx, characteristicIds)
	if err != nil {
		return result, total, err
	}
	characteristicsMap := map[string]models.Characteristic{}
	for _, characteristic := range characteristics {
		characteristicsMap[characteristic.Id] = characteristic
	}
	for _, concept := range temp {
		element := models.ConceptWithCharacteristics{
			Id:                   concept.Id,
			Name:                 concept.Name,
			BaseCharacteristicId: concept.BaseCharacteristicId,
			Characteristics:      []models.Characteristic{},
			Conversions:          concept.Conversions,
		}
		for _, characteristicId := range concept.CharacteristicIds {
			element.Characteristics = append(element.Characteristics, characteristicsMap[characteristicId])
		}
		result = append(result, element)
	}
	return result, total, nil
}

func (this *Mongo) ListConcepts(ctx context.Context, listOptions model.ConceptListOptions) (result []models.Concept, total int64, err error) {
	opt := options.Find()
	opt.SetLimit(listOptions.Limit)
	opt.SetSkip(listOptions.Offset)

	parts := strings.Split(listOptions.SortBy, ".")
	sortby := ConceptBson.Id
	switch parts[0] {
	case "id":
		sortby = ConceptBson.Id
	case "name":
		sortby = ConceptBson.Name
	default:
		sortby = ConceptBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{}
	if listOptions.Ids != nil {
		filter[ConceptBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter[ConceptBson.Name] = bson.M{"$regex": escapedSearch, "$options": "i"}
	}

	cursor, err := this.conceptCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.conceptCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) GetConceptWithCharacteristics(ctx context.Context, id string) (concept models.ConceptWithCharacteristics, exists bool, err error) {
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
	concept.Conversions = temp.Conversions
	return concept, exists, err
}

func (this *Mongo) GetConceptWithoutCharacteristics(ctx context.Context, id string) (concept models.Concept, exists bool, err error) {
	result := this.conceptCollection().FindOne(ctx, bson.M{ConceptBson.Id: id})
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
