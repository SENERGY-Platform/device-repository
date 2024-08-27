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
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permission-search/lib/client"
	permmodel "github.com/SENERGY-Platform/permission-search/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"log"
	"runtime/debug"
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

func (this *Security) GetAdminUsers(token string, topic string, resourceId string) (admins []string, err error) {
	rights, _, err := this.GetResourceRights(token, topic, resourceId, "a")
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return admins, err
	}
	return rights.PermissionHolders.AdminUsers, nil
}

func (this *Security) GetResourceRights(token string, kind string, id string, rights string) (result permmodel.EntryResult, found bool, err error) {
	temp, _, err := client.Query[[]permmodel.EntryResult](this.permissionsearch, token, client.QueryMessage{
		Resource: kind,
		ListIds: &permmodel.QueryListIds{
			QueryListCommons: permmodel.QueryListCommons{
				Limit:  1,
				Offset: 0,
				Rights: rights,
			},
			Ids: []string{id},
		},
	})
	if err != nil {
		return result, false, err
	}
	if len(temp) == 0 {
		return result, false, nil
	}
	return temp[0], true, nil
}

type IdWrapper struct {
	Id string `json:"id"`
}

func (this *Security) ListAccessibleResourceIds(token string, topic string, limit int64, offset int64, action model.AuthAction) (result []string, err error) {
	list, err := client.List[[]IdWrapper](this.permissionsearch, token, topic, client.ListOptions{
		QueryListCommons: permmodel.QueryListCommons{
			Limit:  int(limit),
			Offset: int(offset),
			Rights: string(action),
		},
	})
	if err != nil {
		return result, err
	}
	for _, element := range list {
		result = append(result, element.Id)
	}
	return result, err
}

func (this *Security) GetPermissionsInfo(token string, kind string, id string) (requestingUser string, permissions models.Permissions, err error) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return requestingUser, permissions, err
	}
	requestingUser = jwtToken.GetUserId()
	permentry, _, err := this.GetResourceRights(token, kind, id, "r")
	if err != nil {
		return requestingUser, permissions, err
	}
	permissions = models.Permissions{
		Read:         permentry.Permissions["r"],
		Write:        permentry.Permissions["w"],
		Execute:      permentry.Permissions["x"],
		Administrate: permentry.Permissions["a"],
	}
	return requestingUser, permissions, nil
}
