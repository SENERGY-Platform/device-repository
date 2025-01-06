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
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"strconv"
)

func (c *Client) ReadProtocol(id string, token string) (result models.Protocol, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/protocols/"+id, nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[models.Protocol](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ListProtocols(token string, limit int64, offset int64, sort string) (result []models.Protocol, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/protocols?limit="+strconv.FormatInt(limit, 10)+
		"&offset="+strconv.FormatInt(offset, 10)+"&sort="+sort, nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[[]models.Protocol](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ValidateProtocol(protocol models.Protocol) (err error, code int) {
	return c.validate("/protocols", protocol)
}
