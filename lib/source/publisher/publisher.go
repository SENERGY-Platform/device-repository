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
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/source/messages"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"github.com/SmartEnergyPlatform/amqp-wrapper-lib"
	"log"
)

type Publisher struct {
	conn   *amqp_wrapper_lib.Connection
	config config.Config
}

func New(conn *amqp_wrapper_lib.Connection, config config.Config) (*Publisher, error) {
	return &Publisher{conn: conn, config: config}, nil
}

func NewMute(ignored *amqp_wrapper_lib.Connection, config config.Config) (*Publisher, error) {
	return &Publisher{config: config}, nil
}

func (this *Publisher) PublishDevice(device model.DeviceInstance, owner string) error {
	if this.conn == nil {
		log.Println("WARNING: use mute publisher to publish", device)
		return nil
	}
	msg, err := json.Marshal(messages.DeviceinstanceCommand{DeviceInstance: device, Id: device.Id, Command: "PUT", Owner: owner})
	if err != nil {
		return err
	}
	return this.conn.Publish(this.config.DeviceInstanceTopic, msg)
}

func (this *Publisher) PublishHub(hub model.Hub, owner string) error {
	if this.conn == nil {
		log.Println("WARNING: use mute publisher to publish", hub)
		return nil
	}
	msg, err := json.Marshal(messages.GatewayCommand{Command: "PUT", Id: hub.Id, Name: hub.Name, Hash: hub.Hash, Owner: owner, Devices: hub.Devices})
	if err != nil {
		return err
	}
	return this.conn.Publish(this.config.HubTopic, msg)
}

func (this *Publisher) PublishValueType(valueType model.ValueType, owner string) error {
	if valueType.Id == "" {
		log.Println("WARNING: missing id in valuetype --> no publish")
		return nil
	}
	if this.conn == nil {
		log.Println("WARNING: use mute publisher to publish", valueType)
		return nil
	}
	msg, err := json.Marshal(messages.ValueTypeCommand{ValueType: valueType, Id: valueType.Id, Command: "PUT", Owner: owner})
	if err != nil {
		return err
	}
	return this.conn.Publish(this.config.ValueTypeTopic, msg)
}
