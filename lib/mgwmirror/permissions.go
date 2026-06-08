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

package mgwmirror

import (
	"context"
	"fmt"

	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	model2 "github.com/SENERGY-Platform/permissions-v2/pkg/model"
)

type MgwMirrorPerm struct {
	db     database.Database
	config configuration.Config
}

func NewMgwMirrorPerm(config configuration.Config, db database.Database) *MgwMirrorPerm {
	return &MgwMirrorPerm{
		db:     db,
		config: config,
	}
}

func (this *MgwMirrorPerm) CheckPermission(token string, topicId string, id string, permissions ...model2.Permission) (access bool, err error, code int) {
	return true, nil, 200
}

func (this *MgwMirrorPerm) CheckMultiplePermissions(token string, topicId string, ids []string, permissions ...model2.Permission) (access map[string]bool, err error, code int) {
	access = map[string]bool{}
	for _, id := range ids {
		access[id] = true
	}
	return access, nil, 200
}

func (this *MgwMirrorPerm) ListAccessibleResourceIds(token string, topicId string, options model2.ListOptions, permissions ...model2.Permission) (ids []string, err error, code int) {
	switch topicId {
	case this.config.DeviceTopic:
		devices, _, err := this.db.ListDevices(context.Background(), model.DeviceListOptions{Ids: options.Ids, Limit: options.Limit, Offset: options.Offset}, false)
		if err != nil {
			return nil, err, 500
		}
		for _, device := range devices {
			ids = append(ids, device.Id)
		}
		return ids, nil, 200
	case this.config.DeviceGroupTopic:
		groups, _, err := this.db.ListDeviceGroups(context.Background(), model.DeviceGroupListOptions{Ids: options.Ids, Limit: options.Limit, Offset: options.Offset})
		if err != nil {
			return nil, err, 500
		}
		for _, group := range groups {
			ids = append(ids, group.Id)
		}
		return ids, nil, 200
	case this.config.HubTopic:
		hubs, _, err := this.db.ListHubs(context.Background(), model.HubListOptions{Ids: options.Ids, Limit: options.Limit, Offset: options.Offset}, false)
		if err != nil {
			return nil, err, 500
		}
		for _, hub := range hubs {
			ids = append(ids, hub.Id)
		}
		return ids, nil, 200
	case this.config.LocationTopic:
		locations, _, err := this.db.ListLocations(context.Background(), model.LocationListOptions{Ids: options.Ids, Limit: options.Limit, Offset: options.Offset})
		if err != nil {
			return nil, err, 500
		}
		for _, location := range locations {
			ids = append(ids, location.Id)
		}
		return ids, nil, 200
	}
	return nil, fmt.Errorf("unknown topic"), 404
}

func (this *MgwMirrorPerm) ListComputedPermissions(token string, topic string, ids []string) (result []model2.ComputedPermissions, err error, code int) {
	for _, id := range ids {
		result = append(result, model2.ComputedPermissions{
			Id: id,
			PermissionsMap: model2.PermissionsMap{
				Read:         true,
				Write:        true,
				Execute:      true,
				Administrate: true,
			},
		})
	}
	return result, nil, 200
}

func (this *MgwMirrorPerm) ListTopics(token string, options model2.ListOptions) (result []model2.Topic, err error, code int) {
	return nil, nil, 200
}

func (this *MgwMirrorPerm) GetTopic(token string, id string) (result model2.Topic, err error, code int) {
	return model2.Topic{Id: id}, nil, 200
}

func (this *MgwMirrorPerm) RemoveTopic(token string, id string) (err error, code int) {
	return nil, 200
}

func (this *MgwMirrorPerm) SetTopic(token string, topic model2.Topic) (result model2.Topic, err error, code int) {
	return topic, nil, 200
}

func (this *MgwMirrorPerm) AdminListResourceIds(tokenStr string, topicId string, options model2.ListOptions) (ids []string, err error, code int) {
	return this.ListAccessibleResourceIds(tokenStr, topicId, options)
}

func (this *MgwMirrorPerm) AdminLoadFromPermissionSearch(req model2.AdminLoadPermSearchRequest) (updateCount int, err error, code int) {
	return 0, nil, 200
}

func (this *MgwMirrorPerm) Export(token string, options model2.ImportExportOptions) (result model2.ImportExport, err error, code int) {
	return model2.ImportExport{}, nil, 200
}

func (this *MgwMirrorPerm) Import(token string, importModel model2.ImportExport, options model2.ImportExportOptions) (err error, code int) {
	return nil, 200
}

func (this *MgwMirrorPerm) ListResourcesWithAdminPermission(token string, topicId string, options model2.ListOptions) (result []model2.Resource, err error, code int) {
	return nil, nil, 200
}

func (this *MgwMirrorPerm) GetResource(token string, topicId string, id string) (result model2.Resource, err error, code int) {
	userPerm := map[string]model2.PermissionsMap{}
	userId, _ := this.config.GetMgwMirrorUserId()
	if userId != "" {
		userPerm[userId] = model2.PermissionsMap{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		}
	}
	return model2.Resource{
		Id:      id,
		TopicId: topicId,
		ResourcePermissions: model2.ResourcePermissions{
			UserPermissions: userPerm,
		},
	}, nil, 200
}

func (this *MgwMirrorPerm) RemoveResource(token string, topicId string, id string) (err error, code int) {
	return nil, 200
}

func (this *MgwMirrorPerm) SetPermission(token string, topicId string, id string, permissions model2.ResourcePermissions) (result model2.ResourcePermissions, err error, code int) {
	return permissions, nil, 200
}
