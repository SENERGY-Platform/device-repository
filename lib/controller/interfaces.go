/*
 * Copyright 2025 InfAI (CC SES)
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

import "github.com/SENERGY-Platform/models/go/models"

type Publisher interface {
	PublishDevice(device models.Device) (err error)
	PublishDeviceDelete(device models.Device) error
	PublishDeviceDeletes(devices []models.Device) error

	PublishDeviceType(device models.DeviceType) (err error)
	PublishDeviceTypeDelete(id string) error
	PublishDeviceTypeDeletes(ids []string) error

	PublishDeviceGroup(dg models.DeviceGroup) (err error)
	PublishDeviceGroupDelete(id string) error
	PublishDeviceGroupDeletes(ids []string) error

	PublishProtocol(device models.Protocol) (err error)
	PublishProtocolDelete(id string) error
	PublishProtocolDeletes(ids []string) error

	PublishHub(hub models.Hub) (err error)
	PublishHubDelete(hub models.Hub) error
	PublishHubDeletes(hubs []models.Hub) error

	PublishConcept(concept models.Concept) (err error)
	PublishConceptDelete(id string) error
	PublishConceptDeletes(ids []string) error

	PublishCharacteristic(characteristic models.Characteristic) (err error)
	PublishCharacteristicDelete(id string) error
	PublishCharacteristicDeletes(ids []string) error

	PublishAspect(device models.Aspect) (err error)
	PublishAspectDelete(id string) error
	PublishAspectDeletes(ids []string) error
	PublishAspectClass(aspectClass models.AspectClass) (err error)
	PublishAspectClassDelete(id string) error
	PublishAspectClassDeletes(ids []string) error

	PublishFunction(device models.Function) (err error)
	PublishFunctionDelete(id string) error
	PublishFunctionDeletes(ids []string) error

	PublishDeviceClass(device models.DeviceClass) (err error)
	PublishDeviceClassDelete(id string) error
	PublishDeviceClassDeletes(ids []string) error

	PublishLocation(device models.Location) (err error)
	PublishLocationDelete(id string) error
	PublishLocationDeletes(ids []string) error

	PublishGraph(graph models.Graph) (err error)
	PublishGraphDelete(id string) error
	PublishGraphDeletes(ids []string) error
	PublishThing(thing models.Thing) (err error)
	PublishThingDelete(id string) error
	PublishThingDeletes(ids []string) error
}
