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

package listener

import "github.com/SENERGY-Platform/device-repository/lib/model"

type Controller interface {
	SetDevice(device model.Device, owner string) error
	DeleteDevice(id string) error
	SetHub(hub model.Hub, owner string) error
	DeleteHub(id string) error
	SetDeviceType(deviceType model.DeviceType, owner string) error
	DeleteDeviceType(id string) error
	SetDeviceGroup(deviceGroup model.DeviceGroup, owner string) error
	DeleteDeviceGroup(id string) error
	SetProtocol(protocol model.Protocol, owner string) error
	DeleteProtocol(id string) error
}
