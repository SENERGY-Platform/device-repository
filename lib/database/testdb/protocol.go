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

package testdb

import (
	"context"
	"github.com/SENERGY-Platform/models/go/models"
	"golang.org/x/exp/maps"
	"time"
)

func (db *DB) SetProtocol(ctx context.Context, protocol models.Protocol, syncHandler func(models.Protocol) error) error {
	return set(protocol.Id, db.protocols, protocol, syncHandler)
}

func (db *DB) RemoveProtocol(ctx context.Context, id string, syncDeleteHandler func(models.Protocol) error) error {
	return del(id, db.protocols, syncDeleteHandler)
}

func (db *DB) RetryProtocolSync(lockduration time.Duration, syncDeleteHandler func(models.Protocol) error, syncHandler func(models.Protocol) error) error {
	return nil
}

func (db *DB) GetProtocol(_ context.Context, id string) (result models.Protocol, exists bool, err error) {
	return get(id, db.protocols)
}
func (db *DB) ListProtocols(_ context.Context, limit int64, offset int64, sort string) ([]models.Protocol, error) {
	return maps.Values(db.protocols), nil
}
