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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"net/http"
	"strings"
)

func (this *Controller) setConceptSyncHandler(c models.Concept) error {
	return this.publisher.PublishConcept(c)
}

func (this *Controller) SetConcept(token string, concept models.Concept) (result models.Concept, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return result, errors.New("token is not an admin"), http.StatusUnauthorized
	}

	//ensure ids
	concept.GenerateId()
	if !this.config.DisableStrictValidationForTesting {
		err, code = this.ValidateConcept(concept)
		if err != nil {
			return result, err, code
		}
	}
	err = this.setConcept(concept)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return concept, nil, http.StatusOK
}

func (this *Controller) setConcept(concept models.Concept) (err error) {
	ctx, _ := getTimeoutContext()
	err = this.db.SetConcept(ctx, concept, this.setConceptSyncHandler)
	return err
}

func (this *Controller) deleteConceptSyncHandler(c models.Concept) error {
	return this.publisher.PublishConceptDelete(c.Id)
}

func (this *Controller) DeleteConcept(token string, id string) (error, int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return errors.New("token is not an admin"), http.StatusUnauthorized
	}
	err, code := this.ValidateConceptDelete(id)
	if err != nil {
		return err, code
	}
	ctx, _ := getTimeoutContext()
	err = this.db.RemoveConcept(ctx, id, this.deleteConceptSyncHandler)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
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

func (this *Controller) ListConceptsWithCharacteristics(listOptions model.ConceptListOptions) (result []models.ConceptWithCharacteristics, total int64, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListConceptsWithCharacteristics(ctx, listOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) ListConcepts(listOptions model.ConceptListOptions) (result []models.Concept, total int64, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListConcepts(ctx, listOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
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
