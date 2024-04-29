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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type DeviceCommand struct {
	Command string                `json:"command"`
	Id      string                `json:"id"`
	Owner   string                `json:"owner"`
	Device  models.Device         `json:"device"`
	Rights  *model.ResourceRights `json:"rights,omitempty"`
}

type HubCommand struct {
	Command string                `json:"command"`
	Id      string                `json:"id"`
	Owner   string                `json:"owner"`
	Hub     models.Hub            `json:"hub"`
	Rights  *model.ResourceRights `json:"rights,omitempty"`
}

type DeviceTypeCommand struct {
	Command    string            `json:"command"`
	Id         string            `json:"id"`
	Owner      string            `json:"owner"`
	DeviceType models.DeviceType `json:"device_type"`
}

type ProtocolCommand struct {
	Command  string          `json:"command"`
	Id       string          `json:"id"`
	Owner    string          `json:"owner"`
	Protocol models.Protocol `json:"protocol"`
}

type DeviceGroupCommand struct {
	Command     string                `json:"command"`
	Id          string                `json:"id"`
	Owner       string                `json:"owner"`
	DeviceGroup models.DeviceGroup    `json:"device_group"`
	Rights      *model.ResourceRights `json:"rights,omitempty"`
}

type ConceptCommand struct {
	Command string         `json:"command"`
	Id      string         `json:"id"`
	Owner   string         `json:"owner"`
	Concept models.Concept `json:"concept"`
}

type CharacteristicCommand struct {
	Command        string                `json:"command"`
	Id             string                `json:"id"`
	Owner          string                `json:"owner"`
	Characteristic models.Characteristic `json:"characteristic"`
}

type DeviceClassCommand struct {
	Command     string             `json:"command"`
	Id          string             `json:"id"`
	Owner       string             `json:"owner"`
	DeviceClass models.DeviceClass `json:"device_class"`
}

type AspectCommand struct {
	Command string        `json:"command"`
	Id      string        `json:"id"`
	Owner   string        `json:"owner"`
	Aspect  models.Aspect `json:"aspect"`
}

type FunctionCommand struct {
	Command  string          `json:"command"`
	Id       string          `json:"id"`
	Owner    string          `json:"owner"`
	Function models.Function `json:"function"`
}

type LocationCommand struct {
	Command  string          `json:"command"`
	Id       string          `json:"id"`
	Owner    string          `json:"owner"`
	Location models.Location `json:"location"`
}
