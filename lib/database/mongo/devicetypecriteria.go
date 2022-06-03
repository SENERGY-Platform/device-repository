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

var DeviceTypeCriteriaBson = getBsonFieldObject[model.DeviceTypeCriteria]()

var deviceTypeCriteriaIsControllingFunctionFieldName, deviceTypeCriteriaIsControllingFunctionKey = "IsControllingFunction", ""
var deviceTypeCriteriaIsLeafFieldName, deviceTypeCriteriaIsLeafKey = "IsLeaf", ""
var deviceTypeCriteriaIsVoidFieldName, deviceTypeCriteriaIsVoidKey = "IsVoid", ""
var deviceTypeCriteriaIsInputFieldName, deviceTypeCriteriaIsInputKey = "IsInput", ""

func getDeviceTypeCriteriaCollectionName(config config.Config) string {
	return config.MongoDeviceTypeCollection + "_criteria"
}

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error

		deviceTypeCriteriaIsControllingFunctionKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaIsControllingFunctionFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaIsLeafKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaIsLeafFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaIsVoidKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaIsVoidFieldName)
		if err != nil {
			return err
		}
		deviceTypeCriteriaIsInputKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaIsInputFieldName)
		if err != nil {
			return err
		}
		collection := db.client.Database(db.config.MongoTable).Collection(getDeviceTypeCriteriaCollectionName(db.config))

		err = db.ensureIndex(collection, "deviceTypeCriteriaDeviceTypeIdIndex", DeviceTypeCriteriaBson.DeviceTypeId, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaServiceIdIndex", DeviceTypeCriteriaBson.ServiceId, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaContentVariableIdIndex", DeviceTypeCriteriaBson.ContentVariableId, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaFunctionIdIndex", DeviceTypeCriteriaBson.FunctionId, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaDeviceClassIdIndex", DeviceTypeCriteriaBson.DeviceClassId, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceTypeCriteriaAspectIdIndex", DeviceTypeCriteriaBson.AspectId, true, false)
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
	_, err := this.deviceTypeCriteriaCollection().DeleteMany(ctx, bson.M{DeviceTypeCriteriaBson.DeviceTypeId: deviceTypeId})
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
	isLeaf := len(variable.SubContentVariables) == 0
	isCtrlFun := isControllingFunction(variable.FunctionId)
	isInputControllingFunction := (variable.FunctionId != "" && (!strings.HasPrefix(variable.FunctionId, model.URN_PREFIX) || isCtrlFun == isInput))
	isConfigurableCandidate := isLeaf && isInput
	if isInputControllingFunction || isConfigurableCandidate {
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
			Value:                 variable.Value,
			Type:                  variable.Type,
			IsLeaf:                isLeaf,
			IsInput:               isInput,
		})
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
		DeviceTypeCriteriaBson.DeviceTypeId: bson.M{"$in": deviceTypeIds},
	}
	if criteria.DeviceClassId != "" {
		filter[DeviceTypeCriteriaBson.DeviceClassId] = criteria.DeviceClassId
	}
	if criteria.FunctionId != "" {
		filter[DeviceTypeCriteriaBson.FunctionId] = criteria.FunctionId
	}
	if criteria.Interaction != "" {
		switch criteria.Interaction {
		case model.REQUEST:
			filter[DeviceTypeCriteriaBson.Interaction] = bson.M{"$in": []string{string(model.REQUEST), string(model.EVENT_AND_REQUEST)}}
		case model.EVENT:
			filter[DeviceTypeCriteriaBson.Interaction] = bson.M{"$in": []string{string(model.EVENT), string(model.EVENT_AND_REQUEST)}}
		default:
			filter[DeviceTypeCriteriaBson.Interaction] = string(criteria.Interaction)
		}
	}
	if criteria.AspectId != "" {
		node, exists, err := this.GetAspectNode(ctx, criteria.AspectId)
		if err != nil {
			return result, err
		}
		if exists {
			filter[DeviceTypeCriteriaBson.AspectId] = bson.M{"$in": append(node.DescendentIds, node.Id)}
		} else {
			//return result, errors.New("unknown AspectId: "+criteria.AspectId)
			log.Println("WARNING: filterDeviceTypeIdsByFilterCriteria() aspect id not found as aspect-node", criteria.AspectId)
			filter[DeviceTypeCriteriaBson.AspectId] = criteria.AspectId
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

func (this *Mongo) GetConfigurableCandidates(ctx context.Context, serviceId string) (result []model.DeviceTypeCriteria, err error) {
	filter := bson.M{
		DeviceTypeCriteriaBson.ServiceId: serviceId,
		deviceTypeCriteriaIsLeafKey:      true,
		deviceTypeCriteriaIsInputKey:     true,
		deviceTypeCriteriaIsVoidKey:      false,
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

func (this *Mongo) AspectIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	filter := bson.M{
		DeviceTypeCriteriaBson.AspectId: id,
	}
	temp := this.deviceTypeCriteriaCollection().FindOne(ctx, filter)
	err = temp.Err()
	if err == mongo.ErrNoDocuments {
		return false, nil, nil
	}
	if err != nil {
		return result, nil, err
	}
	criteria := model.DeviceTypeCriteria{}
	_ = temp.Decode(&criteria)
	return true, []string{criteria.DeviceTypeId, criteria.ContentVariableId, criteria.ContentVariablePath}, nil
}

func (this *Mongo) FunctionIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	filter := bson.M{
		DeviceTypeCriteriaBson.FunctionId: id,
	}
	temp := this.deviceTypeCriteriaCollection().FindOne(ctx, filter)
	err = temp.Err()
	if err == mongo.ErrNoDocuments {
		return false, nil, nil
	}
	if err != nil {
		return result, nil, err
	}
	criteria := model.DeviceTypeCriteria{}
	_ = temp.Decode(&criteria)
	return true, []string{criteria.DeviceTypeId, criteria.ContentVariableId, criteria.ContentVariablePath}, nil
}

func (this *Mongo) DeviceClassIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	filter := bson.M{
		DeviceTypeCriteriaBson.DeviceClassId: id,
	}
	temp := this.deviceTypeCriteriaCollection().FindOne(ctx, filter)
	err = temp.Err()
	if err == mongo.ErrNoDocuments {
		return false, nil, nil
	}
	if err != nil {
		return result, nil, err
	}
	criteria := model.DeviceTypeCriteria{}
	_ = temp.Decode(&criteria)
	return true, []string{criteria.DeviceTypeId, criteria.ContentVariableId, criteria.ContentVariablePath}, nil
}
