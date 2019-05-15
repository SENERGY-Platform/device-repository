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

package mongo

import (
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
)

func (this *Mongo) GetDevice(id string) (device model.DeviceInstance, exists bool, err error) {
	panic("implement me") //TODO
}

func (this *Mongo) SetDevice(device model.DeviceInstance) error {
	panic("implement me") //TODO
}

func (this *Mongo) RemoveDevice(id string) error {
	panic("implement me") //TODO
}

func (this *Mongo) ListDevicesOfDeviceType(deviceTypeId string) ([]model.DeviceInstance, error) {
	panic("implement me") //TODO
}

func (this *Mongo) ListDevicesWithHub(id string) ([]model.DeviceInstance, error) {
	panic("implement me") //TODO
}
