/*
 * Copyright 2026 InfAI (CC SES)
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

type GraphCommand struct {
	Command string       `json:"command"`
	Id      string       `json:"id"`
	Graph   models.Graph `json:"graph"`
}

func (this *Publisher) PublishGraph(graph models.Graph) (err error) {
	cmd := GraphCommand{Command: "PUT", Id: graph.Id, Graph: graph}
	return this.PublishGraphCommand(cmd)
}

func (this *Publisher) PublishGraphDelete(id string) error {
	cmd := GraphCommand{Command: "DELETE", Id: id}
	return this.PublishGraphCommand(cmd)
}

func (this *Publisher) PublishGraphCommand(cmd GraphCommand) error {
	this.config.GetLogger().Debug("publish graph command", "command", cmd)
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.graphs.WriteMessages(
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
