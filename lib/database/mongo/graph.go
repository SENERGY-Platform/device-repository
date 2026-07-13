/*
 * Copyright 2026 InfAI (CC SES)
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
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var GraphBson = getBsonFieldObject[models.Graph]()
var GraphNodeBson = getBsonFieldObject[models.Node]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		collection := db.graphCollection()
		err := db.ensureIndex(collection, "graphidindex", GraphBson.Id, true, true)
		if err != nil {
			return err
		}
		err = db.ensureCompoundIndex(collection, "graphnodenodedeviceindex", true, false, GraphBson.Nodes[0].ResourceType, GraphBson.Nodes[0].ResourceId)
		if err != nil {
			return err
		}
		return nil
	})
}

func (this *Mongo) graphCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoGraphCollection)
}

func (this *Mongo) ListGraphs(ctx context.Context, listOptions model.GraphListOptions) (result []models.Graph, total int64, err error) {
	opt := options.Find()
	if listOptions.Limit > 0 {
		opt.SetLimit(listOptions.Limit)
	}
	if listOptions.Offset > 0 {
		opt.SetSkip(listOptions.Offset)
	}

	if listOptions.SortBy == "" {
		listOptions.SortBy = GraphBson.Id + ".asc"
	}

	sortby := listOptions.SortBy
	sortby = strings.TrimSuffix(sortby, ".asc")
	sortby = strings.TrimSuffix(sortby, ".desc")

	direction := int32(1)
	if strings.HasSuffix(listOptions.SortBy, ".desc") {
		direction = int32(-1)
	}
	opt.SetSort(bson.D{{sortby, direction}})

	andFilter := []interface{}{bson.M{NotDeletedFilterKey: NotDeletedFilterValue}}
	filter := bson.M{}
	if listOptions.Ids != nil {
		filter[GraphBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	if listOptions.Attributes != nil {
		for _, attr := range listOptions.Attributes {
			attrFilter := bson.M{}
			attrFilter["key"] = attr.Key
			if attr.Value != "" {
				attrFilter["value"] = attr.Value
			}
			if attr.Origin != "" {
				attrFilter["origin"] = attr.Origin
			}
			andFilter = append(andFilter, bson.M{"attributes": bson.M{"$elemMatch": attrFilter}})
		}
	}
	if listOptions.DeviceIds != nil {
		orFilter := []interface{}{}
		for _, deviceId := range listOptions.DeviceIds {
			nodeFilter := bson.M{
				GraphNodeBson.ResourceType: models.GraphResourceTypeDevice,
				GraphNodeBson.ResourceId:   deviceId,
			}
			orFilter = append(orFilter, bson.M{"nodes": bson.M{"$elemMatch": nodeFilter}})
		}
		if len(orFilter) > 0 {
			andFilter = append(andFilter, bson.M{"$or": orFilter})
		}
	}

	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		orFilter := bson.M{"$or": []interface{}{
			bson.M{GraphBson.Id: bson.M{"$regex": escapedSearch, "$options": "i"}},
			bson.M{GraphBson.Attributes[0].Value: bson.M{"$regex": escapedSearch, "$options": "i"}},
		}}
		andFilter = append(andFilter, orFilter)
	}

	filter["$and"] = andFilter
	cursor, err := this.graphCollection().Find(ctx, filter, opt)
	if err != nil {
		return nil, 0, err
	}
	err = cursor.All(ctx, &result)
	if err != nil {
		return nil, 0, err
	}
	total, err = this.graphCollection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

func (this *Mongo) GetGraph(ctx context.Context, id string) (graph models.Graph, exists bool, err error) {
	result := this.graphCollection().FindOne(ctx, bson.M{GraphBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return graph, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&graph)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return graph, false, nil
	}
	return graph, true, err
}

type GraphWithSyncInfo struct {
	models.Graph `bson:",inline"`
	SyncInfo     `bson:",inline"`
}

func (this *Mongo) SetGraph(ctx context.Context, graph models.Graph, syncHandler func(models.Graph) error) error {
	timestamp := time.Now().Unix()
	collection := this.graphCollection()
	_, err := collection.ReplaceOne(ctx, bson.M{GraphBson.Id: graph.Id}, GraphWithSyncInfo{
		Graph: graph,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = this.SetLastUpdateTimestamp(ctx, this.config.MongoGraphCollection, graph.Owner)
	if err != nil {
		err = nil
	}
	err = syncHandler(graph)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in SetGraph::syncHandler %v, will be retried later\n", err))
		return nil
	}
	err = this.setSynced(ctx, collection, GraphBson.Id, graph.Id, timestamp)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in SetGraph::setSynced %v, will be retried later\n", err))
		return nil
	}
	return nil
}

func (this *Mongo) RemoveGraph(ctx context.Context, id string, syncDeleteHandler func(models.Graph) error) error {
	old, exists, err := this.GetGraph(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.graphCollection()
	err = this.setDeleted(ctx, collection, GraphBson.Id, id)
	if err != nil {
		return err
	}
	err = this.SetLastUpdateTimestamp(ctx, this.config.MongoGraphCollection, old.Owner)
	if err != nil {
		err = nil
	}
	err = syncDeleteHandler(old)
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in RemoveGraph::syncDeleteHandler %v, will be retried later\n", err))
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{GraphBson.Id: id})
	if err != nil {
		this.config.GetLogger().Warn(fmt.Sprintf("error in RemoveGraph::DeleteOne %v, will be retried later\n", err))
		return nil
	}
	return nil
}

func (this *Mongo) RetryGraphSync(lockduration time.Duration, syncDeleteHandler func(models.Graph) error, syncHandler func(models.Graph) error) error {
	collection := this.graphCollection()
	jobs, err := FetchSyncJobs[GraphWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.Graph)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryGraphSync::syncDeleteHandler %v, will be retried later\n", err))
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{GraphBson.Id: job.Id})
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryGraphSync::DeleteOne %v, will be retried later\n", err))
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.Graph)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryGraphSync::syncHandler %v, will be retried later\n", err))
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, GraphBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				this.config.GetLogger().Warn(fmt.Sprintf("error in RetryGraphSync::setSynced %v, will be retried later\n", err))
				continue
			}
		}
	}
	return nil
}

func (this *Mongo) DesyncUnknownGraphs(ctx context.Context, knownGraphs []string) (err error) {
	return this.desyncUnknown(ctx, this.graphCollection(), GraphBson.Id, knownGraphs)
}
