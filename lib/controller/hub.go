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

package controller

import (
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"log"
	"net/http"
	"runtime/debug"
	"slices"
	"sync"
)

func (this *Controller) ListHubs(token string, options model.HubListOptions) (result []models.Hub, err error, errCode int) {
	ids := []string{}
	permissionFlag := options.Permission
	if permissionFlag == models.UnsetPermissionFlag {
		permissionFlag = models.Read
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if options.Ids == nil {
		if jwtToken.IsAdmin() {
			ids = nil //no auth check for admins -> no id filter
		} else {
			ids, err, _ = this.permissionsV2Client.ListAccessibleResourceIds(token, this.config.HubTopic, client.ListOptions{}, client.Permission(permissionFlag))
			if err != nil {
				return result, err, http.StatusInternalServerError
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err, _ := this.permissionsV2Client.CheckMultiplePermissions(token, this.config.HubTopic, options.Ids, client.Permission(permissionFlag))
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		for id, ok := range idMap {
			if ok {
				ids = append(ids, id)
			}
		}
	}
	options.Ids = ids

	if options.LocalDeviceId != "" && options.OwnerId == "" {
		options.OwnerId = jwtToken.GetUserId()
	}

	ctx, _ := getTimeoutContext()
	hubs, _, err := this.db.ListHubs(ctx, options, false)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	result = []models.Hub{}
	for _, hub := range hubs {
		result = append(result, hub.Hub)
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ListExtendedHubs(token string, options model.HubListOptions) (result []models.ExtendedHub, total int64, err error, errCode int) {
	ids := []string{}
	permissionFlag := options.Permission
	if permissionFlag == models.UnsetPermissionFlag {
		permissionFlag = models.Read
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, total, err, http.StatusBadRequest
	}
	if options.Ids == nil {
		if jwtToken.IsAdmin() {
			ids = nil //no auth check for admins -> no id filter
		} else {
			ids, err, _ = this.permissionsV2Client.ListAccessibleResourceIds(token, this.config.HubTopic, client.ListOptions{}, client.Permission(permissionFlag))
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err, _ := this.permissionsV2Client.CheckMultiplePermissions(token, this.config.HubTopic, options.Ids, client.Permission(permissionFlag))
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
		for id, ok := range idMap {
			if ok {
				ids = append(ids, id)
			}
		}
	}
	options.Ids = ids

	if options.LocalDeviceId != "" && options.OwnerId == "" {
		options.OwnerId = jwtToken.GetUserId()
	}

	ctx, _ := getTimeoutContext()
	hubs, total, err := this.db.ListHubs(ctx, options, true)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	result = make([]models.ExtendedHub, len(hubs))
	wg := sync.WaitGroup{}
	mux := sync.Mutex{}
	for i, hub := range hubs {
		wg.Add(1)
		go func(h model.HubWithConnectionState, resultIndex int) {
			defer wg.Done()
			extended, temperr := this.extendHub(token, h)
			mux.Lock()
			defer mux.Unlock()
			err = errors.Join(err, temperr)
			result[resultIndex] = extended
		}(hub, i)
	}
	wg.Wait()
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) ReadExtendedHub(id string, token string, action model.AuthAction) (result models.ExtendedHub, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	hub, exists, err := this.db.GetHub(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.HubTopic, id, client.Permission(action))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	result, err = this.extendHub(token, hub)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}

func (this *Controller) extendHub(token string, hub model.HubWithConnectionState) (models.ExtendedHub, error) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return models.ExtendedHub{}, err
	}
	requestingUser := jwtToken.GetUserId()
	computedPermList, err, _ := this.permissionsV2Client.ListComputedPermissions(token, this.config.HubTopic, []string{hub.Id})
	if err != nil {
		return models.ExtendedHub{}, err
	}
	if len(computedPermList) == 0 {
		return models.ExtendedHub{}, errors.New("no computation permissions")
	}
	permissions := models.Permissions{
		Read:         computedPermList[0].Read,
		Write:        computedPermList[0].Write,
		Execute:      computedPermList[0].Execute,
		Administrate: computedPermList[0].Administrate,
	}
	return models.ExtendedHub{
		Hub:             hub.Hub,
		ConnectionState: hub.ConnectionState,
		Shared:          requestingUser != hub.OwnerId,
		Permissions:     permissions,
	}, nil
}

func (this *Controller) ReadHub(id string, token string, action model.AuthAction) (result models.Hub, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	hub, exists, err := this.db.GetHub(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.HubTopic, id, client.Permission(action))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	return hub.Hub, nil, http.StatusOK
}

func (this *Controller) ListHubDeviceIds(id string, token string, action model.AuthAction, asLocalId bool) (result []string, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	hub, exists, err := this.db.GetHub(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.HubTopic, id, client.Permission(action))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	if asLocalId {
		return hub.DeviceLocalIds, nil, http.StatusOK
	} else {
		return hub.DeviceIds, nil, http.StatusOK
	}
}

func (this *Controller) ValidateHub(token string, hub models.Hub) (err error, code int) {
	if hub.Id == "" {
		return errors.New("missing hub id"), http.StatusBadRequest
	}
	if hub.Name == "" {
		return errors.New("missing hub name"), http.StatusBadRequest
	}
	if hub.OwnerId == "" {
		return errors.New("missing hub owner_id"), http.StatusBadRequest
	}
	for _, deviceId := range hub.DeviceIds {
		ctx, _ := getTimeoutContext()
		device, exists, err := this.db.GetDevice(ctx, deviceId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown device id: " + deviceId), http.StatusBadRequest
		}
		if device.OwnerId != hub.OwnerId {
			return errors.New("all devices in the hub must have the same owner_id as the hub"), http.StatusBadRequest
		}
	}
	ctx, _ := getTimeoutContext()
	original, exists, err := this.db.GetHub(ctx, hub.Id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if exists {
		if hub.OwnerId != original.OwnerId {
			resource, err, code := this.permissionsV2Client.GetResource(token, this.config.HubTopic, hub.Id)
			if err != nil {
				return err, code
			}
			admins := []string{}
			for user, perm := range resource.UserPermissions {
				if perm.Administrate {
					admins = append(admins, user)
				}
			}
			if len(admins) == 0 {
				//o admins indicates the requesting user has not the needed admin rights to see other admins
				return errors.New("requesting user must have admin rights to change owner"), http.StatusBadRequest
			}
			//new device owner-id must be existing admin user (ignore for new devices or devices with unchanged owner)
			if !slices.Contains(admins, hub.OwnerId) {
				return errors.New("new owner must have existing user admin rights"), http.StatusBadRequest
			}
		}
	}
	return this.ValidateHubDevices(hub)
}

func (this *Controller) ValidateHubDevices(hub models.Hub) (err error, code int) {
	for _, localId := range hub.DeviceLocalIds {
		//device exists?
		ctx, _ := getTimeoutContext()
		device, exists, err := this.db.GetDeviceByLocalId(ctx, hub.OwnerId, localId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown device local id: " + localId + "for owner " + hub.OwnerId), http.StatusBadRequest
		}
		if !slices.Contains(hub.DeviceIds, device.Id) {
			return fmt.Errorf("missing device.id %s in device_ids (found by device_local_ids %s)", device.Id, localId), http.StatusBadRequest
		}
	}
	for _, id := range hub.DeviceIds {
		//device exists?
		ctx, _ := getTimeoutContext()
		device, exists, err := this.db.GetDevice(ctx, id)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown device id: " + id), http.StatusBadRequest
		}
		if !slices.Contains(hub.DeviceIds, device.Id) {
			return fmt.Errorf("missing device.local_id %s in device_local_ids (found by device_ids %s)", device.LocalId, id), http.StatusBadRequest
		}
	}
	if len(hub.DeviceIds) != len(hub.DeviceLocalIds) {
		return errors.New("hub.device_ids length does not match hub.device_local_ids length"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func (this *Controller) SetHub(token string, hub models.Hub) (result models.Hub, err error, code int) {
	hub, err, code = this.completeHub(token, hub)
	if err != nil {
		return hub, err, code
	}
	if hub.Id == "" {
		hub.GenerateId()
	}
	ctx, _ := getTimeoutContext()
	old, exists, err := this.db.GetHub(ctx, hub.Id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists && hub.OwnerId != "" && hub.OwnerId != jwtToken.GetUserId() {
		return hub, errors.New("new devices must be initialised with the requesting user as owner-id"), http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() && exists {
		ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.HubTopic, hub.Id, client.Write)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		if !ok {
			return result, errors.New("access denied"), http.StatusForbidden
		}
	}

	//set device owner-id if none is given
	//prefer existing owner, fallback to requesting user
	if hub.OwnerId == "" {
		hub.OwnerId = old.OwnerId //may be empty for new devices
	}
	if hub.OwnerId == "" {
		hub.OwnerId = jwtToken.GetUserId()
	}

	if exists && old.OwnerId != hub.OwnerId && !jwtToken.IsAdmin() {
		ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.HubTopic, hub.Id, client.Administrate)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		if !ok {
			return hub, fmt.Errorf("only admins may set new owner: %w", err), http.StatusBadRequest
		}
	}

	permissions, err, code := this.permissionsV2Client.GetResource(token, this.config.HubTopic, hub.Id)
	if err != nil && code != http.StatusNotFound {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return hub, err, code
	}

	//new device owner-id must be existing admin user (ignore for new devices or devices with unchanged owner)
	if code != http.StatusNotFound && hub.OwnerId != old.OwnerId && !permissions.UserPermissions[hub.OwnerId].Administrate {
		return hub, errors.New("new owner must have existing user admin permissions"), http.StatusBadRequest
	}

	err, code = this.ValidateHub(token, hub)
	if err != nil {
		return hub, err, code
	}
	err = this.setHub(model.HubWithConnectionState{
		Hub:             hub,
		ConnectionState: old.ConnectionState,
	}, hub.OwnerId)
	if err != nil {
		return hub, err, http.StatusInternalServerError
	}
	return hub, nil, http.StatusOK
}

func (this *Controller) completeHub(token string, edit models.Hub) (result models.Hub, err error, code int) {
	result = edit
	if result.DeviceLocalIds == nil {
		result.DeviceLocalIds = []string{}
	}
	if result.DeviceIds == nil {
		result.DeviceIds = []string{}
	}

	if len(edit.DeviceLocalIds) > 0 {
		devices, err, code := this.ListDevices(token, model.DeviceListOptions{LocalIds: edit.DeviceLocalIds, Owner: edit.OwnerId})
		if err != nil {
			return result, err, code
		}
		if len(devices) != len(edit.DeviceLocalIds) {
			return result, errors.New("not all local device ids found"), http.StatusBadRequest
		}
		for _, device := range devices {
			if !slices.Contains(result.DeviceLocalIds, device.LocalId) {
				result.DeviceLocalIds = append(result.DeviceLocalIds, device.LocalId)
			}
			if !slices.Contains(result.DeviceIds, device.Id) {
				result.DeviceIds = append(result.DeviceIds, device.Id)
			}
		}
	}

	if len(edit.DeviceIds) > 0 {
		devices, err, code := this.ListDevices(token, model.DeviceListOptions{Ids: edit.DeviceIds})
		if err != nil {
			return result, err, code
		}
		if len(devices) != len(edit.DeviceIds) {
			return result, errors.New("not all device ids found"), http.StatusBadRequest
		}
		for _, device := range devices {
			if !slices.Contains(result.DeviceLocalIds, device.LocalId) {
				result.DeviceLocalIds = append(result.DeviceLocalIds, device.LocalId)
			}
			if !slices.Contains(result.DeviceIds, device.Id) {
				result.DeviceIds = append(result.DeviceIds, device.Id)
			}
		}
	}

	if len(result.DeviceLocalIds) != len(result.DeviceIds) {
		return result, errors.New("DeviceLocalIds DeviceIds count mismatch"), http.StatusBadRequest
	}
	return result, err, code
}

func (this *Controller) setHubSyncHandler(hub model.HubWithConnectionState) (err error) {
	err = this.EnsureInitialRights(this.config.HubTopic, hub.Id, hub.OwnerId)
	if err != nil {
		return err
	}

	//remove devices fom other hubs
	hubIndex := map[string]model.HubWithConnectionState{}
	for _, id := range hub.DeviceIds {
		ctx, _ := getTimeoutContext()
		hubs, err := this.db.GetHubsByDeviceId(ctx, id)
		if err != nil {
			return err
		}
		for _, otherHub := range hubs {
			if otherHub.Id != hub.Id {
				hubIndex[otherHub.Id] = otherHub
			}
		}
	}
	for _, lid := range hub.DeviceLocalIds {
		for _, otherHub := range hubIndex {
			otherHub.DeviceLocalIds = filter(otherHub.DeviceLocalIds, lid)
			hubIndex[otherHub.Id] = otherHub
		}
	}
	for _, id := range hub.DeviceIds {
		for _, otherHub := range hubIndex {
			otherHub.DeviceIds = filter(otherHub.DeviceIds, id)
			hubIndex[otherHub.Id] = otherHub
		}
	}
	for _, otherHub := range hubIndex {
		err := this.setHub(otherHub, otherHub.OwnerId)
		if err != nil {
			return err
		}
	}
	return this.publisher.PublishHub(hub.Hub)
}

func (this *Controller) setHub(hub model.HubWithConnectionState, owner string) (err error) {
	if hub.Id == "" {
		log.Println("ERROR: received hub without id")
		return nil
	}
	ctx, _ := getTimeoutContext()
	err = this.db.SetHub(ctx, hub, this.setHubSyncHandler)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) DeleteHub(token string, id string) (err error, code int) {
	ctx, _ := getTimeoutContext()
	_, exists, err := this.db.GetHub(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return nil, http.StatusOK
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.HubTopic, id, client.Administrate)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("access denied"), http.StatusForbidden
	}
	err = this.deleteHub(id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteHubSyncHandler(hub model.HubWithConnectionState) (err error) {
	err = this.RemoveRights(this.config.HubTopic, hub.Id)
	if err != nil {
		return err
	}
	return this.publisher.PublishHubDelete(hub.Id)
}

func (this *Controller) deleteHub(id string) (err error) {
	ctx, _ := getTimeoutContext()
	err = this.db.RemoveHub(ctx, id, this.deleteHubSyncHandler)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) SetHubConnectionState(token string, id string, connected bool) (error, int) {
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.DeviceTopic, id, client.Write)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("access denied"), http.StatusForbidden
	}
	state := models.ConnectionStateOffline
	if connected {
		state = models.ConnectionStateOnline
	}
	ctx, _ := getTimeoutContext()
	err = this.db.SetHubConnectionState(ctx, id, state)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
