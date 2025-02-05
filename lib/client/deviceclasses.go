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

func (c *Client) SetDeviceClass(token string, deviceClass models.DeviceClass) (result models.DeviceClass, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(deviceClass)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if deviceClass.Id == "" {
		req, err = http.NewRequest(http.MethodPost, c.baseUrl+"/device-classes", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequest(http.MethodPut, c.baseUrl+"/device-classes/"+url.PathEscape(deviceClass.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.DeviceClass](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) DeleteDeviceClass(token string, id string) (err error, code int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/device-classes/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListDeviceClasses(options model.DeviceClassListOptions) (result []models.DeviceClass, total int64, err error, errCode int) {
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
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v2/device-classes"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	return doWithTotalInResult[[]models.DeviceClass](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetDeviceClasses() ([]models.DeviceClass, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes", nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.DeviceClass](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetDeviceClassesWithControllingFunctions() ([]models.DeviceClass, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes?function=controlling-function", nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.DeviceClass](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetDeviceClassesFunctions(id string) (result []models.Function, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes/"+id+"/functions", nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.Function](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetDeviceClassesControllingFunctions(id string) (result []models.Function, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes/"+id+"/controlling-functions", nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[[]models.Function](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) GetDeviceClass(id string) (result models.DeviceClass, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes/"+id, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[models.DeviceClass](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ValidateDeviceClass(deviceclass models.DeviceClass) (err error, code int) {
	return c.validate("/device-classes", deviceclass)
}

func (c *Client) ValidateDeviceClassDelete(id string) (err error, code int) {
	return c.validateDelete("/device-classes/" + id)
}
