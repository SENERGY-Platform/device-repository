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

package client

import (
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
)

func (c *Client) GetConceptWithCharacteristics(id string) (models.ConceptWithCharacteristics, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/concepts/"+id+"?sub-class=true", nil)
	if err != nil {
		return models.ConceptWithCharacteristics{}, err, http.StatusInternalServerError
	}
	return do[models.ConceptWithCharacteristics](req)
}

func (c *Client) GetConceptWithoutCharacteristics(id string) (models.Concept, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/concepts/"+id+"?sub-class=false", nil)
	if err != nil {
		return models.Concept{}, err, http.StatusInternalServerError
	}
	return do[models.Concept](req)
}

func (c *Client) ValidateConcept(concept models.Concept) (err error, code int) {
	return c.validate("/concepts", concept)
}

func (c *Client) ValidateConceptDelete(id string) (err error, code int) {
	return c.validateDelete("/concepts/" + id)
}
