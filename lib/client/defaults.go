/*
 * Copyright 2025 InfAI (CC SES)
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

func (c *Client) GetDefaultDeviceAttributes(token string) (attributes []models.Attribute, err error, code int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/defaults/devices/attributes", nil)
	if err != nil {
		return attributes, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[[]models.Attribute](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) SetDefaultDeviceAttributes(token string, attributes []models.Attribute) (err error, code int) {
	b, err := json.Marshal(attributes)
	if err != nil {
		return err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+"/defaults/devices/attributes", bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}
