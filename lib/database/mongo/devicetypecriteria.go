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
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var DeviceTypeCriteriaBson = getBsonFieldObject[model.DeviceTypeCriteria]()

var deviceTypeCriteriaIsControllingFunctionFieldName, deviceTypeCriteriaIsControllingFunctionKey = "IsControllingFunction", ""
var deviceTypeCriteriaIsLeafFieldName, deviceTypeCriteriaIsLeafKey = "IsLeaf", ""
var deviceTypeCriteriaIsVoidFieldName, deviceTypeCriteriaIsVoidKey = "IsVoid", ""
var deviceTypeCriteriaIsInputFieldName, deviceTypeCriteriaIsInputKey = "IsInput", ""
var deviceTypeCriteriaIsIdModifiedFieldName, deviceTypeCriteriaIsIdModifiedKey = "IsIdModified", ""

func getDeviceTypeCriteriaCollectionName(config configuration.Config) string {
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
		deviceTypeCriteriaIsIdModifiedKey, err = getBsonFieldName(model.DeviceTypeCriteria{}, deviceTypeCriteriaIsIdModifiedFieldName)
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
	_, err := this.deviceTypeCriteriaCollection().DeleteMany(ctx, bson.M{DeviceTypeCriteriaBson.PureDeviceTypeId: deviceTypeId})
	return err
}

func (this *Mongo) setDeviceTypeCriteria(ctx context.Context, dt models.DeviceType) error {
	err := this.removeDeviceTypeCriteriaByDeviceType(ctx, dt.Id)
	if err != nil {
		return err
	}
	return this.addDeviceTypeCriteria(ctx, createCriteriaListFromDeviceType(dt))
}

func createCriteriaListFromDeviceType(dt models.DeviceType) (result []model.DeviceTypeCriteria) {
	for _, s := range dt.Services {
		result = append(result, createCriteriaFromService(dt.Id, dt.Id, dt.DeviceClassId, s)...)
	}
	servicesByServiceGroup := map[string][]models.Service{}
	unassignedServices := []models.Service{}
	for _, s := range dt.Services {
		if s.ServiceGroupKey == "" {
			unassignedServices = append(unassignedServices, s)
		} else {
			servicesByServiceGroup[s.ServiceGroupKey] = append(servicesByServiceGroup[s.ServiceGroupKey], s)
		}
	}
	for key, sg := range servicesByServiceGroup {
		modifiedId := dt.Id + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {key}})
		services := append(sg, unassignedServices...)
		for _, s := range services {
			result = append(result, createCriteriaFromService(dt.Id, modifiedId, dt.DeviceClassId, s)...)
		}
	}
	return result
}

func createCriteriaFromService(pureDeviceTypeId string, deviceTypeId string, deviceClassId string, service models.Service) (result []model.DeviceTypeCriteria) {
	for _, content := range service.Inputs {
		result = append(result, createCriteriaFromContentVariables(pureDeviceTypeId, deviceTypeId, deviceClassId, service.Id, service.Interaction, content.ContentVariable, true, []string{})...)
	}
	for _, content := range service.Outputs {
		result = append(result, createCriteriaFromContentVariables(pureDeviceTypeId, deviceTypeId, deviceClassId, service.Id, service.Interaction, content.ContentVariable, false, []string{})...)
	}
	return result
}

