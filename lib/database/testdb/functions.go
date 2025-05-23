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
	"slices"
	"strings"
	"time"
)

func (db *DB) SetFunction(ctx context.Context, function models.Function, syncHandler func(models.Function) error) error {
	return set(function.Id, db.functions, function, syncHandler)
}

func (db *DB) RemoveFunction(ctx context.Context, id string, syncDeleteHandler func(models.Function) error) error {
	return del(id, db.functions, syncDeleteHandler)
}

func (db *DB) RetryFunctionSync(lockduration time.Duration, syncDeleteHandler func(models.Function) error, syncHandler func(models.Function) error) error {
	return nil
}

func (db *DB) GetFunction(_ context.Context, id string) (result models.Function, exists bool, err error) {
	return get(id, db.functions)
}

func (db *DB) ListFunctions(ctx context.Context, options model.FunctionListOptions) (result []models.Function, total int64, err error) {
	for _, f := range db.functions {
		if (options.RdfType == "" || f.RdfType == options.RdfType) &&
			(options.Search == "" || strings.Contains(strings.ToLower(f.Name), strings.ToLower(options.Search))) &&
			(options.Ids == nil || slices.Contains(options.Ids, f.Id)) {
			result = append(result, f)
		}
	}
	limit := options.Limit
	offset := options.Offset
	if offset >= int64(len(result)) {
		return []models.Function{}, int64(len(result)), nil
	}
	return result[offset:min(len(result), int(offset+limit))], int64(len(result)), nil
}

func (db *DB) ListAllFunctionsByType(_ context.Context, rdfType string) (result []models.Function, err error) {
	for _, f := range db.functions {
		if f.RdfType == rdfType {
			result = append(result, f)
		}
	}
	return
}
func (db *DB) ListAllMeasuringFunctionsByAspect(_ context.Context, aspect string, ancestors bool, descendants bool) ([]models.Function, error) {
	panic("not implemented")
}

func (db *DB) ListAllFunctionsByDeviceClass(_ context.Context, class string) ([]models.Function, error) {
	panic("not implemented")
}

func (db *DB) ListAllControllingFunctionsByDeviceClass(_ context.Context, class string) ([]models.Function, error) {
	panic("not implemented")
}

func (db *DB) FunctionIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	panic("implement me")
}
