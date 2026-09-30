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
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/client"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/controller"
	"github.com/SENERGY-Platform/device-repository/v3/lib/database"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
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
		seen := map[string]bool{}
		complete := true
		for p, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Protocol, err error) {
			list, err, _ = c.ListProtocols(token, limit, offset, "")
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing protocols for mgw mirror pull", "error", err)
				complete = false
				break
			}
			err = db.SetProtocol(context.Background(), p, func(models.Protocol) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting protocol for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[p.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoProtocolCollection, seen,
				func(limit int64, offset int64) ([]models.Protocol, error) {
					return db.ListProtocols(context.Background(), limit, offset, "")
				},
				func(e models.Protocol) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.ReadProtocol(id, token)
					return
				},
				func(e models.Protocol) error {
					return db.RemoveProtocol(context.Background(), e.Id, func(models.Protocol) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoAspectClassCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.AspectClass, err error) {
			list, _, err, _ = c.ListAspectClasses(client.AspectClassListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing aspect-classes for mgw mirror pull", "error", err)
				complete = false
				break
			}
			err = db.SetAspectClass(context.Background(), e, func(models.AspectClass) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting aspect-classes for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoAspectClassCollection, seen,
				func(limit int64, offset int64) (list []models.AspectClass, err error) {
					list, _, err = db.ListAspectClasses(context.Background(), model.AspectClassListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.AspectClass) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.GetAspectClass(id)
					return
				},
				func(e models.AspectClass) error {
					return db.RemoveAspectClass(context.Background(), e.Id, func(models.AspectClass) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoAspectCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Aspect, err error) {
			list, _, err, _ = c.ListAspects(client.AspectListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing aspects for mgw mirror pull", "error", err)
				complete = false
				break
			}
			//an older source may not resolve the aspect-class down the hierarchy yet
			resolveErr, _ := controller.ResolveAspectClassIds(&e)
			if resolveErr != nil {
				config.GetLogger().Error("error while resolving aspect classes for mgw mirror pull", "error", resolveErr)
				complete = false
				break
			}
			err = db.SetAspect(context.Background(), e, func(models.Aspect) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting aspects for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoAspectCollection, seen,
				func(limit int64, offset int64) (list []models.Aspect, err error) {
					list, _, err = db.ListAspects(context.Background(), model.AspectListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.Aspect) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.GetAspect(id)
					return
				},
				func(e models.Aspect) error {
					return db.RemoveAspect(context.Background(), e.Id, func(models.Aspect) error { return nil })
				})
		}

		seenNodes := map[string]bool{}
		sourceNodesByRoot := map[string][]models.AspectNode{}
		completeNodes := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.AspectNode, err error) {
			list, _, err, _ = c.ListAspectNodes(client.AspectListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing aspect-nodes for mgw mirror pull", "error", err)
				completeNodes = false
				break
			}
			err = db.SetAspectNode(context.Background(), e)
			if err != nil {
				config.GetLogger().Error("error while setting aspect-nodes for mgw mirror pull", "error", err)
				completeNodes = false
				break
			}
			seenNodes[e.Id] = true
			sourceNodesByRoot[e.RootId] = append(sourceNodesByRoot[e.RootId], e)
		}
		if completeNodes {
			//the database can only remove aspect-nodes by their root; like the controller does on an aspect update,
			//the nodes of an affected root are removed and the ones the source still lists are written again
			rootsWithRemovedNodes := map[string]bool{}
			removeMissing(config, config.MongoAspectCollection+"_node", seenNodes,
				func(limit int64, offset int64) (list []models.AspectNode, err error) {
					list, _, err = db.ListAspectNodes(context.Background(), model.AspectListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.AspectNode) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.GetAspectNode(id)
					return
				},
				func(e models.AspectNode) error {
					rootsWithRemovedNodes[e.RootId] = true
					return nil
				})
			for rootId := range rootsWithRemovedNodes {
				err := db.RemoveAspectNodesByRootId(context.Background(), rootId)
				if err != nil {
					config.GetLogger().Error("error while removing aspect-nodes for mgw mirror pull", "error", err, "root_id", rootId)
					break
				}
				for _, node := range sourceNodesByRoot[rootId] {
					err = db.SetAspectNode(context.Background(), node)
					if err != nil {
						config.GetLogger().Error("error while setting aspect-nodes for mgw mirror pull", "error", err)
						break
					}
				}
				if err != nil {
					break
				}
			}
		}
	}

	if checkLastUpdateF(config.MongoCharacteristicCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Characteristic, err error) {
			list, _, err, _ = c.ListCharacteristics(client.CharacteristicListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing characteristics for mgw mirror pull", "error", err)
				complete = false
				break
			}
			err = db.SetCharacteristic(context.Background(), e, func(models.Characteristic) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting characteristics for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoCharacteristicCollection, seen,
				func(limit int64, offset int64) (list []models.Characteristic, err error) {
					list, _, err = db.ListCharacteristics(context.Background(), model.CharacteristicListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.Characteristic) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.GetCharacteristic(id)
					return
				},
				func(e models.Characteristic) error {
					return db.RemoveCharacteristic(context.Background(), e.Id, func(models.Characteristic) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoConceptCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Concept, err error) {
			list, _, err, _ = c.ListConcepts(client.ConceptListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing concepts for mgw mirror pull", "error", err)
				complete = false
				break
			}
			err = db.SetConcept(context.Background(), e, func(models.Concept) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting concepts for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoConceptCollection, seen,
				func(limit int64, offset int64) (list []models.Concept, err error) {
					list, _, err = db.ListConcepts(context.Background(), model.ConceptListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.Concept) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.GetConceptWithoutCharacteristics(id)
					return
				},
				func(e models.Concept) error {
					return db.RemoveConcept(context.Background(), e.Id, func(models.Concept) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoDeviceClassCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.DeviceClass, err error) {
			list, _, err, _ = c.ListDeviceClasses(client.DeviceClassListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing device-class for mgw mirror pull", "error", err)
				complete = false
				break
			}
			err = db.SetDeviceClass(context.Background(), e, func(models.DeviceClass) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting device-class for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoDeviceClassCollection, seen,
				func(limit int64, offset int64) (list []models.DeviceClass, err error) {
					list, _, err = db.ListDeviceClasses(context.Background(), model.DeviceClassListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.DeviceClass) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.GetDeviceClass(id)
					return
				},
				func(e models.DeviceClass) error {
					return db.RemoveDeviceClass(context.Background(), e.Id, func(models.DeviceClass) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoFunctionCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Function, err error) {
			list, _, err, _ = c.ListFunctions(client.FunctionListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing functions for mgw mirror pull", "error", err)
				complete = false
				break
			}
			err = db.SetFunction(context.Background(), e, func(models.Function) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting functions for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoFunctionCollection, seen,
				func(limit int64, offset int64) (list []models.Function, err error) {
					list, _, err = db.ListFunctions(context.Background(), model.FunctionListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.Function) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.GetFunction(id)
					return
				},
				func(e models.Function) error {
					return db.RemoveFunction(context.Background(), e.Id, func(models.Function) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoDeviceTypeCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(100, func(limit int64, offset int64) (list []models.DeviceType, err error) {
			list, _, err, _ = c.ListDeviceTypesV3(token, client.DeviceTypeListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing device-type for mgw mirror pull", "error", err)
				complete = false
				break
			}
			//the source may still send only the deprecated ContentVariable.AspectId;
			//the device-type criteria written by db.SetDeviceType() read AspectIds
			controller.SetContentVariableAspectIdsOnWrite(&e)
			err = db.SetDeviceType(context.Background(), e, func(models.DeviceType) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting device-type for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoDeviceTypeCollection, seen,
				func(limit int64, offset int64) (list []models.DeviceType, err error) {
					list, _, err = db.ListDeviceTypesV3(context.Background(), model.DeviceTypeListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.DeviceType) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.ReadDeviceType(id, token)
					return
				},
				func(e models.DeviceType) error {
					return db.RemoveDeviceType(context.Background(), e.Id, func(models.DeviceType) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoDeviceCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Device, err error) {
			list, err, _ = c.ListDevices(token, client.DeviceListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing devices for mgw mirror pull", "error", err)
				complete = false
				break
			}
			err = db.SetDevice(context.Background(), client.DeviceWithConnectionState{Device: e}, func(client.DeviceWithConnectionState, client.DeviceWithConnectionState) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting devices for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoDeviceCollection, seen,
				func(limit int64, offset int64) (list []model.DeviceWithConnectionState, err error) {
					list, _, err = db.ListDevices(context.Background(), model.DeviceListOptions{Limit: limit, Offset: offset}, false)
					return
				},
				func(e model.DeviceWithConnectionState) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.ReadDevice(id, token, model.READ)
					return
				},
				func(e model.DeviceWithConnectionState) error {
					return db.RemoveDevice(context.Background(), e.Id, func(model.DeviceWithConnectionState) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoDeviceGroupCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.DeviceGroup, err error) {
			list, _, err, _ = c.ListDeviceGroups(token, client.DeviceGroupListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing device-groups for mgw mirror pull", "error", err)
				complete = false
				break
			}
			//the source may still send only the deprecated DeviceGroupFilterCriteria.AspectId;
			//everything reading a stored device-group evaluates AspectIds
			controller.SetDeviceGroupCriteriaAspectIdsOnWrite(&e)
			err = db.SetDeviceGroup(context.Background(), e, func(models.DeviceGroup, string) error { return nil }, userId)
			if err != nil {
				config.GetLogger().Error("error while setting device-groups for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoDeviceGroupCollection, seen,
				func(limit int64, offset int64) (list []models.DeviceGroup, err error) {
					list, _, err = db.ListDeviceGroups(context.Background(), model.DeviceGroupListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.DeviceGroup) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.ReadDeviceGroup(id, token, false)
					return
				},
				func(e models.DeviceGroup) error {
					return db.RemoveDeviceGroup(context.Background(), e.Id, func(models.DeviceGroup) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoHubCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Hub, err error) {
			list, err, _ = c.ListHubs(token, client.HubListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing hubs for mgw mirror pull", "error", err)
				complete = false
				break
			}
			err = db.SetHub(context.Background(), model.HubWithConnectionState{Hub: e}, func(model.HubWithConnectionState) error { return nil })
			if err != nil {
				config.GetLogger().Error("error while setting hubs for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoHubCollection, seen,
				func(limit int64, offset int64) (list []model.HubWithConnectionState, err error) {
					list, _, err = db.ListHubs(context.Background(), model.HubListOptions{Limit: limit, Offset: offset}, false)
					return
				},
				func(e model.HubWithConnectionState) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.ReadHub(id, token, model.READ)
					return
				},
				func(e model.HubWithConnectionState) error {
					return db.RemoveHub(context.Background(), e.Id, func(model.HubWithConnectionState) error { return nil })
				})
		}
	}

	if checkLastUpdateF(config.MongoLocationCollection) {
		seen := map[string]bool{}
		complete := true
		for e, err := range util.IterBatch(500, func(limit int64, offset int64) (list []models.Location, err error) {
			list, _, err, _ = c.ListLocations(token, client.LocationListOptions{
				Limit:  limit,
				Offset: offset,
			})
			return
		}) {
			if err != nil {
				config.GetLogger().Error("error while listing locations for mgw mirror pull", "error", err)
				complete = false
				break
			}
			err = db.SetLocation(context.Background(), e, func(models.Location, string) error { return nil }, userId)
			if err != nil {
				config.GetLogger().Error("error while setting locations for mgw mirror pull", "error", err)
				complete = false
				break
			}
			seen[e.Id] = true
		}
		if complete {
			removeMissing(config, config.MongoLocationCollection, seen,
				func(limit int64, offset int64) (list []models.Location, err error) {
					list, _, err = db.ListLocations(context.Background(), model.LocationListOptions{Limit: limit, Offset: offset})
					return
				},
				func(e models.Location) string { return e.Id },
				func(id string) (err error, code int) {
					_, err, code = c.GetLocation(id, token)
					return
				},
				func(e models.Location) error {
					return db.RemoveLocation(context.Background(), e.Id, func(models.Location) error { return nil })
				})
		}
	}

}

// removeMissing removes local entries of a collection that the source listing did not contain.
// It may only be called after the source listing completed without error; seen holds the ids of that listing.
// The offset paging of the source listing may skip entries if the source changes during the pull,
// so every candidate is confirmed by a single read: only 404 (deleted) or 403 (no longer visible for the mirror user)
// lead to the removal. The local entries are listed completely before the first removal, so the local paging stays stable.
func removeMissing[T any](config configuration.Config, collection string, seen map[string]bool, listLocal func(limit int64, offset int64) ([]T, error), getId func(T) string, readSource func(id string) (err error, code int), remove func(T) error) {
	candidates := []T{}
	for e, err := range util.IterBatch(500, listLocal) {
		if err != nil {
			config.GetLogger().Error("error while listing local entries for mgw mirror pull", "collection", collection, "error", err)
			return
		}
		if !seen[getId(e)] {
			candidates = append(candidates, e)
		}
	}
	for _, e := range candidates {
		id := getId(e)
		err, code := readSource(id)
		if err == nil {
			config.GetLogger().Debug("entry missing in source listing but readable --> keep", "collection", collection, "id", id)
			continue
		}
		if code != http.StatusNotFound && code != http.StatusForbidden {
			config.GetLogger().Warn("unable to confirm removal in mgw mirror source --> keep", "collection", collection, "id", id, "status", code, "error", err)
			continue
		}
		err = remove(e)
		if err != nil {
			config.GetLogger().Error("error while removing entry for mgw mirror pull", "collection", collection, "id", id, "error", err)
			return
		}
		config.GetLogger().Info("removed entry missing in mgw mirror source", "collection", collection, "id", id, "status", code)
	}
}
