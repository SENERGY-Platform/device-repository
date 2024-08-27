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
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"slices"
)

func (this *Mongo) RunStartupMigrations(producer model.MigrationPublisher) error {
	if !this.config.RunStartupMigrations {
		log.Println("INFO: skip startup migration because config.RunStartupMigrations=false")
		return nil
	}
	if this.config.SecurityImpl != config.DbSecurity {
		log.Println("WARNING: migration only with internal security (security_impl=db) supported")
		return nil
	}
	err := this.runDeviceOwnerMigration(producer)
	if err != nil {
		return err
	}
	return this.runHubOwnerMigration(producer)
}

func (this *Mongo) runHubOwnerMigration(producer model.MigrationPublisher) error {
	if producer == nil {
		log.Println("WARNING: skip hub-owner-id migration because producer is nil")
		return nil
	}
	log.Println("start hub-owner-id migration")
	defer log.Println("end hub-owner-id migration")
	err := this.assertEveryDeviceHasOwner()
	if err != nil {
		return err
	}

	cursor, err := this.hubCollection().Find(context.Background(), bson.M{
		"$or": bson.A{
			bson.M{HubBson.OwnerId: bson.M{"$exists": false}},
			bson.M{HubBson.OwnerId: ""},
		},
	})
	if err != nil {
		return err
	}
	hubKind, err := this.getInternalKind(this.config.HubTopic)
	if err != nil {
		return err
	}
	for cursor.Next(context.Background()) {
		element := model.HubWithConnectionState{}
		err = cursor.Decode(&element)
		if err != nil {
			return err
		}
		if element.OwnerId != "" {
			panic("owner must be empty because we searched for hubs without owner")
		}

		ownerCount := map[string]int{}
		devices := []model.DeviceWithConnectionState{}
		for _, deviceId := range element.DeviceIds {
			device, exists, err := this.GetDevice(context.Background(), deviceId)
			if err != nil {
				return err
			}
			devices = append(devices, device)
			if exists && device.OwnerId != "" {
				if device.OwnerId != "" {
					ownerCount[device.OwnerId] = ownerCount[device.OwnerId] + 1
				} else {
					return errors.New("expect all hub devices to have owner")
				}
			}
		}

		rights, err := this.getRights(hubKind, element.Id)
		if err != nil {
			return err
		}
		if len(rights.AdminUsers) == 0 {
			return fmt.Errorf("no admin users found for hub %v", element.Id)
		}

		getMajorityOwnerInAdmin := func(ownerCount map[string]int, admins []string) (majorityOwner string) {
			majorityOwnerCount := 0
			for owner, count := range ownerCount {
				if count > majorityOwnerCount && slices.Contains(admins, owner) {
					majorityOwnerCount = count
					majorityOwner = owner
				}
			}
			return majorityOwner
		}

		element.OwnerId = getMajorityOwnerInAdmin(ownerCount, rights.AdminUsers)
		useMajorityOwner := true
		if element.OwnerId == "" && len(rights.AdminUsers) > 0 {
			element.OwnerId = rights.AdminUsers[0]
			useMajorityOwner = false
		}

		if len(ownerCount) > 1 || !useMajorityOwner {
			log.Printf("WARNING: hub %v (%v) contains devices with multiple different owners.\ndevices-ids=%v\nthe majority owner %v will be used for the hub and as new owner for all devices of the hub", element.Name, element.Id, element.DeviceIds, element.OwnerId)
			for _, device := range devices {
				if device.OwnerId != element.OwnerId {
					err = this.hubOwnerMigrationEnforceDeviceOwner(producer, device, element.OwnerId)
					if err != nil {
						return err
					}
				}
			}
		}

		if element.OwnerId == "" {
			log.Printf("WARNING: no owner for hub %v (%v) found\n", element.Name, element.Id)
		} else {
			log.Println("update hub owner", element.Id, element.OwnerId)
			err = producer.PublishHub(element.Hub, element.OwnerId)
			if err != nil {
				log.Println("ERROR: unable to update hub owner", element.Id, element.OwnerId, err)
				return err
			}
			//locally, so that hubs can check device owners
			ctx, _ := getTimeoutContext()
			err = this.SetHub(ctx, element)
			if err != nil {
				return err
			}
		}
	}
	return cursor.Err()
}

