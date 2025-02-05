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

type DeviceUpdateOptions = model.DeviceUpdateOptions

func (c *Client) SetDeviceConnectionState(token string, id string, connected bool) (error, int) {
	b, err := json.Marshal(connected)
	if err != nil {
		return err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+"/devices/"+url.PathEscape(id)+"/connection-state", bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) SetDevice(token string, device models.Device, options model.DeviceUpdateOptions) (result models.Device, err error, code int) {
	b, err := json.Marshal(device)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	query := url.Values{}
	if options.UpdateOnlySameOriginAttributes != nil {
		query.Set("update-only-same-origin-attributes", strings.Join(options.UpdateOnlySameOriginAttributes, ","))
	}
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+"/devices/"+url.PathEscape(device.Id)+"?"+query.Encode(), bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Device](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) CreateDevice(token string, device models.Device) (result models.Device, err error, code int) {
	b, err := json.Marshal(device)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/devices", bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Device](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) DeleteDevice(token string, id string) (err error, code int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/devices/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListDevices(token string, options DeviceListOptions) (result []models.Device, err error, errCode int) {
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
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/devices"+queryString, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[[]models.Device](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ReadDevice(id string, token string, action model.AuthAction) (result models.Device, err error, errCode int) {
	query := url.Values{}
	if action != models.UnsetPermissionFlag {
		query.Set("p", string(action))
	}
	queryString := ""
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/devices/"+id+queryString, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Device](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ReadDeviceByLocalId(ownerId string, localId string, token string, action model.AuthAction) (result models.Device, err error, errCode int) {
	query := url.Values{}
	if action != models.UnsetPermissionFlag {
		query.Set("p", string(action))
	}
	query.Set("as", "local_id")
	query.Set("owner_id", ownerId)
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/devices/"+url.PathEscape(localId)+"?"+query.Encode(), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Device](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ValidateDevice(token string, device models.Device) (err error, code int) {
	return c.validateWithToken(token, "/devices", device)
}
