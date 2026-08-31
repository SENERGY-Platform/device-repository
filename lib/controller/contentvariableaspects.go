/*
 * Copyright 2026 InfAI (CC SES)
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

package controller

import (
	"slices"

	"github.com/SENERGY-Platform/models/go/models"
)

//models.ContentVariable.AspectId is deprecated in favor of models.ContentVariable.AspectIds.
//Everything behind the write normalization reads AspectIds only; the deprecated field is
//kept in sync at the controller boundary so that clients which know only one of the two
//fields keep working.

// SetContentVariableAspectIdsOnWrite adds the deprecated AspectId to AspectIds,
// so that only AspectIds has to be interpreted afterwards. Exported for the mgw mirror
// and for the device-group criteria, which read device-types straight from the database
// without passing a controller read.
func SetContentVariableAspectIdsOnWrite(deviceType *models.DeviceType) {
	forEachDeviceTypeContentVariable(deviceType, addContentVariableAspectIdToAspectIds)
}

// setContentVariableAspectIdOnRead makes both fields consistent for stored device-types.
func setContentVariableAspectIdOnRead(deviceType *models.DeviceType) {
	forEachDeviceTypeContentVariable(deviceType, syncContentVariableAspects)
}

func setContentVariableAspectIdOnReadList(deviceTypes []models.DeviceType) {
	for i := range deviceTypes {
		setContentVariableAspectIdOnRead(&deviceTypes[i])
	}
}

func setContentVariableAspectIdOnReadService(service *models.Service) {
	forEachServiceContentVariable(service, syncContentVariableAspects)
}

// syncContentVariableAspects fills AspectIds from the deprecated AspectId first, because
// device-types written before the migration carry only AspectId and everything behind a
// read evaluates AspectIds. AspectId is then set to the alphabetically first entry of
// AspectIds, to serve clients that do not know AspectIds yet.
func syncContentVariableAspects(variable *models.ContentVariable) {
	addContentVariableAspectIdToAspectIds(variable)
	if len(variable.AspectIds) == 0 {
		return
	}
	sorted := slices.Clone(variable.AspectIds)
	slices.Sort(sorted)
	variable.AspectId = sorted[0]
}

func addContentVariableAspectIdToAspectIds(variable *models.ContentVariable) {
	if variable.AspectId != "" && !slices.Contains(variable.AspectIds, variable.AspectId) {
		variable.AspectIds = append(variable.AspectIds, variable.AspectId)
	}
}

func forEachDeviceTypeContentVariable(deviceType *models.DeviceType, f func(variable *models.ContentVariable)) {
	for i := range deviceType.Services {
		forEachServiceContentVariable(&deviceType.Services[i], f)
	}
}

func forEachServiceContentVariable(service *models.Service, f func(variable *models.ContentVariable)) {
	for i := range service.Inputs {
		forEachContentVariable(&service.Inputs[i].ContentVariable, f)
	}
	for i := range service.Outputs {
		forEachContentVariable(&service.Outputs[i].ContentVariable, f)
	}
}

func forEachContentVariable(variable *models.ContentVariable, f func(variable *models.ContentVariable)) {
	f(variable)
	for i := range variable.SubContentVariables {
		forEachContentVariable(&variable.SubContentVariables[i], f)
	}
}
