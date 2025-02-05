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

package publisher

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/segmentio/kafka-go"
	"log"
	"runtime/debug"
	"time"
)

type DeviceGroupCommand struct {
	Command     string             `json:"command"`
	Id          string             `json:"id"`
	DeviceGroup models.DeviceGroup `json:"device_group"`

	//field has been removed but can still exist as value in kafka
	//StrictWaitBeforeDone bool          `json:"strict_wait_before_done"`
}

func (this *Publisher) PublishDeviceGroup(group models.DeviceGroup) (err error) {
	cmd := DeviceGroupCommand{Command: "PUT", Id: group.Id, DeviceGroup: group}
	return this.PublishDeviceGroupCommand(cmd)
}

func (this *Publisher) PublishDeviceGroupDelete(id string) error {
	cmd := DeviceGroupCommand{Command: "DELETE", Id: id}
	return this.PublishDeviceGroupCommand(cmd)
}

func (this *Publisher) PublishDeviceGroupCommand(cmd DeviceGroupCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce deviceGroup", cmd)
	}

	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.devicegroups.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(cmd.Id),
			Value: message,
			Time:  time.Now(),
		},
	)
	if err != nil {
		debug.PrintStack()
	}
	return err
}
