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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
)

func (c *Client) GetFunctionsByType(rdfType string) (result []models.Function, err error, errCode int) {
	var path string
	switch rdfType {
	case model.SES_ONTOLOGY_CONTROLLING_FUNCTION:
		path = "/controlling-functions"
	case model.SES_ONTOLOGY_MEASURING_FUNCTION:
		path = "/measuring-functions"
	default:
		return result, errors.New("unknown rdfType"), http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+path, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[[]models.Function](req)
}

func (c *Client) GetFunction(id string) (result models.Function, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/functions/"+id, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[models.Function](req)
}

func (c *Client) ValidateFunction(function models.Function) (err error, code int) {
	return c.validate("/functions", function)
}

func (c *Client) ValidateFunctionDelete(id string) (err error, code int) {
	return c.validateDelete("/functions/" + id)
}
