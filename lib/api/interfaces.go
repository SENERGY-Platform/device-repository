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
	"github.com/SmartEnergyPlatform/jwt-http-router"
)

type Controller interface {
	ReadDeviceType(id string, jwt jwt_http_router.Jwt) (result model.DeviceType, err error, errCode int)
	ListDeviceTypes(jwt jwt_http_router.Jwt, limit int64, offset int64, sort string) (result []model.DeviceType, err error, errCode int)
}
