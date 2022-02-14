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

func (this *Database) ListAspectsWithMeasuringFunction(ctx context.Context) ([]model.Aspect, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListAllDeviceClassesUsedWithControllingFunctions(ctx context.Context) ([]model.DeviceClass, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListAllMeasuringFunctionsByAspect(ctx context.Context, aspect string) ([]model.Function, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListAllFunctionsByDeviceClass(ctx context.Context, class string) ([]model.Function, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListAllControllingFunctionsByDeviceClass(ctx context.Context, class string) ([]model.Function, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string, filter []model.FilterCriteria) (result []model.DeviceType, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) GetAspect(ctx context.Context, id string) (result model.Aspect, exists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) SetAspect(ctx context.Context, aspect model.Aspect) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) RemoveAspect(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListAllAspects(ctx context.Context) ([]model.Aspect, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) GetFunction(ctx context.Context, id string) (result model.Function, exists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) SetCharacteristic(ctx context.Context, characteristic model.Characteristic) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) RemoveCharacteristic(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) GetCharacteristic(ctx context.Context, id string) (result model.Characteristic, exists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListAllCharacteristics(ctx context.Context) ([]model.Characteristic, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) SetConcept(ctx context.Context, concept model.Concept) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) RemoveConcept(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) GetConceptWithCharacteristics(ctx context.Context, id string) (result model.ConceptWithCharacteristics, exists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) GetConceptWithoutCharacteristics(ctx context.Context, id string) (result model.Concept, exists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) SetDeviceClass(ctx context.Context, class model.DeviceClass) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) RemoveDeviceClass(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListAllDeviceClasses(ctx context.Context) ([]model.DeviceClass, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) GetDeviceClass(ctx context.Context, id string) (result model.DeviceClass, exists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) SetFunction(ctx context.Context, function model.Function) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) RemoveFunction(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListAllFunctionsByType(ctx context.Context, rdfType string) ([]model.Function, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) ListFunctions(ctx context.Context, limit int, offset int, search string, direction string) (result []model.Function, count int, err error) {
	//TODO implement me
	panic("implement me")
}

func (this *Database) SetLocation(ctx context.Context, location model.Location) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) RemoveLocation(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Database) GetLocation(ctx context.Context, id string) (result model.Location, exists bool, err error) {
	//TODO implement me
	panic("implement me")
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
