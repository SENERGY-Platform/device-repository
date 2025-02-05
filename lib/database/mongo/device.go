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
	"errors"
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

var DeviceBson = getBsonFieldObject[model.DeviceWithConnectionState]()

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoDeviceCollection)
		err = db.ensureIndex(collection, "deviceidindex", DeviceBson.Id, true, true)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicelocalidindex", DeviceBson.LocalId, true, false)
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicenameindex", DeviceBson.Name, true, false) //to support faster sort
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicedisplaynameindex", DeviceBson.DisplayName, true, false) //to support faster sort
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "deviceconnectionstateindex", DeviceBson.ConnectionState, true, false) //to support faster sort
		if err != nil {
			return err
		}
		err = db.ensureIndex(collection, "devicedevicetypeindex", DeviceBson.DeviceTypeId, true, false)
		if err != nil {
			return err
		}
		err = db.ensureCompoundIndex(collection, "deviceownerlocalidindex", true, false, DeviceBson.OwnerId, DeviceBson.LocalId)
		if err != nil {
			return err
		}
		err = migrateDisplayName(db, collection)
		if err != nil {
			return err
		}
		return nil
	})
}

func migrateDisplayName(db *Mongo, collection *mongo.Collection) error {
	ctx, _ := getTimeoutContext()
	cursor, err := collection.Find(ctx, bson.M{DeviceBson.DisplayName: bson.M{"$exists": false}})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var device model.DeviceWithConnectionState
		err := cursor.Decode(&device)
		if err != nil {
			return err
		}
		device.DisplayName = getDisplayName(device)
		_, err = collection.ReplaceOne(ctx, bson.M{DeviceBson.Id: device.Id}, device, options.Replace().SetUpsert(true))
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Mongo) deviceCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoDeviceCollection)
}

func (this *Mongo) GetDevice(ctx context.Context, id string) (device model.DeviceWithConnectionState, exists bool, err error) {
	result := this.deviceCollection().FindOne(ctx, bson.M{DeviceBson.Id: id, NotDeletedFilterKey: NotDeletedFilterValue})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return device, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&device)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return device, false, nil
	}
	return device, true, err
}

type DeviceWithSyncInfo struct {
	model.DeviceWithConnectionState `bson:",inline"`
	SyncInfo                        `bson:",inline"`
}

