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
	"strings"
	"time"
)

var CharacteristicBson = getBsonFieldObject[models.Characteristic]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoCharacteristicCollection)
		err = db.ensureIndex(collection, "characteristicidindex", CharacteristicBson.Id, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) characteristicCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoCharacteristicCollection)
}

func (this *Mongo) ListCharacteristics(ctx context.Context, listOptions model.CharacteristicListOptions) (result []models.Characteristic, total int64, err error) {
	opt := options.Find()
	opt.SetLimit(listOptions.Limit)
	opt.SetSkip(listOptions.Offset)

	parts := strings.Split(listOptions.SortBy, ".")
	sortby := CharacteristicBson.Id
	switch parts[0] {
	case "id":
		sortby = CharacteristicBson.Id
	case "name":
		sortby = CharacteristicBson.Name
	default:
		sortby = CharacteristicBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{NotDeletedFilterKey: NotDeletedFilterValue}
	if listOptions.Ids != nil {
		filter[CharacteristicBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter[CharacteristicBson.Name] = bson.M{"$regex": escapedSearch, "$options": "i"}
	}

	cursor, err := this.characteristicCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.characteristicCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) GetCharacteristic(ctx context.Context, id string) (characteristic models.Characteristic, exists bool, err error) {
	result := this.characteristicCollection().FindOne(ctx, bson.M{CharacteristicBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return characteristic, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&characteristic)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return characteristic, false, nil
	}
	return characteristic, true, err
}

func (this *Mongo) SetCharacteristic(ctx context.Context, characteristic models.Characteristic, syncHandler func(models.Characteristic) error) error {
	timestamp := time.Now().Unix()
	collection := this.characteristicCollection()
	_, err := this.characteristicCollection().ReplaceOne(ctx, bson.M{CharacteristicBson.Id: characteristic.Id}, CharacteristicWithSyncInfo{
		Characteristic: characteristic,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = syncHandler(characteristic)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::syncHandler %v, will be retried later\n", err)
		return nil
	}
	err = this.setSynced(ctx, collection, CharacteristicBson.Id, characteristic.Id, timestamp)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::setSynced %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RemoveCharacteristic(ctx context.Context, id string, syncDeleteHandler func(models.Characteristic) error) error {
	old, exists, err := this.GetCharacteristic(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.characteristicCollection()
	err = this.setDeleted(ctx, collection, CharacteristicBson.Id, id)
	if err != nil {
		return err
	}
	err = syncDeleteHandler(old)
	if err != nil {
		log.Printf("WARNING: error in RemoveCharacteristic::syncDeleteHandler %v, will be retried later\n", err)
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{CharacteristicBson.Id: id})
	if err != nil {
		log.Printf("WARNING: error in RemoveCharacteristic::DeleteOne %v, will be retried later\n", err)
		return nil
	}
	return nil
}

type CharacteristicWithSyncInfo struct {
	models.Characteristic `bson:",inline"`
	SyncInfo              `bson:",inline"`
}

func (this *Mongo) RetryCharacteristicSync(lockduration time.Duration, syncDeleteHandler func(models.Characteristic) error, syncHandler func(models.Characteristic) error) error {
	collection := this.characteristicCollection()
	jobs, err := FetchSyncJobs[CharacteristicWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.Characteristic)
			if err != nil {
				log.Printf("WARNING: error in RetryCharacteristicSync::syncDeleteHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{CharacteristicBson.Id: job.Id})
			if err != nil {
				log.Printf("WARNING: error in RetryCharacteristicSync::DeleteOne %v, will be retried later\n", err)
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.Characteristic)
			if err != nil {
				log.Printf("WARNING: error in RetryCharacteristicSync::syncHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, CharacteristicBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				log.Printf("WARNING: error in RetryCharacteristicSync::setSynced %v, will be retried later\n", err)
				continue
			}
		}
	}
	return nil
}

func (this *Mongo) ListAllCharacteristics(ctx context.Context) (result []models.Characteristic, err error) {
	cursor, err := this.characteristicCollection().Find(ctx, bson.M{NotDeletedFilterKey: NotDeletedFilterValue})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		characteristic := models.Characteristic{}
		err = cursor.Decode(&characteristic)
		if err != nil {
			return nil, err
		}
		result = append(result, characteristic)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) getCharacteristicsByIds(ctx context.Context, ids []string) (result []models.Characteristic, err error) {
	if len(ids) == 0 {
		return []models.Characteristic{}, nil
	}
	cursor, err := this.characteristicCollection().Find(ctx, bson.M{CharacteristicBson.Id: bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		characteristic := models.Characteristic{}
		err = cursor.Decode(&characteristic)
		if err != nil {
			return nil, err
		}
		result = append(result, characteristic)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) CharacteristicIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	//used in device-type
	filter := bson.M{
		DeviceTypeCriteriaBson.CharacteristicId: id,
	}
	temp := this.deviceTypeCriteriaCollection().FindOne(ctx, filter)
	err = temp.Err()
	if err != nil && err != mongo.ErrNoDocuments {
		return result, nil, err
	}
	if err == nil {
		criteria := model.DeviceTypeCriteria{}
		_ = temp.Decode(&criteria)
		return true, []string{criteria.DeviceTypeId, criteria.ContentVariableId, criteria.ContentVariablePath}, nil
	}

	//used in concept
	temp = this.conceptCollection().FindOne(ctx, bson.M{
		ConceptBson.CharacteristicIds[0]: id,
	})
	err = temp.Err()
	if err != nil && err != mongo.ErrNoDocuments {
		return result, nil, err
	}
	if err == nil {
		concept := models.Concept{}
		_ = temp.Decode(&concept)
		return true, []string{concept.Id, concept.Name}, nil
	}
	return false, nil, nil
}

func (this *Mongo) CharacteristicIsUsedWithConceptInDeviceType(ctx context.Context, characteristicId string, conceptId string) (result bool, where []string, err error) {
	filter := bson.M{
		DeviceTypeCriteriaBson.CharacteristicId: characteristicId,
	}
	temp := this.deviceTypeCriteriaCollection().FindOne(ctx, filter)
	err = temp.Err()
	if err != nil && err != mongo.ErrNoDocuments {
		return result, nil, err
	}
	if err == nil {
		criteria := model.DeviceTypeCriteria{}
		_ = temp.Decode(&criteria)
		if criteria.FunctionId != "" {
			f, exists, err := this.GetFunction(ctx, criteria.FunctionId)
			if err != nil {
				return result, where, err
			}
			if exists && f.ConceptId == conceptId {
				return true, []string{criteria.DeviceTypeId, criteria.ContentVariableId, criteria.ContentVariablePath}, nil
			}
		}
	}
	return false, nil, nil
}
