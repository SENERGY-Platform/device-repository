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
)

type Database interface {
	RunStartupMigrations(methods mongo.GeneratedDeviceGroupMigrationMethods) error
	Disconnect()

	GetDevice(ctx context.Context, id string) (device model.DeviceWithConnectionState, exists bool, err error)
	ListDevices(ctx context.Context, options model.DeviceListOptions, withTotal bool) (devices []model.DeviceWithConnectionState, total int64, err error)
	SetDevice(ctx context.Context, device model.DeviceWithConnectionState) error
	RemoveDevice(ctx context.Context, id string) error
	GetDeviceByLocalId(ctx context.Context, ownerId string, localId string) (device model.DeviceWithConnectionState, exists bool, err error)
	SetDeviceConnectionState(ctx context.Context, id string, state models.ConnectionState) error

	GetHub(ctx context.Context, id string) (hub model.HubWithConnectionState, exists bool, err error)
	ListHubs(ctx context.Context, options model.HubListOptions, withTotal bool) (hubs []model.HubWithConnectionState, total int64, err error)
	SetHub(ctx context.Context, hub model.HubWithConnectionState) error
	RemoveHub(ctx context.Context, id string) error
	GetHubsByDeviceId(ctx context.Context, deviceId string) (hubs []model.HubWithConnectionState, err error)
	SetHubConnectionState(ctx context.Context, id string, state models.ConnectionState) error

	GetDeviceType(ctx context.Context, id string) (deviceType models.DeviceType, exists bool, err error)
	SetDeviceType(ctx context.Context, deviceType models.DeviceType) error
	RemoveDeviceType(ctx context.Context, id string) error
	ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string, includeModified bool) (result []models.DeviceType, err error)
	ListDeviceTypesV2(ctx context.Context, limit int64, offset int64, sort string, filter []model.FilterCriteria, includeModified bool) (result []models.DeviceType, err error)
	ListDeviceTypesV3(ctx context.Context, listOptions model.DeviceTypeListOptions) (result []models.DeviceType, err error)
	GetDeviceTypesByServiceId(ctx context.Context, serviceId string) ([]models.DeviceType, error)

	GetDeviceTypeCriteriaByAspectIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error)
	GetDeviceTypeCriteriaByFunctionIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error)
	GetDeviceTypeCriteriaByDeviceClassIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error)
	GetDeviceTypeCriteriaByCharacteristicIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error)

	GetDeviceTypeCriteriaForDeviceTypeIdsAndFilterCriteria(ctx context.Context, deviceTypeIds []interface{}, criteria model.FilterCriteria, includeModified bool) (result []model.DeviceTypeCriteria, err error)
	GetDeviceTypeIdsByFilterCriteria(ctx context.Context, criteria []model.FilterCriteria, interactionsFilter []string, includeModified bool) (result []interface{}, err error)
	GetDeviceTypeIdsByFilterCriteriaV2(ctx context.Context, criteria []model.FilterCriteria, includeModified bool) (result []interface{}, err error)
	GetConfigurableCandidates(ctx context.Context, serviceId string) (result []model.DeviceTypeCriteria, err error)

	GetDeviceGroup(ctx context.Context, id string) (deviceGroup models.DeviceGroup, exists bool, err error)
	SetDeviceGroup(ctx context.Context, deviceGroup models.DeviceGroup) error
	RemoveDeviceGroup(ctx context.Context, id string) error
	ListDeviceGroups(ctx context.Context, limit int64, offset int64, sort string) (result []models.DeviceGroup, err error)

	GetProtocol(ctx context.Context, id string) (result models.Protocol, exists bool, err error)
	ListProtocols(ctx context.Context, limit int64, offset int64, sort string) ([]models.Protocol, error)
	SetProtocol(ctx context.Context, protocol models.Protocol) error
	RemoveProtocol(ctx context.Context, id string) error

	GetAspect(ctx context.Context, id string) (result models.Aspect, exists bool, err error)
	SetAspect(ctx context.Context, aspect models.Aspect) error
	RemoveAspect(ctx context.Context, id string) error
	ListAllAspects(ctx context.Context) ([]models.Aspect, error)
	ListAspectsWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) ([]models.Aspect, error) //returns all aspects used in combination with measuring functions

	SetAspectNode(ctx context.Context, node models.AspectNode) error
	RemoveAspectNodesByRootId(ctx context.Context, id string) error
	GetAspectNode(ctx context.Context, id string) (result models.AspectNode, exists bool, err error)
	ListAllAspectNodes(ctx context.Context) ([]models.AspectNode, error)
	ListAspectNodesWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) ([]models.AspectNode, error) //returns all aspects used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
	ListAspectNodesByIdList(ctx context.Context, ids []string) ([]models.AspectNode, error)

	SetCharacteristic(ctx context.Context, characteristic models.Characteristic) error
	RemoveCharacteristic(ctx context.Context, id string) error
	GetCharacteristic(ctx context.Context, id string) (result models.Characteristic, exists bool, err error)
	ListAllCharacteristics(ctx context.Context) ([]models.Characteristic, error)

	SetConcept(ctx context.Context, concept models.Concept) error
	RemoveConcept(ctx context.Context, id string) error
	GetConceptWithCharacteristics(ctx context.Context, id string) (result models.ConceptWithCharacteristics, exists bool, err error)
	GetConceptWithoutCharacteristics(ctx context.Context, id string) (result models.Concept, exists bool, err error)

	SetDeviceClass(ctx context.Context, class models.DeviceClass) error
	RemoveDeviceClass(ctx context.Context, id string) error
	ListAllDeviceClasses(ctx context.Context) ([]models.DeviceClass, error)
	ListAllDeviceClassesUsedWithControllingFunctions(ctx context.Context) ([]models.DeviceClass, error) //returns all device-classes used in combination with controlling functions
	GetDeviceClass(ctx context.Context, id string) (result models.DeviceClass, exists bool, err error)

	SetFunction(ctx context.Context, function models.Function) error
	GetFunction(ctx context.Context, id string) (result models.Function, exists bool, err error)
	RemoveFunction(ctx context.Context, id string) error
	ListAllFunctionsByType(ctx context.Context, rdfType string) ([]models.Function, error)
	ListAllMeasuringFunctionsByAspect(ctx context.Context, aspect string, ancestors bool, descendants bool) ([]models.Function, error) //returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
	ListAllFunctionsByDeviceClass(ctx context.Context, class string) ([]models.Function, error)                                        //returns all functions used in combination with given device-class
	ListAllControllingFunctionsByDeviceClass(ctx context.Context, class string) ([]models.Function, error)                             //returns all controlling functions used in combination with given device-class

	SetLocation(ctx context.Context, location models.Location) error
	RemoveLocation(ctx context.Context, id string) error
	GetLocation(ctx context.Context, id string) (result models.Location, exists bool, err error)

	AspectIsUsed(ctx context.Context, id string) (result bool, where []string, err error)
	FunctionIsUsed(ctx context.Context, id string) (result bool, where []string, err error)
	DeviceClassIsUsed(ctx context.Context, id string) (result bool, where []string, err error)
	CharacteristicIsUsed(ctx context.Context, id string) (result bool, where []string, err error)
	CharacteristicIsUsedWithConceptInDeviceType(ctx context.Context, characteristicId string, conceptId string) (result bool, where []string, err error)
	ConceptIsUsed(ctx context.Context, id string) (result bool, where []string, err error)

	Security
}

type Security interface {
	CheckBool(token string, kind string, id string, action model.AuthAction) (allowed bool, err error)
	CheckMultiple(token string, kind string, ids []string, action model.AuthAction) (map[string]bool, error)
	GetAdminUsers(token string, topic string, resourceId string) (admins []string, err error)
	ListAccessibleResourceIds(token string, topic string, limit int64, offset int64, action model.AuthAction) (result []string, err error)
	GetPermissionsInfo(token string, topic string, id string) (requestingUser string, permissions models.Permissions, err error)
	RightsElementExists(topic string, resourceId string) (exists bool, err error)
	SetRights(resourceKind string, resourceId string, rights model.ResourceRights) error
	RemoveRights(topic string, id string) error
}
