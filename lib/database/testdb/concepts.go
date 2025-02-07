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

func (db *DB) SetConcept(ctx context.Context, concept models.Concept, syncHandler func(models.Concept) error) error {
	return set(concept.Id, db.concepts, concept, syncHandler)
}

func (db *DB) RemoveConcept(ctx context.Context, id string, syncDeleteHandler func(models.Concept) error) error {
	return del(id, db.concepts, syncDeleteHandler)
}

func (db *DB) RetryConceptSync(lockduration time.Duration, syncDeleteHandler func(models.Concept) error, syncHandler func(models.Concept) error) error {
	return nil
}

func (db *DB) GetConceptWithCharacteristics(_ context.Context, id string) (result models.ConceptWithCharacteristics, exists bool, err error) {
	panic("not implemented")
}
func (db *DB) GetConceptWithoutCharacteristics(_ context.Context, id string) (result models.Concept, exists bool, err error) {
	return get(id, db.concepts)
}
func (db *DB) ConceptIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	panic("implement me")
}

func (db *DB) ListConceptsWithCharacteristics(ctx context.Context, options model.ConceptListOptions) ([]models.ConceptWithCharacteristics, int64, error) {
	panic("implement me")
}

func (db *DB) ListConcepts(ctx context.Context, options model.ConceptListOptions) (result []models.Concept, total int64, err error) {
	for _, concept := range db.concepts {
		if (options.Search == "" || strings.Contains(strings.ToLower(concept.Name), strings.ToLower(options.Search))) &&
			(options.Ids == nil || slices.Contains(options.Ids, concept.Id)) {
			result = append(result, concept)
		}
	}
	limit := options.Limit
	offset := options.Offset
	if offset >= int64(len(result)) {
		return []models.Concept{}, int64(len(result)), nil
	}
	return result[offset:min(len(result), int(offset+limit))], int64(len(result)), nil
}
