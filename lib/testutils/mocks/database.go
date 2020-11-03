/*
 * Copyright 2020 InfAI (CC SES)
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

package mocks

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/model"
)

type Database struct {
	devices      map[string]model.Device
	deviceTypes  map[string]model.DeviceType
	deviceGroups map[string]model.DeviceGroup
}

func NewDatabase() *Database {
	return &Database{
		devices:      map[string]model.Device{},
		deviceTypes:  map[string]model.DeviceType{},
		deviceGroups: map[string]model.DeviceGroup{},
	}
}

func (this *Database) Disconnect() {
	return
}

func (this *Database) GetDevice(ctx context.Context, id string) (device model.Device, exists bool, err error) {
	device, exists = this.devices[id]
	return
}

func (this *Database) SetDevice(ctx context.Context, device model.Device) error {
	this.devices[device.Id] = device
	return nil
}

func (this *Database) RemoveDevice(ctx context.Context, id string) error {
	delete(this.devices, id)
	return nil
}

func (this *Database) GetDeviceByLocalId(ctx context.Context, localId string) (device model.Device, exists bool, err error) {
	panic("implement me")
}

func (this *Database) GetHub(ctx context.Context, id string) (hub model.Hub, exists bool, err error) {
	panic("implement me")
}

func (this *Database) SetHub(ctx context.Context, hub model.Hub) error {
	panic("implement me")
}

func (this *Database) RemoveHub(ctx context.Context, id string) error {
	panic("implement me")
}

func (this *Database) GetHubsByDeviceLocalId(ctx context.Context, localId string) (hubs []model.Hub, err error) {
	panic("implement me")
}

func (this *Database) GetDeviceType(ctx context.Context, id string) (deviceType model.DeviceType, exists bool, err error) {
	deviceType, exists = this.deviceTypes[id]
	return
}

func (this *Database) SetDeviceType(ctx context.Context, deviceType model.DeviceType) error {
	this.deviceTypes[deviceType.Id] = deviceType
	return nil
}

func (this *Database) RemoveDeviceType(ctx context.Context, id string) error {
	delete(this.deviceTypes, id)
	return nil
}

func (this *Database) ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string) (result []model.DeviceType, err error) {
	panic("implement me")
}

func (this *Database) GetDeviceTypesByServiceId(ctx context.Context, serviceId string) ([]model.DeviceType, error) {
	panic("implement me")
}

func (this *Database) GetDeviceGroup(ctx context.Context, id string) (deviceGroup model.DeviceGroup, exists bool, err error) {
	deviceGroup, exists = this.deviceGroups[id]
	return
}

func (this *Database) SetDeviceGroup(ctx context.Context, deviceGroup model.DeviceGroup) error {
	this.deviceGroups[deviceGroup.Id] = deviceGroup
	return nil
}

func (this *Database) RemoveDeviceGroup(ctx context.Context, id string) error {
	delete(this.deviceGroups, id)
	return nil
}

func (this *Database) ListDeviceGroups(ctx context.Context, limit int64, offset int64, sort string) (result []model.DeviceGroup, err error) {
	panic("implement me")
}

func (this *Database) GetProtocol(ctx context.Context, id string) (result model.Protocol, exists bool, err error) {
	panic("implement me")
}

func (this *Database) ListProtocols(ctx context.Context, limit int64, offset int64, sort string) ([]model.Protocol, error) {
	panic("implement me")
}

func (this *Database) SetProtocol(ctx context.Context, protocol model.Protocol) error {
	panic("implement me")
}

func (this *Database) RemoveProtocol(ctx context.Context, id string) error {
	panic("implement me")
}
