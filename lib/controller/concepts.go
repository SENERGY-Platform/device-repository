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
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func (this *Controller) setConceptSyncHandler(ctx context.Context, c models.Concept) error {
	return this.publisher.PublishConcept(c)
}

func (this *Controller) SetConcept(ctx context.Context, token string, concept models.Concept) (result models.Concept, err error, code int) {
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
		err, code = this.ValidateConcept(ctx, concept)
		if err != nil {
			return result, err, code
		}
	}
	//read before the write, because whether the concept is new and what it was called
	//decides whether its functions are created or renamed
	ctx, _ = getTimeoutContext(ctx)
	old, existed, err := this.db.GetConceptWithoutCharacteristics(ctx, concept.Id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	err = this.setConcept(ctx, concept)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	err = this.ensureConceptFunctions(ctx, old, existed, concept)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return concept, nil, http.StatusOK
}

func (this *Controller) setConcept(ctx context.Context, concept models.Concept) (err error) {
	ctx, _ = getTimeoutContext(ctx)
	err = this.db.SetConcept(ctx, concept, withContext(ctx, this.setConceptSyncHandler))
	return err
}

func (this *Controller) deleteConceptSyncHandler(ctx context.Context, c models.Concept) error {
	return this.publisher.PublishConceptDelete(c.Id)
}

func (this *Controller) DeleteConcept(ctx context.Context, token string, id string) (error, int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return errors.New("token is not an admin"), http.StatusUnauthorized
	}
	err, code := this.ValidateConceptDelete(ctx, id)
	if err != nil {
		return err, code
	}
	ctx, _ = getTimeoutContext(ctx)
	err = this.db.RemoveConcept(ctx, id, withContext(ctx, this.deleteConceptSyncHandler))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) ValidateConcept(ctx context.Context, concept models.Concept) (err error, code int) {
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
	ctx, _ = getTimeoutContext(ctx)
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

func (this *Controller) ListConceptsWithCharacteristics(ctx context.Context, listOptions model.ConceptListOptions) (result []models.ConceptWithCharacteristics, total int64, err error, errCode int) {
	ctx, _ = getTimeoutContext(ctx)
	result, total, err = this.db.ListConceptsWithCharacteristics(ctx, listOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) ListConcepts(ctx context.Context, listOptions model.ConceptListOptions) (result []models.Concept, total int64, err error, errCode int) {
	ctx, _ = getTimeoutContext(ctx)
	result, total, err = this.db.ListConcepts(ctx, listOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) GetConceptWithCharacteristics(ctx context.Context, id string) (result models.ConceptWithCharacteristics, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	result, exists, err := this.db.GetConceptWithCharacteristics(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) GetConceptWithoutCharacteristics(ctx context.Context, id string) (result models.Concept, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	result, exists, err := this.db.GetConceptWithoutCharacteristics(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ValidateConceptDelete(ctx context.Context, id string) (err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	isUsed, where, err := this.db.ConceptIsUsed(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if isUsed {
		return errors.New("still in use: " + strings.Join(where, ",")), http.StatusBadRequest
	}
	return nil, http.StatusOK
}
