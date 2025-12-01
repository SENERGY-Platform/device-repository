/*
 * Copyright 2024 InfAI (CC SES)
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
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

const extendedDevicePath = "extended-devices"

func (c *Client) ListExtendedDevices(token string, options model.ExtendedDeviceListOptions) (result []models.ExtendedDevice, total int64, err error, errCode int) {
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
	if options.LocalIds != nil {
		query.Set("local_ids", strings.Join(options.LocalIds, ","))
	}
	if options.Owner != "" {
		query.Set("owner", options.Owner)
	}
	if options.DeviceTypeIds != nil {
		query.Set("device-type-ids", strings.Join(options.DeviceTypeIds, ","))
	}
	if options.ConnectionState != nil {
		query.Set("connection-state", *options.ConnectionState)
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
	if options.AttributeKeys != nil {
		query.Set("attr-keys", strings.Join(options.AttributeKeys, ","))
	}
	if options.AttributeValues != nil {
		query.Set("attr-values", strings.Join(options.AttributeValues, ","))
	}
	if options.FullDt {
		query.Set("fulldt", "true")
	}
	if options.DeviceAttributeBlacklist != nil {
		b, err := json.Marshal(options.DeviceAttributeBlacklist)
		if err != nil {
			return result, 0, err, http.StatusBadRequest
		}
		query.Set("device-attribute-blacklist", url.QueryEscape(string(b)))
	}
	queryString := ""
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/"+extendedDevicePath+queryString, nil)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doWithTotalInResult[[]models.ExtendedDevice](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ReadExtendedDevice(id string, token string, action model.AuthAction, fullDt bool) (result models.ExtendedDevice, err error, errCode int) {
	query := url.Values{}
	if action != models.UnsetPermissionFlag {
		query.Set("p", string(action))
	}
	if fullDt {
		query.Set("fulldt", "true")
	}
	queryString := ""
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/"+extendedDevicePath+"/"+id+queryString, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.ExtendedDevice](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ReadExtendedDeviceByLocalId(ownerId string, localId string, token string, action model.AuthAction, fullDt bool) (result models.ExtendedDevice, err error, errCode int) {
	query := url.Values{}
	if action != models.UnsetPermissionFlag {
		query.Set("p", string(action))
	}
	query.Set("as", "local_id")
	query.Set("owner_id", ownerId)
	if fullDt {
		query.Set("fulldt", "true")
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/"+extendedDevicePath+"/"+url.PathEscape(localId)+"?"+query.Encode(), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.ExtendedDevice](req, c.optionalAuthTokenForApiGatewayRequest)
}
