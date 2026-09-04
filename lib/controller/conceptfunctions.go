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

package controller

import (
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

// conceptFunctionListLimit bounds the functions a rename looks at per concept and type. A
// concept owns one; more than that only exists where the startup migration found several and
// deprecated them, and those no longer carry a name the rename would touch.
const conceptFunctionListLimit = 1000

//A concept owns one measuring and one controlling function, named "Get-<concept>" and
//"Set-<concept>" by model.ConceptFunctionName. They are created with the concept and
//renamed with it, but only ever while they still carry the name the convention would have
//given them — a name a user has set is theirs and is not overwritten.

// ensureConceptFunctions creates the pair for a new concept and renames it for a renamed
// one. Deliberately not a repair: a concept whose function was deleted does not get it back
// on the next write, because that would undo the deletion rather than honour it. The
// existing stock is handled once by runConceptFunctionsMigration.
func (this *Controller) ensureConceptFunctions(old models.Concept, existed bool, concept models.Concept) error {
	if !existed {
		return this.createConceptFunctions(concept)
	}
	if old.Name == concept.Name {
		return nil
	}
	return this.renameConceptFunctions(old, concept)
}

func (this *Controller) createConceptFunctions(concept models.Concept) error {
	for _, rdfType := range model.ConceptFunctionRdfTypes {
		function := models.Function{
			Name:        model.ConceptFunctionName(rdfType, concept.Name),
			DisplayName: model.ConceptFunctionName(rdfType, concept.Name),
			ConceptId:   concept.Id,
			RdfType:     rdfType,
		}
		function.GenerateId()
		err := this.setFunction(function)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Controller) renameConceptFunctions(old models.Concept, concept models.Concept) error {
	ctx, _ := getTimeoutContext()
	for _, rdfType := range model.ConceptFunctionRdfTypes {
		functions, _, err := this.db.ListFunctions(ctx, model.FunctionListOptions{
			ConceptIds: []string{concept.Id},
			RdfType:    rdfType,
			Limit:      conceptFunctionListLimit,
			SortBy:     "id.asc",
		})
		if err != nil {
			return err
		}
		oldName := model.ConceptFunctionName(rdfType, old.Name)
		newName := model.ConceptFunctionName(rdfType, concept.Name)
		for _, function := range functions {
			if function.Name != oldName {
				continue
			}
			//the display name follows while it agrees with the name, for the same reason the
			//name follows while it matches the convention. An empty one follows too — it is
			//not a name a user chose, and a function without a display name renders badly.
			if function.DisplayName == "" || function.DisplayName == oldName {
				function.DisplayName = newName
			}
			function.Name = newName
			err = this.setFunction(function)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// PublishFunction hands the sync handler of a function write to the startup migration,
// which creates functions for concepts and cannot reach the publisher from
// lib/database/mongo. Unlike the migrations that convert a stored shape, that one produces
// resources consumers have never seen, so it has to publish them.
func (this *Controller) PublishFunction(function models.Function) error {
	return this.setFunctionSyncHandler(function)
}
