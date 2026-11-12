/*
 * Copyright 2026 InfAI (CC SES)
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

type Thing = models.Thing
type ThingDeviceLink = models.ThingDeviceLink

type ThingListOptions = model.ThingListOptions
type ThingDeviceLinkCreate = model.ThingDeviceLinkCreate
type ThingDeviceLinkEnd = model.ThingDeviceLinkEnd

func (c *Client) ListThings(ctx context.Context, token string, options model.ThingListOptions) (result []models.Thing, total int64, err error, errCode int) {
	queryString := ""
	query := url.Values{}
	if options.Permission != models.UnsetPermissionFlag {
		query.Set("p", string(options.Permission))
	}
	if options.Search != "" {
		query.Set("search", options.Search)
	}
	if options.Ids != nil {
		query.Set("ids", strings.Join(options.Ids, ","))
	}
	if options.DeviceIds != nil {
		query.Set("device_ids", strings.Join(options.DeviceIds, ","))
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/things"+queryString, nil)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doCachedWithTotal[[]models.Thing](c, req)
}

func (c *Client) ReadThing(ctx context.Context, token string, id string) (result models.Thing, err error, errCode int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/things/"+url.PathEscape(id), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doCached[models.Thing](c, req)
}

// SetThing creates the thing with POST if thing.Id is empty and updates it with PUT otherwise.
// The device links of the thing are ignored by the service, use AddThingDeviceLink and EndThingDeviceLink.
func (c *Client) SetThing(ctx context.Context, token string, thing models.Thing) (result models.Thing, err error, code int) {
	method := http.MethodPost
	endpoint := c.baseUrl + "/things"
	if thing.Id != "" {
		method = http.MethodPut
		endpoint = c.baseUrl + "/things/" + url.PathEscape(thing.Id)
	}
	b, err := json.Marshal(thing)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Thing](c, req)
}

func (c *Client) DeleteThing(ctx context.Context, token string, id string) (error, int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+"/things/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}

func (c *Client) AddThingDeviceLink(ctx context.Context, token string, thingId string, link model.ThingDeviceLinkCreate) (result models.Thing, err error, code int) {
	b, err := json.Marshal(link)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/things/"+url.PathEscape(thingId)+"/device-links", bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Thing](c, req)
}

func (c *Client) EndThingDeviceLink(ctx context.Context, token string, thingId string, linkId string, end model.ThingDeviceLinkEnd) (result models.Thing, err error, code int) {
	b, err := json.Marshal(end)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/things/"+url.PathEscape(thingId)+"/device-links/"+url.PathEscape(linkId)+"/end", bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Thing](c, req)
}
