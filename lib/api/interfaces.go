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
)

type Controller interface {
	ReadDevice(id string, token string, action model.AuthAction) (result model.Device, err error, errCode int)
	ReadDeviceByLocalId(localId string, token string, action model.AuthAction) (result model.Device, err error, errCode int)
	ValidateDevice(device model.Device) (err error, code int)

	ReadHub(id string, token string, action model.AuthAction) (result model.Hub, err error, errCode int)
	ListHubDeviceIds(id string, token string, action model.AuthAction, asLocalId bool) (result []string, err error, errCode int)
	ValidateHub(hub model.Hub) (err error, code int)

	ReadDeviceType(id string, token string) (result model.DeviceType, err error, errCode int)
	ListDeviceTypes(token string, limit int64, offset int64, sort string, filter []model.FilterCriteria) (result []model.DeviceType, err error, errCode int)
	ValidateDeviceType(deviceType model.DeviceType) (err error, code int)

	GetDeviceTypeSelectables(query []model.FilterCriteria, pathPrefix string) (result []model.DeviceTypeSelectable, err error, code int)

	ReadDeviceGroup(id string, token string) (result model.DeviceGroup, err error, errCode int)
	ValidateDeviceGroup(deviceGroup model.DeviceGroup) (err error, code int)
	CheckAccessToDevicesOfGroup(token string, group model.DeviceGroup) (err error, code int)

	ReadProtocol(id string, token string) (result model.Protocol, err error, errCode int)
	ListProtocols(token string, limit int64, offset int64, sort string) (result []model.Protocol, err error, errCode int)
	ValidateProtocol(protocol model.Protocol) (err error, code int)

	GetService(id string) (result model.Service, err error, code int)

	GetAspects() ([]model.Aspect, error, int)
	GetAspectsWithMeasuringFunction(ancestors bool, descendants bool) ([]model.Aspect, error, int) //returns all aspects used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
	GetAspect(id string) (model.Aspect, error, int)
	ValidateAspect(aspect model.Aspect) (err error, code int)

	GetAspectNode(id string) (model.AspectNode, error, int)
	GetAspectNodes() ([]model.AspectNode, error, int)
	GetAspectNodesMeasuringFunctions(id string, ancestors bool, descendants bool) (result []model.Function, err error, errCode int) //returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
	GetAspectNodesWithMeasuringFunction(ancestors bool, descendants bool) ([]model.AspectNode, error, int)                          //returns all aspect-nodes used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
	GetAspectNodesByIdList(strings []string) ([]model.AspectNode, error, int)

	GetLeafCharacteristics() (result []model.Characteristic, err error, errCode int)
	GetCharacteristic(id string) (result model.Characteristic, err error, errCode int)
	ValidateCharacteristics(characteristic model.Characteristic) (err error, code int)

	GetConceptWithCharacteristics(id string) (model.ConceptWithCharacteristics, error, int)
	GetConceptWithoutCharacteristics(id string) (model.Concept, error, int)
	ValidateConcept(concept model.Concept) (err error, code int)

	GetDeviceClasses() ([]model.DeviceClass, error, int)
	GetDeviceClassesWithControllingFunctions() ([]model.DeviceClass, error, int)                      //returns all device-classes used in combination with controlling functions
	GetDeviceClassesFunctions(id string) (result []model.Function, err error, errCode int)            //returns all functions used in combination with given device-class
	GetDeviceClassesControllingFunctions(id string) (result []model.Function, err error, errCode int) //returns all controlling functions used in combination with given device-class
	GetDeviceClass(id string) (result model.DeviceClass, err error, errCode int)
	ValidateDeviceClass(deviceclass model.DeviceClass) (err error, code int)

	GetFunctionsByType(rdfType string) (result []model.Function, err error, errCode int)
	GetFunction(id string) (result model.Function, err error, errCode int)
	ValidateFunction(function model.Function) (err error, code int)

	GetLocation(id string, token string) (location model.Location, err error, errCode int)
	ValidateLocation(location model.Location) (err error, code int)
}
