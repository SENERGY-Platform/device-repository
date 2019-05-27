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
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"github.com/SmartEnergyPlatform/jwt-http-router"
)

type Publisher interface {
	PublishDevice(device model.DeviceInstance, owner string) error //user has to check for uri collision
	PublishHub(hub model.Hub, owner string) error
	PublishValueType(valueType model.ValueType, owner string) error
	PublishDeviceType(dt model.DeviceType, owner string) error
}

type Security interface {
	CheckBool(jwt jwt_http_router.Jwt, kind string, id string, action model.AuthAction) (allowed bool, err error)
	List(jwt jwt_http_router.Jwt, kind string, action model.AuthAction, limit string, offset string) (ids []string, err error)
	CheckList(jwt jwt_http_router.Jwt, kind string, ids []string, action model.AuthAction) (result map[string]bool, err error)
}
