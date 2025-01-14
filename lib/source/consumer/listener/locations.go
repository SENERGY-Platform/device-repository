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

package listener

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/service-commons/pkg/donewait"
)

func init() {
	Factories = append(Factories, LocationsListenerFactory)
}

func LocationsListenerFactory(config config.Config, control Controller) (topic string, listener Listener, err error) {
	return config.LocationTopic, func(msg []byte) (err error) {
		command := LocationCommand{}
		err = json.Unmarshal(msg, &command)
		if err != nil {
			return
		}
		defer func() {
			if err == nil {
				err = control.SendDone(donewait.DoneMsg{
					ResourceKind: config.LocationTopic,
					ResourceId:   command.Id,
					Command:      command.Command,
				})
			}
		}()
		switch command.Command {
		case "PUT":
			return control.SetLocation(command.Location, command.Owner)
		case "DELETE":
			return control.DeleteLocation(command.Id)
		case "RIGHTS":
			return control.SetRights(config.LocationTopic, command.Id, *command.Rights)
		}
		return errors.New("unable to handle command: " + string(msg))
	}, nil
}