func (this *Mongo) runDeviceOwnerMigration(producer model.MigrationPublisher) error {
	if producer == nil {
		log.Println("WARNING: skip device-owner-id migration because producer is nil")
		return nil
	}
	log.Println("start device-owner-id migration")
	defer log.Println("end device-owner-id migration")

	cursor, err := this.deviceCollection().Find(context.Background(), bson.M{
		"$or": bson.A{
			bson.M{DeviceBson.OwnerId: bson.M{"$exists": false}},
			bson.M{DeviceBson.OwnerId: ""},
		},
	})
	if err != nil {
		return err
	}
	deviceKind, err := this.getInternalKind(this.config.DeviceTopic)
	if err != nil {
		return err
	}
	for cursor.Next(context.Background()) {
		element := model.DeviceWithConnectionState{}
		err = cursor.Decode(&element)
		if err != nil {
			return err
		}
		if element.OwnerId != "" {
			panic("owner must be empty because we searched for devices without owner")
		}
		rights, err := this.getRights(deviceKind, element.Id)
		if err != nil {
			return err
		}
		if len(rights.AdminUsers) == 0 {
			return fmt.Errorf("no admin users found for device %v", element.Id)
		}
		if len(rights.AdminUsers) > 0 {
			element.OwnerId = rights.AdminUsers[0]
		}

		if element.OwnerId == "" {
			log.Printf("WARNING: no owner for device %v (%v) found\n", element.Name, element.Id)
		} else {
			log.Println("update device owner", element.Id, element.OwnerId)

			//check local-id constraints
			ctx, _ := getTimeoutContext()
			existing, found, err := this.GetDeviceByLocalId(ctx, element.OwnerId, element.LocalId)
			if err != nil {
				log.Println("ERROR: unable to check local-id constraint", element.Id, element.OwnerId, err)
				return err
			}
			if found && existing.Id != element.Id {
				return fmt.Errorf("unable to migrate: new device owner breaks local-id constraint (device-to-change=%v(%v), existing-device=%v(%v), owner=%v)", element.Name, element.Id, existing.Name, existing.Id, element.OwnerId)
			}

			//publish so that other services know the new owner immediately
			err = producer.PublishDevice(element.Device, element.OwnerId)
			if err != nil {
				log.Println("ERROR: unable to update device owner", element.Id, element.OwnerId, err)
				return err
			}

			//locally, so that hubs can check device owners
			ctx, _ = getTimeoutContext()
			err = this.SetDevice(ctx, element)
			if err != nil {
				return err
			}
		}

	}
	return cursor.Err()
}

func (this *Mongo) hubOwnerMigrationEnforceDeviceOwner(producer model.MigrationPublisher, device model.DeviceWithConnectionState, owner string) error {
	if owner == "" {
		return errors.New("missing owner")
	}
	if device.Id == "" {
		return errors.New("invalid device")
	}
	if device.OwnerId == owner {
		log.Println("")
		return nil
	}
	log.Printf("force owner %v on device %v %v\n", owner, device.Name, device.Id)

	device.OwnerId = owner

	//check local-id constraints
	ctx, _ := getTimeoutContext()
	existing, found, err := this.GetDeviceByLocalId(ctx, device.OwnerId, device.LocalId)
	if err != nil {
		log.Println("ERROR: unable to check local-id constraint", device.Id, device.OwnerId, err)
		return err
	}
	if found && existing.Id != device.Id {
		return fmt.Errorf("unable to migrate: new device owner breaks local-id constraint (device-to-change=%v(%v), existing-device=%v(%v), owner=%v)", device.Name, device.Id, existing.Name, existing.Id, device.OwnerId)
	}

	deviceKind, err := this.getInternalKind(this.config.DeviceTopic)
	if err != nil {
		return err
	}

	ctx, _ = getTimeoutContext()
	rights, err := this.getRights(deviceKind, device.Id)
	if err != nil {
		return err
	}

	//update admin rights if necessary
	if !slices.Contains(rights.AdminUsers, owner) {
		log.Printf("device %v %v may currently not use owner %v because %v is not an admin --> add %v as admin", device.Name, device.Id, owner, owner, owner)
		resourceRights := rights.ToResourceRights()
		resourceRights.UserRights[owner] = model.Right{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		}
		err = producer.PublishDeviceRights(device.Id, owner, resourceRights)
		if err != nil {
			return err
		}
		err = this.SetRights(this.config.DeviceTopic, device.Id, resourceRights)
		if err != nil {
			return err
		}
	}

	//publish so that other services know the new owner immediately
	err = producer.PublishDevice(device.Device, device.OwnerId)
	if err != nil {
		log.Println("ERROR: unable to update device owner", device.Id, device.OwnerId, err)
		return err
	}

	//update device owner
	ctx, _ = getTimeoutContext()
	err = this.SetDevice(ctx, device)
	if err != nil {
		return err
	}
	return nil
}

func (this *Mongo) assertEveryDeviceHasOwner() error {
	count, err := this.deviceCollection().CountDocuments(context.Background(), bson.M{
		"$or": bson.A{
			bson.M{DeviceBson.OwnerId: bson.M{"$exists": false}},
			bson.M{DeviceBson.OwnerId: ""},
		},
	}, options.Count().SetLimit(1))
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("assertEveryDeviceHasOwner() failed: found devicew without owner")
	}
	return nil
}
