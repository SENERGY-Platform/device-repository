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
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func (c *Client) SetDeviceGroup(ctx context.Context, token string, deviceGroup models.DeviceGroup) (result models.DeviceGroup, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(deviceGroup)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if deviceGroup.Id == "" {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/device-groups", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodPut, c.baseUrl+"/device-groups/"+url.PathEscape(deviceGroup.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.DeviceGroup](c, req)
}

func (c *Client) DeleteDeviceGroup(ctx context.Context, token string, id string) (err error, code int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+"/device-groups/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}

func (c *Client) ListDeviceGroups(ctx context.Context, token string, options model.DeviceGroupListOptions) (result []models.DeviceGroup, total int64, err error, errCode int) {
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
	if options.FilterGenericDuplicateCriteria {
		query.Set("filter_generic_duplicate_criteria", "true")
	}
	if options.Criteria != nil {
		criteriaJson, err := json.Marshal(options.Criteria)
		if err != nil {
			return []models.DeviceGroup{}, total, err, http.StatusInternalServerError
		}
		query.Set("criteria", string(criteriaJson))
	}
	if options.DeviceIds != nil {
		query.Set("device-ids", strings.Join(options.DeviceIds, ","))
	}
	if options.AttributeValues != nil {
		query.Set("attr-values", strings.Join(options.AttributeValues, ","))
	}
	if options.AttributeKeys != nil {
		query.Set("attr-keys", strings.Join(options.AttributeKeys, ","))
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/device-groups"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doCachedWithTotal[[]models.DeviceGroup](c, req)
}

func (c *Client) ReadDeviceGroup(ctx context.Context, id string, token string, filterGenericDuplicateCriteria bool) (result models.DeviceGroup, err error, errCode int) {
	query := ""
	if filterGenericDuplicateCriteria {
		query = "?filter_generic_duplicate_criteria=true"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/device-groups/"+id+query, nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return doCached[models.DeviceGroup](c, req)
}

func (c *Client) ValidateDeviceGroup(ctx context.Context, token string, deviceGroup models.DeviceGroup) (err error, code int) {
	return c.validateWithToken(ctx, token, "/device-groups", deviceGroup)
}

func (c *Client) RecomputeDeviceGroupCriteria(ctx context.Context, token string, ids []string) (err error, code int) {
	endpoint := c.baseUrl + "/device-groups/criteria/recompute"
	if ids != nil {
		endpoint += "?" + url.Values{"ids": {strings.Join(ids, ",")}}.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}

func (c *Client) ValidateDeviceGroupDelete(ctx context.Context, token string, id string) (err error, code int) {
	return c.validateDeleteWithToken(ctx, token, "/device-groups/"+id)
}
