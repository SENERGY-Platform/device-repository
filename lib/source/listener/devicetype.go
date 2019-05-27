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

package listener

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/source/messages"
	"github.com/SmartEnergyPlatform/amqp-wrapper-lib"
)

func init() {
	Factories = append(Factories, DeviceTypeListenerFactory)
}

func DeviceTypeListenerFactory(config config.Config, control Controller) (topic string, listener amqp_wrapper_lib.ConsumerFunc, err error) {
	return config.DeviceTypeTopic, func(msg []byte) (err error) {
		command := messages.DeviceTypeCommand{}
		err = json.Unmarshal(msg, &command)
		if err != nil {
			return
		}
		switch command.Command {
		case "PUT":
			return control.SetDeviceType(command.DeviceType, command.Owner)
		case "DELETE":
			return control.DeleteDeviceType(command.Id)
		}
		return errors.New("unable to handle command: " + string(msg))
	}, nil
}
