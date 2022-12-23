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

func (db *DB) SetConcept(_ context.Context, concept models.Concept) error {
	return set(concept.Id, db.concepts, concept)
}
func (db *DB) RemoveConcept(_ context.Context, id string) error {
	return del(id, db.concepts)
}
func (db *DB) GetConceptWithCharacteristics(_ context.Context, id string) (result models.ConceptWithCharacteristics, exists bool, err error) {
	panic("not implemented") // TODO
}
func (db *DB) GetConceptWithoutCharacteristics(_ context.Context, id string) (result models.Concept, exists bool, err error) {
	return get(id, db.concepts)
}
func (db *DB) ConceptIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	//TODO implement me
	panic("implement me")
}
