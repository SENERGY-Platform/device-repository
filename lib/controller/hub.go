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
	"time"
)

var HubNotFoundError = errors.New("hub not found")

func (this *Controller) SetHub(hub model.Hub, owner string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	transaction, finish, err := this.db.Transaction(ctx)
	if err != nil {
		return err
	}
	err = this.updateDeviceHubs(transaction, hub)
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
func (this *Controller) updateDeviceHubs(ctx context.Context, hub model.Hub) error {
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
			hub.Devices = []string{}
			hub.Hash = ""
			err = this.source.PublishHub(hub, "")
			if err != nil {
				return err
			}
		}
	}
	return nil
}
