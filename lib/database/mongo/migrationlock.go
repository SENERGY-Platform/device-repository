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
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MigrationLock is held by the one instance that runs the startup migrations. Kubernetes
// starts several, and the migrations do not tolerate a second run beside them:
// runConceptFunctionsMigration records itself only after it succeeded, so two instances that
// pass its check together create two function pairs per concept, and the converting ones
// replace whole documents, which drops what the other one wrote in between.
type MigrationLock struct {
	Id string `json:"id"`
	//Holder names the process for the log and keeps refresh and release to their own lock.
	//It is not consulted when the lock is taken, an expired lock belongs to whoever gets it.
	Holder string `json:"holder"`
	//UnixTimestamp is the last heartbeat of the holder, not the time it took the lock. An
	//instance that dies mid migration stops refreshing it, and the lock expires.
	UnixTimestamp int64 `json:"unix_timestamp"`
}

var MigrationLockBson = getBsonFieldObject[MigrationLock]()

// migrationLockUnixTimestampKey is the bson name of the heartbeat. MigrationLockBson only
// carries the string fields.
const migrationLockUnixTimestampKey = "unix_timestamp"

// migrationLockId is constant: there is one set of startup migrations, so one lock.
const migrationLockId = "startup-migrations"

const (
	//migrationLockHeartbeat is how often the holder refreshes its lock while the migrations
	//run, migrationLockExpiration how long a lock survives unrefreshed. The gap between them
	//is what a stalled write may cost before another instance takes the lock over.
	migrationLockHeartbeat  = 10 * time.Second
	migrationLockExpiration = time.Minute

	//migrationLockRetryInterval is how often a waiting instance tries to take the lock.
	migrationLockRetryInterval = 5 * time.Second

	//migrationLockDefaultTimeout bounds the wait when config.MigrationLockTimeout says
	//nothing. It only ever applies to a holder that is demonstrably alive, because a dead one
	//loses the lock after migrationLockExpiration.
	migrationLockDefaultTimeout = time.Hour
)

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoMigrationLockCollection)
		return db.ensureIndex(collection, "migrationlockidindex", MigrationLockBson.Id, true, true)
	})
}

func (this *Mongo) migrationLockCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoMigrationLockCollection)
}

// lockMigrations blocks until this instance holds the migration lock and returns the release.
// Waiting rather than skipping keeps every instance starting against migrated data: once the
// holder is done, the migrations here are the cheap second pass they are on any restart, and
// a holder that died halfway is retried by whoever takes the lock next.
func (this *Mongo) lockMigrations(ctx context.Context) (release func(), err error) {
	holder := this.migrationLockHolder()
	timeout := this.migrationLockTimeout()
	deadline := time.Now().Add(timeout)
	waiting := false
	for {
		acquired, err := this.acquireMigrationLock(ctx, holder)
		if err != nil {
			return nil, err
		}
		if acquired {
			if waiting {
				this.config.GetLogger().Info("got the startup migration lock", "holder", holder)
			}
			return this.holdMigrationLock(holder), nil
		}
		if !waiting {
			this.config.GetLogger().Info("another instance runs the startup migrations, waiting for its lock", "holder", holder, "timeout", timeout.String())
			waiting = true
		}
		if !time.Now().Add(migrationLockRetryInterval).Before(deadline) {
			return nil, errors.New("timeout while waiting for the startup migration lock of another instance")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(migrationLockRetryInterval):
		}
	}
}

// acquireMigrationLock takes the lock in one atomic write. A free or expired lock matches the
// filter and is overwritten; a missing one is inserted by the upsert. A lock someone else
// holds matches neither, so the upsert runs into the unique index on the id — that duplicate
// key is the answer "held", not an error.
func (this *Mongo) acquireMigrationLock(ctx context.Context, holder string) (acquired bool, err error) {
	now := time.Now()
	_, err = this.migrationLockCollection().UpdateOne(ctx,
		bson.M{
			MigrationLockBson.Id:          migrationLockId,
			migrationLockUnixTimestampKey: bson.M{"$lt": now.Add(-migrationLockExpiration).Unix()},
		},
		bson.M{"$set": bson.M{
			MigrationLockBson.Holder:      holder,
			migrationLockUnixTimestampKey: now.Unix(),
		}},
		options.Update().SetUpsert(true))
	if mongo.IsDuplicateKeyError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// holdMigrationLock keeps the lock fresh until the returned release is called. Without the
// heartbeat every migration longer than migrationLockExpiration would hand its lock to the
// next waiting instance while still running.
func (this *Mongo) holdMigrationLock(holder string) (release func()) {
	stop := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(migrationLockHeartbeat)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				err := this.refreshMigrationLock(holder)
				if err != nil {
					this.config.GetLogger().Warn("unable to refresh the startup migration lock", "error", err, "holder", holder)
				}
			}
		}
	}()
	return func() {
		close(stop)
		<-stopped
		ctx, cancel := getTimeoutContext()
		defer cancel()
		_, err := this.migrationLockCollection().DeleteOne(ctx, bson.M{
			MigrationLockBson.Id:     migrationLockId,
			MigrationLockBson.Holder: holder,
		})
		if err != nil {
			//the lock expires on its own, so this costs a delay, not correctness
			this.config.GetLogger().Warn("unable to release the startup migration lock", "error", err, "holder", holder)
		}
	}
}

func (this *Mongo) refreshMigrationLock(holder string) error {
	ctx, cancel := getTimeoutContext()
	defer cancel()
	result, err := this.migrationLockCollection().UpdateOne(ctx,
		bson.M{
			MigrationLockBson.Id:     migrationLockId,
			MigrationLockBson.Holder: holder,
		},
		bson.M{"$set": bson.M{migrationLockUnixTimestampKey: time.Now().Unix()}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		//only reachable if this process was unable to refresh for migrationLockExpiration and
		//another instance took the lock over, so both are migrating now
		return errors.New("the startup migration lock was taken over by another instance")
	}
	return nil
}

// migrationLockHolder identifies one acquisition. The hostname is the pod name in kubernetes,
// which is what a log reader needs; the id distinguishes two acquisitions of one process.
func (this *Mongo) migrationLockHolder() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}
	return hostname + ":" + strconv.Itoa(os.Getpid()) + ":" + this.CreateId()
}

func (this *Mongo) migrationLockTimeout() time.Duration {
	if this.config.MigrationLockTimeout == "" || this.config.MigrationLockTimeout == "-" {
		return migrationLockDefaultTimeout
	}
	timeout, err := time.ParseDuration(this.config.MigrationLockTimeout)
	if err != nil {
		this.config.GetLogger().Warn("unable to parse migration_lock_timeout", "error", err, "used", migrationLockDefaultTimeout.String())
		return migrationLockDefaultTimeout
	}
	return timeout
}
