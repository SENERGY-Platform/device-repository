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

import "github.com/SENERGY-Platform/device-repository/lib/model"

type DeviceCommand struct {
	Command string       `json:"command"`
	Id      string       `json:"id"`
	Owner   string       `json:"owner"`
	Device  model.Device `json:"device"`
}

type HubCommand struct {
	Command string    `json:"command"`
	Id      string    `json:"id"`
	Owner   string    `json:"owner"`
	Hub     model.Hub `json:"hub"`
}

type DeviceTypeCommand struct {
	Command    string           `json:"command"`
	Id         string           `json:"id"`
	Owner      string           `json:"owner"`
	DeviceType model.DeviceType `json:"device_type"`
}

type ProtocolCommand struct {
	Command  string         `json:"command"`
	Id       string         `json:"id"`
	Owner    string         `json:"owner"`
	Protocol model.Protocol `json:"protocol"`
}

type DeviceGroupCommand struct {
	Command     string            `json:"command"`
	Id          string            `json:"id"`
	Owner       string            `json:"owner"`
	DeviceGroup model.DeviceGroup `json:"device_group"`
}
