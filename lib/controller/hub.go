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
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"log"
	"net/http"
	"runtime/debug"
	"slices"
	"sync"
)

/////////////////////////
//		api
/////////////////////////

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
			ids, err = this.db.ListAccessibleResourceIds(token, this.config.HubTopic, 0, 0, permissionFlag)
			if err != nil {
				return result, err, http.StatusInternalServerError
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err := this.db.CheckMultiple(token, this.config.HubTopic, options.Ids, permissionFlag)
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
			ids, err = this.db.ListAccessibleResourceIds(token, this.config.HubTopic, 0, 0, permissionFlag)
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err := this.db.CheckMultiple(token, this.config.HubTopic, options.Ids, permissionFlag)
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
	ok, err := this.db.CheckBool(token, this.config.HubTopic, id, action)
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
	requestingUser, permissions, err := this.db.GetPermissionsInfo(token, this.config.HubTopic, hub.Id)
	if err != nil {
		return models.ExtendedHub{}, err
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
	ok, err := this.db.CheckBool(token, this.config.HubTopic, id, action)
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
	ok, err := this.db.CheckBool(token, this.config.HubTopic, id, action)
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
			admins, err := this.db.GetAdminUsers(token, this.config.HubTopic, hub.Id)
			if errors.Is(err, model.PermissionCheckFailed) {
				return errors.New("requesting user must have admin rights to change owner"), http.StatusBadRequest
			}
			if err != nil {
				return err, http.StatusInternalServerError
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

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetHub(hub models.Hub, owner string) (err error) {
	if hub.Id == "" {
		log.Println("ERROR: received hub without id")
		return nil
	}
	ctx, _ := getTimeoutContext()
	old, exists, err := this.db.GetHub(ctx, hub.Id)
	if err != nil {
		return err
	}
	if hub.OwnerId == "" {
		if exists {
			hub.OwnerId = old.OwnerId
		} else {
			hub.OwnerId = owner
		}
	}
	if err, _ := this.ValidateHubDevices(hub); err != nil {
		log.Printf("ERROR: received invalid hub from kafka\n%v\n%#v\n", err, hub)
		debug.PrintStack()
		if hub.Name == "" {
			hub.Name = "generated-name"
		}
		hub.DeviceIds = []string{}
		hub.DeviceLocalIds = []string{}
		hub.Hash = ""
		if hub.OwnerId == "" {
			hub.OwnerId = owner
		}
		if err, _ = this.ValidateHubDevices(hub); err != nil {
			log.Println("ERROR: unable to fix invalid hub --> ignore: ", hub, err)
			return nil
		}
		return this.producer.PublishHub(hub, owner)
	}
	hubIndex := map[string]models.Hub{}
	for _, id := range hub.DeviceIds {
		ctx, _ := getTimeoutContext()
		hubs, err := this.db.GetHubsByDeviceId(ctx, id)
		if err != nil {
			return err
		}
		for _, hub2 := range hubs {
			if hub2.Id != hub.Id {
				hubIndex[hub2.Id] = hub2.Hub
			}
		}
	}
	for _, lid := range hub.DeviceLocalIds {
		for _, hub2 := range hubIndex {
			hub2.DeviceLocalIds = filter(hub2.DeviceLocalIds, lid)
			hubIndex[hub2.Id] = hub2
		}
	}
	for _, id := range hub.DeviceIds {
		for _, hub2 := range hubIndex {
			hub2.DeviceIds = filter(hub2.DeviceIds, id)
			hubIndex[hub2.Id] = hub2
		}
	}
	for _, hub2 := range hubIndex {
		err := this.producer.PublishHub(hub2, owner)
		if err != nil {
			return err
		}
	}

	ctx, _ = getTimeoutContext()
	return this.db.SetHub(ctx, model.HubWithConnectionState{
		Hub:             hub,
		ConnectionState: old.ConnectionState,
	})
}

func (this *Controller) SetHubConnectionState(id string, connected bool) error {
	state := models.ConnectionStateOffline
	if connected {
		state = models.ConnectionStateOnline
	}
	ctx, _ := getTimeoutContext()
	return this.db.SetHubConnectionState(ctx, id, state)
}

func (this *Controller) DeleteHub(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveHub(ctx, id)
}
