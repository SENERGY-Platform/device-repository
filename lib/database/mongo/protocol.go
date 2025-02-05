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
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strings"
	"time"
)

var ProtocolBson = getBsonFieldObject[models.Protocol]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoProtocolCollection)
		err = db.ensureIndex(collection, "protocolidindex", ProtocolBson.Id, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "protocolnameindex", ProtocolBson.Name, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) protocolCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoProtocolCollection)
}

func (this *Mongo) GetProtocol(ctx context.Context, id string) (protocol models.Protocol, exists bool, err error) {
	result := this.protocolCollection().FindOne(ctx, bson.M{ProtocolBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return protocol, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&protocol)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return protocol, false, nil
	}
	return protocol, true, err
}

func (this *Mongo) ListProtocols(ctx context.Context, limit int64, offset int64, sort string) (result []models.Protocol, err error) {
	opt := options.Find()
	opt.SetLimit(limit)
	opt.SetSkip(offset)

	parts := strings.Split(sort, ".")
	sortby := ProtocolBson.Id
	switch parts[0] {
	case "id":
		sortby = ProtocolBson.Id
	case "name":
		sortby = ProtocolBson.Name
	default:
		sortby = ProtocolBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	cursor, err := this.protocolCollection().Find(ctx, bson.M{NotDeletedFilterKey: NotDeletedFilterValue}, opt)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		protocol := models.Protocol{}
		err = cursor.Decode(&protocol)
		if err != nil {
			return nil, err
		}
		result = append(result, protocol)
	}
	err = cursor.Err()
	return
}

type ProtocolWithSyncInfo struct {
	models.Protocol `bson:",inline"`
	SyncInfo        `bson:",inline"`
}

func (this *Mongo) SetProtocol(ctx context.Context, protocol models.Protocol, syncHandler func(models.Protocol) error) (err error) {
	timestamp := time.Now().Unix()
	collection := this.protocolCollection()
	_, err = this.protocolCollection().ReplaceOne(ctx, bson.M{ProtocolBson.Id: protocol.Id}, ProtocolWithSyncInfo{
		Protocol: protocol,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = syncHandler(protocol)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::syncHandler %v, will be retried later\n", err)
		return nil
	}
	err = this.setSynced(ctx, collection, ProtocolBson.Id, protocol.Id, timestamp)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::setSynced %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RemoveProtocol(ctx context.Context, id string, syncDeleteHandler func(models.Protocol) error) error {
	old, exists, err := this.GetProtocol(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.protocolCollection()
	err = this.setDeleted(ctx, collection, ProtocolBson.Id, id)
	if err != nil {
		return err
	}
	err = syncDeleteHandler(old)
	if err != nil {
		log.Printf("WARNING: error in RemoveProtocol::syncDeleteHandler %v, will be retried later\n", err)
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{ProtocolBson.Id: id})
	if err != nil {
		log.Printf("WARNING: error in RemoveProtocol::DeleteOne %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RetryProtocolSync(lockduration time.Duration, syncDeleteHandler func(models.Protocol) error, syncHandler func(models.Protocol) error) error {
	collection := this.protocolCollection()
	jobs, err := FetchSyncJobs[ProtocolWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.Protocol)
			if err != nil {
				log.Printf("WARNING: error in RetryProtocolSync::syncDeleteHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{ProtocolBson.Id: job.Id})
			if err != nil {
				log.Printf("WARNING: error in RetryProtocolSync::DeleteOne %v, will be retried later\n", err)
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.Protocol)
			if err != nil {
				log.Printf("WARNING: error in RetryProtocolSync::syncHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, ProtocolBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				log.Printf("WARNING: error in RetryProtocolSync::setSynced %v, will be retried later\n", err)
				continue
			}
		}
	}
	return nil
}
