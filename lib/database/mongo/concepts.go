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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"regexp"
	"slices"
	"strings"
	"time"
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

func (this *Mongo) SetConcept(ctx context.Context, concept models.Concept, syncHandler func(models.Concept) error) error {
	timestamp := time.Now().Unix()
	collection := this.conceptCollection()
	_, err := this.conceptCollection().ReplaceOne(ctx, bson.M{ConceptBson.Id: concept.Id}, ConceptWithSyncInfo{
		Concept: concept,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = syncHandler(concept)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::syncHandler %v, will be retried later\n", err)
		return nil
	}
	err = this.setSynced(ctx, collection, ConceptBson.Id, concept.Id, timestamp)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::setSynced %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RemoveConcept(ctx context.Context, id string, syncDeleteHandler func(models.Concept) error) error {
	old, exists, err := this.GetConceptWithoutCharacteristics(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.conceptCollection()
	err = this.setDeleted(ctx, collection, ConceptBson.Id, id)
	if err != nil {
		return err
	}
	err = syncDeleteHandler(old)
	if err != nil {
		log.Printf("WARNING: error in RemoveConcept::syncDeleteHandler %v, will be retried later\n", err)
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{ConceptBson.Id: id})
	if err != nil {
		log.Printf("WARNING: error in RemoveConcept::DeleteOne %v, will be retried later\n", err)
		return nil
	}
	return nil
}

type ConceptWithSyncInfo struct {
	models.Concept `bson:",inline"`
	SyncInfo       `bson:",inline"`
}

func (this *Mongo) RetryConceptSync(lockduration time.Duration, syncDeleteHandler func(models.Concept) error, syncHandler func(models.Concept) error) error {
	collection := this.conceptCollection()
	jobs, err := FetchSyncJobs[ConceptWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.Concept)
			if err != nil {
				log.Printf("WARNING: error in RetryConceptSync::syncDeleteHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{ConceptBson.Id: job.Id})
			if err != nil {
				log.Printf("WARNING: error in RetryConceptSync::DeleteOne %v, will be retried later\n", err)
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.Concept)
			if err != nil {
				log.Printf("WARNING: error in RetryConceptSync::syncHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, ConceptBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				log.Printf("WARNING: error in RetryConceptSync::setSynced %v, will be retried later\n", err)
				continue
			}
		}
	}
	return nil
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

	filter := bson.M{NotDeletedFilterKey: NotDeletedFilterValue}
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
	result := this.conceptCollection().FindOne(ctx, bson.M{ConceptBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return concept, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&concept)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return concept, false, nil
	}
	return concept, true, err
}
