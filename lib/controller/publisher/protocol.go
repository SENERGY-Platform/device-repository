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

type ProtocolCommand struct {
	Command  string          `json:"command"`
	Id       string          `json:"id"`
	Protocol models.Protocol `json:"protocol"`

	//field has been removed but can still exist as value in kafka
	//StrictWaitBeforeDone bool          `json:"strict_wait_before_done"`
}

func (this *Publisher) PublishProtocol(protocol models.Protocol) (err error) {
	cmd := ProtocolCommand{Command: "PUT", Id: protocol.Id, Protocol: protocol}
	return this.PublishProtocolCommand(cmd)
}

func (this *Publisher) PublishProtocolDelete(id string) error {
	cmd := ProtocolCommand{Command: "DELETE", Id: id}
	return this.PublishProtocolCommand(cmd)
}

func (this *Publisher) PublishProtocolCommand(cmd ProtocolCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce protocol", cmd)
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
