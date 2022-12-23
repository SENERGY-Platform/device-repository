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
	"net/http"
	"strconv"
	"strings"
)

func (c *Client) GetDeviceTypeSelectables(query []model.FilterCriteria, pathPrefix string, interactionsFilter []string, includeModified bool) (result []model.DeviceTypeSelectable, err error, code int) {
	body, err := json.Marshal(query)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/query/device-type-selectables?path-prefix="+pathPrefix+
		"&interactions-filter="+strings.Join(interactionsFilter, ",")+"&include_id_modified="+strconv.FormatBool(includeModified), bytes.NewBuffer(body))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[[]model.DeviceTypeSelectable](req)
}

func (c *Client) GetDeviceTypeSelectablesV2(query []model.FilterCriteria, pathPrefix string, includeModified bool) (result []model.DeviceTypeSelectable, err error, code int) {
	body, err := json.Marshal(query)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/v2/query/device-type-selectables?path-prefix="+pathPrefix+
		"&include_id_modified="+strconv.FormatBool(includeModified), bytes.NewBuffer(body))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[[]model.DeviceTypeSelectable](req)
}
