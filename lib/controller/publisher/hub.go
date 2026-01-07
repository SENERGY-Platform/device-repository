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
	"runtime/debug"
	"time"

	"github.com/SENERGY-Platform/models/go/models"
	"github.com/segmentio/kafka-go"
)

type HubCommand struct {
	Command string     `json:"command"`
	Id      string     `json:"id"`
	Hub     models.Hub `json:"hub"`

	//field has been removed but can still exist as value in kafka
	//StrictWaitBeforeDone bool          `json:"strict_wait_before_done"`
}

func (this *Publisher) PublishHub(hub models.Hub) (err error) {
	cmd := HubCommand{Command: "PUT", Id: hub.Id, Hub: hub}
	return this.PublishHubCommand(cmd)
}

func (this *Publisher) PublishHubDelete(id string) error {
	cmd := HubCommand{Command: "DELETE", Id: id}
	return this.PublishHubCommand(cmd)
}

func (this *Publisher) PublishHubCommand(cmd HubCommand) error {
	this.config.GetLogger().Debug("publish hub command", "command", cmd)
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
