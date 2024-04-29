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
	"github.com/SENERGY-Platform/service-commons/pkg/donewait"
)

func init() {
	Factories = append(Factories, DeviceGroupListenerFactory)
}

func DeviceGroupListenerFactory(config config.Config, control Controller, securitySink SecuritySink) (topic string, listener Listener, err error) {
	return config.DeviceGroupTopic, func(msg []byte) (err error) {
		command := DeviceGroupCommand{}
		err = json.Unmarshal(msg, &command)
		if err != nil {
			return
		}
		defer func() {
			if err == nil {
				err = control.SendDone(donewait.DoneMsg{
					ResourceKind: config.DeviceGroupTopic,
					ResourceId:   command.Id,
					Command:      command.Command,
				})
			}
		}()
		switch command.Command {
		case "PUT":
			if securitySink != nil {
				err = securitySink.EnsureInitialRights(config.DeviceGroupTopic, command.Id, command.Owner)
				if err != nil {
					return err
				}
			}
			return control.SetDeviceGroup(command.DeviceGroup, command.Owner)
		case "DELETE":
			err = control.DeleteDeviceGroup(command.Id)
			if err != nil {
				return err
			}
			if securitySink != nil {
				return securitySink.RemoveRights(config.DeviceGroupTopic, command.Id)
			}
			return nil
		case "RIGHTS":
			if securitySink != nil && command.Rights != nil {
				return securitySink.SetRights(config.DeviceGroupTopic, command.Id, *command.Rights)
			}
			return nil
		}
		return errors.New("unable to handle command: " + string(msg))
	}, nil
}
