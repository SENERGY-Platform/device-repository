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
	"context"
	"encoding/json"
	"net/http"

	"github.com/SENERGY-Platform/models/go/models"
)

func (c *Client) GetDefaultDeviceAttributes(ctx context.Context, token string) (attributes []models.Attribute, err error, code int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/defaults/devices/attributes", nil)
	if err != nil {
		return attributes, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doCached[[]models.Attribute](c, req)
}

func (c *Client) SetDefaultDeviceAttributes(ctx context.Context, token string, attributes []models.Attribute) (err error, code int) {
	b, err := json.Marshal(attributes)
	if err != nil {
		return err, http.StatusBadRequest
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseUrl+"/defaults/devices/attributes", bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}
