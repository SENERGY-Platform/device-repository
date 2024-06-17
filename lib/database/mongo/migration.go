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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
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
			bson.M{hubOwnerIdKey: bson.M{"$exists": false}},
			bson.M{hubOwnerIdKey: ""},
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
		element := models.Hub{}
		err = cursor.Decode(&element)
		if err != nil {
			return err
		}
		if element.OwnerId != "" {
			panic("owner must be empty because we searched for hubs without owner")
		}

		ownerCount := map[string]int{}
		foundDeviceWithoutOwner := false
		devices := []models.Device{}
		for _, deviceId := range element.DeviceIds {
			device, exists, err := this.GetDevice(context.Background(), deviceId)
			if err != nil {
				return err
			}
			if exists && device.OwnerId != "" {
				if device.OwnerId != "" {
					ownerCount[device.OwnerId] = ownerCount[device.OwnerId] + 1
				} else {
					foundDeviceWithoutOwner = true
				}
			}
		}
		if foundDeviceWithoutOwner {
			log.Println("WARNING: unable to find the owner of at least one device of the hub", element.Name, element.Id, "--> skip migration")
			continue
		}
		majorityOwnerCount := 0
		for owner, count := range ownerCount {
			if count > majorityOwnerCount {
				majorityOwnerCount = count
				element.OwnerId = owner
			}
		}
		if len(ownerCount) > 1 {
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
		//may only happen if the hub has no devices (or if not all devices have an owner, but with assertEveryDeviceHasOwner() all devices must have an owner)
		if element.OwnerId == "" {
			rights, err := this.getRights(hubKind, element.Id)
			if err != nil {
				return err
			}
			if len(rights.AdminUsers) > 0 {
				element.OwnerId = rights.AdminUsers[0]
			}
		}

		if element.OwnerId == "" {
			log.Printf("WARNING: no owner for hub %v (%v) found\n", element.Name, element.Id)
		} else {
			log.Println("update hub owner", element.Id, element.OwnerId)
			err = producer.PublishHub(element)
			if err != nil {
				log.Println("ERROR: unable to update hub owner", element.Id, element.OwnerId, err)
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
			bson.M{deviceOwnerIdKey: bson.M{"$exists": false}},
			bson.M{deviceOwnerIdKey: ""},
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
		element := models.Device{}
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
		if len(rights.AdminUsers) > 0 {
			element.OwnerId = rights.AdminUsers[0]
		}

		if element.OwnerId == "" {
			log.Printf("WARNING: no owner for device %v (%v) found\n", element.Name, element.Id)
		} else {
			//TODO: ensure no breaking of local-id constraints
			log.Println("update hub owner", element.Id, element.OwnerId)

			//publish so that other services know the new owner immediately
			err = producer.PublishDevice(element)
			if err != nil {
				log.Println("ERROR: unable to update device owner", element.Id, element.OwnerId, err)
				return err
			}

			//locally, so that hubs can check device owners
			ctx, _ := getTimeoutContext()
			log.Println("update device owner", element.Id, element.OwnerId)
			err = this.SetDevice(ctx, element)
			if err != nil {
				return err
			}
		}

	}
	return cursor.Err()
}

func (this *Mongo) assertEveryDeviceHasOwner() error {
	count, err := this.deviceCollection().CountDocuments(context.Background(), bson.M{
		"$or": bson.A{
			bson.M{deviceOwnerIdKey: bson.M{"$exists": false}},
			bson.M{deviceOwnerIdKey: ""},
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

func (this *Mongo) hubOwnerMigrationEnforceDeviceOwner(producer model.MigrationPublisher, device models.Device, owner string) error {
	//TODO: ensure no breaking of local-id constraints
	if owner == "" {
		return errors.New("missing owner")
	}
	if device.Id == "" {
		return errors.New("invalid device")
	}
	if device.OwnerId == owner {
		return nil
	}
	log.Printf("force owner %v on device %v %v\n", owner, device.Name, device.Id)

	deviceKind, err := this.getInternalKind(this.config.DeviceTopic)
	if err != nil {
		return err
	}

	ctx, _ := getTimeoutContext()
	rights, err := this.getRights(deviceKind, device.Id)
	if err != nil {
		return err
	}
	if !slices.Contains(rights.AdminUsers, owner) {
		log.Printf("device %v %v may currently not use owner %v because %v is not an admin --> add %v as admin", device.Name, device.Id, owner, owner, owner)
		rights.AdminUsers = append(rights.AdminUsers, owner)
		resourceRights := rights.ToResourceRights()
		err = producer.PublishDeviceRights(device.Id, owner, resourceRights)
		if err != nil {
			return err
		}
		err = this.SetRights(this.config.DeviceTopic, device.Id, resourceRights)
		if err != nil {
			return err
		}
	}

	ctx, _ = getTimeoutContext()
	device.OwnerId = owner
	err = this.SetDevice(ctx, device)
	if err != nil {
		return err
	}
	return nil
}
