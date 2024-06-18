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
	"log"
	"net/http"
	"runtime/debug"
	"slices"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadHub(id string, token string, action model.AuthAction) (result models.Hub, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	hub, exists, err := this.db.GetHub(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err := this.security.CheckBool(token, this.config.HubTopic, id, action)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	return hub, nil, http.StatusOK
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
	ok, err := this.security.CheckBool(token, this.config.HubTopic, id, action)
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
		admins, err := this.security.GetAdminUsers(token, this.config.DeviceTopic, hub.Id)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		//new device owner-id must be existing admin user (ignore for new devices or devices with unchanged owner)
		if hub.OwnerId != original.OwnerId && !slices.Contains(admins, hub.OwnerId) {
			return errors.New("new owner must have existing user admin rights"), http.StatusBadRequest
		}
		if hub.OwnerId != original.OwnerId && len(admins) == 0 {
			//o admins indicates the requesting user has not the needed admin rights to see other admins
			return errors.New("requesting user must have admin rights"), http.StatusBadRequest
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
		log.Println("ERROR: ", err)
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
		return this.producer.PublishHub(hub)
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
				hubIndex[hub2.Id] = hub2
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
			hub2.DeviceLocalIds = filter(hub2.DeviceIds, id)
			hubIndex[hub2.Id] = hub2
		}
	}
	for _, hub2 := range hubIndex {
		err := this.producer.PublishHub(hub2)
		if err != nil {
			return err
		}
	}

	ctx, _ = getTimeoutContext()
	return this.db.SetHub(ctx, hub)
}

func (this *Controller) DeleteHub(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveHub(ctx, id)
}
