/*
 * Copyright 2026 InfAI (CC SES)
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
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var AspectClassBson = getBsonFieldObject[models.AspectClass]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoAspectClassCollection)
		err = db.ensureIndex(collection, "aspectclassidindex", AspectClassBson.Id, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) aspectClassCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoAspectClassCollection)
}

func (this *Mongo) ListAspectClasses(ctx context.Context, listOptions model.AspectClassListOptions) (result []models.AspectClass, total int64, err error) {
	opt := options.Find()
	opt.SetLimit(listOptions.Limit)
	opt.SetSkip(listOptions.Offset)

	parts := strings.Split(listOptions.SortBy, ".")
	sortby := AspectClassBson.Id
	switch parts[0] {
	case "id":
		sortby = AspectClassBson.Id
	case "name":
		sortby = AspectClassBson.Name
	default:
		sortby = AspectClassBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{NotDeletedFilterKey: NotDeletedFilterValue}
	if listOptions.Ids != nil {
		filter[AspectClassBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter[AspectClassBson.Name] = bson.M{"$regex": escapedSearch, "$options": "i"}
	}

	cursor, err := this.aspectClassCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.aspectClassCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) GetAspectClass(ctx context.Context, id string) (aspectClass models.AspectClass, exists bool, err error) {
	result := this.aspectClassCollection().FindOne(ctx, bson.M{AspectClassBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return aspectClass, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&aspectClass)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return aspectClass, false, nil
	}
	return aspectClass, true, err
}

func (this *Mongo) SetAspectClass(ctx context.Context, aspectClass models.AspectClass, syncHandler func(models.AspectClass) error) error {
	timestamp := time.Now().Unix()
	collection := this.aspectClassCollection()
	_, err := collection.ReplaceOne(ctx, bson.M{AspectClassBson.Id: aspectClass.Id}, AspectClassWithSyncInfo{
		AspectClass: aspectClass,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = this.SetLastUpdateTimestamp(ctx, this.config.MongoAspectClassCollection, "")
	if err != nil {
		err = nil
	}
	err = syncHandler(aspectClass)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in SetAspectClass::syncHandler %v, will be retried later\n", err))
		return nil
	}
	err = this.setSynced(ctx, collection, AspectClassBson.Id, aspectClass.Id, timestamp)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in SetAspectClass::setSynced %v, will be retried later\n", err))
		return nil
	}
	return nil
}

func (this *Mongo) RemoveAspectClass(ctx context.Context, id string, syncDeleteHandler func(models.AspectClass) error) error {
	old, exists, err := this.GetAspectClass(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.aspectClassCollection()
	err = this.setDeleted(ctx, collection, AspectClassBson.Id, id)
	if err != nil {
		return err
	}
	err = this.SetLastUpdateTimestamp(ctx, this.config.MongoAspectClassCollection, "")
	if err != nil {
		err = nil
	}
	err = syncDeleteHandler(old)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in RemoveAspectClass::syncDeleteHandler %v, will be retried later\n", err))
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{AspectClassBson.Id: id})
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in RemoveAspectClass::DeleteOne %v, will be retried later\n", err))
		return nil
	}
	return nil
}

type AspectClassWithSyncInfo struct {
	models.AspectClass `bson:",inline"`
	SyncInfo           `bson:",inline"`
}

func (this *Mongo) RetryAspectClassSync(lockduration time.Duration, syncDeleteHandler func(models.AspectClass) error, syncHandler func(models.AspectClass) error) error {
	collection := this.aspectClassCollection()
	jobs, err := FetchSyncJobs[AspectClassWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.AspectClass)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryAspectClassSync::syncDeleteHandler %v, will be retried later\n", err))
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{AspectClassBson.Id: job.Id})
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryAspectClassSync::DeleteOne %v, will be retried later\n", err))
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.AspectClass)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryAspectClassSync::syncHandler %v, will be retried later\n", err))
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, AspectClassBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryAspectClassSync::setSynced %v, will be retried later\n", err))
				continue
			}
		}
	}
	return nil
}
