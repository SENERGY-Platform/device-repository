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

package listener

import "github.com/SENERGY-Platform/models/go/models"

type Controller interface {
	SetDevice(device models.Device, owner string) error
	DeleteDevice(id string) error
	SetHub(hub models.Hub, owner string) error
	DeleteHub(id string) error
	SetDeviceType(deviceType models.DeviceType, owner string) error
	DeleteDeviceType(id string) error
	SetDeviceGroup(deviceGroup models.DeviceGroup, owner string) error
	DeleteDeviceGroup(id string) error
	SetProtocol(protocol models.Protocol, owner string) error
	DeleteProtocol(id string) error

	SetAspect(aspect models.Aspect, owner string) error
	DeleteAspect(id string) error
	SetCharacteristic(characteristic models.Characteristic, owner string) error
	DeleteCharacteristic(id string) error
	SetConcept(concept models.Concept, owner string) error
	DeleteConcept(id string) error
	SetDeviceClass(class models.DeviceClass, owner string) error
	DeleteDeviceClass(id string) error
	SetFunction(function models.Function, owner string) error
	DeleteFunction(id string) error
	SetLocation(location models.Location, owner string) error
	DeleteLocation(id string) error
}
