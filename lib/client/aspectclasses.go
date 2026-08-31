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

package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type AspectClassListOptions = model.AspectClassListOptions

func (c *Client) ListAspectClasses(options model.AspectClassListOptions) (result []models.AspectClass, total int64, err error, errCode int) {
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
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspect-classes"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	return doWithTotalInResult[[]models.AspectClass](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetAspectClass(id string) (result models.AspectClass, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspect-classes/"+url.PathEscape(id), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[models.AspectClass](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) SetAspectClass(token string, aspectClass models.AspectClass) (result models.AspectClass, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(aspectClass)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if aspectClass.Id == "" {
		req, err = http.NewRequest(http.MethodPost, c.baseUrl+"/aspect-classes", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequest(http.MethodPut, c.baseUrl+"/aspect-classes/"+url.PathEscape(aspectClass.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.AspectClass](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) DeleteAspectClass(token string, id string) (err error, code int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/aspect-classes/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ValidateAspectClass(aspectClass models.AspectClass) (err error, code int) {
	return c.validate("/aspect-classes", aspectClass)
}

func (c *Client) ValidateAspectClassDelete(id string) (err error, code int) {
	return c.validateDelete("/aspect-classes/" + id)
}
