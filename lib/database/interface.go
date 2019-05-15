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

package database

import "github.com/SENERGY-Platform/iot-device-repository/lib/model"

type Database interface {
	CreateId() string
	GetDevice(id string) (device model.DeviceInstance, exists bool, err error)
	SetDevice(device model.DeviceInstance) error
	RemoveDevice(id string) error
	GetDeviceType(id string) (deviceType model.DeviceType, exists bool, err error)
	SetDeviceType(deviceType model.DeviceType) error
	ListDevicesOfDeviceType(deviceTypeId string) ([]model.DeviceInstance, error)
	RemoveDeviceType(id string) error
	ListEndpointsOfDevice(deviceId string) ([]model.Endpoint, error)
	RemoveEndpoint(id string) error
	SetEndpoint(endpoint model.Endpoint) error
	GetHub(id string) (model.Hub, bool, error)
	SetHub(hub model.Hub) error
	RemoveHub(id string) error
	ListDevicesWithHub(id string) ([]model.DeviceInstance, error)
}
