/*
 * Copyright 2025 InfAI (CC SES)
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

import (
	"github.com/SENERGY-Platform/models/go/models"
	permissions "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

type ImportExport struct {
	Protocols       []models.Protocol       `json:"protocols"`
	Functions       []models.Function       `json:"functions"`
	Aspects         []models.Aspect         `json:"aspects"`
	Concepts        []models.Concept        `json:"concepts"`
	Characteristics []models.Characteristic `json:"characteristics"`
	DeviceClasses   []models.DeviceClass    `json:"device_classes"`
	DeviceTypes     []models.DeviceType     `json:"device_types"`

	//include_owned_information == true
	Devices      []models.Device        `json:"devices,omitempty"`
	DeviceGroups []models.DeviceGroup   `json:"device_groups,omitempty"`
	Hubs         []models.Hub           `json:"hubs,omitempty"`
	Locations    []models.Location      `json:"locations,omitempty"`
	Permissions  []permissions.Resource `json:"permissions,omitempty"`
}

type ImportExportOptions struct {
	IncludeOwnedInformation bool `json:"include_owned_information"`
}
