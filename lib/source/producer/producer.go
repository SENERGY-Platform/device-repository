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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/source/util"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"os"
	"strings"
)

type Producer struct {
	config       config.Config
	devices      *kafka.Writer
	deviceGroups *kafka.Writer
	hubs         *kafka.Writer
	aspects      *kafka.Writer
	done         *kafka.Writer
}

func New(conf config.Config) (*Producer, error) {
	devices, err := GetKafkaWriter(conf.KafkaUrl, conf.DeviceTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	devicesGroups, err := GetKafkaWriter(conf.KafkaUrl, conf.DeviceGroupTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	aspects, err := GetKafkaWriter(conf.KafkaUrl, conf.AspectTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	hubs, err := GetKafkaWriter(conf.KafkaUrl, conf.HubTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	err = util.InitTopic(conf.KafkaUrl, conf.DoneTopic)
	if err != nil {
		return nil, err
	}
	done, err := GetKafkaWriter(conf.KafkaUrl, conf.DoneTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	return &Producer{config: conf, devices: devices, deviceGroups: devicesGroups, hubs: hubs, aspects: aspects, done: done}, nil
}

func GetKafkaWriter(broker string, topic string, debug bool) (writer *kafka.Writer, err error) {
	var logger *log.Logger
	if debug {
		logger = log.New(os.Stdout, "[KAFKA-PRODUCER] ", 0)
	} else {
		logger = log.New(io.Discard, "", 0)
	}
	writer = &kafka.Writer{
		Addr:        kafka.TCP(broker),
		Topic:       topic,
		MaxAttempts: 10,
		Logger:      logger,
		BatchSize:   1,
		Balancer:    &KeySeparationBalancer{SubBalancer: &kafka.Hash{}, Seperator: "/"},
	}
	return writer, err
}

type KeySeparationBalancer struct {
	SubBalancer kafka.Balancer
	Seperator   string
}

func (this *KeySeparationBalancer) Balance(msg kafka.Message, partitions ...int) (partition int) {
	key := string(msg.Key)
	if this.Seperator != "" {
		keyParts := strings.Split(key, this.Seperator)
		key = keyParts[0]
	}
	msg.Key = []byte(key)
	return this.SubBalancer.Balance(msg, partitions...)
}
