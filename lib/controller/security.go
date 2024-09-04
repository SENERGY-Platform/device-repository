/*
 * Copyright 2024 InfAI (CC SES)
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

package controller

import (
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

func (c *Controller) GetPermissionsClient() client.Client {
	return c.permissionsV2Client
}

func (this *Controller) EnsureInitialRights(topic string, resourceId string, owner string) error {
	exists, err := this.db.RightsElementExists(topic, resourceId)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	defaultpermissions := this.getDefaultEntryPermissions(topic, owner)
	if this.config.PermissionsV2Url != "" {
		_, err, _ = this.permissionsV2Client.SetPermission(client.InternalAdminToken, topic, resourceId, defaultpermissions.ToPermV2Permissions())
		if err != nil {
			return err
		}
	}
	return this.db.SetRights(topic, resourceId, defaultpermissions)
}

func (this *Controller) SetRights(resourceKind string, resourceId string, rights model.ResourceRights) error {
	return this.db.SetRights(resourceKind, resourceId, rights)
}

func (this *Controller) RemoveRights(topic string, id string) error {
	if this.config.PermissionsV2Url != "" {
		err, _ := this.permissionsV2Client.RemoveResource(client.InternalAdminToken, topic, id)
		if err != nil {
			return err
		}
	}
	return this.db.RemoveRights(topic, id)
}

func (this *Controller) getDefaultEntryPermissions(topic string, owner string) (entry model.ResourceRights) {
	entry = model.ResourceRights{
		UserRights:           map[string]model.Right{},
		GroupRights:          map[string]model.Right{},
		KeycloakGroupsRights: map[string]model.Right{},
	}
	if owner != "" {
		entry.UserRights[owner] = model.Right{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		}
	}
	for group, rights := range this.config.InitialGroupRights[topic] {
		perm := model.Right{}
		for _, right := range rights {
			switch right {
			case 'a':
				perm.Administrate = true
			case 'r':
				perm.Read = true
			case 'w':
				perm.Write = true
			case 'x':
				perm.Execute = true
			}
		}
		entry.GroupRights[group] = perm
	}
	return
}
