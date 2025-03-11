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
	"slices"
	"strings"
)

type ImportExport struct {
	Protocols       []models.Protocol       `json:"protocols,omitempty"`
	Functions       []models.Function       `json:"functions,omitempty"`
	Aspects         []models.Aspect         `json:"aspects,omitempty"`
	Concepts        []models.Concept        `json:"concepts,omitempty"`
	Characteristics []models.Characteristic `json:"characteristics,omitempty"`
	DeviceClasses   []models.DeviceClass    `json:"device_classes,omitempty"`
	DeviceTypes     []models.DeviceType     `json:"device_types,omitempty"`

	//include_owned_information == true
	Devices      []models.Device        `json:"devices,omitempty"`
	DeviceGroups []models.DeviceGroup   `json:"device_groups,omitempty"`
	Hubs         []models.Hub           `json:"hubs,omitempty"`
	Locations    []models.Location      `json:"locations,omitempty"`
	Permissions  []permissions.Resource `json:"permissions,omitempty"`
}

type ImportExportOptions struct {
	IncludeOwnedInformation bool     `json:"include_owned_information"`
	FilterResourceTypes     []string `json:"filter_resource_types,omitempty"` //null->all; []->none
	FilterIds               []string `json:"filter_ids,omitempty"`            //null->all; []->none
}

type ImportFromOptions struct {
	FilterResourceTypes    []string `json:"filter_resource_types,omitempty"` //null->all; []->none
	FilterIds              []string `json:"filter_ids,omitempty"`            //null->all; []->none
	RemoteDeviceRepository string   `json:"remote_device_repository"`        //address of a remote device-repository, where the data is imported from
	RemoteAuthToken        string   `json:"remote_auth_token"`               //auth token used for requests to the remote device-repository
}

func (this *ImportExport) Sort() {
	slices.SortFunc(this.Protocols, func(a, b models.Protocol) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.Functions, func(a, b models.Function) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.Aspects, func(a, b models.Aspect) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.Concepts, func(a, b models.Concept) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.Characteristics, func(a, b models.Characteristic) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.DeviceClasses, func(a, b models.DeviceClass) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.DeviceTypes, func(a, b models.DeviceType) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.Devices, func(a, b models.Device) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.DeviceGroups, func(a, b models.DeviceGroup) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.Hubs, func(a, b models.Hub) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.Locations, func(a, b models.Location) int {
		return strings.Compare(a.Id, b.Id)
	})
	slices.SortFunc(this.Permissions, func(a, b permissions.Resource) int {
		return strings.Compare(a.Id, b.Id)
	})
}
