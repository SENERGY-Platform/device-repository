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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type Graph = models.Graph
type Node = models.Node
type Edge = models.Edge
type GraphSegment = models.GraphSegment

const GraphEditModeChange = models.GraphEditModeChange
const GraphEditModeCorrection = models.GraphEditModeCorrection

const GraphResourceTypeDevice = models.GraphResourceTypeDevice
const GraphEdgeAttrSystemChanged = models.GraphEdgeAttrSystemChanged

type GraphListOptions = model.GraphListOptions

func (c *Client) ListGraphs(ctx context.Context, token string, options model.GraphListOptions) (result []models.Graph, total int64, err error, errCode int) {
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/graphs"+queryString, nil)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doCachedWithTotal[[]models.Graph](c, req)
}

func (c *Client) ReadGraph(ctx context.Context, token string, id string) (result models.Graph, err error, errCode int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/graphs/"+url.PathEscape(id), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doCached[models.Graph](c, req)
}

// ReadGraphAt reads the graph as it is valid at the given instant.
func (c *Client) ReadGraphAt(ctx context.Context, token string, id string, at time.Time) (result models.Graph, err error, errCode int) {
	query := url.Values{}
	query.Set("at", at.UTC().Format(time.RFC3339Nano))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/graphs/"+url.PathEscape(id)+"?"+query.Encode(), nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doCached[models.Graph](c, req)
}

// ReadGraphVersions reads the stored document with all versions of its elements.
func (c *Client) ReadGraphVersions(ctx context.Context, token string, id string) (result models.Graph, err error, errCode int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+"/graphs/"+url.PathEscape(id)+"?versions=all", nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doCached[models.Graph](c, req)
}

// ReadGraphSegments reads the graph cut at every validity bound inside [start, end); nil leaves a side open.
func (c *Client) ReadGraphSegments(ctx context.Context, token string, id string, start *time.Time, end *time.Time) (result []models.GraphSegment, err error, errCode int) {
	query := url.Values{}
	if start != nil {
		query.Set("start", start.UTC().Format(time.RFC3339Nano))
	}
	if end != nil {
		query.Set("end", end.UTC().Format(time.RFC3339Nano))
	}
	endpoint := c.baseUrl + "/graphs/" + url.PathEscape(id) + "/segments"
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doCached[[]models.GraphSegment](c, req)
}

// SetGraph creates a graph, or corrects an existing one at the current time: the given graph is the graph as it is now.
// Versions of elements that are not valid now are kept.
func (c *Client) SetGraph(ctx context.Context, token string, graph models.Graph) (result models.Graph, err error, code int) {
	return c.EditGraph(ctx, token, graph, model.GraphEditOptions{})
}

// EditGraph writes an existing graph (a graph without id is created): the given graph is the edited graph as it is
// valid at options.At, see PUT /graphs/{id}. The result is the stored graph as it is now.
func (c *Client) EditGraph(ctx context.Context, token string, graph models.Graph, options model.GraphEditOptions) (result models.Graph, err error, code int) {
	method := http.MethodPost
	endpoint := c.baseUrl + "/graphs"
	if graph.Id != "" {
		method = http.MethodPut
		endpoint = c.baseUrl + "/graphs/" + url.PathEscape(graph.Id)
		query := url.Values{}
		if options.At != nil {
			query.Set("at", options.At.UTC().Format(time.RFC3339Nano))
		}
		if options.Until != nil {
			query.Set("until", options.Until.UTC().Format(time.RFC3339Nano))
		}
		if options.Mode != "" {
			query.Set("mode", options.Mode)
		}
		if len(query) > 0 {
			endpoint += "?" + query.Encode()
		}
	}
	b, err := json.Marshal(graph)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Graph](c, req)
}

type GraphNodeMove = model.GraphNodeMove
type GraphNodeEnd = model.GraphNodeEnd

// MoveGraphNode moves a node below another node from move.At on, see POST /graphs/{id}/nodes/{node_id}/move.
// The result is the graph as it is now.
func (c *Client) MoveGraphNode(ctx context.Context, token string, graphId string, nodeId string, move model.GraphNodeMove) (result models.Graph, err error, code int) {
	return c.postGraphNodeAction(ctx, token, graphId, nodeId, "move", move)
}

// EndGraphNode takes a node out of the graph from end.At on, see POST /graphs/{id}/nodes/{node_id}/end.
// An open device link of the node's thing is not ended. The result is the graph as it is now.
func (c *Client) EndGraphNode(ctx context.Context, token string, graphId string, nodeId string, end model.GraphNodeEnd) (result models.Graph, err error, code int) {
	return c.postGraphNodeAction(ctx, token, graphId, nodeId, "end", end)
}

func (c *Client) postGraphNodeAction(ctx context.Context, token string, graphId string, nodeId string, action string, body any) (result models.Graph, err error, code int) {
	b, err := json.Marshal(body)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/graphs/"+url.PathEscape(graphId)+"/nodes/"+url.PathEscape(nodeId)+"/"+action, bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[models.Graph](c, req)
}

func (c *Client) DeleteGraph(ctx context.Context, token string, id string) (error, int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+"/graphs/"+url.PathEscape(id), nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}

func (c *Client) DeleteGraphs(ctx context.Context, token string, ids []string) (err error, code int) {
	return c.deleteMany(ctx, token, "/graphs", ids)
}
