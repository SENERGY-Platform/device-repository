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
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
)

var HubNotFoundError = errors.New("hub not found")

func (this *Controller) SetHub(hub model.Hub, owner string) error {
	err := this.updateDeviceHubs(hub)
	if err != nil {
		return err
	}
	return this.db.SetHub(hub)
}

func (this *Controller) DeleteHub(id string, owner string) error {
	hub, exists, err := this.db.GetHub(id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	for _, device := range hub.Devices {
		err = this.removeHubFromDevice(device)
		if err != nil {
			return err
		}
	}
	return this.db.RemoveHub(id)
}

//updates devices using this hub to ether remove or add hub.id to device.gateway
func (this *Controller) updateDeviceHubs(hub model.Hub) error {
	devices, err := this.db.ListDevicesWithHub(hub.Id)
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
			err = this.db.SetDevice(device) //remove hub ref from device
			if err != nil {
				return err
			}
		}
	}

	for _, device := range hub.Devices {
		if !inDeviceList[device] {
			err = this.addHubToDevice(device, hub.Id)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *Controller) removeHubFromDevice(deviceId string) error {
	device, exists, err := this.db.GetDevice(deviceId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("device not found")
	}
	device.Gateway = ""
	return this.db.SetDevice(device)
}

func (this *Controller) addHubToDevice(deviceId string, hubId string) error {
	device, exists, err := this.db.GetDevice(deviceId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("device not found")
	}
	device.Gateway = hubId
	return this.db.SetDevice(device)
}

func (this *Controller) resetHubOfDevice(device model.DeviceInstance) error {
	if device.Gateway != "" {
		hub, exists, err := this.db.GetHub(device.Gateway)
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
