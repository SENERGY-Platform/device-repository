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
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type Graph = models.Graph
type Node = models.Node
type Edge = models.Edge

const GraphResourceTypeDevice = models.GraphResourceTypeDevice
const GraphEdgeAttrSystemChanged = models.GraphEdgeAttrSystemChanged

type GraphListOptions = model.GraphListOptions

func (c *Client) ListGraphs(token string, options model.GraphListOptions) (result []models.Graph, total int64, err error, errCode int) {
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

	if options.Attributes != nil {
		encodeAsJson := slices.ContainsFunc(options.Attributes, func(attribute models.Attribute) bool {
			return attribute.Origin != "" ||
				strings.Contains(attribute.Key, ",") ||
				strings.Contains(attribute.Value, ",") ||
				strings.Contains(attribute.Key, ":") ||
				strings.Contains(attribute.Value, ":")
		})
		if encodeAsJson {
			b, err := json.Marshal(options.Attributes)
			if err != nil {
				return result, total, err, http.StatusBadRequest
			}
			query.Set("attributes_json", url.QueryEscape(string(b)))
		} else {
			parts := []string{}
			for _, attribute := range options.Attributes {
				if attribute.Value == "" {
					parts = append(parts, attribute.Key)
				} else {
					parts = append(parts, attribute.Key+":"+attribute.Value)
				}
			}
			query.Set("attributes", strings.Join(parts, ","))
		}

	}
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/graphs"+queryString, nil)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doWithTotalInResult[[]models.Graph](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ReadGraph(token string, id string) (result models.Graph, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/graphs/"+url.PathEscape(id), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Graph](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) SetGraph(token string, graph models.Graph) (result models.Graph, err error, code int) {
	method := http.MethodPost
	endpoint := c.baseUrl + "/graphs"
	if graph.Id != "" {
		method = http.MethodPut
		endpoint = c.baseUrl + "/graphs/" + url.PathEscape(graph.Id)
	}
	b, err := json.Marshal(graph)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Graph](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) DeleteGraph(token string, id string) (error, int) {
	req, err := http.NewRequest(http.MethodDelete, c.baseUrl+"/graphs/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}