func (this *Mongo) SetDevice(ctx context.Context, device model.DeviceWithConnectionState, syncHandler func(model.DeviceWithConnectionState) error) error {
	device.DisplayName = getDisplayName(device)
	timestamp := time.Now().Unix()
	collection := this.deviceCollection()
	_, err := collection.ReplaceOne(ctx, bson.M{DeviceBson.Id: device.Id}, DeviceWithSyncInfo{
		DeviceWithConnectionState: device,
		SyncInfo: SyncInfo{
			SyncTodo:          true,
			SyncDelete:        false,
			SyncUnixTimestamp: timestamp,
		},
	}, options.Replace().SetUpsert(true))
	if err != nil {
		return err
	}
	err = syncHandler(device)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::syncHandler %v, will be retried later\n", err)
		return nil
	}
	err = this.setSynced(ctx, collection, DeviceBson.Id, device.Id, timestamp)
	if err != nil {
		log.Printf("WARNING: error in SetDevice::setSynced %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RemoveDevice(ctx context.Context, id string, syncDeleteHandler func(model.DeviceWithConnectionState) error) (err error) {
	old, exists, err := this.GetDevice(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	collection := this.deviceCollection()
	err = this.setDeleted(ctx, collection, DeviceBson.Id, id)
	if err != nil {
		return err
	}
	err = syncDeleteHandler(old)
	if err != nil {
		log.Printf("WARNING: error in RemoveDevice::syncDeleteHandler %v, will be retried later\n", err)
		return nil
	}
	_, err = collection.DeleteOne(ctx, bson.M{DeviceBson.Id: id})
	if err != nil {
		log.Printf("WARNING: error in RemoveDevice::DeleteOne %v, will be retried later\n", err)
		return nil
	}
	return nil
}

func (this *Mongo) RetryDeviceSync(lockduration time.Duration, syncDeleteHandler func(model.DeviceWithConnectionState) error, syncHandler func(model.DeviceWithConnectionState) error) error {
	collection := this.deviceCollection()
	jobs, err := FetchSyncJobs[DeviceWithSyncInfo](collection, lockduration, FetchSyncJobsDefaultBatchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.SyncDelete {
			err = syncDeleteHandler(job.DeviceWithConnectionState)
			if err != nil {
				log.Printf("WARNING: error in RetryDeviceSync::syncDeleteHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			_, err = collection.DeleteOne(ctx, bson.M{DeviceBson.Id: job.Id})
			if err != nil {
				log.Printf("WARNING: error in RetryDeviceSync::DeleteOne %v, will be retried later\n", err)
				continue
			}
		} else if job.SyncTodo {
			err = syncHandler(job.DeviceWithConnectionState)
			if err != nil {
				log.Printf("WARNING: error in RetryDeviceSync::syncHandler %v, will be retried later\n", err)
				continue
			}
			ctx, _ := getTimeoutContext()
			err = this.setSynced(ctx, collection, DeviceBson.Id, job.Id, job.SyncUnixTimestamp)
			if err != nil {
				log.Printf("WARNING: error in RetryDeviceSync::setSynced %v, will be retried later\n", err)
				continue
			}
		}
	}
	return nil
}

func getDisplayName(device model.DeviceWithConnectionState) string {
	displayName := device.Name
	for _, attr := range device.Attributes {
		if attr.Key == "shared/nickname" && attr.Value != "" {
			displayName = attr.Value
		}
	}
	return displayName
}

func (this *Mongo) GetDeviceByLocalId(ctx context.Context, ownerId string, localId string) (device model.DeviceWithConnectionState, exists bool, err error) {
	filter := bson.M{DeviceBson.LocalId: localId, NotDeletedFilterKey: NotDeletedFilterValue}
	if this.config.LocalIdUniqueForOwner {
		filter[DeviceBson.OwnerId] = ownerId
	}
	result := this.deviceCollection().FindOne(ctx, filter)
	err = result.Err()
	if err == mongo.ErrNoDocuments {
		return device, false, nil
	}
	if err != nil {
		return
	}
	err = result.Decode(&device)
	if err == mongo.ErrNoDocuments {
		return device, false, nil
	}
	return device, true, err
}

func (this *Mongo) ListDevices(ctx context.Context, listOptions model.DeviceListOptions, withTotal bool) (result []model.DeviceWithConnectionState, total int64, err error) {
	opt := options.Find()
	if listOptions.Limit > 0 {
		opt.SetLimit(listOptions.Limit)
	}
	if listOptions.Offset > 0 {
		opt.SetSkip(listOptions.Offset)
	}

	if listOptions.SortBy == "" {
		listOptions.SortBy = DeviceBson.Name + ".asc"
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
		filter[DeviceBson.Id] = bson.M{"$in": listOptions.Ids}
	}
	if listOptions.Owner != "" {
		filter[DeviceBson.OwnerId] = listOptions.Owner
	}
	if listOptions.DeviceTypeIds != nil {
		filter[DeviceBson.DeviceTypeId] = bson.M{"$in": listOptions.DeviceTypeIds}
	}
	if listOptions.AttributeKeys != nil {
		filter[DeviceBson.Attributes[0].Key] = bson.M{"$in": listOptions.AttributeKeys}
	}
	if listOptions.AttributeValues != nil {
		filter[DeviceBson.Attributes[0].Value] = bson.M{"$in": listOptions.AttributeValues}
	}
	search := strings.TrimSpace(listOptions.Search)
	if search != "" {
		escapedSearch := regexp.QuoteMeta(search)
		orFilter := bson.M{"$or": []interface{}{
			bson.M{DeviceBson.Name: bson.M{"$regex": escapedSearch, "$options": "i"}},
			bson.M{DeviceBson.DisplayName: bson.M{"$regex": escapedSearch, "$options": "i"}},
		}}
		andFilter = append(andFilter, orFilter)
	}
	if listOptions.ConnectionState != nil {
		filter[DeviceBson.ConnectionState] = listOptions.ConnectionState
	}

	filter["$and"] = andFilter

	cursor, err := this.deviceCollection().Find(ctx, filter, opt)
	if err != nil {
		return result, total, err
	}
	result, err, _ = readCursorResult[model.DeviceWithConnectionState](ctx, cursor)
	if err != nil {
		return result, total, err
	}
	if withTotal {
		total, err = this.deviceCollection().CountDocuments(ctx, filter)
		if err != nil {
			return result, total, err
		}
	}
	return result, total, err
}

func (this *Mongo) DeviceLocalIdsToIds(ctx context.Context, owner string, localIds []string) (ids []string, err error) {
	cursor, err := this.deviceCollection().Find(ctx, bson.M{
		DeviceBson.LocalId:  bson.M{"$in": localIds},
		DeviceBson.OwnerId:  owner,
		NotDeletedFilterKey: NotDeletedFilterValue,
	})
	if err != nil {
		return ids, err
	}
	defer cursor.Close(ctx)
	err = cursor.Err()
	if err != nil {
		return ids, err
	}
	ids = []string{}
	for cursor.Next(ctx) {
		err = cursor.Err()
		if err != nil {
			return ids, err
		}
		device := models.Device{}
		err = cursor.Decode(&device)
		if err != nil {
			return ids, err
		}
		ids = append(ids, device.Id)
	}
	err = cursor.Err()
	if err != nil {
		return ids, err
	}
	return ids, nil
}

func (this *Mongo) SetDeviceConnectionState(ctx context.Context, id string, state models.ConnectionState) error {
	_, err := this.deviceCollection().UpdateOne(ctx, bson.M{
		DeviceBson.Id: id,
	}, bson.M{
		"$set": bson.M{DeviceBson.ConnectionState: state},
	})
	return err
}
