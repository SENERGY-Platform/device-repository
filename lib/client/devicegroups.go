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
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
)

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

func (c *Client) ValidateDeviceGroup(deviceGroup models.DeviceGroup) (err error, code int) {
	return c.CheckAccessToDevicesOfGroup("", deviceGroup)
}

func (c *Client) CheckAccessToDevicesOfGroup(token string, group models.DeviceGroup) (err error, code int) {
	b, err := json.Marshal(group)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+"/device-groups?dry-run=true", bytes.NewBuffer(b))
	req.Header.Set("Authorization", token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, resp.StatusCode
}
