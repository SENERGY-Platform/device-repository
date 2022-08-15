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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/source/util"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
	"os"
)

type Producer struct {
	config  config.Config
	devices *kafka.Writer
	hubs    *kafka.Writer
	aspects *kafka.Writer
}

func New(conf config.Config) (*Producer, error) {
	broker, err := util.GetBroker(conf.KafkaUrl)
	if err != nil {
		return nil, err
	}
	if len(broker) == 0 {
		return nil, errors.New("missing kafka broker")
	}
	devices, err := GetKafkaWriter(broker, conf.DeviceTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	aspects, err := GetKafkaWriter(broker, conf.AspectTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	hubs, err := GetKafkaWriter(broker, conf.HubTopic, conf.Debug)
	if err != nil {
		return nil, err
	}
	return &Producer{config: conf, devices: devices, hubs: hubs, aspects: aspects}, nil
}

func GetKafkaWriter(broker []string, topic string, debug bool) (writer *kafka.Writer, err error) {
	var logger *log.Logger
	if debug {
		logger = log.New(os.Stdout, "[KAFKA-PRODUCER] ", 0)
	} else {
		logger = log.New(ioutil.Discard, "", 0)
	}
	writer = &kafka.Writer{
		Addr:        kafka.TCP(broker...),
		Topic:       topic,
		MaxAttempts: 10,
		Logger:      logger,
		BatchSize:   1,
		Balancer:    &kafka.Hash{},
	}
	return writer, err
}
