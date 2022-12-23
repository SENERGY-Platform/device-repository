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
	"github.com/SENERGY-Platform/device-repository/lib/source/consumer/listener"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
)

type Producer struct {
	config   config.Config
	listener map[string]listener.Listener
}

func (this *Producer) PublishDeviceDelete(id string, owner string) error {
	//TODO implement me
	panic("implement me")
}

func (this *Producer) PublishHub(hub models.Hub) (err error) {
	//TODO implement me
	panic("implement me")
}

func New(conf config.Config) *Producer {
	return &Producer{config: conf, listener: map[string]listener.Listener{}}
}

func (this *Producer) Handle(topic string, handler listener.Listener) {
	this.listener[topic] = handler
}

func (this *Producer) callListener(topic string, message []byte) error {
	handler, ok := this.listener[topic]
	if !ok {
		return errors.New("unknown topic:" + topic)
	}
	err := handler(message)
	if err != nil {
		log.Println("TEST-WARNING:", err)
	}
	return err
}

func StartSourceMock(config config.Config, control listener.Controller) (prod *Producer, err error) {
	prod = New(config)

	for _, factory := range listener.Factories {
		topic, handler, err := factory(config, control)
		if err != nil {
			log.Println("ERROR: listener.factory", topic, err)
			return prod, err
		}
		prod.Handle(topic, handler)
	}
	return prod, err
}
