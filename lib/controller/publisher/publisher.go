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
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Publisher struct {
	config          configuration.Config
	devicetypes     *kafka.Writer
	devicegroups    *kafka.Writer
	protocols       *kafka.Writer
	devices         *kafka.Writer
	hubs            *kafka.Writer
	concepts        *kafka.Writer
	characteristics *kafka.Writer
	aspects         *kafka.Writer
	functions       *kafka.Writer
	deviceclasses   *kafka.Writer
	locations       *kafka.Writer
}

func New(conf configuration.Config, ctx context.Context) (*Publisher, error) {
	log.Println("ensure kafka topics")
	err := InitTopic(
		conf.KafkaUrl,
		conf.DeviceTypeTopic,
		conf.DeviceGroupTopic,
		conf.ProtocolTopic,
		conf.DeviceTopic,
		conf.HubTopic,
		conf.ConceptTopic,
		conf.CharacteristicTopic,
		conf.AspectTopic,
		conf.FunctionTopic,
		conf.DeviceClassTopic,
		conf.LocationTopic)
	if err != nil {
		return nil, err
	}

	log.Println("Produce to ", conf.DeviceTypeTopic, conf.ProtocolTopic, conf.DeviceTopic, conf.HubTopic, conf.ConceptTopic, conf.CharacteristicTopic, conf.LocationTopic)
	devicetypes := getProducer(ctx, conf.KafkaUrl, conf.DeviceTypeTopic, conf.Debug)
	devicegroups := getProducer(ctx, conf.KafkaUrl, conf.DeviceGroupTopic, conf.Debug)
	devices := getProducer(ctx, conf.KafkaUrl, conf.DeviceTopic, conf.Debug)
	hubs := getProducer(ctx, conf.KafkaUrl, conf.HubTopic, conf.Debug)
	protocol := getProducer(ctx, conf.KafkaUrl, conf.ProtocolTopic, conf.Debug)
	concepts := getProducer(ctx, conf.KafkaUrl, conf.ConceptTopic, conf.Debug)
	characteristics := getProducer(ctx, conf.KafkaUrl, conf.CharacteristicTopic, conf.Debug)
	aspect := getProducer(ctx, conf.KafkaUrl, conf.AspectTopic, conf.Debug)
	function := getProducer(ctx, conf.KafkaUrl, conf.FunctionTopic, conf.Debug)
	deviceclass := getProducer(ctx, conf.KafkaUrl, conf.DeviceClassTopic, conf.Debug)
	location := getProducer(ctx, conf.KafkaUrl, conf.LocationTopic, conf.Debug)
	return &Publisher{
		config:          conf,
		devicetypes:     devicetypes,
		devicegroups:    devicegroups,
		protocols:       protocol,
		devices:         devices,
		hubs:            hubs,
		concepts:        concepts,
		characteristics: characteristics,
		aspects:         aspect,
		functions:       function,
		deviceclasses:   deviceclass,
		locations:       location,
	}, nil
}

func getProducer(ctx context.Context, broker string, topic string, debug bool) (writer *kafka.Writer) {
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
	go func() {
		<-ctx.Done()
		err := writer.Close()
		if err != nil {
			log.Println("ERROR: unable to close producer for", topic, err)
		}
	}()
	return writer
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

func InitTopic(bootstrapUrl string, topics ...string) (err error) {
	conn, err := kafka.Dial("tcp", bootstrapUrl)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{}

	for _, topic := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
			ConfigEntries: []kafka.ConfigEntry{
				{
					ConfigName:  "retention.ms",
					ConfigValue: "2592000000", // 30d
				},
			},
		})
	}

	return controllerConn.CreateTopics(topicConfigs...)
}
