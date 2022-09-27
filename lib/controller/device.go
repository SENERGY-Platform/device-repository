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
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"net/http"
	"strings"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadDevice(id string, token string, action model.AuthAction) (result model.Device, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	pureId, modifier := idmodifier.SplitModifier(id)
	device, exists, err := this.db.GetDevice(ctx, pureId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err := this.security.CheckBool(token, this.config.DeviceTopic, pureId, action)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	device.Id = id
	if modifier != nil && len(modifier) > 0 {
		device, err, errCode = this.modifyDevice(device, modifier)
		if err != nil {
			return result, err, errCode
		}
	}
	return device, nil, http.StatusOK
}

func (this *Controller) ReadDeviceByLocalId(localId string, token string, action model.AuthAction) (result model.Device, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	device, exists, err := this.db.GetDeviceByLocalId(ctx, localId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err := this.security.CheckBool(token, this.config.DeviceTopic, device.Id, action)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	return device, nil, http.StatusOK
}

const DisplayNameAttributeName = "shared/nickname"

func ValidateDeviceName(device model.Device) (err error) {
	if device.Name == "" {
		hasDisplayNameAttribute := false
		for _, attr := range device.Attributes {
			if attr.Key == DisplayNameAttributeName {
				hasDisplayNameAttribute = true
				break
			}
		}
		if !hasDisplayNameAttribute {
			return errors.New("missing device name")
		}
	}
	return nil
}

func (this *Controller) ValidateDevice(device model.Device) (err error, code int) {
	if device.Id == "" {
		return errors.New("missing device id"), http.StatusBadRequest
	}
	err = ValidateDeviceName(device)
	if err != nil {
		return err, http.StatusBadRequest
	}
	if device.LocalId == "" {
		return errors.New("missing device local id"), http.StatusBadRequest
	}
	if strings.ContainsAny(device.LocalId, "+#/") {
		return errors.New("device local id may not contain any +#/"), http.StatusBadRequest
	}
	if device.DeviceTypeId == "" {
		return errors.New("missing device type id"), http.StatusBadRequest
	}

	//device-type exists
	ctx, _ := getTimeoutContext()
	_, ok, err := this.db.GetDeviceType(ctx, device.DeviceTypeId)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("unknown device type id"), http.StatusBadRequest
	}

	//local ids are globally unique
	ctx, _ = getTimeoutContext()
	d, ok, err := this.db.GetDeviceByLocalId(ctx, device.LocalId)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if ok && d.Id != device.Id {
		return errors.New("local id should be empty or globally unique"), http.StatusBadRequest
	}

	return nil, http.StatusOK
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDevice(device model.Device, owner string) (err error) {
	//prevent collision of local ids
	//this if branch should be rarely needed if 2 devices are created at the same time with the same local_id (when the second device is validated before the creation of the first is finished)
	ctx, _ := getTimeoutContext()
	d, collision, err := this.db.GetDeviceByLocalId(ctx, device.LocalId)
	if err != nil {
		return err
	}
	if collision && d.Id != device.Id {

		//handle invalid device
		device.LocalId = ""
		ctx, _ = getTimeoutContext()
		err = this.db.SetDevice(ctx, device)
		if err != nil {
			return err
		}
		return this.PublishDeviceDelete(device.Id, owner)

	} else {

		//update hub about changed device.local_id
		ctx, _ = getTimeoutContext()
		old, exists, err := this.db.GetDevice(ctx, device.Id)
		if err != nil {
			return err
		}
		if exists && old.LocalId != device.LocalId {
			err = this.resetHubsForDeviceUpdate(old)
		}

		//save device
		ctx, _ = getTimeoutContext()
		return this.db.SetDevice(ctx, device)
	}

}

func (this *Controller) DeleteDevice(id string) error {
	ctx, _ := getTimeoutContext()
	old, exists, err := this.db.GetDevice(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		ctx, _ := getTimeoutContext()
		err := this.db.RemoveDevice(ctx, id)
		if err != nil {
			return err
		}
		return this.resetHubsForDeviceUpdate(old)
	}
	return nil
}

func (this *Controller) PublishDeviceDelete(id string, owner string) error {
	return this.producer.PublishDeviceDelete(id, owner)
}

func (this *Controller) resetHubsForDeviceUpdate(old model.Device) error {
	ctx, _ := getTimeoutContext()
	hubs, err := this.db.GetHubsByDeviceLocalId(ctx, old.LocalId)
	if err != nil {
		return err
	}
	for _, hub := range hubs {
		hub.DeviceLocalIds = filter(hub.DeviceLocalIds, old.LocalId)
		hub.Hash = ""
		err = this.producer.PublishHub(hub)
		if err != nil {
			return err
		}
	}
	return nil
}

func filter(in []string, not string) (out []string) {
	for _, str := range in {
		if str != not {
			out = append(out, str)
		}
	}
	return
}
