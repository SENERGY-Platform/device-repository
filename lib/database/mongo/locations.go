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

var LocationBson = getBsonFieldObject[models.Location]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoLocationCollection)
		err = db.ensureIndex(collection, "locationidindex", LocationBson.Id, true, true)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) locationCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoLocationCollection)
}

func (this *Mongo) GetLocation(ctx context.Context, id string) (location models.Location, exists bool, err error) {
	result := this.locationCollection().FindOne(ctx, bson.M{LocationBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return location, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&location)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return location, false, nil
	}
	return location, true, err
}

type LocationWithSyncInfo struct {
	models.Location `bson:",inline"`
	SyncInfo        `bson:",inline"`
	SyncUser        string `bson:"sync_user"`
}

func (this *Mongo) SetLocation(ctx context.Context, location models.Location, syncHandler func(l models.Location, user string) error, user string) (err error) {
	timestamp := time.Now().Unix()
	collection := this.locationCollection()
	_, err = this.locationCollection().ReplaceOne(ctx, bson.M{LocationBson.Id: location.Id}, LocationWithSyncInfo{
		Location: location,
		SyncUser: user,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = syncHandler(location, user)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::syncHandler %v, will be retried later\n", err)
		return nil
	}
	err = this.setSynced(ctx, collection, LocationBson.Id, location.Id, timestamp)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::setSynced %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RemoveLocation(ctx context.Context, id string, syncDeleteHandler func(models.Location) error) error {
	old, exists, err := this.GetLocation(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.locationCollection()
	err = this.setDeleted(ctx, collection, LocationBson.Id, id)
	if err != nil {
		return err
	}
	err = syncDeleteHandler(old)
	if err != nil {
		log.Printf("WARNING: error in RemoveLocation::syncDeleteHandler %v, will be retried later\n", err)
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{LocationBson.Id: id})
	if err != nil {
		log.Printf("WARNING: error in RemoveLocation::DeleteOne %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RetryLocationSync(lockduration time.Duration, syncDeleteHandler func(models.Location) error, syncHandler func(l models.Location, user string) error) error {
	collection := this.locationCollection()
	jobs, err := FetchSyncJobs[LocationWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.Location)
			if err != nil {
				log.Printf("WARNING: error in RetryLocationSync::syncDeleteHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{LocationBson.Id: job.Id})
			if err != nil {
				log.Printf("WARNING: error in RetryLocationSync::DeleteOne %v, will be retried later\n", err)
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.Location, job.SyncUser)
			if err != nil {
				log.Printf("WARNING: error in RetryLocationSync::syncHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, LocationBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				log.Printf("WARNING: error in RetryLocationSync::setSynced %v, will be retried later\n", err)
				continue
			}
		}
	}
	return nil
}

func (this *Mongo) ListLocations(ctx context.Context, listOptions model.LocationListOptions) (result []models.Location, total int64, err error) {
	opt := options.Find()
	if listOptions.Limit > 0 {
		opt.SetLimit(listOptions.Limit)
	}
	if listOptions.Offset > 0 {
		opt.SetSkip(listOptions.Offset)
	}

	if listOptions.SortBy == "" {
		listOptions.SortBy = LocationBson.Name + ".asc"
	}

	sortby := listOptions.SortBy
	sortby = strings.TrimSuffix(sortby, ".asc")
	sortby = strings.TrimSuffix(sortby, ".desc")

	direction := int32(1)
	if strings.HasSuffix(listOptions.SortBy, ".desc") {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{NotDeletedFilterKey: NotDeletedFilterValue}
	if listOptions.Ids != nil {
		filter[LocationBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter["$or"] = []interface{}{
			bson.M{LocationBson.Name: bson.M{"$regex": escapedSearch, "$options": "i"}},
			bson.M{LocationBson.Description: bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	}

	cursor, err := this.locationCollection().Find(ctx, filter, opt)
	if err != nil {
		return result, total, err
	}
	result, err, _ = readCursorResult[models.Location](ctx, cursor)
	if err != nil {
		return result, total, err
	}
	total, err = this.locationCollection().CountDocuments(ctx, filter)
	if err != nil {
		return result, total, err
	}
	return result, total, err
}
