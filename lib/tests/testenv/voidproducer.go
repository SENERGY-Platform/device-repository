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

func (v VoidProducerMock) PublishDeviceGroupDelete(id string, owner string) error {
	return nil
}

func (v VoidProducerMock) PublishDeviceGroup(element models.DeviceGroup, owner string) error {
	return nil
}

func (v VoidProducerMock) PublishDevice(element models.Device, userId string) error {
	panic("implement me")
}

func (v VoidProducerMock) PublishDeviceRights(deviceId string, userId string, rights model.ResourceRights) (err error) {
	panic("implement me")
}

func (v VoidProducerMock) SendDone(msg donewait.DoneMsg) error {
	return nil
}

func (v VoidProducerMock) PublishAspectDelete(id string, owner string) error {
	panic("implement me")
}

func (v VoidProducerMock) PublishAspectUpdate(aspect models.Aspect, owner string) error {
	panic("implement me")
}

func (v VoidProducerMock) PublishDeviceDelete(id string, owner string) error {
	panic("implement me")
}

func (v VoidProducerMock) PublishHub(hub models.Hub, userId string) (err error) {
	panic("implement me")
}
