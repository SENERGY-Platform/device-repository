/*
 * Copyright 2022 InfAI (CC SES)
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

package mongo

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strings"
)

var deviceTypeCriteriaDeviceTypeIdFieldName, deviceTypeCriteriaDeviceTypeIdKey = "DeviceTypeId", ""
var deviceTypeCriteriaServiceIdFieldName, deviceTypeCriteriaServiceIdKey = "ServiceId", ""
var deviceTypeCriteriaContentVariableIdFieldName, deviceTypeCriteriaContentVariableIdKey = "ContentVariableId", ""
var deviceTypeCriteriaFunctionIdFieldName, deviceTypeCriteriaFunctionIdKey = "FunctionId", ""
var deviceTypeCriteriaDeviceClassIdFieldName, deviceTypeCriteriaDeviceClassIdKey = "DeviceClassId", ""
var deviceTypeCriteriaAspectIdFieldName, deviceTypeCriteriaAspectIdKey = "AspectId", ""
var deviceTypeCriteriaIsControllingFunctionFieldName, deviceTypeCriteriaIsControllingFunctionKey = "IsControllingFunction", ""
var deviceTypeCriteriaInteractionFieldName, deviceTypeCriteriaInteractionKey = "Interaction", ""
var deviceTypeCriteriaCharacteristicIdFieldName, deviceTypeCriteriaCharacteristicIdKey = "CharacteristicId", ""

func getDeviceTypeCriteriaCollectionName(config config.Config) string {
	return config.MongoDeviceTypeCollection + "_criteria"
}

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		deviceTypeCriteriaDeviceTypeIdKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaDeviceTypeIdFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaServiceIdKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaServiceIdFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaContentVariableIdKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaContentVariableIdFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaFunctionIdKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaFunctionIdFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaDeviceClassIdKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaDeviceClassIdFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaAspectIdKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaAspectIdFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaIsControllingFunctionKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaIsControllingFunctionFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaInteractionKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaInteractionFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaCharacteristicIdKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaCharacteristicIdFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(getDeviceTypeCriteriaCollectionName(db.config))

		err = db.ensureIndex(collection, "deviceTypeCriteriaDeviceTypeIdIndex", deviceTypeCriteriaDeviceTypeIdKey, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaServiceIdIndex", deviceTypeCriteriaServiceIdKey, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaContentVariableIdIndex", deviceTypeCriteriaContentVariableIdKey, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaFunctionIdIndex", deviceTypeCriteriaFunctionIdKey, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaDeviceClassIdIndex", deviceTypeCriteriaDeviceClassIdKey, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaAspectIdIndex", deviceTypeCriteriaAspectIdKey, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceTypeCriteriaCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(getDeviceTypeCriteriaCollectionName(this.config))
}

func (this *Mongo) addDeviceTypeCriteria(ctx context.Context, deviceTypeCriteria []model.DeviceTypeCriteria) error {
	if len(deviceTypeCriteria) == 0 {
		return nil
	}
	documents := []interface{}{}
	for _, c := range deviceTypeCriteria {
		documents = append(documents, c)
	}
	_, err := this.deviceTypeCriteriaCollection().InsertMany(ctx, documents)
	return err
}

func (this *Mongo) removeDeviceTypeCriteriaByDeviceType(ctx context.Context, deviceTypeId string) error {
	_, err := this.deviceTypeCriteriaCollection().DeleteMany(ctx, bson.M{deviceTypeCriteriaDeviceTypeIdKey: deviceTypeId})
	return err
}

func (this *Mongo) setDeviceTypeCriteria(ctx context.Context, dt model.DeviceType) error {
	err := this.removeDeviceTypeCriteriaByDeviceType(ctx, dt.Id)
	if err != nil {
		return err
	}
	return this.addDeviceTypeCriteria(ctx, createCriteriaListFromDeviceType(dt))
}

func createCriteriaListFromDeviceType(dt model.DeviceType) (result []model.DeviceTypeCriteria) {
	for _, s := range dt.Services {
		result = append(result, createCriteriaFromService(dt.Id, dt.DeviceClassId, s)...)
	}
	return result
}

func createCriteriaFromService(deviceTypeId string, deviceClassId string, service model.Service) (result []model.DeviceTypeCriteria) {
	for _, content := range service.Inputs {
		result = append(result, createCriteriaFromContentVariables(deviceTypeId, deviceClassId, service.Id, service.Interaction, content.ContentVariable, true, []string{})...)
	}
	for _, content := range service.Outputs {
		result = append(result, createCriteriaFromContentVariables(deviceTypeId, deviceClassId, service.Id, service.Interaction, content.ContentVariable, false, []string{})...)
	}
	return result
}

func createCriteriaFromContentVariables(deviceTypeId string, deviceClassId string, serviceId string, interaction model.Interaction, variable model.ContentVariable, isInput bool, pathParts []string) (result []model.DeviceTypeCriteria) {
	currentPath := append(pathParts, variable.Name)
	if variable.FunctionId != "" {
		isCtrlFun := isControllingFunction(variable.FunctionId)
		if !strings.HasPrefix(variable.FunctionId, model.URN_PREFIX) || isCtrlFun == isInput {
			result = append(result, model.DeviceTypeCriteria{
				DeviceTypeId:          deviceTypeId,
				ServiceId:             serviceId,
				ContentVariableId:     variable.Id,
				ContentVariablePath:   strings.Join(currentPath, "."),
				FunctionId:            variable.FunctionId,
				Interaction:           string(interaction),
				IsControllingFunction: isCtrlFun,
				DeviceClassId:         deviceClassId,
				AspectId:              variable.AspectId,
				CharacteristicId:      variable.CharacteristicId,
				IsVoid:                variable.IsVoid,
			})
		}
	}
	for _, sub := range variable.SubContentVariables {
		result = append(result, createCriteriaFromContentVariables(deviceTypeId, deviceClassId, serviceId, interaction, sub, isInput, currentPath)...)
	}
	return
}

func isControllingFunction(functionId string) bool {
	if strings.HasPrefix(functionId, "urn:infai:ses:controlling-function:") {
		return true
	}
	return false
}

func (this *Mongo) GetDeviceTypeCriteriaForDeviceTypeIdsAndFilterCriteria(ctx context.Context, deviceTypeIds []interface{}, criteria model.FilterCriteria) (result []model.DeviceTypeCriteria, err error) {
	filter := bson.M{
		deviceTypeCriteriaDeviceTypeIdKey: bson.M{"$in": deviceTypeIds},
	}
	if criteria.DeviceClassId != "" {
		filter[deviceTypeCriteriaDeviceClassIdKey] = criteria.DeviceClassId
	}
	if criteria.FunctionId != "" {
		filter[deviceTypeCriteriaFunctionIdKey] = criteria.FunctionId
	}
	if criteria.AspectId != "" {
		node, exists, err := this.GetAspectNode(ctx, criteria.AspectId)
		if err != nil {
			return result, err
		}
		if exists {
			filter[deviceTypeCriteriaAspectIdKey] = bson.M{"$in": append(node.DescendentIds, node.Id)}
		} else {
			//return result, errors.New("unknown AspectId: "+criteria.AspectId)
			log.Println("WARNING: filterDeviceTypeIdsByFilterCriteria() aspect id not found as aspect-node", criteria.AspectId)
			filter[deviceTypeCriteriaAspectIdKey] = criteria.AspectId
		}
	}

	cursor, err := this.deviceTypeCriteriaCollection().Find(ctx, filter)
	if err != nil {
		return result, err
	}
	for cursor.Next(context.Background()) {
		dtCriteria := model.DeviceTypeCriteria{}
		err = cursor.Decode(&dtCriteria)
		if err != nil {
			return nil, err
		}
		result = append(result, dtCriteria)
	}
	err = cursor.Err()
	return
}
