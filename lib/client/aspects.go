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

func (c *Client) SetAspect(token string, aspect models.Aspect) (result models.Aspect, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(aspect)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if aspect.Id == "" {
		req, err = http.NewRequest(http.MethodPost, c.baseUrl+"/aspects", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequest(http.MethodPut, c.baseUrl+"/aspects/"+url.PathEscape(aspect.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Aspect](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) DeleteAspect(token string, id string) (err error, code int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/aspects/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListAspects(options model.AspectListOptions) (result []models.Aspect, total int64, err error, errCode int) {
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
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v2/aspects"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	return doWithTotalInResult[[]models.Aspect](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetAspects() ([]models.Aspect, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspects", nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.Aspect](req, c.optionalAuthTokenForApiGatewayRequest)
}
func (c *Client) GetAspectsWithMeasuringFunction(ancestors bool, descendants bool) ([]models.Aspect, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspects?function=measuring-function&ancestors="+strconv.FormatBool(ancestors)+"&descendants="+strconv.FormatBool(descendants), nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.Aspect](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetAspect(id string) (models.Aspect, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspects/"+id, nil)
	if err != nil {
		return models.Aspect{}, err, http.StatusInternalServerError
	}
	return do[models.Aspect](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ValidateAspect(aspect models.Aspect) (err error, code int) {
	return c.validate("/aspects", aspect)
}

func (c *Client) ValidateAspectDelete(id string) (err error, code int) {
	return c.validateDelete("/aspects/" + id)
}
