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

package messages

import "github.com/SENERGY-Platform/iot-device-repository/lib/model"

type DeviceinstanceCommand struct {
	Command        string               `json:"command"`
	Id             string               `json:"id"`
	Owner          string               `json:"owner"`
	DeviceInstance model.DeviceInstance `json:"device_instance"`
}

type DeviceTypeCommand struct {
	Command    string           `json:"command"`
	Id         string           `json:"id"`
	Owner      string           `json:"owner"`
	DeviceType model.DeviceType `json:"device_type"`
}

type GatewayCommand struct {
	Command string   `json:"command"`
	Id      string   `json:"id"`
	Owner   string   `json:"owner"`
	Name    string   `json:"name"`
	Hash    string   `json:"hash"`
	Devices []string `json:"devices"`
}

type ValueTypeCommand struct {
	Command   string          `json:"command"`
	Id        string          `json:"id"`
	Owner     string          `json:"owner"`
	ValueType model.ValueType `json:"value_type"`
}