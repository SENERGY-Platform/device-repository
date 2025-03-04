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

func (c *Client) SetHubConnectionState(token string, id string, connected bool) (error, int) {
	b, err := json.Marshal(connected)
	if err != nil {
		return err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+"/hubs/"+url.PathEscape(id)+"/connection-state", bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) SetHub(token string, hub models.Hub) (result models.Hub, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(hub)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if hub.Id == "" {
		req, err = http.NewRequest(http.MethodPost, c.baseUrl+"/hubs", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequest(http.MethodPut, c.baseUrl+"/hubs/"+url.PathEscape(hub.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Hub](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) DeleteHub(token string, id string) (err error, code int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/hubs/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ReadHub(id string, token string, action model.AuthAction) (result models.Hub, err error, errCode int) {
	query := url.Values{}
	if action != models.UnsetPermissionFlag {
		query.Set("p", string(action))
	}
	queryString := ""
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/hubs/"+id+queryString, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Hub](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListHubs(token string, options model.HubListOptions) (result []models.Hub, err error, errCode int) {
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
	if options.LocalDeviceId != "" {
		query.Set("local-device-id", options.LocalDeviceId)
	}
	if options.OwnerId != "" {
		query.Set("owner", options.OwnerId)
	}
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/hubs"+queryString, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[[]models.Hub](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListHubDeviceIds(id string, token string, action model.AuthAction, asLocalId bool) (result []string, err error, errCode int) {
	query := url.Values{}
	if action != models.UnsetPermissionFlag {
		query.Set("p", string(action))
	}
	if asLocalId {
		query.Set("as", "local_id")
	}
	queryString := ""
	url := c.baseUrl + "/hubs/" + id + queryString
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[[]string](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ValidateHub(token string, hub models.Hub) (err error, code int) {
	return c.validateWithToken(token, "/hubs", hub)
}
