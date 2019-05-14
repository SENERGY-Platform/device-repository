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

package source

import (
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/source/listener"
	"github.com/SmartEnergyPlatform/amqp-wrapper-lib"
	"log"
)

func Start(config config.Config, control listener.Controller) (conn *amqp_wrapper_lib.Connection, err error) {
	topics := []string{}
	handlers := []amqp_wrapper_lib.ConsumerFunc{}
	for _, factory := range listener.Factories {
		topic, handler, err := factory(config, control)
		if err != nil {
			log.Println("ERROR: listener.factory", topic, err)
			return conn, err
		}
		topics = append(topics, topic)
		handlers = append(handlers, handler)
	}

	conn, err = amqp_wrapper_lib.Init(config.AmqpUrl, topics, config.AmqpReconnectTimeout)
	if err != nil {
		log.Fatal("ERROR: while initializing amqp connection", err)
		return
	}

	for i := 0; i < len(topics); i++ {
		err = conn.Consume(config.AmqpConsumerName+"_"+topics[i], topics[i], handlers[i])
		if err != nil {
			log.Println("ERROR: source.consume", topics[i], err)
			conn.Close()
			return conn, err
		}
	}
	return conn, err
}
