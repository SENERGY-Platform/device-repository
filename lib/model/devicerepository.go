/*
 * Copyright 2022 InfAI (CC SES)
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

package model

import "github.com/SENERGY-Platform/models/go/models"

type DeviceTypeCriteria struct {
	IsIdModified          bool        `json:"is_id_modified"`
	PureDeviceTypeId      string      `json:"pure_device_type_id"`
	DeviceTypeId          string      `json:"device_type_id"`
	ServiceId             string      `json:"service_id"`
	ContentVariableId     string      `json:"content_variable_id"`
	ContentVariablePath   string      `json:"content_variable_path"`
	FunctionId            string      `json:"function_id"`
	Interaction           string      `json:"interaction"`
	IsControllingFunction bool        `json:"controlling_function"`
	DeviceClassId         string      `json:"device_class_id"`
	AspectId              string      `json:"aspect_id"`
	CharacteristicId      string      `json:"characteristic_id"`
	IsVoid                bool        `json:"is_void"`
	Value                 interface{} `json:"value"`
	Type                  models.Type `json:"type"`
	IsLeaf                bool        `json:"is_leaf"`
	IsInput               bool        `json:"is_input"`
}

type DeviceTypeSelectable struct {
	DeviceTypeId       string                         `json:"device_type_id,omitempty"`
	Services           []models.Service               `json:"services,omitempty"`
	ServicePathOptions map[string][]ServicePathOption `json:"service_path_options,omitempty"`
}

type ServicePathOption struct {
	ServiceId             string             `json:"service_id"`
	Path                  string             `json:"path"`
	CharacteristicId      string             `json:"characteristic_id"`
	AspectNode            models.AspectNode  `json:"aspect_node"`
	FunctionId            string             `json:"function_id"`
	IsVoid                bool               `json:"is_void"`
	Value                 interface{}        `json:"value,omitempty"`
	IsControllingFunction bool               `json:"is_controlling_function"`
	Configurables         []Configurable     `json:"configurables,omitempty"`
	Type                  models.Type        `json:"type,omitempty"`
	Interaction           models.Interaction `json:"interaction"`
}

type Configurable struct {
	Path             string            `json:"path"`
	CharacteristicId string            `json:"characteristic_id"`
	AspectNode       models.AspectNode `json:"aspect_node"`
	FunctionId       string            `json:"function_id"`
	Value            interface{}       `json:"value,omitempty"`
	Type             models.Type       `json:"type,omitempty"`
}
