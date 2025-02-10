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

import (
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"net/url"
	"strconv"
)

var True = true
var TruePtr = &True

type ValidationOptions struct {
	AllowNoneLeafAspectNodesInDeviceTypes *bool
}

func (this *ValidationOptions) CheckAllowNoneLeafAspectNodesInDeviceTypes(defaults configuration.Config) bool {
	if this != nil && this.AllowNoneLeafAspectNodesInDeviceTypes != nil {
		return *this.AllowNoneLeafAspectNodesInDeviceTypes
	}
	return defaults.AllowNoneLeafAspectNodesInDeviceTypesDefault
}

const allowNoneLeafAspectNodesInDeviceTypesQueryField string = "allow_none_leaf_aspect_nodes_in_device_types"

func (this ValidationOptions) AsUrlValues() url.Values {
	result := url.Values{}
	if this.AllowNoneLeafAspectNodesInDeviceTypes != nil {
		result.Set(allowNoneLeafAspectNodesInDeviceTypesQueryField, strconv.FormatBool(*this.AllowNoneLeafAspectNodesInDeviceTypes))
	}
	return result
}

func LoadDeviceTypeValidationOptions(query url.Values) (result ValidationOptions, err error) {
	if val := query.Get(allowNoneLeafAspectNodesInDeviceTypesQueryField); val != "" {
		parsedVal, err := strconv.ParseBool(val)
		if err != nil {
			return result, err
		}
		result.AllowNoneLeafAspectNodesInDeviceTypes = &parsedVal
	}
	return result, nil
}
