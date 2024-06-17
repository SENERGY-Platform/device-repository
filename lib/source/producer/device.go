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

package producer

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/segmentio/kafka-go"
	"log"
	"runtime/debug"
	"time"
)

type DeviceCommand struct {
	Command string                `json:"command"`
	Id      string                `json:"id"`
	Owner   string                `json:"owner"`
	Device  models.Device         `json:"device"`
	Rights  *model.ResourceRights `json:"rights,omitempty"`
}

func (this *Producer) PublishDeviceDelete(id string, userId string) error {
	cmd := DeviceCommand{Command: "DELETE", Id: id, Owner: userId}
	return this.PublishDeviceCommand(cmd)
}

func (this *Producer) PublishDeviceRights(deviceId string, userId string, rights model.ResourceRights) (err error) {
	cmd := DeviceCommand{Command: "RIGHTS", Id: deviceId, Owner: userId, Rights: &rights}
	return this.PublishDeviceCommand(cmd)
}

func (this *Producer) PublishDevice(device models.Device) error {
	cmd := DeviceCommand{Command: "PUT", Id: device.Id, Device: device}
	return this.PublishDeviceCommand(cmd)
}

func (this *Producer) PublishDeviceCommand(cmd DeviceCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce device", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.devices.WriteMessages(
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
