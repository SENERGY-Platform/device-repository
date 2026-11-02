/*
 * Copyright 2025 InfAI (CC SES)
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

package controller

import (
	"context"
	"errors"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
)

func (this *Controller) StartSyncLoop(ctx context.Context, interval time.Duration, lockduration time.Duration) {
	err := this.Sync(ctx, lockduration)
	if err != nil {
		this.config.GetLogger().Error("error while db sync run", "error", err)
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := this.Sync(ctx, lockduration)
				if err != nil {
					this.config.GetLogger().Error("error while db sync run", "error", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (this *Controller) Sync(ctx context.Context, lockduration time.Duration) (err error) {
	err = errors.Join(err, this.db.RetryDeviceSync(lockduration, withContext(ctx, this.deleteDeviceSyncHandler), func(state model.DeviceWithConnectionState) error {
		return this.setDeviceSyncHandler(ctx, model.DeviceWithConnectionState{}, state)
	}))
	err = errors.Join(err, this.db.RetryAspectSync(lockduration, withContext(ctx, this.deleteAspectSyncHandler), withContext(ctx, this.setAspectSyncHandler)))
	err = errors.Join(err, this.db.RetryCharacteristicSync(lockduration, withContext(ctx, this.deleteCharacteristicSyncHandler), withContext(ctx, this.setCharacteristicSyncHandler)))
	err = errors.Join(err, this.db.RetryConceptSync(lockduration, withContext(ctx, this.deleteConceptSyncHandler), withContext(ctx, this.setConceptSyncHandler)))
	err = errors.Join(err, this.db.RetryAspectClassSync(lockduration, withContext(ctx, this.deleteAspectClassSyncHandler), withContext(ctx, this.setAspectClassSyncHandler)))
	err = errors.Join(err, this.db.RetryDeviceClassSync(lockduration, withContext(ctx, this.deleteDeviceClassSyncHandler), withContext(ctx, this.setDeviceClassSyncHandler)))
	err = errors.Join(err, this.db.RetryDeviceGroupSync(lockduration, withContext(ctx, this.deleteDeviceGroupSyncHandler), withContext2(ctx, this.setDeviceGroupSyncHandler)))
	err = errors.Join(err, this.db.RetryDeviceTypeSync(lockduration, withContext(ctx, this.deleteDeviceTypeSyncHandler), withContext(ctx, this.setDeviceTypeSyncHandler)))
	err = errors.Join(err, this.db.RetryFunctionSync(lockduration, withContext(ctx, this.deleteFunctionSyncHandler), withContext(ctx, this.setFunctionSyncHandler)))
	err = errors.Join(err, this.db.RetryGraphSync(lockduration, withContext(ctx, this.deleteGraphSyncHandler), withContext(ctx, this.setGraphSyncHandler)))
	err = errors.Join(err, this.db.RetryHubSync(lockduration, withContext(ctx, this.deleteHubSyncHandler), withContext(ctx, this.setHubSyncHandler)))
	err = errors.Join(err, this.db.RetryLocationSync(lockduration, withContext(ctx, this.deleteLocationSyncHandler), withContext2(ctx, this.setLocationSyncHandler)))
	err = errors.Join(err, this.db.RetryProtocolSync(lockduration, withContext(ctx, this.deleteProtocolSyncHandler), withContext(ctx, this.setProtocolSyncHandler)))
	return err
}
