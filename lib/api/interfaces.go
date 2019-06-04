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
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"github.com/SmartEnergyPlatform/jwt-http-router"
)

type Controller interface {
	ReadDevice(id string, jwt jwt_http_router.Jwt) (device model.DeviceInstance, err error, errCode int)
	ReadDeviceByUri(uri string, permission string, jwt jwt_http_router.Jwt) (device model.DeviceInstance, err error, errCode int)
	ListDevices(jwt jwt_http_router.Jwt, options listoptions.ListOptions) (result []model.DeviceInstance, err error, errCode int)
	ListEndpoints(jwt jwt_http_router.Jwt, options listoptions.ListOptions) (result []model.Endpoint, err error, errCode int)
	ReadHub(jwt jwt_http_router.Jwt, id string) (result model.Hub, err error, errCode int)
	ReadHubDevices(jwt jwt_http_router.Jwt, id string, as string) (result []string, err error, errCode int)
	ReadDeviceType(id string, jwt jwt_http_router.Jwt) (result model.DeviceType, err error, errCode int)
	ListDeviceTypes(jwt jwt_http_router.Jwt, options listoptions.ListOptions) (result []model.DeviceType, err error, errCode int)
	ReadService(id string, jwt jwt_http_router.Jwt) (result model.Service, err error, errCode int)
	ReadValueType(id string, jwt jwt_http_router.Jwt) (result model.ValueType, err error, errCode int)
	ListValueTypes(jwt jwt_http_router.Jwt, options listoptions.ListOptions) (result []model.ValueType, err error, errCode int)
}
