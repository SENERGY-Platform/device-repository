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
)

func (c *Client) ReadDevice(id string, token string, action model.AuthAction) (result models.Device, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/devices/"+id+"?p="+string(action), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Device](req)
}

func (c *Client) ReadDeviceByLocalId(localId string, token string, action model.AuthAction) (result models.Device, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/devices/"+localId+"?p="+string(action)+"&as=local_id", nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Device](req)
}

func (c *Client) ValidateDevice(token string, device models.Device) (err error, code int) {
	return c.validateWithToken(token, "/devices", device)
}
