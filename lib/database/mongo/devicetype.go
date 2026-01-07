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

package mongo

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DeviceTypeBson = getBsonFieldObject[models.DeviceType]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error

		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceTypeCollection)
		err = db.ensureIndex(collection, "devicetypeidindex", DeviceTypeBson.Id, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicetypenameindex", DeviceTypeBson.Name, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicetypeserviceindex", DeviceTypeBson.Services[0].Id, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) deviceTypeCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceTypeCollection)
}

func (this *Mongo) GetDeviceType(ctx context.Context, id string) (deviceType models.DeviceType, exists bool, err error) {
	result := this.deviceTypeCollection().FindOne(ctx, bson.M{DeviceTypeBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return deviceType, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&deviceType)
	if err == mongo.ErrNoDocuments {
		return deviceType, false, nil
	}
	return deviceType, true, err
}

func (this *Mongo) ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string, filterCriteria []model.FilterCriteria, interactionsFilter []string, includeModified bool) (result []models.DeviceType, err error) {
	result = []models.DeviceType{}
	opt := options.Find()
	opt.SetLimit(limit)
	opt.SetSkip(offset)

	parts := strings.Split(sort, ".")
	sortby := DeviceTypeBson.Id
	switch parts[0] {
	case "id":
		sortby = DeviceTypeBson.Id
	case "name":
		sortby = DeviceTypeBson.Name
	default:
		sortby = DeviceTypeBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{NotDeletedFilterKey: NotDeletedFilterValue}
	var deviceTypeIds []interface{}
	if len(filterCriteria) > 0 {
		deviceTypeIds, err = this.GetDeviceTypeIdsByFilterCriteria(ctx, filterCriteria, interactionsFilter, includeModified)
		if err != nil {
			return nil, err
		}
		filter = bson.M{DeviceTypeBson.Id: bson.M{"$in": deviceTypeIds}}
	}

	cursor, err := this.deviceTypeCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		deviceType := models.DeviceType{}
		err = cursor.Decode(&deviceType)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceType)
	}
	err = cursor.Err()
	if err != nil {
		return result, err
	}
	if includeModified {
		result = addElementsWithModifiedId(result, deviceTypeIds)
	}
	return
}

func (this *Mongo) ListDeviceTypesV2(ctx context.Context, limit int64, offset int64, sort string, filterCriteria []model.FilterCriteria, includeModified bool) (result []models.DeviceType, err error) {
	result = []models.DeviceType{}
	opt := options.Find()
	opt.SetLimit(limit)
	opt.SetSkip(offset)

	parts := strings.Split(sort, ".")
	sortby := DeviceTypeBson.Id
	switch parts[0] {
	case "id":
		sortby = DeviceTypeBson.Id
	case "name":
		sortby = DeviceTypeBson.Name
	default:
		sortby = DeviceTypeBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{NotDeletedFilterKey: NotDeletedFilterValue}
	var deviceTypeIds []interface{}
	if len(filterCriteria) > 0 {
		deviceTypeIds, err = this.GetDeviceTypeIdsByFilterCriteriaV2(ctx, filterCriteria, includeModified)
		if err != nil {
			return nil, err
		}
		filter = bson.M{DeviceTypeBson.Id: bson.M{"$in": deviceTypeIds}}
	} else if includeModified {
		deviceTypeIds, err = this.filterDeviceTypeIdsByFilterCriteriaV2(ctx, nil, model.FilterCriteria{}, includeModified)
		if err != nil {
			return nil, err
		}
	}

	cursor, err := this.deviceTypeCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		deviceType := models.DeviceType{}
		err = cursor.Decode(&deviceType)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceType)
	}
	err = cursor.Err()
	if err != nil {
		return result, err
	}
	if includeModified {
		result = addElementsWithModifiedId(result, deviceTypeIds)
	}
	return
}

