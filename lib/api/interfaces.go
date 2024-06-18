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

package api

import (
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type Controller interface {
	ListDevices(token string, options model.DeviceListOptions) (result []models.Device, err error, errCode int)
	ReadDevice(id string, token string, action model.AuthAction) (result models.Device, err error, errCode int)
	ReadDeviceByLocalId(ownerId string, localId string, token string, action model.AuthAction) (result models.Device, err error, errCode int)
	ValidateDevice(token string, device models.Device) (err error, code int)

	ReadHub(id string, token string, action model.AuthAction) (result models.Hub, err error, errCode int)
	ListHubDeviceIds(id string, token string, action model.AuthAction, asLocalId bool) (result []string, err error, errCode int)
	ValidateHub(token string, hub models.Hub) (err error, code int)

	ReadDeviceType(id string, token string) (result models.DeviceType, err error, errCode int)
	ListDeviceTypes(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int)
	ListDeviceTypesV2(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int)
	ValidateDeviceType(deviceType models.DeviceType, options model.ValidationOptions) (err error, code int)

	GetDeviceTypeSelectables(query []model.FilterCriteria, pathPrefix string, interactionsFilter []string, includeModified bool) (result []model.DeviceTypeSelectable, err error, code int)
	GetDeviceTypeSelectablesV2(query []model.FilterCriteria, pathPrefix string, includeModified bool, servicesMustMatchAllCriteria bool) (result []model.DeviceTypeSelectable, err error, code int)

	ReadDeviceGroup(id string, token string, filterGenericDuplicateCriteria bool) (result models.DeviceGroup, err error, errCode int)
	ValidateDeviceGroup(deviceGroup models.DeviceGroup) (err error, code int)
	CheckAccessToDevicesOfGroup(token string, group models.DeviceGroup) (err error, code int)

	ReadProtocol(id string, token string) (result models.Protocol, err error, errCode int)
	ListProtocols(token string, limit int64, offset int64, sort string) (result []models.Protocol, err error, errCode int)
	ValidateProtocol(protocol models.Protocol) (err error, code int)

	GetService(id string) (result models.Service, err error, code int)

	GetAspects() ([]models.Aspect, error, int)
	GetAspectsWithMeasuringFunction(ancestors bool, descendants bool) ([]models.Aspect, error, int) //returns all aspects used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
	GetAspect(id string) (models.Aspect, error, int)
	ValidateAspect(aspect models.Aspect) (err error, code int)
	ValidateAspectDelete(id string) (err error, code int)

	GetAspectNode(id string) (models.AspectNode, error, int)
	GetAspectNodes() ([]models.AspectNode, error, int)
	GetAspectNodesMeasuringFunctions(id string, ancestors bool, descendants bool) (result []models.Function, err error, errCode int) //returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
	GetAspectNodesWithMeasuringFunction(ancestors bool, descendants bool) ([]models.AspectNode, error, int)                          //returns all aspect-nodes used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
	GetAspectNodesByIdList(strings []string) ([]models.AspectNode, error, int)

	GetLeafCharacteristics() (result []models.Characteristic, err error, errCode int)
	GetCharacteristic(id string) (result models.Characteristic, err error, errCode int)
	ValidateCharacteristics(characteristic models.Characteristic) (err error, code int)
	ValidateCharacteristicDelete(id string) (err error, code int)

	GetConceptWithCharacteristics(id string) (models.ConceptWithCharacteristics, error, int)
	GetConceptWithoutCharacteristics(id string) (models.Concept, error, int)
	ValidateConcept(concept models.Concept) (err error, code int)
	ValidateConceptDelete(id string) (err error, code int)

	GetDeviceClasses() ([]models.DeviceClass, error, int)
	GetDeviceClassesWithControllingFunctions() ([]models.DeviceClass, error, int)                      //returns all device-classes used in combination with controlling functions
	GetDeviceClassesFunctions(id string) (result []models.Function, err error, errCode int)            //returns all functions used in combination with given device-class
	GetDeviceClassesControllingFunctions(id string) (result []models.Function, err error, errCode int) //returns all controlling functions used in combination with given device-class
	GetDeviceClass(id string) (result models.DeviceClass, err error, errCode int)
	ValidateDeviceClass(deviceclass models.DeviceClass) (err error, code int)
	ValidateDeviceClassDelete(id string) (err error, code int)

	GetFunctionsByType(rdfType string) (result []models.Function, err error, errCode int)
	GetFunction(id string) (result models.Function, err error, errCode int)
	ValidateFunction(function models.Function) (err error, code int)
	ValidateFunctionDelete(id string) (err error, code int)

	GetLocation(id string, token string) (location models.Location, err error, errCode int)
	ValidateLocation(location models.Location) (err error, code int)
	GetUsedInDeviceType(query model.UsedInDeviceTypeQuery) (result model.UsedInDeviceTypeResponse, err error, errCode int)
}
