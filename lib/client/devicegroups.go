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
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) ListDeviceGroups(token string, options model.DeviceGroupListOptions) (result []models.DeviceGroup, total int64, err error, errCode int) {
	queryString := ""
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
	if options.IgnoreGenerated {
		query.Set("ignore_generated", strconv.FormatBool(options.IgnoreGenerated))
	}
	if options.Criteria != nil {
		criteriaJson, err := json.Marshal(options.Criteria)
		if err != nil {
			return []models.DeviceGroup{}, total, err, http.StatusInternalServerError
		}
		query.Set("criteria", string(criteriaJson))
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
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-groups"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doWithTotalInResult[[]models.DeviceGroup](req)
}

func (c *Client) ReadDeviceGroup(id string, token string, filterGenericDuplicateCriteria bool) (result models.DeviceGroup, err error, errCode int) {
	query := ""
	if filterGenericDuplicateCriteria {
		query = "?filter_generic_duplicate_criteria=true"
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-groups/"+id+query, nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[models.DeviceGroup](req)
}

func (c *Client) ValidateDeviceGroup(token string, deviceGroup models.DeviceGroup) (err error, code int) {
	return c.validateWithToken(token, "/device-groups", deviceGroup)
}

func (c *Client) ValidateDeviceGroupDelete(token string, id string) (err error, code int) {
	return c.validateDeleteWithToken(token, "/device-groups/"+id)
}
