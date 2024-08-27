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

package testdb

import (
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"slices"
	"strings"
)

func (this *DB) EnsureInitialRights(resourceKind string, resourceId string, owner string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	_, err := this.getSecurityResource(resourceKind, resourceId, GetOptions{CheckPermission: false})
	if errors.Is(err, ErrNotFound) {
		err = nil
		initialRights, ok := this.config.InitialGroupRights[resourceKind]
		if ok {
			resource := Resource{
				Id:      resourceId,
				TopicId: resourceKind,
				ResourceRights: model.ResourceRights{
					UserRights:  map[string]model.Right{owner: {Read: true, Write: true, Execute: true, Administrate: true}},
					GroupRights: map[string]model.Right{},
				},
			}
			for group, rights := range initialRights {
				resource.GroupRights[group] = model.Right{
					Read:         strings.Contains(rights, "r"),
					Write:        strings.Contains(rights, "w"),
					Execute:      strings.Contains(rights, "x"),
					Administrate: strings.Contains(rights, "a"),
				}
			}
			this.permissions = append(this.permissions, resource)
		}
	}
	return err
}

func (this *DB) SetRights(resourceKind string, resourceId string, rights model.ResourceRights) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	for i, element := range this.permissions {
		if element.Id == resourceId && element.TopicId == resourceKind {
			this.permissions[i] = Resource{
				Id:             resourceId,
				TopicId:        resourceKind,
				ResourceRights: rights,
			}
			return nil
		}
	}
	this.permissions = append(this.permissions, Resource{
		Id:             resourceId,
		TopicId:        resourceKind,
		ResourceRights: rights,
	})
	return nil
}

func (this *DB) RemoveRights(topic string, id string) error {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.permissions = slices.DeleteFunc(this.permissions, func(element Resource) bool {
		return element.Id == id && element.TopicId == topic
	})
	return nil
}

func (this *DB) CheckBool(token string, kind string, id string, action model.AuthAction) (allowed bool, err error) {
	parsedToken, err := jwt.Parse(token)
	if err != nil {
		return false, err
	}
	_, err = this.getSecurityResource(kind, id, GetOptions{
		CheckPermission: true,
		UserId:          parsedToken.GetUserId(),
		GroupIds:        parsedToken.GetRoles(),
		Permissions:     []model.AuthAction{action},
	})
	if errors.Is(err, ErrNotFound) || errors.Is(err, PermissionCheckFailed) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (this *DB) CheckMultiple(token string, kind string, ids []string, action model.AuthAction) (result map[string]bool, err error) {
	result = map[string]bool{}
	parsedToken, err := jwt.Parse(token)
	if err != nil {
		return result, err
	}
	for _, id := range ids {
		_, err = this.getSecurityResource(kind, id, GetOptions{
			CheckPermission: true,
			UserId:          parsedToken.GetUserId(),
			GroupIds:        parsedToken.GetRoles(),
			Permissions:     []model.AuthAction{action},
		})
		if errors.Is(err, PermissionCheckFailed) {
			result[id] = false
			err = nil
			continue
		}
		if errors.Is(err, ErrNotFound) {
			err = nil
			continue
		}
		if err != nil {
			return result, err
		}
		result[id] = true
	}
	return result, nil
}

type GetOptions struct {
	CheckPermission bool
	UserId          string
	GroupIds        []string
	Permissions     []model.AuthAction
}

var PermissionCheckFailed = errors.New("permission check failed")
var ErrNotFound = errors.New("not found")

func (this *DB) getSecurityResource(topicId string, id string, options GetOptions) (resource Resource, err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	for _, element := range this.permissions {
		if element.TopicId == topicId && element.Id == id {
			if options.CheckPermission && !checkPerms(element, options.UserId, options.GroupIds, options.Permissions...) {
				return resource, PermissionCheckFailed
			}
			return element, nil
		}
	}
	return resource, ErrNotFound
}

func checkPerms(element Resource, user string, groups []string, permissions ...model.AuthAction) bool {
	for _, p := range permissions {
		if !checkPerm(element, user, groups, p) {
			return false
		}
	}
	return true
}

func checkPerm(element Resource, user string, groups []string, permission model.AuthAction) bool {
	switch permission {
	case model.READ:
		if element.UserRights[user].Read {
			return true
		}
		for _, g := range groups {
			if element.GroupRights[g].Read {
				return true
			}
		}
	case model.WRITE:
		if element.UserRights[user].Write {
			return true
		}
		for _, g := range groups {
			if element.GroupRights[g].Write {
				return true
			}
		}
	case model.EXECUTE:
		if element.UserRights[user].Execute {
			return true
		}
		for _, g := range groups {
			if element.GroupRights[g].Execute {
				return true
			}
		}
	case model.ADMINISTRATE:
		if element.UserRights[user].Administrate {
			return true
		}
		for _, g := range groups {
			if element.GroupRights[g].Administrate {
				return true
			}
		}
	}
	return false
}

func (this *DB) GetAdminUsers(token string, topic string, resourceId string) (admins []string, err error) {
	parsedToken, err := jwt.Parse(token)
	if err != nil {
		return admins, err
	}
	resource, err := this.getSecurityResource(topic, resourceId, GetOptions{
		CheckPermission: true,
		UserId:          parsedToken.GetUserId(),
		GroupIds:        parsedToken.GetRoles(),
		Permissions:     []model.AuthAction{model.ADMINISTRATE},
	})
	if err != nil {
		return admins, err
	}
	for user, rights := range resource.UserRights {
		if rights.Administrate {
			admins = append(admins, user)
		}
	}
	return admins, nil
}

func (this *DB) ListAccessibleResourceIds(token string, topic string, limit int64, offset int64, action model.AuthAction) (result []string, err error) {
	parsedToken, err := jwt.Parse(token)
	if err != nil {
		return result, err
	}
	for _, element := range this.permissions {
		if checkPerms(element, parsedToken.GetUserId(), parsedToken.GetRoles(), action) {
			result = append(result, element.Id)
		}
	}
	return result, nil
}

func (this *DB) GetPermissionsInfo(token string, kind string, id string) (requestingUser string, permissions models.Permissions, err error) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return requestingUser, permissions, err
	}
	requestingUser = jwtToken.GetUserId()
	resourceRights, err := this.getSecurityResource(kind, id, GetOptions{})
	if err != nil {
		return requestingUser, permissions, err
	}
	permissions = models.Permissions{
		Read:         resourceRights.UserRights[requestingUser].Read,
		Write:        resourceRights.UserRights[requestingUser].Write,
		Execute:      resourceRights.UserRights[requestingUser].Execute,
		Administrate: resourceRights.UserRights[requestingUser].Administrate,
	}
	for _, role := range jwtToken.GetRoles() {
		rolePermissions := resourceRights.GroupRights[role]
		if rolePermissions.Read {
			permissions.Read = true
		}
		if rolePermissions.Write {
			permissions.Write = true
		}
		if rolePermissions.Execute {
			permissions.Execute = true
		}
		if rolePermissions.Administrate {
			permissions.Administrate = true
		}
	}
	for _, group := range jwtToken.GetGroups() {
		groupPermissions := resourceRights.KeycloakGroupsRights[group]
		if groupPermissions.Read {
			permissions.Read = true
		}
		if groupPermissions.Write {
			permissions.Write = true
		}
		if groupPermissions.Execute {
			permissions.Execute = true
		}
		if groupPermissions.Administrate {
			permissions.Administrate = true
		}
	}
	return requestingUser, permissions, err
}
