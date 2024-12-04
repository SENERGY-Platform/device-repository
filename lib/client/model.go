/*
 * Copyright 2024 InfAI (CC SES)
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

package client

import (
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

const READ = model.READ
const WRITE = model.WRITE
const EXECUTE = model.EXECUTE
const ADMINISTRATE = model.ADMINISTRATE

type DeviceListOptions = model.DeviceListOptions
type ExtendedDeviceListOptions = model.ExtendedDeviceListOptions
type HubListOptions = model.HubListOptions
type DeviceTypeListOptions = model.DeviceTypeListOptions
type DeviceGroupListOptions = model.DeviceGroupListOptions
type FilterCriteria = model.FilterCriteria
type LocationListOptions = model.LocationListOptions
type FunctionListOptions = model.FunctionListOptions
type AspectListOptions = model.AspectListOptions

const SES_ONTOLOGY_CONTROLLING_FUNCTION = model.SES_ONTOLOGY_CONTROLLING_FUNCTION
const SES_ONTOLOGY_MEASURING_FUNCTION = model.SES_ONTOLOGY_MEASURING_FUNCTION

var Event = models.EVENT
var Request = models.REQUEST
var EventAndRequest = models.EVENT_AND_REQUEST

var connectionStateOnline = models.ConnectionStateOnline
var ConnectionStateOnline = &connectionStateOnline

var connectionStateOffline = models.ConnectionStateOffline
var ConnectionStateOffline = &connectionStateOffline

var connectionStateUnknown = models.ConnectionStateUnknown
var ConnectionStateUnknown = &connectionStateUnknown

type DeviceWithConnectionState = model.DeviceWithConnectionState
