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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"regexp"
	"strings"
	"time"
)

var FunctionBson = getBsonFieldObject[models.Function]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoFunctionCollection)
		err = db.ensureIndex(collection, "functionidindex", FunctionBson.Id, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "functionrdftypeindex", FunctionBson.RdfType, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "functionconceptindex", FunctionBson.ConceptId, true, false)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) functionCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoFunctionCollection)
}

func (this *Mongo) ListFunctions(ctx context.Context, listOptions model.FunctionListOptions) (result []models.Function, total int64, err error) {
	opt := options.Find()
	opt.SetLimit(listOptions.Limit)
	opt.SetSkip(listOptions.Offset)

	parts := strings.Split(listOptions.SortBy, ".")
	sortby := FunctionBson.Id
	switch parts[0] {
	case "id":
		sortby = FunctionBson.Id
	case "name":
		sortby = FunctionBson.Name
	default:
		sortby = FunctionBson.Id
	}
	direction := int32(1)
	if len(parts) > 1 && parts[1] == "desc" {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	filter := bson.M{}
	if listOptions.Ids != nil {
		filter[FunctionBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	if listOptions.RdfType != "" {
		filter[FunctionBson.RdfType] = listOptions.RdfType
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		filter["$or"] = []interface{}{
			bson.M{FunctionBson.Name: bson.M{"$regex": escapedSearch, "$options": "i"}},
			bson.M{FunctionBson.DisplayName: bson.M{"$regex": escapedSearch, "$options": "i"}},
			bson.M{FunctionBson.Description: bson.M{"$regex": escapedSearch, "$options": "i"}},
		}
	}

	cursor, err := this.functionCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.functionCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) GetFunction(ctx context.Context, id string) (function models.Function, exists bool, err error) {
	result := this.functionCollection().FindOne(ctx, bson.M{FunctionBson.Id: id})
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return function, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&function)
	if err == mongo.ErrNoDocuments {
		return function, false, nil
	}
	return function, true, err
}

type FunctionWithSyncInfo struct {
	models.Function `bson:",inline"`
	SyncInfo        `bson:",inline"`
}

func (this *Mongo) SetFunction(ctx context.Context, function models.Function, syncHandler func(models.Function) error) (err error) {
	timestamp := time.Now().Unix()
	collection := this.functionCollection()
	_, err = this.functionCollection().ReplaceOne(ctx, bson.M{FunctionBson.Id: function.Id}, FunctionWithSyncInfo{
		Function: function,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = syncHandler(function)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::syncHandler %v, will be retried later\n", err)
		return nil
	}
	err = this.setSynced(ctx, collection, FunctionBson.Id, function.Id, timestamp)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::setSynced %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RemoveFunction(ctx context.Context, id string, syncDeleteHandler func(models.Function) error) error {
	old, exists, err := this.GetFunction(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.functionCollection()
	err = this.setDeleted(ctx, collection, FunctionBson.Id, id)
	if err != nil {
		return err
	}
	err = syncDeleteHandler(old)
	if err != nil {
		log.Printf("WARNING: error in RemoveFunction::syncDeleteHandler %v, will be retried later\n", err)
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{FunctionBson.Id: id})
	if err != nil {
		log.Printf("WARNING: error in RemoveFunction::DeleteOne %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RetryFunctionSync(lockduration time.Duration, syncDeleteHandler func(models.Function) error, syncHandler func(models.Function) error) error {
	collection := this.functionCollection()
	jobs, err := FetchSyncJobs[FunctionWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.Function)
			if err != nil {
				log.Printf("WARNING: error in RetryFunctionSync::syncDeleteHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{FunctionBson.Id: job.Id})
			if err != nil {
				log.Printf("WARNING: error in RetryFunctionSync::DeleteOne %v, will be retried later\n", err)
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.Function)
			if err != nil {
				log.Printf("WARNING: error in RetryFunctionSync::syncHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, FunctionBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				log.Printf("WARNING: error in RetryFunctionSync::setSynced %v, will be retried later\n", err)
				continue
			}
		}
	}
	return nil
}

func (this *Mongo) ListAllFunctionsByType(ctx context.Context, rdfType string) (result []models.Function, err error) {
	cursor, err := this.functionCollection().Find(ctx, bson.M{FunctionBson.RdfType: rdfType}, options.Find().SetSort(bson.D{{FunctionBson.Id, 1}}))
	if err != nil {
		return nil, err
	}
	result = []models.Function{}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		function := models.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

// returns all measuring functions used in combination with given aspect (and optional its descendants and ancestors)
func (this *Mongo) ListAllMeasuringFunctionsByAspect(ctx context.Context, aspect string, ancestors bool, descendants bool) (result []models.Function, err error) {
	var aspectFilter interface{}
	if ancestors || descendants {
		relatedIds := []string{aspect}
		node, exists, err := this.GetAspectNode(ctx, aspect)
		if err != nil {
			return nil, err
		}
		if exists {
			if ancestors {
				relatedIds = append(relatedIds, node.AncestorIds...)
			}
			if descendants {
				relatedIds = append(relatedIds, node.DescendentIds...)
			}
			aspectFilter = bson.M{"$in": relatedIds}
		}
	} else {
		aspectFilter = aspect
	}
	functionIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.FunctionId, bson.M{
		deviceTypeCriteriaIsControllingFunctionKey: false,
		DeviceTypeCriteriaBson.AspectId:            aspectFilter,
		DeviceTypeCriteriaBson.FunctionId:          bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.functionCollection().Find(ctx, bson.M{FunctionBson.Id: bson.M{"$in": functionIds}}, options.Find().SetSort(bson.D{{FunctionBson.Id, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.Function{}
	for cursor.Next(context.Background()) {
		function := models.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListAllFunctionsByDeviceClass(ctx context.Context, class string) (result []models.Function, err error) {
	functionIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.FunctionId, bson.M{
		DeviceTypeCriteriaBson.DeviceClassId: class,
		DeviceTypeCriteriaBson.FunctionId:    bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.functionCollection().Find(ctx, bson.M{FunctionBson.Id: bson.M{"$in": functionIds}}, options.Find().SetSort(bson.D{{FunctionBson.Id, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.Function{}
	for cursor.Next(context.Background()) {
		function := models.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ListAllControllingFunctionsByDeviceClass(ctx context.Context, class string) (result []models.Function, err error) {
	functionIds, err := this.deviceTypeCriteriaCollection().Distinct(ctx, DeviceTypeCriteriaBson.FunctionId, bson.M{
		DeviceTypeCriteriaBson.DeviceClassId:       class,
		deviceTypeCriteriaIsControllingFunctionKey: true,
		DeviceTypeCriteriaBson.FunctionId:          bson.M{"$exists": true, "$ne": ""},
	})
	if err != nil {
		return nil, err
	}
	cursor, err := this.functionCollection().Find(ctx, bson.M{FunctionBson.Id: bson.M{"$in": functionIds}}, options.Find().SetSort(bson.D{{FunctionBson.Id, 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	result = []models.Function{}
	for cursor.Next(context.Background()) {
		function := models.Function{}
		err = cursor.Decode(&function)
		if err != nil {
			return nil, err
		}
		result = append(result, function)
	}
	err = cursor.Err()
	return
}

func (this *Mongo) ConceptIsUsed(ctx context.Context, id string) (result bool, where []string, err error) {
	filter := bson.M{
		FunctionBson.ConceptId: id,
	}
	temp := this.functionCollection().FindOne(ctx, filter)
	err = temp.Err()
	if err == mongo.ErrNoDocuments {
		return false, nil, nil
	}
	if err != nil {
		return result, nil, err
	}
	function := models.Function{}
	_ = temp.Decode(&function)
	return true, []string{function.Id}, nil
}
