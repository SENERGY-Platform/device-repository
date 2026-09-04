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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MigrationState records that a one-shot migration has run. Migrations that convert a
// stored shape do not need it — they recognise the old shape and skip. One that creates
// resources does: a second run would create them again, and there is nothing in the data
// to tell a resource this migration made from one a user made afterwards.
type MigrationState struct {
	Name          string `json:"name"`
	UnixTimestamp int64  `json:"unix_timestamp"`
}

var MigrationStateBson = getBsonFieldObject[MigrationState]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoMigrationStateCollection)
		return db.ensureIndex(collection, "migrationstatenameindex", MigrationStateBson.Name, true, true)
	})
}

func (this *Mongo) migrationStateCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoMigrationStateCollection)
}

func (this *Mongo) migrationHasRun(ctx context.Context, name string) (bool, error) {
	err := this.migrationStateCollection().FindOne(ctx, bson.M{MigrationStateBson.Name: name}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// setMigrationHasRun is written after the migration succeeded, not before. A migration that
// dies halfway therefore runs again on the next start, which is why the ones using this have
// to survive seeing their own partial result.
func (this *Mongo) setMigrationHasRun(ctx context.Context, name string) error {
	_, err := this.migrationStateCollection().ReplaceOne(ctx,
		bson.M{MigrationStateBson.Name: name},
		MigrationState{Name: name, UnixTimestamp: time.Now().Unix()},
		options.Replace().SetUpsert(true))
	return err
}
