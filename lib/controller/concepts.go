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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"net/http"
)

func (this *Controller) SetConcept(concept model.Concept, owner string) error {
	ctx, _ := getTimeoutContext()
	return this.db.SetConcept(ctx, concept)
}

func (this *Controller) DeleteConcept(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveConcept(ctx, id)
}

func (this *Controller) ValidateConcept(concept model.Concept) (err error, code int) {
	if concept.Id == "" {
		return errors.New("missing concept id"), http.StatusBadRequest
	}
	if concept.Name == "" {
		return errors.New("missing concept name"), http.StatusBadRequest
	}
	for _, charId := range concept.CharacteristicIds {
		if charId == "" {
			return errors.New("missing char id"), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) GetConceptWithCharacteristics(id string) (result model.ConceptWithCharacteristics, err error, code int) {
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

func (this *Controller) GetConceptWithoutCharacteristics(id string) (result model.Concept, err error, code int) {
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
