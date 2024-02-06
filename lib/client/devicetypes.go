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
	return do[models.DeviceType](req)
}

func (c *Client) ListDeviceTypes(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int) {
	filterStr, err := json.Marshal(filter)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-types?limit="+strconv.FormatInt(limit, 10)+
		"&offset="+strconv.FormatInt(offset, 10)+"&sort="+sort+"&filter="+string(filterStr)+
		"&interactions-filter="+strings.Join(interactionsFilter, ",")+"&include_id_modified="+
		strconv.FormatBool(includeModified)+"&include_id_unmodified="+strconv.FormatBool(includeUnmodified), nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[[]models.DeviceType](req)
}

func (c *Client) ListDeviceTypesV2(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int) {
	filterStr, err := json.Marshal(filter)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-types?limit="+strconv.FormatInt(limit, 10)+
		"&offset="+strconv.FormatInt(offset, 10)+"&sort="+sort+"&filter="+string(filterStr)+
		"&include_id_modified="+
		strconv.FormatBool(includeModified)+"&include_id_unmodified="+strconv.FormatBool(includeUnmodified), nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[[]models.DeviceType](req)
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
	return do[model.UsedInDeviceTypeResponse](req)
}
