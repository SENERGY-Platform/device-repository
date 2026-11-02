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
	"errors"
	"net/http"
	"net/url"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func (c *Client) SetFunction(ctx context.Context, token string, function models.Function) (result models.Function, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(function)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if function.Id == "" {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/functions", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodPut, c.baseUrl+"/functions/"+url.PathEscape(function.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Function](c, req)
}

func (c *Client) DeleteFunction(ctx context.Context, token string, id string) (err error, code int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+"/functions/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}

func (c *Client) ListFunctions(ctx context.Context, options model.FunctionListOptions) (result []models.Function, total int64, err error, errCode int) {
	buf := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(buf).Encode(options)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/query/functions", buf)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	return doCachedWithTotal[[]models.Function](c, req)
}

func (c *Client) GetFunctionsByType(ctx context.Context, rdfType string) (result []models.Function, err error, errCode int) {
	var path string
	switch rdfType {
	case model.SES_ONTOLOGY_CONTROLLING_FUNCTION:
		path = "/controlling-functions"
	case model.SES_ONTOLOGY_MEASURING_FUNCTION:
		path = "/measuring-functions"
	default:
		return result, errors.New("unknown rdfType"), http.StatusBadRequest
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+path, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return doCached[[]models.Function](c, req)
}

func (c *Client) GetFunction(ctx context.Context, id string) (result models.Function, err error, errCode int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/functions/"+id, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return doCached[models.Function](c, req)
}

func (c *Client) ValidateFunction(ctx context.Context, function models.Function) (err error, code int) {
	return c.validate(ctx, "/functions", function)
}

func (c *Client) ValidateFunctionDelete(ctx context.Context, id string) (err error, code int) {
	return c.validateDelete(ctx, "/functions/"+id)
}
