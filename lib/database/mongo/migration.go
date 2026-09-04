/*
 * Copyright 2024 InfAI (CC SES)
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
	"runtime/debug"
	"slices"
	"strings"

	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (this *Mongo) RunStartupMigrations(helper MigrationMethods) error {
	if !this.config.RunStartupMigrations {
		this.config.GetLogger().Info("skip startup migration because config.RunStartupMigrations=false")
		return nil
	}
	err := this.runDeviceGroupMigration(helper)
	if err != nil {
		return err
	}
	err = this.runDeviceTypeCriteriaAspectIdsMigration(helper)
	if err != nil {
		return err
	}
	err = this.runDeviceGroupCriteriaAspectIdsMigration()
	if err != nil {
		return err
	}
	err = this.runGeneratedDeviceGroupCriteriaMigration(helper)
	if err != nil {
		return err
	}
	return nil
}

type MigrationMethods interface {
	DeviceIdToGeneratedDeviceGroupId(deviceId string) string
	EnsureGeneratedDeviceGroup(old models.Device, device models.Device) (err error)
	//SetContentVariableAspectIdsOnWrite is needed to read stored device-types that predate
	//models.ContentVariable.AspectIds. It lives in the controller, which this package may
	//not import, so it is handed in.
	SetContentVariableAspectIdsOnWrite(deviceType *models.DeviceType)
	GetDeviceGroupCriteria(deviceIds []string) (result []models.DeviceGroupFilterCriteria, err error, code int)
}

func (this *Mongo) runDeviceGroupMigration(helper MigrationMethods) error {
	this.config.GetLogger().Info("start runDeviceGroupMigration()")
	cursor, err := this.deviceCollection().Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		if cursor.Err() != nil {
			debug.PrintStack()
			return cursor.Err()
		}
		var device models.Device
		err = cursor.Decode(&device)
		if err != nil {
			debug.PrintStack()
			return err
		}
		id := helper.DeviceIdToGeneratedDeviceGroupId(device.Id)
		_, exists, err := this.GetDeviceGroup(context.Background(), id)
		if err != nil {
			debug.PrintStack()
			return err
		}
		if !exists {
			this.config.GetLogger().Debug("generate device-group", "deviceId", device.Id, "deviceName", device.Name)
			err = helper.EnsureGeneratedDeviceGroup(device, device)
			if err != nil {
				debug.PrintStack()
				return err
			}
		}
	}
	return nil
}

// deviceTypeCriteriaLegacyAspectIdKey is the bson name of the single aspect field that
// model.DeviceTypeCriteria carried before AspectIds. It only exists in stored documents.
const deviceTypeCriteriaLegacyAspectIdKey = "aspectid"

// runDeviceTypeCriteriaAspectIdsMigration replaces the criteria rows that hold one aspect
// each by rows that hold the whole aspect list of their content variable. The rows are
// derived data, so they are not converted but rebuilt from the stored device-types. A
// device-type written before models.ContentVariable.AspectIds carries only the deprecated
// AspectId, which is why the normalization of the controller is needed here.
func (this *Mongo) runDeviceTypeCriteriaAspectIdsMigration(helper MigrationMethods) error {
	ctx := context.Background()
	legacy := this.deviceTypeCriteriaCollection().FindOne(ctx, bson.M{deviceTypeCriteriaLegacyAspectIdKey: bson.M{"$exists": true}})
	err := legacy.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	if err != nil {
		debug.PrintStack()
		return err
	}
	this.config.GetLogger().Info("start runDeviceTypeCriteriaAspectIdsMigration()")
	//the index on the field that is about to disappear would keep indexing a missing field
	_, err = this.deviceTypeCriteriaCollection().Indexes().DropOne(ctx, "deviceTypeCriteriaAspectIdIndex")
	if err != nil {
		this.config.GetLogger().Warn("unable to drop the index of the deprecated aspect field", "error", err)
	}
	cursor, err := this.deviceTypeCollection().Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		if cursor.Err() != nil {
			debug.PrintStack()
			return cursor.Err()
		}
		deviceType := models.DeviceType{}
		err = cursor.Decode(&deviceType)
		if err != nil {
			debug.PrintStack()
			return err
		}
		helper.SetContentVariableAspectIdsOnWrite(&deviceType)
		err = this.setDeviceTypeCriteria(ctx, deviceType)
		if err != nil {
			debug.PrintStack()
			return err
		}
	}
	return nil
}

// runDeviceGroupCriteriaAspectIdsMigration fills AspectIds of the stored device-group
// criteria from the deprecated AspectId. A criteria names at most one aspect, so the
// criteria_short of the group does not change and the group keeps matching the same
// queries; what changes is that the aspect is readable without the deprecated field.
func (this *Mongo) runDeviceGroupCriteriaAspectIdsMigration() error {
	ctx := context.Background()
	criteriaKey, err := getBsonFieldName(models.DeviceGroup{}, "Criteria")
	if err != nil {
		return err
	}
	aspectIdKey, err := getBsonFieldName(models.DeviceGroupFilterCriteria{}, "AspectId")
	if err != nil {
		return err
	}
	aspectIdsKey, err := getBsonFieldName(models.DeviceGroupFilterCriteria{}, "AspectIds")
	if err != nil {
		return err
	}
	//a criteria written after the migration always has AspectIds set next to a non empty
	//AspectId, so an unset list next to one is the pre-migration shape
	filter := bson.M{criteriaKey: bson.M{"$elemMatch": bson.M{
		aspectIdKey:  bson.M{"$ne": ""},
		aspectIdsKey: bson.M{"$in": bson.A{nil}},
	}}}
	cursor, err := this.deviceGroupCollection().Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	migrated := []DeviceGroupWithSyncInfo{}
	for cursor.Next(ctx) {
		if cursor.Err() != nil {
			debug.PrintStack()
			return cursor.Err()
		}
		deviceGroup := DeviceGroupWithSyncInfo{}
		err = cursor.Decode(&deviceGroup)
		if err != nil {
			debug.PrintStack()
			return err
		}
		for i, criteria := range deviceGroup.Criteria {
			if criteria.AspectId != "" && !slices.Contains(criteria.AspectIds, criteria.AspectId) {
				deviceGroup.Criteria[i].AspectIds = append(criteria.AspectIds, criteria.AspectId)
			}
		}
		migrated = append(migrated, deviceGroup)
	}
	if len(migrated) == 0 {
		return nil
	}
	this.config.GetLogger().Info("start runDeviceGroupCriteriaAspectIdsMigration()", "count", len(migrated))
	for _, deviceGroup := range migrated {
		//replace the whole document to keep the sync info, so that no consumer is notified
		//about a change that is invisible to it
		_, err = this.deviceGroupCollection().ReplaceOne(ctx, bson.M{DeviceGroupBson.Id: deviceGroup.Id}, deviceGroup, options.Replace().SetUpsert(false))
		if err != nil {
			debug.PrintStack()
			return err
		}
	}
	return nil
}

// runGeneratedDeviceGroupCriteriaMigration rebuilds the criteria of the auto generated
// device-groups. A content variable with several aspects used to produce one criteria per
// aspect, which does not record that a single variable carries all of them; the generator
// now writes that list next to the single aspects, and a query over several aspects looks
// for it. Only the generated groups are rebuilt: their criteria are derived from their
// device by definition, while a manually created group may carry hand written ones and
// converges on the next write of one of its device-types.
//
// There is no marker for a group that has already been rebuilt, so this recomputes and
// compares on every start and writes only what actually differs. That costs one device-type
// read per generated group at startup, in the order of what runDeviceGroupMigration above
// already spends.
func (this *Mongo) runGeneratedDeviceGroupCriteriaMigration(helper MigrationMethods) error {
	ctx := context.Background()
	cursor, err := this.deviceGroupCollection().Find(ctx, bson.M{
		DeviceGroupBson.AutoGeneratedByDevice: bson.M{"$nin": bson.A{nil, ""}},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	outdated := []DeviceGroupWithSyncInfo{}
	for cursor.Next(ctx) {
		if cursor.Err() != nil {
			debug.PrintStack()
			return cursor.Err()
		}
		deviceGroup := DeviceGroupWithSyncInfo{}
		err = cursor.Decode(&deviceGroup)
		if err != nil {
			debug.PrintStack()
			return err
		}
		criteria, err, _ := helper.GetDeviceGroupCriteria(deviceGroup.DeviceIds)
		if err != nil {
			debug.PrintStack()
			return err
		}
		if slices.Equal(criteriaShorts(deviceGroup.Criteria), criteriaShorts(criteria)) {
			continue
		}
		slices.SortFunc(criteria, func(a, b models.DeviceGroupFilterCriteria) int {
			return strings.Compare(a.Short(), b.Short())
		})
		deviceGroup.Criteria = criteria
		deviceGroup.SetShortCriteria()
		outdated = append(outdated, deviceGroup)
	}
	err = cursor.Err()
	if err != nil {
		debug.PrintStack()
		return err
	}
	if len(outdated) == 0 {
		return nil
	}
	this.config.GetLogger().Info("start runGeneratedDeviceGroupCriteriaMigration()", "count", len(outdated))
	for _, deviceGroup := range outdated {
		//replace the whole document to keep the sync info, so that no consumer is notified
		//about a change of a representation it does not read yet
		_, err = this.deviceGroupCollection().ReplaceOne(ctx, bson.M{DeviceGroupBson.Id: deviceGroup.Id}, deviceGroup, options.Replace().SetUpsert(false))
		if err != nil {
			debug.PrintStack()
			return err
		}
	}
	return nil
}

// criteriaShorts renders a criteria list as a comparable set. Short() holds every field of a
// criteria, so two lists with the same shorts hold the same criteria.
func criteriaShorts(criteria []models.DeviceGroupFilterCriteria) []string {
	result := make([]string, len(criteria))
	for i, c := range criteria {
		result[i] = c.Short()
	}
	slices.Sort(result)
	return result
}
