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
	"strings"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func (c *Client) SetCharacteristic(ctx context.Context, token string, characteristic models.Characteristic) (result models.Characteristic, err error, code int) {
	var req *http.Request
	b, err := json.Marshal(characteristic)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if characteristic.Id == "" {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/characteristics", bytes.NewBuffer(b))
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodPut, c.baseUrl+"/characteristics/"+url.PathEscape(characteristic.Id), bytes.NewBuffer(b))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Characteristic](c, req)
}

func (c *Client) DeleteCharacteristic(ctx context.Context, token string, id string) (err error, code int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+"/characteristics/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}

func (c *Client) ListCharacteristics(ctx context.Context, options model.CharacteristicListOptions) (result []models.Characteristic, total int64, err error, errCode int) {
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/v2/characteristics"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	return doCachedWithTotal[[]models.Characteristic](c, req)
}

func (c *Client) GetCharacteristics(ctx context.Context, leafsOnly bool) (result []models.Characteristic, err error, errCode int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/characteristics?leafsOnly="+strconv.FormatBool(leafsOnly), nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return doCached[[]models.Characteristic](c, req)
}

func (c *Client) GetCharacteristic(ctx context.Context, id string) (result models.Characteristic, err error, errCode int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/characteristics/"+id, nil)
	if err != nil {
		return models.Characteristic{}, err, http.StatusInternalServerError
	}
	return doCached[models.Characteristic](c, req)
}

func (c *Client) ValidateCharacteristics(ctx context.Context, characteristic models.Characteristic) (err error, code int) {
	return c.validate(ctx, "/characteristics", characteristic)
}

func (c *Client) ValidateCharacteristicDelete(ctx context.Context, id string) (err error, code int) {
	return c.validateDelete(ctx, "/characteristics/"+id)
}
