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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"log"
	"time"
)

func (this *Controller) StartSyncLoop(ctx context.Context, interval time.Duration, lockduration time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := this.Sync(lockduration)
				if err != nil {
					log.Printf("ERROR: while db sync run: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (this *Controller) Sync(lockduration time.Duration) (err error) {
	err = errors.Join(err, this.db.RetryDeviceSync(lockduration, this.deleteDeviceSyncHandler, func(state model.DeviceWithConnectionState) error {
		return this.setDeviceSyncHandler(model.DeviceWithConnectionState{}, state)
	}))
	err = errors.Join(err, this.db.RetryAspectSync(lockduration, this.deleteAspectSyncHandler, this.setAspectSyncHandler))
	err = errors.Join(err, this.db.RetryCharacteristicSync(lockduration, this.deleteCharacteristicSyncHandler, this.setCharacteristicSyncHandler))
	err = errors.Join(err, this.db.RetryConceptSync(lockduration, this.deleteConceptSyncHandler, this.setConceptSyncHandler))
	err = errors.Join(err, this.db.RetryDeviceClassSync(lockduration, this.deleteDeviceClassSyncHandler, this.setDeviceClassSyncHandler))
	err = errors.Join(err, this.db.RetryDeviceGroupSync(lockduration, this.deleteDeviceGroupSyncHandler, this.setDeviceGroupSyncHandler))
	err = errors.Join(err, this.db.RetryDeviceTypeSync(lockduration, this.deleteDeviceTypeSyncHandler, this.setDeviceTypeSyncHandler))
	err = errors.Join(err, this.db.RetryFunctionSync(lockduration, this.deleteFunctionSyncHandler, this.setFunctionSyncHandler))
	err = errors.Join(err, this.db.RetryHubSync(lockduration, this.deleteHubSyncHandler, this.setHubSyncHandler))
	err = errors.Join(err, this.db.RetryLocationSync(lockduration, this.deleteLocationSyncHandler, this.setLocationSyncHandler))
	err = errors.Join(err, this.db.RetryProtocolSync(lockduration, this.deleteProtocolSyncHandler, this.setProtocolSyncHandler))
	return err
}
