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

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadDevice(id string, jwt jwt_http_router.Jwt) (device model.DeviceInstance, err error, errCode int) {
	panic("implement me")
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDevice(device model.DeviceInstance) (err error) {
	old, exists, err := this.db.ReadDevice(device.Id)
	if err != nil {
		return
	}
	if exists {
		err = this.updateEndpointsOfDevice(old, device)
		if err != nil {
			return
		}
		err = this.updateHubOfDevice(old, device)
		if err != nil {
			return
		}
	}
	err = this.db.SetDevice(device)
	return
}

func (this *Controller) DeleteDevice(id string, owner string) (err error) {
	old, exists, err := this.db.ReadDevice(id)
	if err != nil || !exists {
		return
	}
	err = this.removeEndpointsOfDevice(old)
	if err != nil {
		return
	}
	err = this.resetHubOfDevice(old)
	if err != nil {
		return
	}
	return this.db.RemoveDevice(id)
}
