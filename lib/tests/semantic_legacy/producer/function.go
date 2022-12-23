/*
 *
 * Copyright 2020 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *
 */

package producer

import (
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/source/consumer/listener"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"runtime/debug"
)

func (this *Producer) PublishFunction(function models.Function, userId string) (err error) {
	cmd := listener.FunctionCommand{Command: "PUT", Function: function, Owner: userId}
	return this.PublishFunctionCommand(cmd)
}

func (this *Producer) PublishFunctionDelete(id string, userId string) error {
	cmd := listener.FunctionCommand{Command: "DELETE", Id: id, Owner: userId}
	return this.PublishFunctionCommand(cmd)
}

func (this *Producer) PublishFunctionCommand(cmd listener.FunctionCommand) error {
	if this.config.Debug {
		log.Println("DEBUG: produce function", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	return this.callListener(this.config.FunctionTopic, message)
}
