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
)

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDeviceType(deviceType model.DeviceType, owner string) (err error) {
	err = this.publishMissingValueTypesOfDeviceType(deviceType, owner)
	if err != nil {
		return err
	}

	old, exists, err := this.db.GetDeviceType(deviceType.Id)
	if err != nil {
		return
	}

	err = this.updateEndpointsOfDeviceType(old, deviceType)
	if err != nil {
		return err
	}

	if exists && old.ImgUrl != deviceType.ImgUrl {
		err = this.updateDefaultDeviceImages(deviceType.Id, old.ImgUrl, deviceType.ImgUrl)
		if err != nil {
			return
		}
	}

	return this.db.SetDeviceType(deviceType)
}

func (this *Controller) DeleteDeviceType(id string) error {
	return this.db.RemoveDeviceType(id)
}
