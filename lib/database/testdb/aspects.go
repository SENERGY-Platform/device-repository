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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"golang.org/x/exp/maps"
	"slices"
	"strings"
)

func (db *DB) ListAspects(ctx context.Context, options model.AspectListOptions) (result []models.Aspect, total int64, err error) {
	for _, aspect := range db.aspects {
		if (options.Search == "" || strings.Contains(strings.ToLower(aspect.Name), strings.ToLower(options.Search))) &&
			(options.Ids == nil || slices.Contains(options.Ids, aspect.Id)) {
			result = append(result, aspect)
		}
	}
	limit := options.Limit
	offset := options.Offset
	if offset >= int64(len(result)) {
		return []models.Aspect{}, int64(len(result)), nil
	}
	return result[offset:min(len(result), int(offset+limit))], int64(len(result)), nil
}

func (db *DB) GetAspect(_ context.Context, id string) (result models.Aspect, exists bool, err error) {
	return get(id, db.aspects)
}
func (db *DB) SetAspect(_ context.Context, aspect models.Aspect) error {
	return set(aspect.Id, db.aspects, aspect)
}
func (db *DB) RemoveAspect(_ context.Context, id string) error {
	return del(id, db.aspects)
}
func (db *DB) ListAllAspects(_ context.Context) ([]models.Aspect, error) {
	return maps.Values(db.aspects), nil
}
func (db *DB) ListAspectsWithMeasuringFunction(_ context.Context, ancestors bool, descendants bool) ([]models.Aspect, error) {
	panic("not implemented")
}

func (db *DB) AspectIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {

	panic("implement me")
}
