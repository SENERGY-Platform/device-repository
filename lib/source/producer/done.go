/*
 * Copyright 2024 InfAI (CC SES)
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
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/service-commons/pkg/donewait"
	"github.com/segmentio/kafka-go"
	"time"
)

func (this *Producer) SendDone(msg donewait.DoneMsg) error {
	if this.done == nil {
		return nil
	}
	msg.Handler = "github.com/SENERGY-Platform/device-repository"
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return this.done.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(msg.ResourceId),
			Value: payload,
			Time:  time.Now(),
		},
	)
}
