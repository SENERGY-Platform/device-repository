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

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ThingBson = getBsonFieldObject[models.Thing]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.thingCollection()
		err := db.ensureIndex(collection, "thingidindex", ThingBson.Id, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "thingdevicelinkdeviceindex", ThingBson.DeviceLinks[0].DeviceId, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) thingCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoDatabase).Collection(this.config.MongoThingCollection)
}

func (this *Mongo) ListThings(ctx context.Context, listOptions model.ThingListOptions) (result []models.Thing, total int64, err error) {
	opt := options.Find()
	if listOptions.Limit > 0 {
		opt.SetLimit(listOptions.Limit)
	}
	if listOptions.Offset > 0 {
		opt.SetSkip(listOptions.Offset)
	}

	if listOptions.SortBy == "" {
		listOptions.SortBy = ThingBson.Id + ".asc"
	}

	sortby := listOptions.SortBy
	sortby = strings.TrimSuffix(sortby, ".asc")
	sortby = strings.TrimSuffix(sortby, ".desc")
	switch sortby {
	case "id":
		sortby = ThingBson.Id
	case "name":
		sortby = ThingBson.Name
	default:
		return nil, 0, fmt.Errorf("%w %q, expected id or name", model.ErrUnsupportedSort, sortby)
	}

	direction := int32(1)
	if strings.HasSuffix(listOptions.SortBy, ".desc") {
		direction = int32(-1)
	}
	sort := bson.D{{Key: sortby, Value: direction}}
	if sortby != ThingBson.Id {
		//names repeat, the id keeps pages stable
		sort = append(sort, bson.E{Key: ThingBson.Id, Value: direction})
	}
	opt.SetSort(sort)

	andFilter := []interface{}{bson.M{NotDeletedFilterKey: NotDeletedFilterValue}}
	filter := bson.M{}
	if listOptions.Ids != nil {
		filter[ThingBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	if len(listOptions.DeviceIds) > 0 {
		andFilter = append(andFilter, bson.M{ThingBson.DeviceLinks[0].DeviceId: bson.M{"$in": listOptions.DeviceIds}})
	}

	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		orFilter := bson.M{"$or": []interface{}{
			bson.M{ThingBson.Id: bson.M{"$regex": escapedSearch, "$options": "i"}},
			bson.M{ThingBson.Name: bson.M{"$regex": escapedSearch, "$options": "i"}},
		}}
		andFilter = append(andFilter, orFilter)
	}

	filter["$and"] = andFilter
	cursor, err := this.thingCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.thingCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) GetThing(ctx context.Context, id string) (thing models.Thing, exists bool, err error) {
	result := this.thingCollection().FindOne(ctx, bson.M{ThingBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return thing, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&thing)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return thing, false, nil
	}
	return thing, true, err
}

type ThingWithSyncInfo struct {
	models.Thing `bson:",inline"`
	SyncInfo     `bson:",inline"`
}

func (this *Mongo) SetThing(ctx context.Context, thing models.Thing, syncHandler func(models.Thing) error) error {
	timestamp := time.Now().Unix()
	collection := this.thingCollection()
	_, err := collection.ReplaceOne(ctx, bson.M{ThingBson.Id: thing.Id}, ThingWithSyncInfo{
		Thing: thing,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = this.SetLastUpdateTimestamp(ctx, this.config.MongoThingCollection, thing.OwnerId)
	if err != nil {
		err = nil
	}
	err = syncHandler(thing)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in SetThing::syncHandler %v, will be retried later\n", err))
		return nil
	}
	err = this.setSynced(ctx, collection, ThingBson.Id, thing.Id, timestamp)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in SetThing::setSynced %v, will be retried later\n", err))
		return nil
	}
	return nil
}

func (this *Mongo) RemoveThing(ctx context.Context, id string, syncDeleteHandler func(models.Thing) error) error {
	old, exists, err := this.GetThing(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.thingCollection()
	err = this.setDeleted(ctx, collection, ThingBson.Id, id)
	if err != nil {
		return err
	}
	err = this.SetLastUpdateTimestamp(ctx, this.config.MongoThingCollection, old.OwnerId)
	if err != nil {
		err = nil
	}
	err = syncDeleteHandler(old)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in RemoveThing::syncDeleteHandler %v, will be retried later\n", err))
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{ThingBson.Id: id})
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in RemoveThing::DeleteOne %v, will be retried later\n", err))
		return nil
	}
	return nil
}

func (this *Mongo) RetryThingSync(lockduration time.Duration, syncDeleteHandler func(models.Thing) error, syncHandler func(models.Thing) error) error {
	collection := this.thingCollection()
	jobs, err := FetchSyncJobs[ThingWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.Thing)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryThingSync::syncDeleteHandler %v, will be retried later\n", err))
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{ThingBson.Id: job.Id})
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryThingSync::DeleteOne %v, will be retried later\n", err))
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.Thing)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryThingSync::syncHandler %v, will be retried later\n", err))
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, ThingBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryThingSync::setSynced %v, will be retried later\n", err))
				continue
			}
		}
	}
	return nil
}

func (this *Mongo) DesyncUnknownThings(ctx context.Context, knownThings []string) (err error) {
	return this.desyncUnknown(ctx, this.thingCollection(), ThingBson.Id, knownThings)
}
