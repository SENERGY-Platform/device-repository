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
	"bytes"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) SetConcept(token string, concept models.Concept) (result models.Concept, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(concept)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if concept.Id == "" {
		req, err = http.NewRequest(http.MethodPost, c.baseUrl+"/concepts", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequest(http.MethodPut, c.baseUrl+"/concepts/"+url.PathEscape(concept.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Concept](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) DeleteConcept(token string, id string) (err error, code int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/concepts/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListConcepts(options model.ConceptListOptions) (result []models.Concept, total int64, err error, errCode int) {
	queryString := ""
	query := url.Values{}
	if options.Search != "" {
		query.Set("search", options.Search)
	}
	if options.Ids != nil {
		query.Set("ids", strings.Join(options.Ids, ","))
	}
	if options.SortBy != "" {
		query.Set("sort", options.SortBy)
	}
	if options.Limit != 0 {
		query.Set("limit", strconv.FormatInt(options.Limit, 10))
	}
	if options.Offset != 0 {
		query.Set("offset", strconv.FormatInt(options.Offset, 10))
	}
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v2/concepts"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	return doWithTotalInResult[[]models.Concept](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListConceptsWithCharacteristics(options model.ConceptListOptions) (result []models.ConceptWithCharacteristics, total int64, err error, errCode int) {
	queryString := ""
	query := url.Values{}
	if options.Search != "" {
		query.Set("search", options.Search)
	}
	if options.Ids != nil {
		query.Set("ids", strings.Join(options.Ids, ","))
	}
	if options.SortBy != "" {
		query.Set("sort", options.SortBy)
	}
	if options.Limit != 0 {
		query.Set("limit", strconv.FormatInt(options.Limit, 10))
	}
	if options.Offset != 0 {
		query.Set("offset", strconv.FormatInt(options.Offset, 10))
	}
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v2/concepts-with-characteristics"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	return doWithTotalInResult[[]models.ConceptWithCharacteristics](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetConceptWithCharacteristics(id string) (models.ConceptWithCharacteristics, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/concepts/"+id+"?sub-class=true", nil)
	if err != nil {
		return models.ConceptWithCharacteristics{}, err, http.StatusInternalServerError
	}
	return do[models.ConceptWithCharacteristics](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetConceptWithoutCharacteristics(id string) (models.Concept, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/concepts/"+id+"?sub-class=false", nil)
	if err != nil {
		return models.Concept{}, err, http.StatusInternalServerError
	}
	return do[models.Concept](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ValidateConcept(concept models.Concept) (err error, code int) {
	return c.validate("/concepts", concept)
}

func (c *Client) ValidateConceptDelete(id string) (err error, code int) {
	return c.validateDelete("/concepts/" + id)
}
