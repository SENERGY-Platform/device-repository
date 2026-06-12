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

package lib

import (
	"context"
	"sync"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/api"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/controller/publisher"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/mgwmirror"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/util"
)

// set wg if you want to wait for clean disconnects after ctx is done
func Start(baseCtx context.Context, wg *sync.WaitGroup, conf configuration.Config) (err error) {
	ctx, cancel := context.WithCancel(baseCtx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	db, err := database.New(conf)
	if err != nil {
		conf.GetLogger().Error("unable to connect to database", "error", err)
		return err
	}
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		<-ctx.Done()
		db.Disconnect()
		if wg != nil {
			wg.Done()
		}
	}()

	var permClient client.Client
	if conf.AsMgwMirror {
		permClient = mgwmirror.NewMgwMirrorPerm(conf, db)
	} else {
		permClient = client.New(conf.PermissionsV2Url)
		if conf.EnablePermResourceSyncOnStartup {
			err = SyncPermResources(ctx, conf, permClient, db)
			if err != nil {
				conf.GetLogger().Error("unable to sync permissions resources", "error", err)
			}
		}
	}

	var p controller.Publisher
	if conf.KafkaUrl == "" || conf.KafkaUrl == "-" {
		publisher.VoidPublisherError = nil
		p = publisher.Void{}
		conf.GetLogger().Warn("kafka not configured, no publishing of events")
	} else {
		p, err = publisher.New(conf, ctx)
	}

	if err != nil {
		db.Disconnect()
		conf.GetLogger().Error("unable to start control", "error", err)
		return err
	}

	ctrl, err := controller.New(conf, db, p, permClient)
	if err != nil {
		db.Disconnect()
		conf.GetLogger().Error("unable to start control", "error", err)
		return err
	}

	if conf.RunStartupMigrations && !conf.AsMgwMirror {
		err = db.RunStartupMigrations(ctrl)
		if err != nil {
			db.Disconnect()
			conf.GetLogger().Error("unable to run startup migrations", "error", err)
			return err
		}
	}

	syncInterval := 10 * time.Minute
	if conf.SyncInterval != "" && conf.SyncInterval != "-" {
		syncInterval, err = time.ParseDuration(conf.SyncInterval)
	}
	syncLockDuration := time.Minute
	if conf.SyncLockDuration != "" && conf.SyncLockDuration != "-" {
		syncLockDuration, err = time.ParseDuration(conf.SyncLockDuration)
	}

	ctrl.StartSyncLoop(ctx, syncInterval, syncLockDuration)

	if conf.AsMgwMirror {
		err = mgwmirror.StartSourcePullWorker(ctx, wg, conf, db)
		if err != nil {
			conf.GetLogger().Error("unable to start mgw mirror source pull worker", "error", err)
			return err
		}
		ctrl.SetMirrorPullCallback(mgwmirror.Pull)
	}

	err = api.Start(ctx, conf, ctrl)
	if err != nil {
		conf.GetLogger().Error("unable to start api", "error", err)
		return err
	}

	return err
}

func deleteUnknownPermissions(ctx context.Context, config configuration.Config, permClient client.Client, topic string, check func(id string) (bool, error)) error {
	for id, err := range util.IterBatch(500, func(limit int64, offset int64) (ids []string, err error) {
		ids, err, _ = permClient.AdminListResourceIds(client.InternalAdminToken, topic, client.ListOptions{
			Limit:  limit,
			Offset: offset,
		})
		return ids, err
	}) {
		if err != nil {
			config.GetLogger().Error("error while listing permission resource for syncPermResource", "error", err, "topic", topic)
			return err
		}
		exists, err := check(id)
		if err != nil {
			config.GetLogger().Error("error while checking permission resource existence in local db for syncPermResource", "error", err, "topic", topic, "id", id)
			return err
		}
		if !exists {
			config.GetLogger().Info("removed permission resource for syncPermResource", "topic", topic, "id", id)
			err, _ = permClient.RemoveResource(client.InternalAdminToken, topic, id)
			if err != nil {
				config.GetLogger().Error("error while removing permission resource for syncPermResource", "error", err, "topic", topic, "id", id)
				return err
			}
		}
	}
	return nil
}

func SyncPermResources(ctx context.Context, config configuration.Config, permClient client.Client, db database.Database) error {
	devicesKnownInPermissions := []string{}
	err := deleteUnknownPermissions(ctx, config, permClient, config.DeviceTopic, func(id string) (bool, error) {
		_, exists, err := db.GetDevice(ctx, id)
		if err != nil {
			return false, err
		}
		devicesKnownInPermissions = append(devicesKnownInPermissions, id)
		return exists, nil
	})
	if err != nil {
		return err
	}
	err = db.DesyncUnknownDevices(ctx, devicesKnownInPermissions)
	if err != nil {
		return err
	}

	deviceGroupsKnownInPermissions := []string{}
	err = deleteUnknownPermissions(ctx, config, permClient, config.DeviceGroupTopic, func(id string) (bool, error) {
		_, exists, err := db.GetDeviceGroup(ctx, id)
		if err != nil {
			return false, err
		}
		deviceGroupsKnownInPermissions = append(deviceGroupsKnownInPermissions, id)
		return exists, nil
	})
	if err != nil {
		return err
	}
	err = db.DesyncUnknownDeviceGroups(ctx, deviceGroupsKnownInPermissions)
	if err != nil {
		return err
	}

	hubsKnownInPermissions := []string{}
	err = deleteUnknownPermissions(ctx, config, permClient, config.HubTopic, func(id string) (bool, error) {
		_, exists, err := db.GetHub(ctx, id)
		if err != nil {
			return false, err
		}
		hubsKnownInPermissions = append(hubsKnownInPermissions, id)
		return exists, nil
	})
	if err != nil {
		return err
	}
	err = db.DesyncUnknownHubs(ctx, hubsKnownInPermissions)
	if err != nil {
		return err
	}

	locationsKnownInPermissions := []string{}
	err = deleteUnknownPermissions(ctx, config, permClient, config.LocationTopic, func(id string) (bool, error) {
		_, exists, err := db.GetLocation(ctx, id)
		if err != nil {
			return false, err
		}
		locationsKnownInPermissions = append(locationsKnownInPermissions, id)
		return exists, nil
	})
	if err != nil {
		return err
	}
	err = db.DesyncUnknownLocations(ctx, locationsKnownInPermissions)
	if err != nil {
		return err
	}
	return nil
}
