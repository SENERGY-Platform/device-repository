/*
 * Copyright 2022 InfAI (CC SES)
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

package testutils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/source/producer"
	"github.com/SENERGY-Platform/device-repository/lib/source/util"
	"github.com/segmentio/kafka-go"
	"log"
	"runtime/debug"
	"time"
)

type Publisher struct {
	config       config.Config
	devicetypes  *kafka.Writer
	protocols    *kafka.Writer
	devices      *kafka.Writer
	devicegroups *kafka.Writer
	hubs         *kafka.Writer
	aspects      *kafka.Writer
	functions    *kafka.Writer
}

func NewPublisher(conf config.Config) (*Publisher, error) {
	broker, err := util.GetBroker(conf.KafkaUrl)
	if err != nil {
		return nil, err
	}
	if len(broker) == 0 {
		return nil, errors.New("missing kafka broker")
	}
	devicetypes, err := producer.GetKafkaWriter(broker, conf.DeviceTypeTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	devices, err := producer.GetKafkaWriter(broker, conf.DeviceTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	devicegroups, err := producer.GetKafkaWriter(broker, conf.DeviceGroupTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	protocols, err := producer.GetKafkaWriter(broker, conf.ProtocolTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	hubs, err := producer.GetKafkaWriter(broker, conf.HubTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	aspects, err := producer.GetKafkaWriter(broker, conf.AspectTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	functions, err := producer.GetKafkaWriter(broker, conf.FunctionTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	return &Publisher{config: conf, devicetypes: devicetypes, protocols: protocols, devices: devices, hubs: hubs, devicegroups: devicegroups, aspects: aspects, functions: functions}, nil
}

type DeviceTypeCommand struct {
	Command    string           `json:"command"`
	Id         string           `json:"id"`
	Owner      string           `json:"owner"`
	DeviceType model.DeviceType `json:"device_type"`
}

type DeviceGroupCommand struct {
	Command     string            `json:"command"`
	Id          string            `json:"id"`
	Owner       string            `json:"owner"`
	DeviceGroup model.DeviceGroup `json:"device_group"`
}

type DeviceCommand struct {
	Command string       `json:"command"`
	Id      string       `json:"id"`
	Owner   string       `json:"owner"`
	Device  model.Device `json:"device"`
}

func (this *Publisher) PublishDeviceGroup(dg model.DeviceGroup, userId string) (err error) {
	cmd := DeviceGroupCommand{Command: "PUT", Id: dg.Id, DeviceGroup: dg, Owner: userId}
	return this.PublishDeviceGroupCommand(cmd)
}

func (this *Publisher) PublishDeviceGroupCommand(cmd DeviceGroupCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce", cmd)
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

func (this *Publisher) PublishDeviceType(device model.DeviceType, userId string) (err error) {
	cmd := DeviceTypeCommand{Command: "PUT", Id: device.Id, DeviceType: device, Owner: userId}
	return this.PublishDeviceTypeCommand(cmd)
}

func (this *Publisher) PublishDeviceTypeDelete(id string, userId string) error {
	cmd := DeviceTypeCommand{Command: "DELETE", Id: id, Owner: userId}
	return this.PublishDeviceTypeCommand(cmd)
}

func (this *Publisher) PublishDeviceTypeCommand(cmd DeviceTypeCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.devicetypes.WriteMessages(
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

func (this *Publisher) PublishDevice(device model.Device, userId string) (err error) {
	cmd := DeviceCommand{Command: "PUT", Id: device.Id, Device: device, Owner: userId}
	return this.PublishDeviceCommand(cmd)
}

func (this *Publisher) PublishDeviceCommand(cmd DeviceCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce", cmd)
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

type ProtocolCommand struct {
	Command  string         `json:"command"`
	Id       string         `json:"id"`
	Owner    string         `json:"owner"`
	Protocol model.Protocol `json:"protocol"`
}

func (this *Publisher) PublishProtocol(device model.Protocol, userId string) (err error) {
	cmd := ProtocolCommand{Command: "PUT", Id: device.Id, Protocol: device, Owner: userId}
	return this.PublishProtocolCommand(cmd)
}

func (this *Publisher) PublishProtocolDelete(id string, userId string) error {
	cmd := ProtocolCommand{Command: "DELETE", Id: id, Owner: userId}
	return this.PublishProtocolCommand(cmd)
}

func (this *Publisher) PublishProtocolCommand(cmd ProtocolCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.protocols.WriteMessages(
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

type HubCommand struct {
	Command string    `json:"command"`
	Id      string    `json:"id"`
	Owner   string    `json:"owner"`
	Hub     model.Hub `json:"hub"`
}

func (this *Publisher) PublishHub(hub model.Hub, userId string) (err error) {
	cmd := HubCommand{Command: "PUT", Id: hub.Id, Hub: hub, Owner: userId}
	return this.PublishHubCommand(cmd)
}

func (this *Publisher) PublishHubDelete(id string, userId string) error {
	cmd := HubCommand{Command: "DELETE", Id: id, Owner: userId}
	return this.PublishHubCommand(cmd)
}

func (this *Publisher) PublishHubCommand(cmd HubCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce hub", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.hubs.WriteMessages(
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

func (this *Publisher) PublishAspect(aspect model.Aspect, userid string) error {
	cmd := AspectCommand{Command: "PUT", Id: aspect.Id, Aspect: aspect, Owner: userid}
	return this.PublishAspectCommand(cmd)
}

func (this *Publisher) PublishAspectCommand(cmd AspectCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce hub", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.aspects.WriteMessages(
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

type AspectCommand struct {
	Command string       `json:"command"`
	Id      string       `json:"id"`
	Owner   string       `json:"owner"`
	Aspect  model.Aspect `json:"aspect"`
}

func (this *Publisher) PublishFunction(function model.Function, userid string) error {
	cmd := FunctionCommand{Command: "PUT", Id: function.Id, Function: function, Owner: userid}
	return this.PublishFunctionCommand(cmd)
}

func (this *Publisher) PublishFunctionCommand(cmd FunctionCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce hub", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.functions.WriteMessages(
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

type FunctionCommand struct {
	Command  string         `json:"command"`
	Id       string         `json:"id"`
	Owner    string         `json:"owner"`
	Function model.Function `json:"function"`
}
