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
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/SENERGY-Platform/models/go/models"
)

func (c *Client) SetProtocol(ctx context.Context, token string, protocol models.Protocol) (result models.Protocol, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(protocol)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if protocol.Id == "" {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/protocols", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodPut, c.baseUrl+"/protocols/"+url.PathEscape(protocol.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Protocol](c, req)
}

func (c *Client) DeleteProtocol(ctx context.Context, token string, id string) (err error, code int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+"/protocols/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}

func (c *Client) ReadProtocol(ctx context.Context, id string, token string) (result models.Protocol, err error, errCode int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/protocols/"+id, nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return doCached[models.Protocol](c, req)
}

func (c *Client) ListProtocols(ctx context.Context, token string, limit int64, offset int64, sort string) (result []models.Protocol, err error, errCode int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/protocols?limit="+strconv.FormatInt(limit, 10)+
		"&offset="+strconv.FormatInt(offset, 10)+"&sort="+sort, nil)
	req.Header.Set("Authorization", token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return doCached[[]models.Protocol](c, req)
}

func (c *Client) ValidateProtocol(ctx context.Context, protocol models.Protocol) (err error, code int) {
	return c.validate(ctx, "/protocols", protocol)
}
