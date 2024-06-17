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

func (c *Client) ReadHub(id string, token string, action model.AuthAction) (result models.Hub, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/hubs/"+id+"&p="+string(action), nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[models.Hub](req)
}

func (c *Client) ListHubDeviceIds(id string, token string, action model.AuthAction, asLocalId bool) (result []string, err error, errCode int) {
	url := c.baseUrl + "/hubs/" + id + "?p=" + string(action)
	if asLocalId {
		url += "&as=local_id"
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[[]string](req)
}

func (c *Client) ValidateHub(token string, hub models.Hub) (err error, code int) {
	return c.validateWithToken(token, "/hubs", hub)
}
