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
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"github.com/SmartEnergyPlatform/jwt-http-router"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadDevice(id string, jwt jwt_http_router.Jwt) (device model.DeviceInstance, err error, errCode int) {
	var exists bool
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	device, exists, err = this.db.GetDevice(ctx, id)
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

func (this *Controller) ListDevices(jwt jwt_http_router.Jwt, options listoptions.ListOptions) (result []model.DeviceInstance, err error, errCode int) {
	limit, withLimit := options.GetLimit()
	offset, withOffset := options.GetOffset()
	if !withLimit {
		limit = 100
	}
	if !withOffset {
		offset = 0
	}

	right, withPermission := options.Get("permission")

	action := model.READ
	if withPermission {
		switch right {
		case "w":
			action = model.WRITE
		case "x":
			action = model.EXECUTE
		case "a":
			action = model.ADMINISTRATE
		}
	}

	sort, withSort := options.Get("sort")
	var sortby string
	var sortdirection string
	if withSort {
		sortstr, ok := sort.(string)
		if !ok {
			return result, errors.New("unable to interpret sort as string"), http.StatusInternalServerError
		}
		parts := strings.Split(sortstr, ".")
		sortby = parts[0]
		sortdirection = parts[1]
	}

	err = options.EvalStrict()
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	var ids []string
	if withSort {
		ids, err = this.security.SortedList(jwt, this.config.DeviceInstanceTopic, action, strconv.FormatInt(limit, 10), strconv.FormatInt(offset, 10), sortby, sortdirection)
	} else {
		ids, err = this.security.List(jwt, this.config.DeviceInstanceTopic, action, strconv.FormatInt(limit, 10), strconv.FormatInt(offset, 10))
	}
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	for _, id := range ids {
		device, exists, err := this.db.GetDevice(ctx, id)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		if exists {
			result = append(result, device)
		} else {
			log.Println("WARNING: security returned id of a nonexistent device: ", id)
		}
	}
	return
}

func (this *Controller) ReadDeviceByUri(uri string, jwt jwt_http_router.Jwt) (device model.DeviceInstance, err error, errCode int) {
	var exists bool
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	device, exists, err = this.db.GetDeviceByUri(ctx, uri)
	if err != nil {
		return device, err, http.StatusInternalServerError
	}
	if !exists {
		return model.DeviceInstance{}, errors.New("not found"), http.StatusNotFound
	}
	allowed, err := this.security.CheckBool(jwt, this.config.DeviceInstanceTopic, device.Id, model.READ)
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
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	transaction, finish, err := this.db.Transaction(ctx)
	if err != nil {
		return err
	}
	ok, err := this.validateDevice(transaction, device)
	if err != nil {
		return err
	}
	if !ok {
		log.Println("ERROR: invalid device command; ignore", device)
		return
	}
	old, exists, err := this.db.GetDevice(transaction, device.Id)
	if err != nil {
		_ = finish(false)
		return err
	}
	device.Gateway = old.Gateway //device.gateway may only be changed by updating hub
	err = this.updateEndpointsOfDevice(transaction, old, device)
	if err != nil {
		_ = finish(false)
		return err
	}
	if exists && old.Gateway != "" && (old.Url != device.Url || tagRemovedOrChanged(old.Tags, device.Tags)) {
		err = this.resetHubOfDevice(transaction, old)
		if err == HubNotFoundError {
			log.Println("WARNING: inconsistency will be removed by over overwriting device", device)
			device.Gateway = "" //inconsistent state to consistent state
			err = nil
		}
		if err != nil {
			_ = finish(false)
			return
		}
	}
	err = this.db.SetDevice(transaction, device)
	if err != nil {
		_ = finish(false)
		return err
	}
	return finish(true)
}

func (this *Controller) DeleteDevice(id string) (err error) {
	ctx, finish, err := this.db.Transaction(context.Background())
	if err != nil {
		return err
	}
	old, exists, err := this.db.GetDevice(ctx, id)
	if err != nil {
		_ = finish(false)
		return err
	}
	if !exists {
		_ = finish(true)
		return
	}
	err = this.removeEndpointsOfDevice(ctx, old)
	if err != nil {
		_ = finish(false)
		return
	}
	err = this.resetHubOfDevice(ctx, old)
	if err == HubNotFoundError {
		err = nil //ignore inconsistency because it will be deleted
	}
	if err != nil {
		_ = finish(false)
		return
	}
	err = this.db.RemoveDevice(ctx, id)
	if err != nil {
		_ = finish(false)
		return
	}
	return finish(true)
}

func (this *Controller) updateDefaultDeviceImages(ctx context.Context, deviceTypeId string, oldImage string, newImage string) error {
	if oldImage == newImage {
		return nil
	}
	devices, err := this.db.ListDevicesOfDeviceType(ctx, deviceTypeId)
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

func (this *Controller) validateDevice(ctx context.Context, instance model.DeviceInstance) (ok bool, err error) {
	_, ok, err = this.db.GetDeviceType(ctx, instance.DeviceType)
	return
}
