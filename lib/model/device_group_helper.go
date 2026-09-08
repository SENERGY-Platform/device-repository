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

package model

import "github.com/SENERGY-Platform/models/go/models"

type DeviceGroupHelperResult struct {
	Criteria []models.DeviceGroupFilterCriteria `json:"criteria"`
	Options  []DeviceGroupOption                `json:"options"`
}

type DeviceGroupOption struct {
	Device                  models.ExtendedDevice              `json:"device"`
	RemovesCriteria         []models.DeviceGroupFilterCriteria `json:"removes_criteria"`
	MaintainsGroupUsability bool                               `json:"maintains_group_usability"`
}

// DeviceGroupHelperOptions selects the devices listed as DeviceGroupHelperResult.Options.
type DeviceGroupHelperOptions struct {
	Search                  string
	Limit                   int64    //default 100
	Offset                  int64    //default 0
	MaintainsGroupUsability bool     //filter; list only options that leave the group with at least one counting criteria
	FunctionBlockList       []string //function ids that do not count towards keeping the group usable; ignored if empty
}
