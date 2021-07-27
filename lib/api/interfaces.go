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

package api

import (
	"github.com/SENERGY-Platform/device-repository/lib/model"
)

type Controller interface {
	ReadDevice(id string, token string, action model.AuthAction) (result model.Device, err error, errCode int)
	ReadDeviceByLocalId(localId string, token string, action model.AuthAction) (result model.Device, err error, errCode int)
	ValidateDevice(device model.Device) (err error, code int)

	ReadHub(id string, token string, action model.AuthAction) (result model.Hub, err error, errCode int)
	ListHubDeviceIds(id string, token string, action model.AuthAction, asLocalId bool) (result []string, err error, errCode int)
	ValidateHub(hub model.Hub) (err error, code int)

	ReadDeviceType(id string, token string) (result model.DeviceType, err error, errCode int)
	ListDeviceTypes(token string, limit int64, offset int64, sort string) (result []model.DeviceType, err error, errCode int)
	ValidateDeviceType(deviceType model.DeviceType) (err error, code int)

	ReadDeviceGroup(id string, token string) (result model.DeviceGroup, err error, errCode int)
	ValidateDeviceGroup(deviceGroup model.DeviceGroup) (err error, code int)
	CheckAccessToDevicesOfGroup(token string, group model.DeviceGroup) (err error, code int)

	ReadProtocol(id string, token string) (result model.Protocol, err error, errCode int)
	ListProtocols(token string, limit int64, offset int64, sort string) (result []model.Protocol, err error, errCode int)
	ValidateProtocol(protocol model.Protocol) (err error, code int)

	GetService(id string) (result model.Service, err error, code int)
}
