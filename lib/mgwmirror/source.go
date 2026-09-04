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

package mgwmirror

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/util"
)

func StartSourcePullWorker(ctx context.Context, wg *sync.WaitGroup, config configuration.Config, db database.Database) error {
	if config.MgwMirrorSourceUrl == "" {
		return fmt.Errorf("mgwmirror source url not set")
	}
	interval, err := time.ParseDuration(config.MgwMirrorUpdateInterval)
	if err != nil {
		return err
	}
	go Pull(config, db, false)
	ticker := time.NewTicker(interval)
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		if wg != nil {
			defer wg.Done()
		}
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				Pull(config, db, true)
			}
		}
	}()
	return nil
}

func Pull(config configuration.Config, db database.Database, checkLastUpdate bool) {
	config.GetLogger().Info("start mgw mirror pull")
	defer config.GetLogger().Info("finished mgw mirror pull")
	c := client.NewClient(config.MgwMirrorSourceUrl, nil)
	token := ""

	userId, err := config.GetMgwMirrorUserId()
	if err != nil {
		config.GetLogger().Error("error while getting mgw mirror user id", "error", err)
		return
	}

	checkLastUpdateF := func(collection string) (doUpdate bool) {
		config.GetLogger().Info("mgw mirror update", "collection", collection)
		return true
	}

	if checkLastUpdate {
		sourceLastUpdateTimestamps, err, _ := c.GetLastUpdateTimestamps(token, "")
		if err != nil {
			config.GetLogger().Error("error while getting source last update timestamps for mgw mirror pull", "error", err)
			return
		}
		config.GetLogger().Debug("source last update timestamps", "source_last_update_timestamps", fmt.Sprintf("%#v", sourceLastUpdateTimestamps))

		localLastUpdateTimestamps, err := db.GetLastUpdateTimestampsForUser(context.Background(), userId)
		if err != nil {
			config.GetLogger().Error("error while getting local last update timestamps for mgw mirror pull", "error", err)
			return
		}
		checkLastUpdateF = func(collection string) (doUpdate bool) {
			defer func() {
				if doUpdate {
					config.GetLogger().Info("mgw mirror update", "collection", collection)
				} else {
					config.GetLogger().Info("skip mgw mirror update", "collection", collection)
				}
			}()
			defer func() {
				if r := recover(); r != nil {
					config.GetLogger().Error("panic in mirror pull checkLastUpdateF()", "error", r)
					doUpdate = true
				}
			}()
			localIndex := slices.IndexFunc(localLastUpdateTimestamps, func(timestamp model.LastUpdateTimestamp) bool {
				return timestamp.Collection == collection
			})
			sourceIndex := slices.IndexFunc(sourceLastUpdateTimestamps, func(timestamp model.LastUpdateTimestamp) bool {
				return timestamp.Collection == collection
			})
			if sourceIndex == -1 {
				config.GetLogger().Debug("no source last update timestamps for collection --> skip update", "collection", collection)
				return false
			}
			if localIndex == -1 {
				config.GetLogger().Debug("no local last update timestamps for collection --> update", "collection", collection)
				return true
			}
			if localLastUpdateTimestamps[localIndex].UnixTimestamp < sourceLastUpdateTimestamps[sourceIndex].UnixTimestamp {
				config.GetLogger().Debug("local last update timestamp is older than source last update timestamp --> update", "collection", collection, "local_timestamp", localLastUpdateTimestamps[localIndex], "source_timestamp", sourceLastUpdateTimestamps[sourceIndex])
				return true
			}
			return false
		}
	}

	if checkLastUpdateF(config.MongoProtocolCollection) {
		for p, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Protocol, err error) {
			list, err, _ = c.ListProtocols(token, limit, offset, "")
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing protocols for mgw mirror pull", "error", err)
				break
			}
			err = db.SetProtocol(context.Background(), p, func(models.Protocol) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting protocol for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoAspectClassCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.AspectClass, err error) {
			list, _, err, _ = c.ListAspectClasses(client.AspectClassListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing aspect-classes for mgw mirror pull", "error", err)
				break
			}
			err = db.SetAspectClass(context.Background(), e, func(models.AspectClass) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting aspect-classes for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoAspectCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Aspect, err error) {
			list, _, err, _ = c.ListAspects(client.AspectListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing aspects for mgw mirror pull", "error", err)
				break
			}
			//an older source may not resolve the aspect-class down the hierarchy yet
			resolveErr, _ := controller.ResolveAspectClassIds(&e)
			if resolveErr != nil {
				config.GetLogger().Error("error while resolving aspect classes for mgw mirror pull", "error", resolveErr)
				break
			}
			err = db.SetAspect(context.Background(), e, func(models.Aspect) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting aspects for mgw mirror pull", "error", err)
				break
			}
		}

		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.AspectNode, err error) {
			list, _, err, _ = c.ListAspectNodes(client.AspectListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing aspect-nodes for mgw mirror pull", "error", err)
				break
			}
			err = db.SetAspectNode(context.Background(), e)
			if err != nil {
				config.GetLogger().Error("error while setting aspect-nodes for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoCharacteristicCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Characteristic, err error) {
			list, _, err, _ = c.ListCharacteristics(client.CharacteristicListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing characteristics for mgw mirror pull", "error", err)
				break
			}
			err = db.SetCharacteristic(context.Background(), e, func(models.Characteristic) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting characteristics for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoConceptCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Concept, err error) {
			list, _, err, _ = c.ListConcepts(client.ConceptListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing concepts for mgw mirror pull", "error", err)
				break
			}
			err = db.SetConcept(context.Background(), e, func(models.Concept) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting concepts for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoDeviceClassCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.DeviceClass, err error) {
			list, _, err, _ = c.ListDeviceClasses(client.DeviceClassListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing device-class for mgw mirror pull", "error", err)
				break
			}
			err = db.SetDeviceClass(context.Background(), e, func(models.DeviceClass) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting device-class for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoFunctionCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Function, err error) {
			list, _, err, _ = c.ListFunctions(client.FunctionListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing functions for mgw mirror pull", "error", err)
				break
			}
			err = db.SetFunction(context.Background(), e, func(models.Function) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting functions for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoDeviceTypeCollection) {
		for e, err := range util.IterBatch(100, func(limit int64, offset int64) (list []models.DeviceType, err error) {
			list, _, err, _ = c.ListDeviceTypesV3(token, client.DeviceTypeListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing device-type for mgw mirror pull", "error", err)
				break
			}
			//the source may still send only the deprecated ContentVariable.AspectId;
			//the device-type criteria written by db.SetDeviceType() read AspectIds
			controller.SetContentVariableAspectIdsOnWrite(&e)
			err = db.SetDeviceType(context.Background(), e, func(models.DeviceType) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting device-type for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoDeviceCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Device, err error) {
			list, err, _ = c.ListDevices(token, client.DeviceListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing devices for mgw mirror pull", "error", err)
				break
			}
			err = db.SetDevice(context.Background(), client.DeviceWithConnectionState{Device: e}, func(client.DeviceWithConnectionState, client.DeviceWithConnectionState) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting devices for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoDeviceGroupCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.DeviceGroup, err error) {
			list, _, err, _ = c.ListDeviceGroups(token, client.DeviceGroupListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing device-groups for mgw mirror pull", "error", err)
				break
			}
			//the source may still send only the deprecated DeviceGroupFilterCriteria.AspectId;
			//everything reading a stored device-group evaluates AspectIds
			controller.SetDeviceGroupCriteriaAspectIdsOnWrite(&e)
			err = db.SetDeviceGroup(context.Background(), e, func(models.DeviceGroup, string) error { return nil }, userId)
			if err != nil {
				config.GetLogger().Error("error while setting device-groups for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoHubCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Hub, err error) {
			list, err, _ = c.ListHubs(token, client.HubListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing hubs for mgw mirror pull", "error", err)
				break
			}
			err = db.SetHub(context.Background(), model.HubWithConnectionState{Hub: e}, func(model.HubWithConnectionState) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting hubs for mgw mirror pull", "error", err)
				break
			}
		}
	}

	if checkLastUpdateF(config.MongoLocationCollection) {
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Location, err error) {
			list, _, err, _ = c.ListLocations(token, client.LocationListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing locations for mgw mirror pull", "error", err)
				break
			}
			err = db.SetLocation(context.Background(), e, func(models.Location, string) error { return nil }, userId)
			if err != nil {
				config.GetLogger().Error("error while setting locations for mgw mirror pull", "error", err)
				break
			}
		}
	}

}
