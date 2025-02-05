/*
 * Copyright 2025 InfAI (CC SES)
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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type SyncInfo struct {
	SyncTodo          bool  `bson:"sync_todo"`
	SyncDelete        bool  `bson:"sync_delete"`
	SyncUnixTimestamp int64 `bson:"sync_unix_timestamp"`
}

const SyncTodoBson = "sync_todo"
const SyncDeleteBson = "sync_delete"
const SyncUnixTimestampBson = "sync_unix_timestamp"

var NotDeletedFilterKey = "$or"
var NotDeletedFilterValue = []interface{}{
	bson.M{SyncDeleteBson: bson.M{"$exists": false}},
	bson.M{SyncDeleteBson: false},
}

func (this *Mongo) setSynced(ctx context.Context, collection *mongo.Collection, idField string, idValue string, unixTimestampWhereElementIsUnsynced int64) error {
	_, err := collection.UpdateOne(ctx, bson.M{
		idField:               idValue,
		SyncUnixTimestampBson: unixTimestampWhereElementIsUnsynced,
	}, bson.M{
		"$set": bson.M{
			SyncTodoBson:          false,
			SyncUnixTimestampBson: time.Now().Unix(),
		},
	})
	return err
}

func (this *Mongo) setDeleted(ctx context.Context, collection *mongo.Collection, idField string, idValue string) error {
	_, err := collection.UpdateOne(ctx, bson.M{
		idField: idValue,
	}, bson.M{
		"$set": bson.M{
			SyncTodoBson:          true,
			SyncDeleteBson:        true,
			SyncUnixTimestampBson: time.Now().Unix(),
		},
	})
	return err
}

const FetchSyncJobsDefaultBatchSize = 1000

func FetchSyncJobs[OutputType any](collection *mongo.Collection, syncLockDuration time.Duration, maxBatchSize int) (jobs []OutputType, err error) {
	now := time.Now()
	cutoff := now.Add(-syncLockDuration)
	loopBreakTime := now.Add(syncLockDuration)
	for {
		//should never happen, emergency break
		if !time.Now().Before(loopBreakTime) {
			return jobs, nil
		}

		job, err := fetchSyncJob[OutputType](collection, now, cutoff)

		//no more jobs in queue
		if errors.Is(err, mongo.ErrNoDocuments) {
			return jobs, nil
		}

		//normal error handling
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)

		if len(jobs) >= maxBatchSize {
			return jobs, nil
		}
	}
}

func fetchSyncJob[OutputType any](collection *mongo.Collection, now time.Time, cutoff time.Time) (job OutputType, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result := collection.FindOneAndUpdate(ctx,
		bson.M{
			SyncTodoBson:          true,
			SyncUnixTimestampBson: bson.M{"$lt": cutoff.Unix()},
		},
		bson.M{
			"$set": bson.M{
				SyncUnixTimestampBson: now.Unix(),
			},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	if result.Err() != nil {
		err = result.Err()
		return job, err
	}
	err = result.Decode(&job)
	return job, err
}
