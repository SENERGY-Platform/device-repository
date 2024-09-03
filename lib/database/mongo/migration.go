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
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"sync"
	"time"
)

func (this *Mongo) RunStartupMigrations() error {
	if !this.config.RunStartupMigrations {
		log.Println("INFO: skip startup migration because config.RunStartupMigrations=false")
		return nil
	}
	err := this.runPermissionsV2Migration()
	if err != nil {
		return err
	}
	return nil
}

func (this *Mongo) runPermissionsV2Migration() (err error) {
	if this.config.PermissionsV2Url == "" || this.config.PermissionsV2Url == "-" {
		log.Println("skip permissions-v2 migration because PermissionsV2Url is not configured")
		return nil
	}
	log.Println("start permissions-v2 migration")
	c := client.New(this.config.PermissionsV2Url)
	topics := []string{this.config.DeviceTopic, this.config.DeviceGroupTopic, this.config.HubTopic}

	workerChan := make(chan RightsEntry, 100)
	wg := &sync.WaitGroup{}
	mux := sync.Mutex{}

	deviceKind, _ := this.getInternalKind(this.config.DeviceTopic)
	deviceGroupKind, _ := this.getInternalKind(this.config.DeviceGroupTopic)
	hubKind, _ := this.getInternalKind(this.config.HubTopic)

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range workerChan {
				topic := this.config.DeviceTopic
				if entry.Kind == deviceKind {
					topic = this.config.DeviceTopic
				}
				if entry.Kind == deviceGroupKind {
					topic = this.config.DeviceGroupTopic
				}
				if entry.Kind == hubKind {
					topic = this.config.HubTopic
				}
				temp := entry.ToResourceRights()
				_, temperr, _ := c.SetPermission(client.InternalAdminToken, topic, entry.Id, temp.ToPermV2Permissions())
				if temperr != nil {
					mux.Lock()
					defer mux.Unlock()
					err = errors.Join(err, temperr)
					return
				}
			}
		}()
	}

	for _, topic := range topics {
		log.Printf("start permissions-v2 %v migration", topic)
		_, err, code := c.GetTopic(client.InternalAdminToken, topic)
		if code == http.StatusOK {
			return nil
		}
		if err != nil && code != http.StatusNotFound {
			return err
		}
		_, err, _ = c.SetTopic(client.InternalAdminToken, client.Topic{Id: topic})
		if err != nil {
			return err
		}

		kind, err := this.getInternalKind(topic)
		if err != nil {
			return err
		}
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		cursor, err := this.rightsCollection().Find(ctx, bson.M{"kind": kind})
		if err != nil {
			return err
		}
		for cursor.Next(context.Background()) {
			entry := RightsEntry{}
			err = cursor.Decode(&entry)
			if err != nil {
				return err
			}
			workerChan <- entry
		}
		err = cursor.Err()
		if err != nil {
			return err
		}
	}
	close(workerChan)
	wg.Wait()
	return err
}
