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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) ListAspectNodes(options model.AspectListOptions) (result []models.AspectNode, total int64, err error, errCode int) {
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
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/v2/aspect-nodes"+queryString, nil)
	if err != nil {
		return result, 0, err, http.StatusInternalServerError
	}
	return doWithTotalInResult[[]models.AspectNode](req)
}

func (c *Client) GetAspectNode(id string) (models.AspectNode, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspect-nodes/"+id, nil)
	if err != nil {
		return models.AspectNode{}, err, http.StatusInternalServerError
	}
	return do[models.AspectNode](req)
}

func (c *Client) GetAspectNodes() ([]models.AspectNode, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspect-nodes", nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.AspectNode](req)
}

func (c *Client) GetAspectNodesWithMeasuringFunction(ancestors bool, descendants bool) ([]models.AspectNode, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspect-nodes?function=measuring-function&ancestors="+strconv.FormatBool(ancestors)+"&descendants="+strconv.FormatBool(descendants), nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.AspectNode](req)
}

func (c *Client) GetAspectNodesMeasuringFunctions(id string, ancestors bool, descendants bool) (result []models.Function, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspect-nodes/"+id+"/measuring-functions?ancestors="+
		strconv.FormatBool(ancestors)+"&descendants="+strconv.FormatBool(descendants), nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.Function](req)
}

func (c *Client) GetAspectNodesWithFunction(function string, ancestors bool, descendants bool) ([]models.AspectNode, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspect-nodes?function="+function+"&ancestors="+strconv.FormatBool(ancestors)+"&descendants="+strconv.FormatBool(descendants), nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.AspectNode](req)
}

func (c *Client) GetAspectNodesByIdList(ids []string) (result []models.AspectNode, err error, code int) {
	b, err := json.Marshal(ids)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/query/aspect-nodes", bytes.NewBuffer(b))
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.AspectNode](req)
}
