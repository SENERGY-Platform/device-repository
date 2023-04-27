/*
 * Copyright 2019 InfAI (CC SES)
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

package com

import (
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/permission-search/lib/client"
)

func NewSecurity(config config.Config) (*Security, error) {
	return &Security{config: config, permissionsearch: client.NewClient(config.PermissionsUrl)}, nil
}

type Security struct {
	config           config.Config
	permissionsearch client.Client
}

func (this *Security) CheckBool(token string, kind string, id string, action model.AuthAction) (allowed bool, err error) {
	err = this.permissionsearch.CheckUserOrGroup(token, kind, id, string(action))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, client.ErrAccessDenied) {
		return false, nil
	}
	if errors.Is(err, client.ErrNotFound) {
		return false, nil
	}
	return allowed, err
}

func (this *Security) CheckMultiple(token string, kind string, ids []string, action model.AuthAction) (result map[string]bool, err error) {
	query := client.QueryMessage{
		Resource: kind,
		CheckIds: &client.QueryCheckIds{
			Ids:    ids,
			Rights: string(action),
		},
	}
	result, _, err = client.Query[map[string]bool](this.permissionsearch, token, query)
	return result, err
}