func (this *Mongo) ListDeviceTypesV3(ctx context.Context, listOptions model.DeviceTypeListOptions) (result []models.DeviceType, total int64, err error) {
	result = []models.DeviceType{}

	opt := options.Find()
	if listOptions.Limit > 0 {
		opt.SetLimit(listOptions.Limit)
	}
	if listOptions.Offset > 0 {
		opt.SetSkip(listOptions.Offset)
	}

	if listOptions.SortBy == "" {
		listOptions.SortBy = DeviceTypeBson.Name + ".asc"
	}

	sortby := listOptions.SortBy
	sortby = strings.TrimSuffix(sortby, ".asc")
	sortby = strings.TrimSuffix(sortby, ".desc")
	switch sortby {
	case "id":
		sortby = DeviceTypeBson.Id
	case "name":
		sortby = DeviceTypeBson.Name
	default:
		sortby = DeviceTypeBson.Id
	}

	direction := int32(1)
	if strings.HasSuffix(listOptions.SortBy, ".desc") {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{NotDeletedFilterKey: NotDeletedFilterValue}
	if listOptions.Ids != nil {
		filter[DeviceTypeBson.Id] = bson.M{"$in": listOptions.Ids}
	}

	if listOptions.AttributeKeys != nil {
		filter[DeviceTypeBson.Attributes[0].Key] = bson.M{"$in": listOptions.AttributeKeys}
	}
	if listOptions.AttributeValues != nil {
		filter[DeviceTypeBson.Attributes[0].Value] = bson.M{"$in": listOptions.AttributeValues}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter["$or"] = []interface{}{
			bson.M{DeviceTypeBson.Name: bson.M{"$regex": escapedSearch, "$options": "i"}},
			bson.M{DeviceTypeBson.Description: bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	}

	if listOptions.ProtocolIds != nil {
		filter[DeviceTypeBson.Services[0].ProtocolId] = bson.M{"$in": listOptions.ProtocolIds}
	}

	var deviceTypeIdsWithModifier []interface{}
	if len(listOptions.Criteria) > 0 {
		deviceTypeIdsWithModifier, err = this.GetDeviceTypeIdsByFilterCriteriaV2(ctx, listOptions.Criteria, listOptions.IncludeModified)
		if err != nil {
			return nil, 0, err
		}
		filter = bson.M{DeviceTypeBson.Id: bson.M{"$in": mergeDeviceIdFilter(deviceTypeIdsWithModifier, listOptions.Ids)}}
	} else if listOptions.IncludeModified {
		deviceTypeIdsWithModifier, err = this.filterDeviceTypeIdsByFilterCriteriaV2(ctx, nil, model.FilterCriteria{}, listOptions.IncludeModified)
		if err != nil {
			return nil, 0, err
		}
	}

	cursor, err := this.deviceTypeCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		deviceType := models.DeviceType{}
		err = cursor.Decode(&deviceType)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, deviceType)
	}
	err = cursor.Err()
	if err != nil {
		return result, 0, err
	}
	if listOptions.IncludeModified {
		result = addElementsWithModifiedId(result, deviceTypeIdsWithModifier)
	}
	total, err = this.deviceTypeCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func mergeDeviceIdFilter(ids []interface{}, ids2 []string) []string {
	result := []string{}
	if ids == nil {
		return ids2
	}
	for _, id := range ids {
		idStr, ok := id.(string)
		if ok && (ids2 == nil || slices.Contains(ids2, idStr)) {
			result = append(result, idStr)
		}
	}
	return result
}

func addElementsWithModifiedId(deviceTypes []models.DeviceType, ids []interface{}) (result []models.DeviceType) {
	modifiedIndex := map[string][]string{}
	for _, idInterface := range ids {
		id, ok := idInterface.(string)
		if !ok {
			debug.PrintStack()
			continue
		}
		pure, _ := idmodifier.SplitModifier(id)
		if id != pure {
			modifiedIndex[pure] = append(modifiedIndex[pure], id)
		}
	}
	for _, dt := range deviceTypes {
		result = append(result, dt)
		modifiedIds := modifiedIndex[dt.Id]
		if len(modifiedIds) > 0 {
			for _, modifiedId := range modifiedIds {
				temp := dt
				temp.Id = modifiedId
				result = append(result, temp)
			}
		}
	}
	return result
}

type DeviceTypeWithSyncInfo struct {
	models.DeviceType `bson:",inline"`
	SyncInfo          `bson:",inline"`
}

func (this *Mongo) SetDeviceType(ctx context.Context, deviceType models.DeviceType, syncHandler func(models.DeviceType) error) error {
	timestamp := time.Now().Unix()
	collection := this.deviceTypeCollection()
	_, err := this.deviceTypeCollection().ReplaceOne(ctx, bson.M{DeviceTypeBson.Id: deviceType.Id}, DeviceTypeWithSyncInfo{
		DeviceType: deviceType,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = this.setDeviceTypeCriteria(ctx, deviceType)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in SetDeviceType::setDeviceTypeCriteria %v, will be retried later\n", err))
		return nil
	}
	err = syncHandler(deviceType)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in SetDeviceType::syncHandler %v, will be retried later\n", err))
		return nil
	}
	err = this.setSynced(ctx, collection, DeviceTypeBson.Id, deviceType.Id, timestamp)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in SetDeviceType::setSynced %v, will be retried later\n", err))
		return nil
	}
	return nil
}

func (this *Mongo) RemoveDeviceType(ctx context.Context, id string, syncDeleteHandler func(models.DeviceType) error) error {
	old, exists, err := this.GetDeviceType(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.deviceTypeCollection()
	err = this.setDeleted(ctx, collection, DeviceTypeBson.Id, id)
	if err != nil {
		return err
	}
	err = this.removeDeviceTypeCriteriaByDeviceType(ctx, id)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in RemoveDeviceType::removeDeviceTypeCriteriaByDeviceType %v, will be retried later\n", err))
		return nil
	}
	err = syncDeleteHandler(old)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in RemoveDeviceType::syncDeleteHandler %v, will be retried later\n", err))
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{DeviceTypeBson.Id: id})
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in RemoveDeviceType::DeleteOne %v, will be retried later\n", err))
		return nil
	}
	return nil
}

