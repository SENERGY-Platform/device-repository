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

import "github.com/SENERGY-Platform/iot-device-repository/lib/model"

func (this *Controller) removeEndpointsOfDevice(device model.DeviceInstance) error {
	panic("todo") //TODO
}

func (this *Controller) updateEndpointsOfDevice(oldDevice, newDevice model.DeviceInstance) error {
	panic("todo") //TODO
	if oldDevice.Url != newDevice.Url {
		/*
			    deviceType, err := this.GetDeviceTypeById(device.DeviceType, 3)
				if err != nil {
					return err
				}

				//delete old
				endpoints, err := this.getEndpointsByDevice(device.Id)
				if err != nil {
					return err
				}
				for _, endpoint := range endpoints {
					this.ordf.Delete(endpoint)
					if err != nil {
						return err
					}
				}

				//create new
				for _, service := range deviceType.Services {
					endpoint := model.Endpoint{
						ProtocolHandler: service.Protocol.ProtocolHandlerUrl,
						Service:         service.Id,
						Device:          device.Id,
						Endpoint:        createEndpointString(service.EndpointFormat, device.Url, service.Url, device.Config),
					}
					if endpoint.Endpoint != "" {
						tempErr := this.ordf.SetIdDeep(&endpoint)
						if tempErr != nil {
							err = tempErr
						} else {
							_, tempErr = this.ordf.Insert(endpoint)
							if tempErr != nil {
								err = tempErr
							}
						}
					}
				}
				return err
		*/
	}
	return nil
}

func (this *Controller) updateEndpointsOfDeviceType(oldDeviceType, newDeviceType model.DeviceType) error {
	panic("todo") //TODO

	return nil
}
