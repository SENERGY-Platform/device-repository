/*
 * Copyright 2021 InfAI (CC SES)
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

package publisher

import (
	"errors"
	"github.com/SENERGY-Platform/models/go/models"
)

type Void struct{}

var VoidPublisherError = errors.New("try to use void publisher")

func (this Void) PublishDevice(device models.Device) (err error) {
	return VoidPublisherError
}

func (this Void) PublishDeviceDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishDeviceType(device models.DeviceType) (err error) {
	return VoidPublisherError
}

func (this Void) PublishDeviceTypeDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishDeviceGroup(device models.DeviceGroup) (err error) {
	return VoidPublisherError
}

func (this Void) PublishDeviceGroupDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishProtocol(device models.Protocol) (err error) {
	return VoidPublisherError
}

func (this Void) PublishProtocolDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishHub(hub models.Hub) (err error) {
	return VoidPublisherError
}

func (this Void) PublishHubDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishConcept(concept models.Concept) (err error) {
	return VoidPublisherError
}

func (this Void) PublishConceptDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishCharacteristic(characteristic models.Characteristic) (err error) {
	return VoidPublisherError
}

func (this Void) PublishCharacteristicDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishAspect(device models.Aspect) (err error) {
	return VoidPublisherError
}

func (this Void) PublishAspectDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishFunction(device models.Function) (err error) {
	return VoidPublisherError
}

func (this Void) PublishFunctionDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishDeviceClass(device models.DeviceClass) (err error) {
	return VoidPublisherError
}

func (this Void) PublishDeviceClassDelete(id string) error {
	return VoidPublisherError
}

func (this Void) PublishLocation(device models.Location) (err error) {
	return VoidPublisherError
}

func (this Void) PublishLocationDelete(id string) error {
	return VoidPublisherError
}
