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
	"context"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type Controller interface {
	ListDevices(ctx context.Context, token string, options model.DeviceListOptions) (result []models.Device, err error, errCode int)
	ReadDevice(ctx context.Context, id string, token string, action model.AuthAction) (result models.Device, err error, errCode int)
	ReadDeviceByLocalId(ctx context.Context, ownerId string, localId string, token string, action model.AuthAction) (result models.Device, err error, errCode int)
	ValidateDevice(ctx context.Context, token string, device models.Device) (err error, code int)
	SetDevice(ctx context.Context, token string, device models.Device, options model.DeviceUpdateOptions) (result models.Device, err error, code int)
	CreateDevice(ctx context.Context, token string, device models.Device) (result models.Device, err error, code int)
	DeleteDevice(ctx context.Context, token string, id string) (err error, code int)

	ListExtendedDevices(ctx context.Context, token string, options model.ExtendedDeviceListOptions) (result []models.ExtendedDevice, total int64, err error, errCode int)
	ReadExtendedDevice(ctx context.Context, id string, token string, action model.AuthAction, fullDt bool) (result models.ExtendedDevice, err error, errCode int)
	ReadExtendedDeviceByLocalId(ctx context.Context, ownerId string, localId string, token string, action model.AuthAction, fullDt bool) (result models.ExtendedDevice, err error, errCode int)

	ReadHub(ctx context.Context, id string, token string, action model.AuthAction) (result models.Hub, err error, errCode int)
	ListHubs(ctx context.Context, token string, options model.HubListOptions) (result []models.Hub, err error, errCode int)
	ListHubDeviceIds(ctx context.Context, id string, token string, action model.AuthAction, asLocalId bool) (result []string, err error, errCode int)
	ValidateHub(ctx context.Context, token string, hub models.Hub) (err error, code int)
	SetHub(ctx context.Context, token string, hub models.Hub, options model.HubUpdateOptions) (result models.Hub, err error, errCode int)
	DeleteHub(ctx context.Context, token string, id string) (err error, code int)

	ListExtendedHubs(ctx context.Context, token string, options model.HubListOptions) (result []models.ExtendedHub, total int64, err error, errCode int)
	ReadExtendedHub(ctx context.Context, id string, token string, action model.AuthAction) (result models.ExtendedHub, err error, errCode int)

	ReadDeviceType(ctx context.Context, id string, token string) (result models.DeviceType, err error, errCode int)
	ListDeviceTypes(ctx context.Context, token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, interactionsFilter []string, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int)
	ListDeviceTypesV2(ctx context.Context, token string, limit int64, offset int64, sort string, filter []model.FilterCriteria, includeModified bool, includeUnmodified bool) (result []models.DeviceType, err error, errCode int)
	ListDeviceTypesV3(ctx context.Context, token string, listOptions model.DeviceTypeListOptions) (result []models.DeviceType, total int64, err error, errCode int)
	ListDeviceTypesUsedByUser(ctx context.Context, token string, listOptions model.DeviceTypeListOptions) (result []models.DeviceType, total int64, err error, errCode int)
	ValidateDeviceType(ctx context.Context, deviceType models.DeviceType, options model.ValidationOptions) (err error, code int)
	SetDeviceType(ctx context.Context, token string, dt models.DeviceType, options model.DeviceTypeUpdateOptions) (result models.DeviceType, err error, errCode int)
	DeleteDeviceType(ctx context.Context, token string, id string) (err error, code int)

	GetDeviceTypeSelectables(ctx context.Context, query []model.FilterCriteria, pathPrefix string, interactionsFilter []string, includeModified bool) (result []model.DeviceTypeSelectable, err error, code int)
	GetDeviceTypeSelectablesV2(ctx context.Context, query []model.FilterCriteria, pathPrefix string, includeModified bool, servicesMustMatchAllCriteria bool) (result []model.DeviceTypeSelectable, err error, code int)

	ReadDeviceGroup(ctx context.Context, id string, token string, filterGenericDuplicateCriteria bool) (result models.DeviceGroup, err error, errCode int)
	ListDeviceGroups(ctx context.Context, token string, options model.DeviceGroupListOptions) (result []models.DeviceGroup, total int64, err error, errCode int)
	ValidateDeviceGroup(ctx context.Context, token string, deviceGroup models.DeviceGroup) (err error, code int)
	ValidateDeviceGroupDelete(ctx context.Context, token string, id string) (err error, code int)
	SetDeviceGroup(ctx context.Context, token string, dg models.DeviceGroup) (result models.DeviceGroup, err error, errCode int)
	DeleteDeviceGroup(ctx context.Context, token string, id string) (err error, code int)
	DeviceGroupHelper(ctx context.Context, token string, deviceIds []string, options model.DeviceGroupHelperOptions) (result model.DeviceGroupHelperResult, err error, code int)
	RecomputeDeviceGroupCriteria(ctx context.Context, token string, ids []string) (err error, code int)

	ReadProtocol(ctx context.Context, id string, token string) (result models.Protocol, err error, errCode int)
	ListProtocols(ctx context.Context, token string, limit int64, offset int64, sort string) (result []models.Protocol, err error, errCode int)
	ValidateProtocol(ctx context.Context, protocol models.Protocol) (err error, code int)
	SetProtocol(ctx context.Context, token string, p models.Protocol) (result models.Protocol, err error, errCode int)
	DeleteProtocol(ctx context.Context, token string, id string) (err error, code int)

	GetService(ctx context.Context, id string) (result models.Service, err error, code int)

	ListAspects(ctx context.Context, listOptions model.AspectListOptions) (result []models.Aspect, total int64, err error, errCode int)
	GetAspects(ctx context.Context) ([]models.Aspect, error, int)
	GetAspectsWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) ([]models.Aspect, error, int)   //returns all aspects used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
	GetAspectsWithControllingFunction(ctx context.Context, ancestors bool, descendants bool) ([]models.Aspect, error, int) //returns all aspects used in combination with controlling functions (usage may optionally be by its descendants or ancestors)
	GetAspect(ctx context.Context, id string) (models.Aspect, error, int)
	ValidateAspect(ctx context.Context, aspect models.Aspect) (err error, code int)
	ValidateAspectDelete(ctx context.Context, id string) (err error, code int)
	SetAspect(ctx context.Context, token string, aspect models.Aspect) (models.Aspect, error, int)
	DeleteAspect(ctx context.Context, token string, id string) (err error, code int)

	ListAspectNodes(ctx context.Context, listOptions model.AspectListOptions) (result []models.AspectNode, total int64, err error, errCode int)
	GetAspectNode(ctx context.Context, id string) (models.AspectNode, error, int)
	GetAspectNodes(ctx context.Context) ([]models.AspectNode, error, int)
	GetAspectNodesMeasuringFunctions(ctx context.Context, id string, ancestors bool, descendants bool) (result []models.Function, err error, errCode int)   //returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
	GetAspectNodesWithMeasuringFunction(ctx context.Context, ancestors bool, descendants bool) ([]models.AspectNode, error, int)                            //returns all aspect-nodes used in combination with measuring functions (usage may optionally be by its descendants or ancestors)
	GetAspectNodesControllingFunctions(ctx context.Context, id string, ancestors bool, descendants bool) (result []models.Function, err error, errCode int) //returns all controlling functions used in combination with given aspect (and optional its descendants and ancestors)
	GetAspectNodesWithControllingFunction(ctx context.Context, ancestors bool, descendants bool) ([]models.AspectNode, error, int)                          //returns all aspect-nodes used in combination with controlling functions (usage may optionally be by its descendants or ancestors)
	GetAspectNodesByIdList(ctx context.Context, strings []string) ([]models.AspectNode, error, int)

	ListCharacteristics(ctx context.Context, listOptions model.CharacteristicListOptions) (result []models.Characteristic, total int64, err error, errCode int)
	GetCharacteristics(ctx context.Context, leafsOnly bool) (result []models.Characteristic, err error, errCode int)
	GetCharacteristic(ctx context.Context, id string) (result models.Characteristic, err error, errCode int)
	ValidateCharacteristics(ctx context.Context, characteristic models.Characteristic) (err error, code int)
	ValidateCharacteristicDelete(ctx context.Context, id string) (err error, code int)
	SetCharacteristic(ctx context.Context, token string, characteristic models.Characteristic) (result models.Characteristic, err error, errCode int)
	DeleteCharacteristic(ctx context.Context, token string, id string) (err error, code int)

	ListConceptsWithCharacteristics(ctx context.Context, listOptions model.ConceptListOptions) (result []models.ConceptWithCharacteristics, total int64, err error, errCode int)
	ListConcepts(ctx context.Context, listOptions model.ConceptListOptions) (result []models.Concept, total int64, err error, errCode int)
	GetConceptWithCharacteristics(ctx context.Context, id string) (models.ConceptWithCharacteristics, error, int)
	GetConceptWithoutCharacteristics(ctx context.Context, id string) (models.Concept, error, int)
	ValidateConcept(ctx context.Context, concept models.Concept) (err error, code int)
	ValidateConceptDelete(ctx context.Context, id string) (err error, code int)
	SetConcept(ctx context.Context, token string, concept models.Concept) (result models.Concept, err error, errCode int)
	DeleteConcept(ctx context.Context, token string, id string) (err error, code int)

	ListAspectClasses(ctx context.Context, listOptions model.AspectClassListOptions) (result []models.AspectClass, total int64, err error, errCode int)
	GetAspectClass(ctx context.Context, id string) (result models.AspectClass, err error, errCode int)
	ValidateAspectClass(ctx context.Context, aspectClass models.AspectClass) (err error, code int)
	ValidateAspectClassDelete(ctx context.Context, id string) (err error, code int)
	SetAspectClass(ctx context.Context, token string, aspectClass models.AspectClass) (result models.AspectClass, err error, errCode int)
	DeleteAspectClass(ctx context.Context, token string, id string) (err error, code int)

	ListDeviceClasses(ctx context.Context, listOptions model.DeviceClassListOptions) (result []models.DeviceClass, total int64, err error, errCode int)
	GetDeviceClasses(ctx context.Context) ([]models.DeviceClass, error, int)
	GetDeviceClassesWithControllingFunctions(ctx context.Context) ([]models.DeviceClass, error, int)                        //returns all device-classes used in combination with controlling functions
	GetDeviceClassesFunctions(ctx context.Context, id string) (result []models.Function, err error, errCode int)            //returns all functions used in combination with given device-class
	GetDeviceClassesControllingFunctions(ctx context.Context, id string) (result []models.Function, err error, errCode int) //returns all controlling functions used in combination with given device-class
	GetDeviceClass(ctx context.Context, id string) (result models.DeviceClass, err error, errCode int)
	ValidateDeviceClass(ctx context.Context, deviceclass models.DeviceClass) (err error, code int)
	ValidateDeviceClassDelete(ctx context.Context, id string) (err error, code int)
	SetDeviceClass(ctx context.Context, token string, dc models.DeviceClass) (result models.DeviceClass, err error, errCode int)
	DeleteDeviceClass(ctx context.Context, token string, id string) (err error, code int)

	ListFunctions(ctx context.Context, options model.FunctionListOptions) (result []models.Function, total int64, err error, errCode int)
	GetFunctionsByType(ctx context.Context, rdfType string) (result []models.Function, err error, errCode int)
	GetFunction(ctx context.Context, id string) (result models.Function, err error, errCode int)
	ValidateFunction(ctx context.Context, function models.Function) (err error, code int)
	ValidateFunctionDelete(ctx context.Context, id string) (err error, code int)
	SetFunction(ctx context.Context, token string, f models.Function) (result models.Function, err error, errCode int)
	DeleteFunction(ctx context.Context, token string, id string) (err error, code int)

	GetLocation(ctx context.Context, id string, token string) (location models.Location, err error, errCode int)
	ValidateLocation(ctx context.Context, location models.Location) (err error, code int)
	ListLocations(ctx context.Context, token string, options model.LocationListOptions) (result []models.Location, total int64, err error, errCode int)
	ListExtendedLocations(ctx context.Context, token string, options model.LocationListOptions) (result []models.ExtendedLocation, total int64, err error, errCode int)
	GetUsedInDeviceType(ctx context.Context, query model.UsedInDeviceTypeQuery) (result model.UsedInDeviceTypeResponse, err error, errCode int)
	SetLocation(ctx context.Context, token string, location models.Location) (result models.Location, err error, errCode int)
	DeleteLocation(ctx context.Context, token string, id string) (err error, code int)

	DeleteUser(ctx context.Context, adminToken string, userId string) (err error, errCode int)

	SetHubConnectionState(ctx context.Context, token string, id string, connected bool) (error, int)
	SetHubConnectionStates(ctx context.Context, token string, states map[string]bool) (error, int)
	SetDeviceConnectionState(ctx context.Context, token string, id string, connected bool) (error, int)
	SetDeviceConnectionStates(ctx context.Context, token string, states map[string]bool) (error, int)

	Export(ctx context.Context, token string, options model.ImportExportOptions) (result model.ImportExport, err error, code int)
	Import(ctx context.Context, token string, importModel model.ImportExport, options model.ImportExportOptions) (err error, code int)

	ImportFrom(ctx context.Context, token string, includeOwnedInformation bool, options model.ImportFromOptions) (err error, code int)

	GetDefaultDeviceAttributes(ctx context.Context, token string) (attributes []models.Attribute, err error, code int)
	SetDefaultDeviceAttributes(ctx context.Context, token string, attributes []models.Attribute) (err error, code int)

	ListGraphs(ctx context.Context, token string, options model.GraphListOptions) (result []models.Graph, total int64, err error, errCode int)
	ReadGraph(ctx context.Context, token string, id string) (result models.Graph, err error, errCode int)
	SetGraph(ctx context.Context, token string, graph models.Graph) (result models.Graph, err error, code int)
	DeleteGraph(ctx context.Context, token string, id string) (error, int)

	ListThings(ctx context.Context, token string, options model.ThingListOptions) (result []models.Thing, total int64, err error, errCode int)
	ReadThing(ctx context.Context, token string, id string) (result models.Thing, err error, errCode int)
	SetThing(ctx context.Context, token string, thing models.Thing) (result models.Thing, err error, code int)
	DeleteThing(ctx context.Context, token string, id string) (error, int)
	AddThingDeviceLink(ctx context.Context, token string, thingId string, link model.ThingDeviceLinkCreate) (result models.Thing, err error, code int)
	EndThingDeviceLink(ctx context.Context, token string, thingId string, linkId string, end model.ThingDeviceLinkEnd) (result models.Thing, err error, code int)

	GetLastUpdateTimestamps(ctx context.Context, token string, userId string) (result []model.LastUpdateTimestamp, err error, code int)

	MirrorUpdate(ctx context.Context, uncheckedCollections ...string) error
}
