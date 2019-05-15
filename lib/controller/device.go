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
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"github.com/SmartEnergyPlatform/jwt-http-router"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strings"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadDevice(id string, jwt jwt_http_router.Jwt) (device model.DeviceInstance, err error, errCode int) {
	var exists bool
	device, exists, err = this.db.GetDevice(device.Id)
	if err != nil {
		return device, err, http.StatusInternalServerError
	}
	if !exists {
		return model.DeviceInstance{}, errors.New("not found"), http.StatusNotFound
	}
	allowed, err := this.security.CheckBool(jwt, this.config.DeviceInstanceTopic, id, model.READ)
	if err != nil {
		return device, err, http.StatusInternalServerError
	}
	if !allowed {
		return model.DeviceInstance{}, errors.New("access denied"), http.StatusForbidden
	}
	return device, nil, http.StatusOK
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDevice(device model.DeviceInstance) (err error) {
	old, exists, err := this.db.GetDevice(device.Id)
	if err != nil {
		return
	}
	device.Gateway = old.Gateway //device.gateway may only be changed by updating hub
	err = this.updateEndpointsOfDevice(old, device)
	if err != nil {
		return
	}
	if exists && old.Gateway != "" && (old.Url != device.Url || tagRemovedOrChanged(old.Tags, device.Tags)) {
		err = this.resetHubOfDevice(old)
		if err == HubNotFoundError {
			log.Println("WARNING: inconsistency will be removed by over overwriting device", device)
			device.Gateway = "" //inconsistent state to consistent state
			err = nil
		}
		if err != nil {
			return
		}
	}
	err = this.db.SetDevice(device)
	return
}

func (this *Controller) DeleteDevice(id string) (err error) {
	old, exists, err := this.db.GetDevice(id)
	if err != nil || !exists {
		return
	}
	err = this.removeEndpointsOfDevice(old)
	if err != nil {
		return
	}
	err = this.resetHubOfDevice(old)
	if err == HubNotFoundError {
		err = nil //ignore inconsistency because it will be deleted
	}
	if err != nil {
		return
	}
	return this.db.RemoveDevice(id)
}

func (this *Controller) updateDefaultDeviceImages(deviceTypeId string, oldImage string, newImage string) error {
	if oldImage == newImage {
		return nil
	}
	devices, err := this.db.ListDevicesOfDeviceType(deviceTypeId)
	if err != nil {
		return err
	}
	for _, device := range devices {
		if device.ImgUrl == "" || device.ImgUrl == oldImage {
			device.ImgUrl = newImage
			err = this.source.PublishDevice(device, "")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func tagRemovedOrChanged(oldTags []string, newTags []string) bool {
	oldTagsIndex := indexTags(oldTags)
	newTagsIndex := indexTags(newTags)
	for key, oldVal := range oldTagsIndex {
		newVal, ok := newTagsIndex[key]
		if !ok || newVal != oldVal {
			return true
		}
	}
	return false
}

func indexTags(tags []string) (result map[string]string) {
	result = map[string]string{}
	for _, tag := range tags {
		parts := strings.SplitN(tag, ":", 2)
		if len(parts) != 2 {
			log.Println("ERROR: wrong tag syntax; ", tag)
			continue
		}
		result[parts[0]] = parts[1]
	}
	return result
}
