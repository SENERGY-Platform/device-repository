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
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/v2/lib/model"
)

func (c *Client) DeviceGroupHelper(token string, deviceIds []string, options model.DeviceGroupHelperOptions) (result model.DeviceGroupHelperResult, err error, code int) {
	query := url.Values{}
	if options.Search != "" {
		query.Set("search", options.Search)
	}
	if options.Limit != 0 {
		query.Set("limit", strconv.FormatInt(options.Limit, 10))
	}
	if options.Offset != 0 {
		query.Set("offset", strconv.FormatInt(options.Offset, 10))
	}
	if options.MaintainsGroupUsability {
		query.Set("maintains_group_usability", "true")
	}
	if len(options.FunctionBlockList) > 0 {
		query.Set("function_block_list", strings.Join(options.FunctionBlockList, ","))
	}
	queryString := ""
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	if deviceIds == nil {
		deviceIds = []string{}
	}
	b, err := json.Marshal(deviceIds)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/device-group-helper"+queryString, bytes.NewBuffer(b))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[model.DeviceGroupHelperResult](req, c.optionalAuthTokenForApiGatewayRequest)
}
