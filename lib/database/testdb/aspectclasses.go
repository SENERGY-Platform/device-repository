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

package testdb

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func (db *DB) SetAspectClass(ctx context.Context, aspectClass models.AspectClass, syncHandler func(models.AspectClass) error) error {
	return set(aspectClass.Id, db.aspectClasses, aspectClass, syncHandler)
}

func (db *DB) RemoveAspectClass(ctx context.Context, id string, syncDeleteHandler func(models.AspectClass) error) error {
	return del(id, db.aspectClasses, syncDeleteHandler)
}

func (db *DB) RetryAspectClassSync(lockduration time.Duration, syncDeleteHandler func(models.AspectClass) error, syncHandler func(models.AspectClass) error) error {
	return nil
}

func (db *DB) ListAspectClasses(ctx context.Context, options model.AspectClassListOptions) (result []models.AspectClass, total int64, err error) {
	for _, ac := range db.aspectClasses {
		if (options.Search == "" || strings.Contains(strings.ToLower(ac.Name), strings.ToLower(options.Search))) &&
			(options.Ids == nil || slices.Contains(options.Ids, ac.Id)) {
			result = append(result, ac)
		}
	}
	limit := options.Limit
	offset := options.Offset
	if offset >= int64(len(result)) {
		return []models.AspectClass{}, int64(len(result)), nil
	}
	return result[offset:min(len(result), int(offset+limit))], int64(len(result)), nil
}

func (db *DB) GetAspectClass(_ context.Context, id string) (result models.AspectClass, exists bool, err error) {
	return get(id, db.aspectClasses)
}
