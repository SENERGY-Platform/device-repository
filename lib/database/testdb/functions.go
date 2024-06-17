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
)

func (db *DB) SetFunction(_ context.Context, function models.Function) error {
	return set(function.Id, db.functions, function)
}
func (db *DB) GetFunction(_ context.Context, id string) (result models.Function, exists bool, err error) {
	return get(id, db.functions)
}
func (db *DB) RemoveFunction(_ context.Context, id string) error {
	return del(id, db.functions)
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
