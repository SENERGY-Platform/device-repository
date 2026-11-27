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
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// removeMany is the batch counterpart of the Remove<Resource> methods: it marks all existing, not yet deleted elements
// with the given ids as deleted, calls the sync handler once with all of them and deletes them afterwards.
// Like in the single variants, a failing sync handler or delete is logged and not returned:
// the marked elements are picked up by the Retry<Resource>Sync loop.
//
// afterMark runs once after the elements are marked (last update timestamps, dependent collections);
// it receives the loaded elements and their ids in the same order. An error from afterMark ends the removal
// before the sync handler is called, which matches the single variants of the resources that need it. afterMark may be nil.
// An error wrapping ErrLastUpdateTimestamp is returned to the caller, any other error is logged and not returned.
// Unknown ids and elements that are already marked as deleted are ignored.
func removeMany[T any](ctx context.Context, logger *slog.Logger, name string, collection *mongo.Collection, idField string, ids []string, afterMark func(elements []T, ids []string) error, syncDeleteHandler func([]T) error) error {
	if len(ids) == 0 {
		return nil
	}
	cursor, err := collection.Find(ctx, bson.M{idField: bson.M{"$in": ids}, NotDeletedFilterKey: NotDeletedFilterValue})
	if err != nil {
		return err
	}
	var raws []bson.Raw
	err = cursor.All(ctx, &raws)
	if err != nil {
		return err
	}
	if len(raws) == 0 {
		return nil
	}
	elements := make([]T, 0, len(raws))
	foundIds := make([]string, 0, len(raws))
	for _, raw := range raws {
		var element T
		err = bson.Unmarshal(raw, &element)
		if err != nil {
			return err
		}
		id, ok := raw.Lookup(idField).StringValueOK()
		if !ok {
			return fmt.Errorf("%v: element without string field %v", name, idField)
		}
		elements = append(elements, element)
		foundIds = append(foundIds, id)
	}

	_, err = collection.UpdateMany(ctx, bson.M{idField: bson.M{"$in": foundIds}}, bson.M{
		"$set": bson.M{
			SyncTodoBson:          true,
			SyncDeleteBson:        true,
			SyncUnixTimestampBson: time.Now().Unix(),
		},
	})
	if err != nil {
		return err
	}

	if afterMark != nil {
		err = afterMark(elements, foundIds)
		if errors.Is(err, ErrLastUpdateTimestamp) {
			return err
		}
		if err != nil {
			logger.Warn(fmt.Sprintf("error in %v::afterMark %v, will be retried later\n", name, err))
			return nil
		}
	}

	err = syncDeleteHandler(elements)
	if err != nil {
		logger.Warn(fmt.Sprintf("error in %v::syncDeleteHandler %v, will be retried later\n", name, err))
		return nil
	}
	// fresh context: the sync handler may have used up the deadline of ctx, and the marked documents are only removed here
	deleteCtx, cancel := getTimeoutContext()
	defer cancel()
	_, err = collection.DeleteMany(deleteCtx, bson.M{idField: bson.M{"$in": foundIds}})
	if err != nil {
		logger.Warn(fmt.Sprintf("error in %v::DeleteMany %v, will be retried later\n", name, err))
		return nil
	}
	return nil
}

// setLastUpdateTimestampForOwners calls SetLastUpdateTimestamp once for every distinct owner.
// Errors are logged by SetLastUpdateTimestamp. Every owner is tried, the first error is returned;
// callers for collections that do not back a 304 answer ignore it, like the single remove methods do.
func (this *Mongo) setLastUpdateTimestampForOwners(ctx context.Context, collection string, owners []string) (err error) {
	seen := map[string]bool{}
	for _, owner := range owners {
		if seen[owner] {
			continue
		}
		seen[owner] = true
		if ownerErr := this.SetLastUpdateTimestamp(ctx, collection, owner); ownerErr != nil && err == nil {
			err = ownerErr
		}
	}
	return err
}

func (this *Mongo) desyncUnknown(ctx context.Context, collection *mongo.Collection, idField string, knownIds []string) error {
	result, err := collection.UpdateOne(ctx, bson.M{
		idField: bson.M{"$nin": knownIds},
	}, bson.M{
		"$set": bson.M{SyncTodoBson: true},
	})
	if err != nil {
		return err
	}
	this.config.GetLogger().Info("desynced unknown elements", "collection", collection.Name(), "affected", result.ModifiedCount)
	return nil
}

const FetchSyncJobsDefaultBatchSize = 1000

func FetchSyncJobs[OutputType any](collection *mongo.Collection, syncLockDuration time.Duration, maxBatchSize int) (jobs []OutputType, err error) {
	defer func() {
		slog.Debug("fetch sync jobs", "collection", collection.Name(), "jobs", len(jobs), "error", err)
	}()
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
