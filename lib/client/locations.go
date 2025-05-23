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

func (c *Client) SetLocation(token string, location models.Location) (result models.Location, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(location)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if location.Id == "" {
		req, err = http.NewRequest(http.MethodPost, c.baseUrl+"/locations", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequest(http.MethodPut, c.baseUrl+"/locations/"+url.PathEscape(location.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Location](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) DeleteLocation(token string, id string) (err error, code int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/locations/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetLocation(id string, token string) (location models.Location, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/locations/"+id, nil)
	if err != nil {
		return location, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Location](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ValidateLocation(location models.Location) (err error, code int) {
	return c.validate("/locations", location)
}

func (c *Client) ListLocations(token string, options model.LocationListOptions) (result []models.Location, total int64, err error, errCode int) {
	query := url.Values{}
	if options.Permission != models.UnsetPermissionFlag {
		query.Set("p", string(options.Permission))
	}
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
	queryString := ""
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/locations"+queryString, nil)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doWithTotalInResult[[]models.Location](req, c.optionalAuthTokenForApiGatewayRequest)
}
