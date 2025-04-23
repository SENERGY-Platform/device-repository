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

var DeviceClassBson = getBsonFieldObject[models.DeviceClass]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceClassCollection)
		err = db.ensureIndex(collection, "deviceclassidindex", DeviceClassBson.Id, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceClassCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceClassCollection)
}

func (this *Mongo) ListDeviceClasses(ctx context.Context, listOptions model.DeviceClassListOptions) (result []models.DeviceClass, total int64, err error) {
	opt := options.Find()
	opt.SetLimit(listOptions.Limit)
	opt.SetSkip(listOptions.Offset)

	parts := strings.Split(listOptions.SortBy, ".")
	sortby := DeviceClassBson.Id
	switch parts[0] {
	case "id":
		sortby = DeviceClassBson.Id
	case "name":
		sortby = DeviceClassBson.Name
	default:
		sortby = DeviceClassBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{NotDeletedFilterKey: NotDeletedFilterValue}
	if listOptions.Ids != nil {
		filter[DeviceClassBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter[DeviceClassBson.Name] = bson.M{"$regex": escapedSearch, "$options": "i"}
	}

	cursor, err := this.deviceClassCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.deviceClassCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) GetDeviceClass(ctx context.Context, id string) (deviceClass models.DeviceClass, exists bool, err error) {
	result := this.deviceClassCollection().FindOne(ctx, bson.M{DeviceClassBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return deviceClass, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&deviceClass)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return deviceClass, false, nil
	}
	return deviceClass, true, err
}

func (this *Mongo) SetDeviceClass(ctx context.Context, deviceClass models.DeviceClass, syncHandler func(models.DeviceClass) error) error {
	timestamp := time.Now().Unix()
	collection := this.deviceClassCollection()
	_, err := this.deviceClassCollection().ReplaceOne(ctx, bson.M{DeviceClassBson.Id: deviceClass.Id}, DeviceClassWithSyncInfo{
		DeviceClass: deviceClass,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = syncHandler(deviceClass)
	if err != nil {
		log.Printf("WARNING: error in SetDeviceClass::syncHandler %v, will be retried later\n", err)
		return nil
	}
	err = this.setSynced(ctx, collection, DeviceClassBson.Id, deviceClass.Id, timestamp)
	if err != nil {
		log.Printf("WARNING: error in SetDeviceClass::setSynced %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RemoveDeviceClass(ctx context.Context, id string, syncDeleteHandler func(models.DeviceClass) error) error {
	old, exists, err := this.GetDeviceClass(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.deviceClassCollection()
	err = this.setDeleted(ctx, collection, DeviceClassBson.Id, id)
	if err != nil {
		return err
	}
	err = syncDeleteHandler(old)
	if err != nil {
		log.Printf("WARNING: error in RemoveDeviceClass::syncDeleteHandler %v, will be retried later\n", err)
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{DeviceClassBson.Id: id})
	if err != nil {
		log.Printf("WARNING: error in RemoveDeviceClass::DeleteOne %v, will be retried later\n", err)
		return nil
	}
	return nil
}

type DeviceClassWithSyncInfo struct {
	models.DeviceClass `bson:",inline"`
	SyncInfo           `bson:",inline"`
}

func (this *Mongo) RetryDeviceClassSync(lockduration time.Duration, syncDeleteHandler func(models.DeviceClass) error, syncHandler func(models.DeviceClass) error) error {
	collection := this.deviceClassCollection()
	jobs, err := FetchSyncJobs[DeviceClassWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.DeviceClass)
			if err != nil {
				log.Printf("WARNING: error in RetryDeviceClassSync::syncDeleteHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{DeviceClassBson.Id: job.Id})
			if err != nil {
				log.Printf("WARNING: error in RetryDeviceClassSync::DeleteOne %v, will be retried later\n", err)
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.DeviceClass)
			if err != nil {
				log.Printf("WARNING: error in RetryDeviceClassSync::syncHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, DeviceClassBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				log.Printf("WARNING: error in RetryDeviceClassSync::setSynced %v, will be retried later\n", err)
				continue
			}
		}
	}
	return nil
}

func (this *Mongo) ListAllDeviceClasses(ctx context.Context) (result []models.DeviceClass, err error) {
	cursor, err := this.deviceClassCollection().Find(ctx, bson.M{NotDeletedFilterKey: NotDeletedFilterValue})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		deviceClass := models.DeviceClass{}
		err = cursor.Decode(&deviceClass)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceClass)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListAllDeviceClassesUsedWithControllingFunctions(ctx context.Context) (result []models.DeviceClass, err error) {
	deviceClassIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.DeviceClassId, bson.M{
		deviceTypeCriteriaIsControllingFunctionKey: true,
		DeviceTypeCriteriaBson.DeviceClassId:       bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.deviceClassCollection().Find(ctx, bson.M{DeviceClassBson.Id: bson.M{"$in": deviceClassIds}, NotDeletedFilterKey: NotDeletedFilterValue})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		deviceClass := models.DeviceClass{}
		err = cursor.Decode(&deviceClass)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceClass)
	}
	err = cursor.Err()
	return
}
