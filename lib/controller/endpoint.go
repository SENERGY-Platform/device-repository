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
	"github.com/cbroglie/mustache"
)

func (this *Controller) removeEndpointsOfDevice(ctx context.Context, device model.DeviceInstance) error {
	endpoints, err := this.db.ListEndpointsOfDevice(ctx, device.Id)
	if err != nil {
		return err
	}
	for _, endpoint := range endpoints {
		err = this.db.RemoveEndpoint(ctx, endpoint.Id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Controller) updateEndpointsOfDevice(ctx context.Context, oldDevice, newDevice model.DeviceInstance) error {
	if oldDevice.Url == newDevice.Url {
		return nil
	}
	deviceType, exists, err := this.db.GetDeviceType(ctx, oldDevice.DeviceType)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("unable to find device-type of device, to update device endpoints")
	}
	return this.updateEndpointsOfDeviceAndDeviceType(ctx, newDevice, deviceType)
}

func (this *Controller) updateEndpointsOfDeviceAndDeviceType(ctx context.Context, device model.DeviceInstance, deviceType model.DeviceType) error {
	err := this.removeEndpointsOfDevice(ctx, device)
	if err != nil {
		return err
	}

	for _, service := range deviceType.Services {
		endpoint := model.Endpoint{
			ProtocolHandler: service.Protocol.ProtocolHandlerUrl,
			Service:         service.Id,
			Device:          device.Id,
			Endpoint:        createEndpointString(service.EndpointFormat, device.Url, service.Url, device.Config),
		}
		if endpoint.Endpoint != "" {
			endpoint.Id = this.db.CreateId()
			err = this.db.SetEndpoint(ctx, endpoint)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createEndpointString(endpointFormat string, deviceUrl string, serviceUrl string, deviceConfig []model.ConfigField) (result string) {
	conf := map[string]string{"device_uri": deviceUrl, "service_uri": serviceUrl}
	for _, field := range deviceConfig {
		conf[field.Name] = field.Value
	}
	result, _ = mustache.Render(endpointFormat, conf)
	return
}

func (this *Controller) updateEndpointsOfDeviceType(ctx context.Context, oldDeviceType, newDeviceType model.DeviceType) error {
	devices, err := this.db.ListDevicesOfDeviceType(ctx, oldDeviceType.Id)
	if err != nil {
		return err
	}
	for _, device := range devices {
		err = this.updateEndpointsOfDeviceAndDeviceType(ctx, device, newDeviceType)
		if err != nil {
			return err
		}
	}
	return nil
}
