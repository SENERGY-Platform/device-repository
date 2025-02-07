/*
 * Copyright 2019 InfAI (CC SES)
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

var HubBson = getBsonFieldObject[model.HubWithConnectionState]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoHubCollection)
		err = db.ensureIndex(collection, "hubidindex", HubBson.Id, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "hubnameindex", HubBson.Name, true, false) //to support faster sort
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "hubdeviceidindex", HubBson.DeviceIds[0], true, false)
		if err != nil {
			return err
		}
		err = db.removeIndex(collection, "hubdevicelocalidindex")
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) hubCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoHubCollection)
}

func (this *Mongo) GetHub(ctx context.Context, id string) (hub model.HubWithConnectionState, exists bool, err error) {
	result := this.hubCollection().FindOne(ctx, bson.M{HubBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return hub, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&hub)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return hub, false, nil
	}
	return hub, true, err
}

type HubWithSyncInfo struct {
	model.HubWithConnectionState `bson:",inline"`
	SyncInfo                     `bson:",inline"`
}

func (this *Mongo) SetHub(ctx context.Context, hub model.HubWithConnectionState, syncHandler func(model.HubWithConnectionState) error) (err error) {
	timestamp := time.Now().Unix()
	collection := this.hubCollection()
	_, err = this.hubCollection().ReplaceOne(ctx, bson.M{HubBson.Id: hub.Id}, HubWithSyncInfo{
		HubWithConnectionState: hub,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = syncHandler(hub)
	if err != nil {
		log.Printf("WARNING: error in SetHub::syncHandler %v, will be retried later\n", err)
		return nil
	}
	err = this.setSynced(ctx, collection, HubBson.Id, hub.Id, timestamp)
	if err != nil {
		log.Printf("WARNING: error in SetHub::setSynced %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RemoveHub(ctx context.Context, id string, syncDeleteHandler func(model.HubWithConnectionState) error) error {
	old, exists, err := this.GetHub(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.hubCollection()
	err = this.setDeleted(ctx, collection, HubBson.Id, id)
	if err != nil {
		return err
	}
	err = syncDeleteHandler(old)
	if err != nil {
		log.Printf("WARNING: error in RemoveHub::syncDeleteHandler %v, will be retried later\n", err)
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{HubBson.Id: id})
	if err != nil {
		log.Printf("WARNING: error in RemoveHub::DeleteOne %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RetryHubSync(lockduration time.Duration, syncDeleteHandler func(model.HubWithConnectionState) error, syncHandler func(model.HubWithConnectionState) error) error {
	collection := this.hubCollection()
	jobs, err := FetchSyncJobs[HubWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.HubWithConnectionState)
			if err != nil {
				log.Printf("WARNING: error in RetryHubSync::syncDeleteHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{HubBson.Id: job.Id})
			if err != nil {
				log.Printf("WARNING: error in RetryHubSync::DeleteOne %v, will be retried later\n", err)
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.HubWithConnectionState)
			if err != nil {
				log.Printf("WARNING: error in RetryHubSync::syncHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, HubBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				log.Printf("WARNING: error in RetryHubSync::setSynced %v, will be retried later\n", err)
				continue
			}
		}
	}
	return nil
}

func (this *Mongo) GetHubsByDeviceId(ctx context.Context, id string) (hubs []model.HubWithConnectionState, err error) {
	cursor, err := this.hubCollection().Find(ctx, bson.M{HubBson.DeviceIds[0]: id, NotDeletedFilterKey: NotDeletedFilterValue})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(ctx) {
		hub := model.HubWithConnectionState{}
		err = cursor.Decode(&hub)
		if err != nil {
			return nil, err
		}
		hubs = append(hubs, hub)
	}
	err = cursor.Err()
	return hubs, err
}

func (this *Mongo) ListHubs(ctx context.Context, listOptions model.HubListOptions, withTotal bool) (result []model.HubWithConnectionState, total int64, err error) {
	opt := options.Find()
	if listOptions.Limit > 0 {
		opt.SetLimit(listOptions.Limit)
	}
	if listOptions.Offset > 0 {
		opt.SetSkip(listOptions.Offset)
	}

	if listOptions.SortBy == "" {
		listOptions.SortBy = HubBson.Name + ".asc"
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
		filter[HubBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		filter[HubBson.Name] = bson.M{"$regex": regexp.QuoteMeta(search), "$options": "i"}
	}
	if listOptions.ConnectionState != nil {
		filter[HubBson.ConnectionState] = listOptions.ConnectionState
	}

	if listOptions.LocalDeviceId != "" {
		filter[HubBson.DeviceLocalIds[0]] = listOptions.LocalDeviceId
	}
	if listOptions.OwnerId != "" {
		filter[HubBson.OwnerId] = listOptions.OwnerId
	}

	cursor, err := this.hubCollection().Find(ctx, filter, opt)
	if err != nil {
		return result, total, err
	}
	result, err, _ = readCursorResult[model.HubWithConnectionState](ctx, cursor)
	if err != nil {
		return result, total, err
	}
	if withTotal {
		total, err = this.hubCollection().CountDocuments(ctx, filter)
		if err != nil {
			return result, total, err
		}
	}
	return result, total, err
}

func (this *Mongo) SetHubConnectionState(ctx context.Context, id string, state models.ConnectionState) error {
	_, err := this.hubCollection().UpdateOne(ctx, bson.M{
		HubBson.Id: id,
	}, bson.M{
		"$set": bson.M{HubBson.ConnectionState: state},
	})
	return err
}