func (this *Mongo) RetryDeviceTypeSync(lockduration time.Duration, syncDeleteHandler func(models.DeviceType) error, syncHandler func(models.DeviceType) error) error {
	collection := this.deviceTypeCollection()
	jobs, err := FetchSyncJobs[DeviceTypeWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			ctx, _ := getTimeoutContext()
			err = this.removeDeviceTypeCriteriaByDeviceType(ctx, job.Id)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryDeviceTypeSync::removeDeviceTypeCriteriaByDeviceType %v, will be retried later\n", err))
				continue
			}
			err = syncDeleteHandler(job.DeviceType)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryDeviceTypeSync::syncDeleteHandler %v, will be retried later\n", err))
				continue
			}
			ctx, _ = getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{DeviceTypeBson.Id: job.Id})
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryDeviceTypeSync::DeleteOne %v, will be retried later\n", err))
				continue
			}
		} else if job.SyncTodo {
			ctx, _ := getTimeoutContext()
			err = this.setDeviceTypeCriteria(ctx, job.DeviceType)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryDeviceTypeSync::setDeviceTypeCriteria %v, will be retried later\n", err))
				return nil
			}
			err = syncHandler(job.DeviceType)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryDeviceTypeSync::syncHandler %v, will be retried later\n", err))
				continue
			}
			ctx, _ = getTimeoutContext()
			err = this.setSynced(ctx, collection, DeviceTypeBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryDeviceTypeSync::setSynced %v, will be retried later\n", err))
				continue
			}
		}
	}
	return nil
}

func (this *Mongo) GetDeviceTypesByServiceId(ctx context.Context, serviceId string) (result []models.DeviceType, err error) {
	opt := options.Find()
	opt.SetLimit(2)
	opt.SetSkip(0)

	cursor, err := this.deviceTypeCollection().Find(ctx, bson.M{DeviceTypeBson.Services[0].Id: serviceId, NotDeletedFilterKey: NotDeletedFilterValue}, opt)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		deviceType := models.DeviceType{}
		err = cursor.Decode(&deviceType)
		if err != nil {
			return nil, err
		}
		result = append(result, deviceType)
	}
	err = cursor.Err()
	return
}

// all criteria must match; if interactionsFilter is used (len > 0), at least one must match
func (this *Mongo) GetDeviceTypeIdsByFilterCriteria(ctx context.Context, criteria []model.FilterCriteria, interactionsFilter []string, includeModified bool) (result []interface{}, err error) {
	for _, c := range criteria {
		result, err = this.filterDeviceTypeIdsByFilterCriteria(ctx, result, c, interactionsFilter, includeModified)
		if err != nil {
			return result, err
		}
	}
	return
}

func (this *Mongo) GetDeviceTypeIdsByFilterCriteriaV2(ctx context.Context, criteria []model.FilterCriteria, includeModified bool) (result []interface{}, err error) {
	for _, c := range criteria {
		result, err = this.filterDeviceTypeIdsByFilterCriteriaV2(ctx, result, c, includeModified)
		if err != nil {
			return result, err
		}
	}
	return
}

func (this *Mongo) filterDeviceTypeIdsByFilterCriteria(ctx context.Context, deviceTypeIds []interface{}, criteria model.FilterCriteria, interactions []string, includeModified bool) (result []interface{}, err error) {
	result = []interface{}{}
	if deviceTypeIds != nil && len(deviceTypeIds) == 0 {
		return result, nil
	}
	filter := bson.M{}
	if deviceTypeIds != nil {
		filter = bson.M{
			DeviceTypeCriteriaBson.DeviceTypeId: bson.M{"$in": deviceTypeIds},
		}
	}
	if !includeModified {
		filter[deviceTypeCriteriaIsIdModifiedKey] = bson.M{"$ne": true}
	}
	if len(interactions) > 0 {
		filter[DeviceTypeCriteriaBson.Interaction] = bson.M{"$in": interactions}
	}
	if criteria.DeviceClassId != "" {
		filter[DeviceTypeCriteriaBson.DeviceClassId] = criteria.DeviceClassId
	}
	if criteria.FunctionId != "" {
		filter[DeviceTypeCriteriaBson.FunctionId] = criteria.FunctionId
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

	temp, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.DeviceTypeId, filter)
	if err != nil {
		return result, err
	}
	if temp != nil {
		result = temp
	}
	return
}

func (this *Mongo) filterDeviceTypeIdsByFilterCriteriaV2(ctx context.Context, deviceTypeIds []interface{}, criteria model.FilterCriteria, includeModified bool) (result []interface{}, err error) {
	result = []interface{}{}
	if deviceTypeIds != nil && len(deviceTypeIds) == 0 {
		return result, nil
	}
	filter := bson.M{}
	if deviceTypeIds != nil {
		filter = bson.M{
			DeviceTypeCriteriaBson.DeviceTypeId: bson.M{"$in": deviceTypeIds},
		}
	}
	if !includeModified {
		filter[deviceTypeCriteriaIsIdModifiedKey] = bson.M{"$ne": true}
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

	temp, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.DeviceTypeId, filter)
	if err != nil {
		return result, err
	}
	if temp != nil {
		result = temp
	}
	return
}
