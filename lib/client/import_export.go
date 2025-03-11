/*
 * Copyright 2025 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	permissions "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"net/http"
	"net/url"
)

type ImportExportOptions = model.ImportExportOptions
type ImportExport = model.ImportExport
type Resource = permissions.Resource
type ResourcePermissions = permissions.ResourcePermissions
type PermissionsMap = permissions.PermissionsMap
type ImportFromOptions = model.ImportFromOptions

func (c *Client) Export(token string, options model.ImportExportOptions) (result model.ImportExport, err error, code int) {
	req, err := controller.GetExportHttpRequest(c.baseUrl, token, options)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[model.ImportExport](req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) Import(token string, importModel model.ImportExport, options model.ImportExportOptions) (err error, code int) {
	queryString := ""
	query := url.Values{}
	if options.IncludeOwnedInformation {
		query.Set("include_owned_information", "true")
	}
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	b, err := json.Marshal(importModel)
	if err != nil {
		return err, http.StatusBadRequest
	}
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+"/import"+queryString, bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}

func (c *Client) ImportFrom(token string, includeOwnedInformation bool, options model.ImportFromOptions) (err error, code int) {
	b, err := json.Marshal(options)
	if err != nil {
		return err, http.StatusBadRequest
	}
	queryString := ""
	query := url.Values{}
	if includeOwnedInformation {
		query.Set("include_owned_information", "true")
	}
	if len(query) > 0 {
		queryString = "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodPost, c.baseUrl+"/import-from"+queryString, bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(req, c.optionalAuthTokenForApiGatewayRequest)
}
