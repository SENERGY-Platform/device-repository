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

func (c *Client) ReadDeviceType(id string, token string) (result models.DeviceType, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-types/"+id, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[models.DeviceType](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListDeviceTypes(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int) {
	options := url.Values{
		"limit":                 {strconv.FormatInt(limit, 10)},
		"offset":                {strconv.FormatInt(offset, 10)},
		"sort":                  {sort},
		"include_id_modified":   {strconv.FormatBool(includeModified)},
		"include_id_unmodified": {strconv.FormatBool(includeUnmodified)},
		"interactions-filter":   {strings.Join(interactionsFilter, ",")},
	}
	if len(filter) > 0 {
		filterStr, err := json.Marshal(filter)
		if err != nil {
			return result, err, http.StatusBadRequest
		}
		options.Add("filter", string(filterStr))
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-types?"+options.Encode(), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[[]models.DeviceType](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListDeviceTypesV2(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int) {
	options := url.Values{
		"limit":                 {strconv.FormatInt(limit, 10)},
		"offset":                {strconv.FormatInt(offset, 10)},
		"sort":                  {sort},
		"include_id_modified":   {strconv.FormatBool(includeModified)},
		"include_id_unmodified": {strconv.FormatBool(includeUnmodified)},
	}
	if len(filter) > 0 {
		filterStr, err := json.Marshal(filter)
		if err != nil {
			return result, err, http.StatusBadRequest
		}
		options.Add("filter", string(filterStr))
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-types?"+options.Encode(), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[[]models.DeviceType](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListDeviceTypesV3(token string, options model.DeviceTypeListOptions) (result []models.DeviceType, total int64, err error, errCode int) {
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
	if options.ProtocolIds != nil {
		query.Set("protocol-ids", strings.Join(options.ProtocolIds, ","))
	}
	if options.AttributeKeys != nil {
		query.Set("attr-keys", strings.Join(options.AttributeKeys, ","))
	}
	if options.AttributeValues != nil {
		query.Set("attr-values", strings.Join(options.AttributeValues, ","))
	}
	if options.IncludeModified {
		query.Set("include-modified", strconv.FormatBool(options.IncludeModified))
	}
	if options.IgnoreUnmodified {
		query.Set("ignore-unmodified", strconv.FormatBool(options.IgnoreUnmodified))
	}
	if len(options.Criteria) > 0 {
		filterStr, err := json.Marshal(options.Criteria)
		if err != nil {
			return result, 0, err, http.StatusBadRequest
		}
		query.Add("criteria", string(filterStr))
	}
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v3/device-types"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doWithTotalInResult[[]models.DeviceType](req, c.optionalAuthTokenForApiGatewayRequest)
}

type DeviceTypeValidationOptions = model.ValidationOptions

func (c *Client) ValidateDeviceType(deviceType models.DeviceType, options model.ValidationOptions) (err error, code int) {
	return c.validateWithOptions("/device-types", deviceType, options.AsUrlValues())
}

func (c *Client) GetUsedInDeviceType(query model.UsedInDeviceTypeQuery) (result model.UsedInDeviceTypeResponse, err error, errCode int) {
	body, err := json.Marshal(query)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/query/used-in-device-type", bytes.NewBuffer(body))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[model.UsedInDeviceTypeResponse](req, c.optionalAuthTokenForApiGatewayRequest)
}
