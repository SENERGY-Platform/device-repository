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

package database

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/model"
)

type Database interface {
	Disconnect()

	GetDevice(ctx context.Context, id string) (device model.Device, exists bool, err error)
	SetDevice(ctx context.Context, device model.Device) error
	RemoveDevice(ctx context.Context, id string) error
	GetDeviceByLocalId(ctx context.Context, localId string) (device model.Device, exists bool, err error)

	GetHub(ctx context.Context, id string) (hub model.Hub, exists bool, err error)
	SetHub(ctx context.Context, hub model.Hub) error
	RemoveHub(ctx context.Context, id string) error
	GetHubsByDeviceLocalId(ctx context.Context, localId string) (hubs []model.Hub, err error)

	GetDeviceType(ctx context.Context, id string) (deviceType model.DeviceType, exists bool, err error)
	SetDeviceType(ctx context.Context, deviceType model.DeviceType) error
	RemoveDeviceType(ctx context.Context, id string) error
	ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string) (result []model.DeviceType, err error)
	GetDeviceTypesByServiceId(ctx context.Context, serviceId string) ([]model.DeviceType, error)

	GetDeviceTypeCriteriaForDeviceTypeIdsAndFilterCriteria(ctx context.Context, deviceTypeIds []interface{}, criteria model.FilterCriteria) (result []model.DeviceTypeCriteria, err error)
	GetDeviceTypeIdsByFilterCriteria(ctx context.Context, criteria []model.FilterCriteria, interactionsFilter []string) (result []interface{}, err error)
	GetConfigurableCandidates(ctx context.Context, serviceId string) (result []model.DeviceTypeCriteria, err error)

	GetDeviceGroup(ctx context.Context, id string) (deviceGroup model.DeviceGroup, exists bool, err error)
	SetDeviceGroup(ctx context.Context, deviceGroup model.DeviceGroup) error
	RemoveDeviceGroup(ctx context.Context, id string) error
	ListDeviceGroups(ctx context.Context, limit int64, offset int64, sort string) (result []model.DeviceGroup, err error)

	GetProtocol(ctx context.Context, id string) (result model.Protocol, exists bool, err error)
	ListProtocols(ctx context.Context, limit int64, offset int64, sort string) ([]model.Protocol, error)
	SetProtocol(ctx context.Context, protocol model.Protocol) error
	RemoveProtocol(ctx context.Context, id string) error

	GetAspect(ctx context.Context, id string) (result model.Aspect, exists bool, err error)
	SetAspect(ctx context.Context, aspect model.Aspect) error
	RemoveAspect(ctx context.Context, id string) error
	ListAllAspects(ctx context.Context) ([]model.Aspect, error)
	ListAspectsWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) ([]model.Aspect, error) //returns all aspects used in combination with measuring functions

	AddAspectNode(ctx context.Context, node model.AspectNode) error
	RemoveAspectNodesByRootId(ctx context.Context, id string) error
	GetAspectNode(ctx context.Context, id string) (result model.AspectNode, exists bool, err error)
	ListAllAspectNodes(ctx context.Context) ([]model.AspectNode, error)
	ListAspectNodesWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) ([]model.AspectNode, error) //returns all aspects used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
	ListAspectNodesByIdList(ctx context.Context, ids []string) ([]model.AspectNode, error)

	SetCharacteristic(ctx context.Context, characteristic model.Characteristic) error
	RemoveCharacteristic(ctx context.Context, id string) error
	GetCharacteristic(ctx context.Context, id string) (result model.Characteristic, exists bool, err error)
	ListAllCharacteristics(ctx context.Context) ([]model.Characteristic, error)

	SetConcept(ctx context.Context, concept model.Concept) error
	RemoveConcept(ctx context.Context, id string) error
	GetConceptWithCharacteristics(ctx context.Context, id string) (result model.ConceptWithCharacteristics, exists bool, err error)
	GetConceptWithoutCharacteristics(ctx context.Context, id string) (result model.Concept, exists bool, err error)

	SetDeviceClass(ctx context.Context, class model.DeviceClass) error
	RemoveDeviceClass(ctx context.Context, id string) error
	ListAllDeviceClasses(ctx context.Context) ([]model.DeviceClass, error)
	ListAllDeviceClassesUsedWithControllingFunctions(ctx context.Context) ([]model.DeviceClass, error) //returns all device-classes used in combination with controlling functions
	GetDeviceClass(ctx context.Context, id string) (result model.DeviceClass, exists bool, err error)

	SetFunction(ctx context.Context, function model.Function) error
	GetFunction(ctx context.Context, id string) (result model.Function, exists bool, err error)
	RemoveFunction(ctx context.Context, id string) error
	ListAllFunctionsByType(ctx context.Context, rdfType string) ([]model.Function, error)
	ListAllMeasuringFunctionsByAspect(ctx context.Context, aspect string, ancestors bool, descendants bool) ([]model.Function, error) //returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
	ListAllFunctionsByDeviceClass(ctx context.Context, class string) ([]model.Function, error)                                        //returns all functions used in combination with given device-class
	ListAllControllingFunctionsByDeviceClass(ctx context.Context, class string) ([]model.Function, error)                             //returns all controlling functions used in combination with given device-class

	SetLocation(ctx context.Context, location model.Location) error
	RemoveLocation(ctx context.Context, id string) error
	GetLocation(ctx context.Context, id string) (result model.Location, exists bool, err error)
}
