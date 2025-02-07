/*
 * Copyright 2024 InfAI (CC SES)
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

package testenv

import (
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/donewait"
)

type VoidProducerMock struct{}

func (v VoidProducerMock) PublishDevice(device models.Device) (err error) {
	return nil
}

func (v VoidProducerMock) PublishDeviceDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishDeviceType(device models.DeviceType) (err error) {
	return nil
}

func (v VoidProducerMock) PublishDeviceTypeDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishDeviceGroup(dg models.DeviceGroup) (err error) {
	return nil
}

func (v VoidProducerMock) PublishDeviceGroupDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishProtocol(device models.Protocol) (err error) {
	return nil
}

func (v VoidProducerMock) PublishProtocolDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishHub(hub models.Hub) (err error) {
	return nil
}

func (v VoidProducerMock) PublishHubDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishConcept(concept models.Concept) (err error) {
	return nil
}

func (v VoidProducerMock) PublishConceptDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishCharacteristic(characteristic models.Characteristic) (err error) {
	return nil
}

func (v VoidProducerMock) PublishCharacteristicDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishAspect(device models.Aspect) (err error) {
	return nil
}

func (v VoidProducerMock) PublishAspectDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishFunction(device models.Function) (err error) {
	return nil
}

func (v VoidProducerMock) PublishFunctionDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishDeviceClass(device models.DeviceClass) (err error) {
	return nil
}

func (v VoidProducerMock) PublishDeviceClassDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishLocation(device models.Location) (err error) {
	return nil
}

func (v VoidProducerMock) PublishLocationDelete(id string) error {
	return nil
}

func (v VoidProducerMock) PublishDeviceRights(deviceId string, userId string, rights model.ResourceRights) (err error) {
	panic("implement me")
}

func (v VoidProducerMock) SendDone(msg donewait.DoneMsg) error {
	return nil
}

func (v VoidProducerMock) PublishAspectUpdate(aspect models.Aspect, owner string) error {
	panic("implement me")
}
