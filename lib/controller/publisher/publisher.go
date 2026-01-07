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
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/segmentio/kafka-go"
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
	if conf.InitTopics {
		conf.GetLogger().Info("ensure kafka topics")
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
	}
	conf.GetLogger().Info(fmt.Sprintf("produce to: %#v", []string{conf.DeviceTypeTopic, conf.ProtocolTopic, conf.DeviceTopic, conf.HubTopic, conf.ConceptTopic, conf.CharacteristicTopic, conf.LocationTopic}))
	devicetypes := getProducer(ctx, conf.KafkaUrl, conf.DeviceTypeTopic, conf.GetLogger())
	devicegroups := getProducer(ctx, conf.KafkaUrl, conf.DeviceGroupTopic, conf.GetLogger())
	devices := getProducer(ctx, conf.KafkaUrl, conf.DeviceTopic, conf.GetLogger())
	hubs := getProducer(ctx, conf.KafkaUrl, conf.HubTopic, conf.GetLogger())
	protocol := getProducer(ctx, conf.KafkaUrl, conf.ProtocolTopic, conf.GetLogger())
	concepts := getProducer(ctx, conf.KafkaUrl, conf.ConceptTopic, conf.GetLogger())
	characteristics := getProducer(ctx, conf.KafkaUrl, conf.CharacteristicTopic, conf.GetLogger())
	aspect := getProducer(ctx, conf.KafkaUrl, conf.AspectTopic, conf.GetLogger())
	function := getProducer(ctx, conf.KafkaUrl, conf.FunctionTopic, conf.GetLogger())
	deviceclass := getProducer(ctx, conf.KafkaUrl, conf.DeviceClassTopic, conf.GetLogger())
	location := getProducer(ctx, conf.KafkaUrl, conf.LocationTopic, conf.GetLogger())
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

func getProducer(ctx context.Context, broker string, topic string, logger *slog.Logger) (writer *kafka.Writer) {
	kafkaLogger := slog.NewLogLogger(logger.Handler(), slog.LevelDebug)
	kafkaLogger.SetPrefix("[KAFKA-PRODUCER] ")
	writer = &kafka.Writer{
		Addr:        kafka.TCP(broker),
		Topic:       topic,
		MaxAttempts: 10,
		Logger:      kafkaLogger,
		BatchSize:   1,
		Balancer:    &KeySeparationBalancer{SubBalancer: &kafka.Hash{}, Seperator: "/"},
		Compression: kafka.Snappy,
	}
	go func() {
		<-ctx.Done()
		err := writer.Close()
		if err != nil {
			logger.Error("ERROR: unable to close producer", "topic", topic, "error", err)
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
