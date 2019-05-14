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

func (this *Controller) updateHubOfDevice(oldDevice model.DeviceInstance, newDevice model.DeviceInstance) error {
	panic("implement me") //TODO
	/*

		if old.Gateway != "" && (old.Url != deviceInstance.Url || tagRemovedOrChanged(old.Tags, deviceInstance.Tags)) {
			//reset gateway hash
			gw, err := this.GetGateway(old.Gateway)
			if err != nil {
				return err
			}
			gw.Hash = ""
			err = eventsourcing.PublishGateway(gw, "")
			if err != nil {
				return err
			}
		}
		deviceInstance.Gateway = old.Gateway
		_, err = this.ordf.Update(old, deviceInstance)
		if err != nil {
			return
		}
	*/
}

func (this *Controller) resetHubOfDevice(device model.DeviceInstance) error {
	panic("implement me") //TODO
}
