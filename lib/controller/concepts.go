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

package controller

import (
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"strings"
)

func (this *Controller) SetConcept(concept models.Concept, owner string) error {
	ctx, _ := getTimeoutContext()
	return this.db.SetConcept(ctx, concept)
}

func (this *Controller) DeleteConcept(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveConcept(ctx, id)
}

func (this *Controller) ValidateConcept(concept models.Concept) (err error, code int) {
	if concept.Id == "" {
		return errors.New("missing concept id"), http.StatusBadRequest
	}
	if concept.Name == "" {
		return errors.New("missing concept name"), http.StatusBadRequest
	}
	if len(concept.CharacteristicIds) > 0 && concept.BaseCharacteristicId == "" {
		return errors.New("missing concept base characteristic"), http.StatusBadRequest
	}
	if concept.BaseCharacteristicId != "" && !contains(concept.CharacteristicIds, concept.BaseCharacteristicId) {
		return errors.New("concept base characteristic not in characteristic ids list"), http.StatusBadRequest
	}
	newCharacteristicIds := map[string]bool{}
	ctx, _ := getTimeoutContext()
	for _, charId := range concept.CharacteristicIds {
		if charId == "" {
			return errors.New("missing char id"), http.StatusBadRequest
		}
		_, exists, err := this.db.GetCharacteristic(ctx, charId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown characteristic: " + charId), http.StatusBadRequest
		}
		newCharacteristicIds[charId] = true
	}
	old, exists, err := this.db.GetConceptWithoutCharacteristics(ctx, concept.Id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if exists {
		for _, charid := range old.CharacteristicIds {
			if !newCharacteristicIds[charid] {
				used, where, err := this.db.CharacteristicIsUsedWithConceptInDeviceType(ctx, charid, concept.Id)
				if err != nil {
					return err, http.StatusInternalServerError
				}
				if used {
					return fmt.Errorf("removed characteristic is still used with this concept in %v", strings.Join(where, ", ")), http.StatusBadRequest
				}
			}
		}
	}

	return nil, http.StatusOK
}

func (this *Controller) GetConceptWithCharacteristics(id string) (result models.ConceptWithCharacteristics, err error, code int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetConceptWithCharacteristics(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) GetConceptWithoutCharacteristics(id string) (result models.Concept, err error, code int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetConceptWithoutCharacteristics(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ValidateConceptDelete(id string) (err error, code int) {
	ctx, _ := getTimeoutContext()
	isUsed, where, err := this.db.ConceptIsUsed(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if isUsed {
		return errors.New("still in use: " + strings.Join(where, ",")), http.StatusBadRequest
	}
	return nil, http.StatusOK
}
