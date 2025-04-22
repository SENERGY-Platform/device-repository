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
	"github.com/SENERGY-Platform/device-repository/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"time"
)

type Database interface {
	RunStartupMigrations(methods mongo.GeneratedDeviceGroupMigrationMethods) error
	Disconnect()

	GetDevice(ctx context.Context, id string) (device model.DeviceWithConnectionState, exists bool, err error)
	ListDevices(ctx context.Context, options model.DeviceListOptions, withTotal bool) (devices []model.DeviceWithConnectionState, total int64, err error)
	GetDeviceByLocalId(ctx context.Context, ownerId string, localId string) (device model.DeviceWithConnectionState, exists bool, err error)
	SetDeviceConnectionState(ctx context.Context, id string, state models.ConnectionState) error
	DeviceLocalIdsToIds(ctx context.Context, owner string, localIds []string) ([]string, error)

	SetDevice(ctx context.Context, device model.DeviceWithConnectionState, syncHandler func(old model.DeviceWithConnectionState, new model.DeviceWithConnectionState) error) error
	RemoveDevice(ctx context.Context, id string, syncDeleteHandler func(model.DeviceWithConnectionState) error) error
	RetryDeviceSync(lockduration time.Duration, syncDeleteHandler func(model.DeviceWithConnectionState) error, syncHandler func(model.DeviceWithConnectionState) error) error

	GetHub(ctx context.Context, id string) (hub model.HubWithConnectionState, exists bool, err error)
	ListHubs(ctx context.Context, options model.HubListOptions, withTotal bool) (hubs []model.HubWithConnectionState, total int64, err error)
	GetHubsByDeviceId(ctx context.Context, deviceId string) (hubs []model.HubWithConnectionState, err error)
	SetHubConnectionState(ctx context.Context, id string, state models.ConnectionState) error

	SetHub(ctx context.Context, hub model.HubWithConnectionState, syncHandler func(model.HubWithConnectionState) error) error
	RemoveHub(ctx context.Context, id string, syncDeleteHandler func(model.HubWithConnectionState) error) error
	RetryHubSync(lockduration time.Duration, syncDeleteHandler func(model.HubWithConnectionState) error, syncHandler func(model.HubWithConnectionState) error) error

	GetDeviceType(ctx context.Context, id string) (deviceType models.DeviceType, exists bool, err error)
	ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string, includeModified bool) (result []models.DeviceType, err error)
	ListDeviceTypesV2(ctx context.Context, limit int64, offset int64, sort string, filter []model.FilterCriteria, includeModified bool) (result []models.DeviceType, err error)
	ListDeviceTypesV3(ctx context.Context, listOptions model.DeviceTypeListOptions) (result []models.DeviceType, total int64, err error)
	GetDeviceTypesByServiceId(ctx context.Context, serviceId string) ([]models.DeviceType, error)

	SetDeviceType(ctx context.Context, deviceType models.DeviceType, syncHandler func(models.DeviceType) error) error
	RemoveDeviceType(ctx context.Context, id string, syncDeleteHandler func(models.DeviceType) error) error
	RetryDeviceTypeSync(lockduration time.Duration, syncDeleteHandler func(models.DeviceType) error, syncHandler func(models.DeviceType) error) error

	GetDeviceTypeCriteriaByAspectIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error)
	GetDeviceTypeCriteriaByFunctionIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error)
	GetDeviceTypeCriteriaByDeviceClassIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error)
	GetDeviceTypeCriteriaByCharacteristicIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error)

	GetDeviceTypeCriteriaForDeviceTypeIdsAndFilterCriteria(ctx context.Context, deviceTypeIds []interface{}, criteria model.FilterCriteria, includeModified bool) (result []model.DeviceTypeCriteria, err error)
	GetDeviceTypeIdsByFilterCriteria(ctx context.Context, criteria []model.FilterCriteria, interactionsFilter []string, includeModified bool) (result []interface{}, err error)
	GetDeviceTypeIdsByFilterCriteriaV2(ctx context.Context, criteria []model.FilterCriteria, includeModified bool) (result []interface{}, err error)
	GetConfigurableCandidates(ctx context.Context, serviceId string) (result []model.DeviceTypeCriteria, err error)

	GetDeviceGroup(ctx context.Context, id string) (deviceGroup models.DeviceGroup, exists bool, err error)
	ListDeviceGroups(ctx context.Context, options model.DeviceGroupListOptions) (result []models.DeviceGroup, total int64, err error)

	GetDeviceGroupSyncUser(ctx context.Context, deviceGroupId string) (syncUser string, exists bool, err error)
	SetDeviceGroup(ctx context.Context, deviceGroup models.DeviceGroup, syncHandler func(dg models.DeviceGroup, user string) error, user string) error
	RemoveDeviceGroup(ctx context.Context, id string, syncDeleteHandler func(models.DeviceGroup) error) error
	RetryDeviceGroupSync(lockduration time.Duration, syncDeleteHandler func(models.DeviceGroup) error, syncHandler func(dg models.DeviceGroup, user string) error) error

	GetProtocol(ctx context.Context, id string) (result models.Protocol, exists bool, err error)
	ListProtocols(ctx context.Context, limit int64, offset int64, sort string) ([]models.Protocol, error)

	SetProtocol(ctx context.Context, protocol models.Protocol, syncHandler func(models.Protocol) error) error
	RemoveProtocol(ctx context.Context, id string, syncDeleteHandler func(models.Protocol) error) error
	RetryProtocolSync(lockduration time.Duration, syncDeleteHandler func(models.Protocol) error, syncHandler func(models.Protocol) error) error

	ListAspects(ctx context.Context, listOptions model.AspectListOptions) (result []models.Aspect, total int64, err error)
	GetAspect(ctx context.Context, id string) (result models.Aspect, exists bool, err error)
	ListAllAspects(ctx context.Context) ([]models.Aspect, error)
	ListAspectsWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) ([]models.Aspect, error) //returns all aspects used in combination with measuring functions

	SetAspect(ctx context.Context, aspect models.Aspect, syncHandler func(models.Aspect) error) error
	RemoveAspect(ctx context.Context, id string, syncDeleteHandler func(models.Aspect) error) error
	RetryAspectSync(lockduration time.Duration, syncDeleteHandler func(models.Aspect) error, syncHandler func(models.Aspect) error) error

	ListAspectNodes(ctx context.Context, listOptions model.AspectListOptions) (result []models.AspectNode, total int64, err error)
	SetAspectNode(ctx context.Context, node models.AspectNode) error
	RemoveAspectNodesByRootId(ctx context.Context, id string) error
	GetAspectNode(ctx context.Context, id string) (result models.AspectNode, exists bool, err error)
	ListAllAspectNodes(ctx context.Context) ([]models.AspectNode, error)
	ListAspectNodesWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) ([]models.AspectNode, error) //returns all aspects used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
	ListAspectNodesByIdList(ctx context.Context, ids []string) ([]models.AspectNode, error)

	ListCharacteristics(ctx context.Context, options model.CharacteristicListOptions) ([]models.Characteristic, int64, error)
	GetCharacteristic(ctx context.Context, id string) (result models.Characteristic, exists bool, err error)
	ListAllCharacteristics(ctx context.Context) ([]models.Characteristic, error)

	SetCharacteristic(ctx context.Context, characteristic models.Characteristic, syncHandler func(models.Characteristic) error) error
	RemoveCharacteristic(ctx context.Context, id string, syncDeleteHandler func(models.Characteristic) error) error
	RetryCharacteristicSync(lockduration time.Duration, syncDeleteHandler func(models.Characteristic) error, syncHandler func(models.Characteristic) error) error

	GetConceptWithCharacteristics(ctx context.Context, id string) (result models.ConceptWithCharacteristics, exists bool, err error)
	GetConceptWithoutCharacteristics(ctx context.Context, id string) (result models.Concept, exists bool, err error)
	ListConceptsWithCharacteristics(ctx context.Context, options model.ConceptListOptions) ([]models.ConceptWithCharacteristics, int64, error)
	ListConcepts(ctx context.Context, options model.ConceptListOptions) ([]models.Concept, int64, error)

	SetConcept(ctx context.Context, concept models.Concept, syncHandler func(models.Concept) error) error
	RemoveConcept(ctx context.Context, id string, syncDeleteHandler func(models.Concept) error) error
	RetryConceptSync(lockduration time.Duration, syncDeleteHandler func(models.Concept) error, syncHandler func(models.Concept) error) error

	ListDeviceClasses(ctx context.Context, options model.DeviceClassListOptions) ([]models.DeviceClass, int64, error)
	ListAllDeviceClasses(ctx context.Context) ([]models.DeviceClass, error)
	ListAllDeviceClassesUsedWithControllingFunctions(ctx context.Context) ([]models.DeviceClass, error) //returns all device-classes used in combination with controlling functions
	GetDeviceClass(ctx context.Context, id string) (result models.DeviceClass, exists bool, err error)

	SetDeviceClass(ctx context.Context, class models.DeviceClass, syncHandler func(models.DeviceClass) error) error
	RemoveDeviceClass(ctx context.Context, id string, syncDeleteHandler func(models.DeviceClass) error) error
	RetryDeviceClassSync(lockduration time.Duration, syncDeleteHandler func(models.DeviceClass) error, syncHandler func(models.DeviceClass) error) error

	ListFunctions(ctx context.Context, options model.FunctionListOptions) (result []models.Function, total int64, err error)
	GetFunction(ctx context.Context, id string) (result models.Function, exists bool, err error)
	ListAllFunctionsByType(ctx context.Context, rdfType string) ([]models.Function, error)
	ListAllMeasuringFunctionsByAspect(ctx context.Context, aspect string, ancestors bool, descendants bool) ([]models.Function, error) //returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
	ListAllFunctionsByDeviceClass(ctx context.Context, class string) ([]models.Function, error)                                        //returns all functions used in combination with given device-class
	ListAllControllingFunctionsByDeviceClass(ctx context.Context, class string) ([]models.Function, error)                             //returns all controlling functions used in combination with given device-class

	SetFunction(ctx context.Context, function models.Function, syncHandler func(models.Function) error) error
	RemoveFunction(ctx context.Context, id string, syncDeleteHandler func(models.Function) error) error
	RetryFunctionSync(lockduration time.Duration, syncDeleteHandler func(models.Function) error, syncHandler func(models.Function) error) error

	GetLocation(ctx context.Context, id string) (result models.Location, exists bool, err error)
	ListLocations(ctx context.Context, options model.LocationListOptions) ([]models.Location, int64, error)

	SetLocation(ctx context.Context, location models.Location, syncHandler func(l models.Location, user string) error, user string) error
	RemoveLocation(ctx context.Context, id string, syncDeleteHandler func(models.Location) error) error
	RetryLocationSync(lockduration time.Duration, syncDeleteHandler func(models.Location) error, syncHandler func(l models.Location, user string) error) error

	AspectIsUsed(ctx context.Context, id string) (result bool, where []string, err error)
	FunctionIsUsed(ctx context.Context, id string) (result bool, where []string, err error)
	DeviceClassIsUsed(ctx context.Context, id string) (result bool, where []string, err error)
	CharacteristicIsUsed(ctx context.Context, id string) (result bool, where []string, err error)
	CharacteristicIsUsedWithConceptInDeviceType(ctx context.Context, characteristicId string, conceptId string) (result bool, where []string, err error)
	ConceptIsUsed(ctx context.Context, id string) (result bool, where []string, err error)
}
