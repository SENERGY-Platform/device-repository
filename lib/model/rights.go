/*
 * Copyright 2024InfAI (CC SES)
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

package model

import (
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	model2 "github.com/SENERGY-Platform/permissions-v2/pkg/model"
)

type ResourceRights struct {
	UserRights           map[string]Right `json:"user_rights"`
	GroupRights          map[string]Right `json:"group_rights"`
	KeycloakGroupsRights map[string]Right `json:"keycloak_groups_rights"`
}

type Right struct {
	Read         bool `json:"read"`
	Write        bool `json:"write"`
	Execute      bool `json:"execute"`
	Administrate bool `json:"administrate"`
}

func (this *ResourceRights) ToPermV2Permissions() client.ResourcePermissions {
	result := client.ResourcePermissions{
		UserPermissions:  map[string]model2.PermissionsMap{},
		GroupPermissions: map[string]model2.PermissionsMap{},
		RolePermissions:  map[string]model2.PermissionsMap{},
	}
	for k, v := range this.UserRights {
		result.UserPermissions[k] = model2.PermissionsMap{
			Read:         v.Read,
			Write:        v.Write,
			Execute:      v.Execute,
			Administrate: v.Administrate,
		}
	}
	for k, v := range this.GroupRights {
		result.RolePermissions[k] = model2.PermissionsMap{
			Read:         v.Read,
			Write:        v.Write,
			Execute:      v.Execute,
			Administrate: v.Administrate,
		}
	}
	for k, v := range this.KeycloakGroupsRights {
		result.GroupPermissions[k] = model2.PermissionsMap{
			Read:         v.Read,
			Write:        v.Write,
			Execute:      v.Execute,
			Administrate: v.Administrate,
		}
	}
	return result
}
