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

package controller

import (
	"errors"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	gojwt "github.com/golang-jwt/jwt"
	"net/http"
	"slices"
	"strings"
	"time"
)

func (this *Controller) DeleteUser(adminToken string, userId string) (err error, errCode int) {
	jwtToken, err := jwt.Parse(adminToken)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return errors.New("token is not an admin"), http.StatusUnauthorized
	}

	token, err := mockUserToken("device-repository", userId)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	//devices
	devicesToDelete, userToDeleteFromDevices, err := this.resourcesEffectedByUserDelete(token, this.config.DeviceTopic)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, id := range devicesToDelete {
		err = this.deleteDevice(id)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}
	for _, r := range userToDeleteFromDevices {
		delete(r.UserPermissions, userId)
		_, err, _ = this.permissionsV2Client.SetPermission(client.InternalAdminToken, this.config.DeviceTopic, r.Id, r.ResourcePermissions)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}
	//device-groups
	deviceGroupToDelete, userToDeleteFromDeviceGroups, err := this.resourcesEffectedByUserDelete(token, this.config.DeviceGroupTopic)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, id := range deviceGroupToDelete {
		err = this.deleteDeviceGroup(id)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}
	for _, r := range userToDeleteFromDeviceGroups {
		delete(r.UserPermissions, userId)
		_, err, _ = this.permissionsV2Client.SetPermission(client.InternalAdminToken, this.config.DeviceGroupTopic, r.Id, r.ResourcePermissions)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}
	//hubs
	hubToDelete, userToDeleteFromHubs, err := this.resourcesEffectedByUserDelete(token, this.config.HubTopic)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, id := range hubToDelete {
		err = this.deleteHub(id)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}
	for _, r := range userToDeleteFromHubs {
		delete(r.UserPermissions, userId)
		_, err, _ = this.permissionsV2Client.SetPermission(client.InternalAdminToken, this.config.HubTopic, r.Id, r.ResourcePermissions)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}
	//locations
	locationToDelete, userToDeleteFromLocations, err := this.resourcesEffectedByUserDelete(token, this.config.LocationTopic)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, id := range locationToDelete {
		err = this.deleteLocation(id)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}
	for _, r := range userToDeleteFromLocations {
		delete(r.UserPermissions, userId)
		_, err, _ = this.permissionsV2Client.SetPermission(client.InternalAdminToken, this.config.LocationTopic, r.Id, r.ResourcePermissions)
		if err != nil {
			return err, http.StatusInternalServerError
		}
	}
	return nil, http.StatusOK
}

var ResourcesEffectedByUserDelete_BATCH_SIZE int64 = 1000

func (this *Controller) resourcesEffectedByUserDelete(token jwt.Token, resource string) (deleteResourceIds []string, deleteUserFromResource []client.Resource, err error) {
	userid := token.GetUserId()
	err = this.iterateResource(token, resource, ResourcesEffectedByUserDelete_BATCH_SIZE, client.Administrate, func(element client.Resource) {
		if containsOtherAdmin(element.UserPermissions, userid) {
			deleteUserFromResource = append(deleteUserFromResource, element)
		} else {
			deleteResourceIds = append(deleteResourceIds, element.Id)
		}
	})
	if err != nil {
		return
	}

	err = this.iterateResource(token, resource, ResourcesEffectedByUserDelete_BATCH_SIZE, client.Read, func(element client.Resource) {
		if !slices.ContainsFunc(deleteUserFromResource, func(resource client.Resource) bool {
			return resource.Id == element.Id
		}) {
			if containsOtherAdmin(element.UserPermissions, userid) {
				deleteUserFromResource = append(deleteUserFromResource, element)
			} else {
				deleteResourceIds = append(deleteResourceIds, element.Id)
			}
		}
	})
	if err != nil {
		return
	}
	err = this.iterateResource(token, resource, ResourcesEffectedByUserDelete_BATCH_SIZE, client.Write, func(element client.Resource) {
		if !slices.ContainsFunc(deleteUserFromResource, func(resource client.Resource) bool {
			return resource.Id == element.Id
		}) {
			if containsOtherAdmin(element.UserPermissions, userid) {
				deleteUserFromResource = append(deleteUserFromResource, element)
			} else {
				deleteResourceIds = append(deleteResourceIds, element.Id)
			}
		}
	})
	if err != nil {
		return
	}
	err = this.iterateResource(token, resource, ResourcesEffectedByUserDelete_BATCH_SIZE, client.Execute, func(element client.Resource) {
		if !slices.ContainsFunc(deleteUserFromResource, func(resource client.Resource) bool {
			return resource.Id == element.Id
		}) {
			if containsOtherAdmin(element.UserPermissions, userid) {
				deleteUserFromResource = append(deleteUserFromResource, element)
			} else {
				deleteResourceIds = append(deleteResourceIds, element.Id)
			}
		}
	})
	if err != nil {
		return
	}
	return deleteResourceIds, deleteUserFromResource, err
}

func (this *Controller) iterateResource(token jwt.Token, resource string, batchsize int64, rights client.Permission, handler func(element client.Resource)) (err error) {
	lastCount := batchsize
	var offset int64 = 0
	for lastCount == batchsize {
		options := client.ListOptions{
			Limit:  batchsize,
			Offset: offset,
		}
		offset += batchsize
		ids, err, _ := this.permissionsV2Client.ListAccessibleResourceIds(token.Jwt(), resource, options, rights)
		if err != nil {
			return err
		}
		lastCount = int64(len(ids))
		for _, id := range ids {
			element, err, _ := this.permissionsV2Client.GetResource(client.InternalAdminToken, resource, id)
			if err != nil {
				return err
			}
			handler(element)
		}
	}
	return err
}

func containsOtherAdmin(m map[string]client.PermissionsMap, notThisKey string) bool {
	for k, v := range m {
		if k != notThisKey && v.Administrate {
			return true
		}
	}
	return false
}

type KeycloakClaims struct {
	RealmAccess map[string][]string `json:"realm_access"`
	gojwt.StandardClaims
}

func mockUserToken(issuer string, userId string) (token jwt.Token, err error) {
	realmAccess := map[string][]string{"roles": {}}
	claims := KeycloakClaims{
		realmAccess,
		gojwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
			Issuer:    issuer,
			Subject:   userId,
		},
	}

	jwtoken := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	unsignedTokenString, err := jwtoken.SigningString()
	if err != nil {
		return token, err
	}
	tokenString := strings.Join([]string{unsignedTokenString, ""}, ".")
	return jwt.Parse("Bearer " + tokenString)
}
