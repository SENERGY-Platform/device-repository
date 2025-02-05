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
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"runtime/debug"
)

func (this *Mongo) RunStartupMigrations(helper GeneratedDeviceGroupMigrationMethods) error {
	if !this.config.RunStartupMigrations {
		log.Println("INFO: skip startup migration because config.RunStartupMigrations=false")
		return nil
	}
	err := this.runDeviceGroupMigration(helper)
	if err != nil {
		return err
	}
	return nil
}

type GeneratedDeviceGroupMigrationMethods interface {
	DeviceIdToGeneratedDeviceGroupId(deviceId string) string
	EnsureGeneratedDeviceGroup(device models.Device) (err error)
}

func (this *Mongo) runDeviceGroupMigration(helper GeneratedDeviceGroupMigrationMethods) error {
	log.Println("start runDeviceGroupMigration()")
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
			log.Printf("generate device-group for %v %v\n", device.Id, device.Name)
			err = helper.EnsureGeneratedDeviceGroup(device)
			if err != nil {
				debug.PrintStack()
				return err
			}
		}
	}
	return nil
}
