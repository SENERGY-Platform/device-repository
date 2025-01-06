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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) ListCharacteristics(options model.CharacteristicListOptions) (result []models.Characteristic, total int64, err error, errCode int) {
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
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v2/characteristics"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	return doWithTotalInResult[[]models.Characteristic](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetCharacteristics(leafsOnly bool) (result []models.Characteristic, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/characteristics?leafsOnly="+strconv.FormatBool(leafsOnly), nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.Characteristic](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetCharacteristic(id string) (result models.Characteristic, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/characteristics/"+id, nil)
	if err != nil {
		return models.Characteristic{}, err, http.StatusInternalServerError
	}
	return do[models.Characteristic](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ValidateCharacteristics(characteristic models.Characteristic) (err error, code int) {
	return c.validate("/characteristics", characteristic)
}

func (c *Client) ValidateCharacteristicDelete(id string) (err error, code int) {
	return c.validateDelete("/characteristics/" + id)
}
