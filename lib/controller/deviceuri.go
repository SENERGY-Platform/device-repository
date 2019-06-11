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
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
)

/*
	update device instance by uri
	id, uri, gateway, user-tags and image in body will be ignored
*/
func (this *Controller) PublishDeviceUriUpdate(jwt jwt_http_router.Jwt, uri string, device model.DeviceInstance) (result model.DeviceInstance, err error, errCode int) {
	target, err, errCode := this.ReadDeviceByUri(uri, "w", jwt)
	if err != nil {
		return result, err, errCode
	}
	device.Id = target.Id
	device.Url = target.Url
	device.UserTags = target.UserTags
	device.ImgUrl = target.ImgUrl
	if err, errCode = this.validateDeviceUpdate(device, target.Id); err != nil {
		return result, err, errCode
	}
	err = this.source.PublishDevice(device, jwt.UserId)
	if err != nil {
		errCode = http.StatusInternalServerError
	}
	result = device
	return
}

func (this *Controller) PublishDeviceUriDelete(jwt jwt_http_router.Jwt, uri string) (err error, errCode int) {
	target, err, errCode := this.ReadDeviceByUri(uri, "a", jwt)
	if err != nil {
		return err, errCode
	}
	err = this.source.PublishDeviceDelete(target.Id)
	if err != nil {
		errCode = http.StatusInternalServerError
	}
	return
}
