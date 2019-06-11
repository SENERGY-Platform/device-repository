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
	"errors"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
	"time"
)

var HubNotFoundError = errors.New("hub not found")

func (this *Controller) PublishHubCreate(jwt jwt_http_router.Jwt, hub model.Hub) (result model.Hub, err error, errCode int) {
	hub.Id = generateId()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	valid, gw, err := this.hubToFlatGateway(ctx, hub)
	if err != nil {
		errCode = http.StatusInternalServerError
		return result, err, errCode
	}
	if !valid {
		return hub, errors.New("invalid"), http.StatusBadRequest
	}
	err = this.source.PublishHub(gw, jwt.UserId)
	if err != nil {
		errCode = http.StatusInternalServerError
	}
	result = hub
	return
}

func (this *Controller) PublishHubUpdate(jwt jwt_http_router.Jwt, id string, hub model.Hub) (result model.Hub, err error, errCode int) {
	if id != hub.Id {
		return hub, errors.New("hub.id different from update id"), http.StatusBadRequest
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	valid, gw, err := this.hubToFlatGateway(ctx, hub)
	if err != nil {
		errCode = http.StatusInternalServerError
		return result, err, errCode
	}
	if !valid {
		return hub, errors.New("invalid"), http.StatusBadRequest
	}
	allowed, err := this.security.CheckBool(jwt, this.config.HubTopic, hub.Id, model.WRITE)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !allowed {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	err = this.source.PublishHub(gw, jwt.UserId)
	if err != nil {
		errCode = http.StatusInternalServerError
	}
	result = hub
	return
}

func (this *Controller) PublishHubDelete(jwt jwt_http_router.Jwt, id string) (err error, errCode int) {
	allowed, err := this.security.CheckBool(jwt, this.config.HubTopic, id, model.ADMINISTRATE)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !allowed {
		return errors.New("access denied"), http.StatusForbidden
	}
	err = this.source.PublishHubDelete(id)
	if err != nil {
		errCode = http.StatusInternalServerError
	}
	return
}

func (this *Controller) ReadHub(jwt jwt_http_router.Jwt, id string) (result model.Hub, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, exists, err := this.db.GetHub(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, HubNotFoundError, http.StatusNotFound
	}
	access, err := this.security.CheckBool(jwt, this.config.HubTopic, id, model.READ)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !access {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ReadHubDevices(jwt jwt_http_router.Jwt, id string, as string) (result []string, err error, errCode int) {
	var hub model.Hub
	hub, err, errCode = this.ReadHub(jwt, id)
	if err != nil {
		return
	}
	if as == "uri" || as == "url" || as == "" {
		return hub.Devices, nil, http.StatusOK
	} else if as == "id" {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		valid, flat, err := this.hubToFlatGateway(ctx, hub)
		if err != nil {
			return result, err, http.StatusBadRequest
		}
		if !valid {
			return result, errors.New("inconsistent data"), http.StatusInternalServerError
		}
		return flat.Devices, err, errCode
	} else {
		return result, errors.New("unknown value for 'as' query parameter"), http.StatusBadRequest
	}
}

func (this *Controller) SetHub(gw model.GatewayFlat, owner string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	transaction, finish, err := this.db.Transaction(ctx)
	if err != nil {
		return err
	}
	ok, hub, err := this.flatGatewayToHub(transaction, gw)
	if err != nil {
		_ = finish(false)
		return err
	}
	if !ok {
		log.Println("ERROR: invalid gateway command; ignore", gw)
		_ = finish(true)
		return nil
	}
	err = this.updateDeviceHubs(transaction, gw)
	if err != nil {
		_ = finish(false)
		return err
	}
	err = this.db.SetHub(transaction, hub)
	if err != nil {
		_ = finish(false)
		return err
	}
	return finish(true)
}

func (this *Controller) DeleteHub(id string, owner string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	transaction, finish, err := this.db.Transaction(ctx)
	if err != nil {
		return err
	}
	hub, exists, err := this.db.GetHub(transaction, id)
	if err != nil {
		_ = finish(false)
		return err
	}
	if !exists {
		return finish(true)
	}
	for _, device := range hub.Devices {
		err = this.removeHubFromDevice(transaction, device)
		if err != nil {
			_ = finish(false)
			return err
		}
	}
	err = this.db.RemoveHub(transaction, id)
	if err != nil {
		_ = finish(false)
		return err
	}
	return finish(true)
}

//updates devices using this hub to ether remove or add hub.id to device.gateway
func (this *Controller) updateDeviceHubs(ctx context.Context, hub model.GatewayFlat) error {
	devices, err := this.db.ListDevicesWithHub(ctx, hub.Id)
	if err != nil {
		return err
	}

	inHubList := map[string]bool{}
	for _, id := range hub.Devices {
		inHubList[id] = true
	}

	inDeviceList := map[string]bool{}
	for _, device := range devices {
		inDeviceList[device.Id] = true
	}

	for _, device := range devices {
		if !inHubList[device.Id] {
			device.Gateway = ""
			err = this.db.SetDevice(ctx, device) //remove hub ref from device
			if err != nil {
				return err
			}
		}
	}

	for _, device := range hub.Devices {
		if !inDeviceList[device] {
			err = this.addHubToDevice(ctx, device, hub.Id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *Controller) removeHubFromDevice(ctx context.Context, deviceId string) error {
	device, exists, err := this.db.GetDevice(ctx, deviceId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("device not found")
	}
	device.Gateway = ""
	return this.db.SetDevice(ctx, device)
}

func (this *Controller) addHubToDevice(ctx context.Context, deviceId string, hubId string) error {
	device, exists, err := this.db.GetDevice(ctx, deviceId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("device not found")
	}
	device.Gateway = hubId
	return this.db.SetDevice(ctx, device)
}

func (this *Controller) resetHubOfDevice(ctx context.Context, device model.DeviceInstance) error {
	if device.Gateway != "" {
		hub, exists, err := this.db.GetHub(ctx, device.Gateway)
		if err != nil {
			return err
		}
		if !exists {
			return HubNotFoundError
		}
		if hub.Name != "" {
			err = this.source.PublishHub(model.GatewayFlat{Id: hub.Id, Name: hub.Name, Hash: "", Devices: []string{}}, "")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *Controller) flatGatewayToHub(ctx context.Context, flat model.GatewayFlat) (valid bool, result model.Hub, err error) {
	result.Id = flat.Id
	result.Name = flat.Name
	result.Hash = flat.Hash
	for _, deviceId := range flat.Devices {
		device, exists, err := this.db.GetDevice(ctx, deviceId)
		if err != nil || !exists {
			return exists, result, err
		}
		result.Devices = append(result.Devices, device.Url)
	}
	return true, result, nil
}

func (this *Controller) hubToFlatGateway(ctx context.Context, hub model.Hub) (valid bool, result model.GatewayFlat, err error) {
	result.Id = hub.Id
	result.Name = hub.Name
	result.Hash = hub.Hash
	for _, deviceUri := range hub.Devices {
		device, exists, err := this.db.GetDeviceByUri(ctx, deviceUri)
		if err != nil || !exists {
			return exists, result, err
		}
		result.Devices = append(result.Devices, device.Id)
	}
	return true, result, nil
}