func createCriteriaFromContentVariables(pureDeviceTypeId string, deviceTypeId string, deviceClassId string, serviceId string, interaction models.Interaction, variable models.ContentVariable, isInput bool, pathParts []string) (result []model.DeviceTypeCriteria) {
	currentPath := append(pathParts, variable.Name)
	isLeaf := len(variable.SubContentVariables) == 0
	isCtrlFun := isControllingFunction(variable.FunctionId)
	isInputControllingFunction := (variable.FunctionId != "" && (!strings.HasPrefix(variable.FunctionId, model.URN_PREFIX) || isCtrlFun == isInput))
	isConfigurableCandidate := isLeaf && isInput
	if isInputControllingFunction || isConfigurableCandidate {
		result = append(result, model.DeviceTypeCriteria{
			IsIdModified:          pureDeviceTypeId != deviceTypeId,
			PureDeviceTypeId:      pureDeviceTypeId,
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
		result = append(result, createCriteriaFromContentVariables(pureDeviceTypeId, deviceTypeId, deviceClassId, serviceId, interaction, sub, isInput, currentPath)...)
	}
	return
}

func isControllingFunction(functionId string) bool {
	if strings.HasPrefix(functionId, "urn:infai:ses:controlling-function:") {
		return true
	}
	return false
}

func (this *Mongo) GetDeviceTypeCriteriaByAspectIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
	filter := bson.M{
		DeviceTypeCriteriaBson.AspectId: bson.M{"$in": ids},
	}
	if !includeModified {
		filter[deviceTypeCriteriaIsIdModifiedKey] = bson.M{"$ne": true}
	}
	cursor, err := this.deviceTypeCriteriaCollection().Find(ctx, filter)
	if err != nil {
		return result, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		dtCriteria := model.DeviceTypeCriteria{}
		err = cursor.Decode(&dtCriteria)
		if err != nil {
			return nil, err
		}
		result = append(result, dtCriteria)
	}
	err = cursor.Err()
	return result, err
}

func (this *Mongo) GetDeviceTypeCriteriaByFunctionIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
	filter := bson.M{
		DeviceTypeCriteriaBson.FunctionId: bson.M{"$in": ids},
	}
	if !includeModified {
		filter[deviceTypeCriteriaIsIdModifiedKey] = bson.M{"$ne": true}
	}
	cursor, err := this.deviceTypeCriteriaCollection().Find(ctx, filter)
	if err != nil {
		return result, err
	}
	defer cursor.Close(context.Background())
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

func (this *Mongo) GetDeviceTypeCriteriaByDeviceClassIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
	filter := bson.M{
		DeviceTypeCriteriaBson.DeviceClassId: bson.M{"$in": ids},
	}
	if !includeModified {
		filter[deviceTypeCriteriaIsIdModifiedKey] = bson.M{"$ne": true}
	}
	cursor, err := this.deviceTypeCriteriaCollection().Find(ctx, filter)
	if err != nil {
		return result, err
	}
	defer cursor.Close(context.Background())
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

func (this *Mongo) GetDeviceTypeCriteriaByCharacteristicIds(ctx context.Context, ids []string, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
	filter := bson.M{
		DeviceTypeCriteriaBson.CharacteristicId: bson.M{"$in": ids},
	}
	if !includeModified {
		filter[deviceTypeCriteriaIsIdModifiedKey] = bson.M{"$ne": true}
	}
	cursor, err := this.deviceTypeCriteriaCollection().Find(ctx, filter)
	if err != nil {
		return result, err
	}
	defer cursor.Close(context.Background())
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

func (this *Mongo) GetDeviceTypeCriteriaForDeviceTypeIdsAndFilterCriteria(ctx context.Context, deviceTypeIds []interface{}, criteria model.FilterCriteria, includeModified bool) (result []model.DeviceTypeCriteria, err error) {
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
		case models.REQUEST:
			filter[DeviceTypeCriteriaBson.Interaction] = bson.M{"$in": []string{string(models.REQUEST), string(models.EVENT_AND_REQUEST)}}
		case models.EVENT:
			filter[DeviceTypeCriteriaBson.Interaction] = bson.M{"$in": []string{string(models.EVENT), string(models.EVENT_AND_REQUEST)}}
		default:
			filter[DeviceTypeCriteriaBson.Interaction] = string(criteria.Interaction)
		}
	}
	if !includeModified {
		filter[deviceTypeCriteriaIsIdModifiedKey] = bson.M{"$ne": true}
	}
	if criteria.AspectId != "" {
		node, exists, err := this.GetAspectNode(ctx, criteria.AspectId)
		if err != nil {
			return result, err
		}
		if exists {
			filter[DeviceTypeCriteriaBson.AspectId] = bson.M{"$in": append(node.DescendentIds, node.Id)}
		} else {
			this.config.GetLogger().Warn("WARNING: filterDeviceTypeIdsByFilterCriteria() aspect id not found as aspect-node", "aspectId", criteria.AspectId)
			filter[DeviceTypeCriteriaBson.AspectId] = criteria.AspectId
		}
	}

	cursor, err := this.deviceTypeCriteriaCollection().Find(ctx, filter)
	if err != nil {
		return result, err
	}
	defer cursor.Close(context.Background())
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
		DeviceTypeCriteriaBson.ServiceId:  serviceId,
		deviceTypeCriteriaIsLeafKey:       true,
		deviceTypeCriteriaIsInputKey:      true,
		deviceTypeCriteriaIsVoidKey:       false,
		deviceTypeCriteriaIsIdModifiedKey: bson.M{"$ne": true},
	}
	cursor, err := this.deviceTypeCriteriaCollection().Find(ctx, filter)
	if err != nil {
		return result, err
	}
	defer cursor.Close(context.Background())
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
