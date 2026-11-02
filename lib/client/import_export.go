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
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/SENERGY-Platform/device-repository/v3/lib/controller"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	permissions "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

type ImportExportOptions = model.ImportExportOptions
type ImportExport = model.ImportExport
type Resource = permissions.Resource
type ResourcePermissions = permissions.ResourcePermissions
type PermissionsMap = permissions.PermissionsMap
type ImportFromOptions = model.ImportFromOptions

func (c *Client) Export(ctx context.Context, token string, options model.ImportExportOptions) (result model.ImportExport, err error, code int) {
	req, err := controller.GetExportHttpRequest(c.baseUrl, token, options)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return do[model.ImportExport](c, req)
}

func (c *Client) Import(ctx context.Context, token string, importModel model.ImportExport, options model.ImportExportOptions) (err error, code int) {
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseUrl+"/import"+queryString, bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}

func (c *Client) ImportFrom(ctx context.Context, token string, includeOwnedInformation bool, options model.ImportFromOptions) (err error, code int) {
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+"/import-from"+queryString, bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	return doVoid(c, req)
}
