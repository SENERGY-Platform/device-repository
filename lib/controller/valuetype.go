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

package controller

import "github.com/SENERGY-Platform/iot-device-repository/lib/model"

func (this *Controller) SetValueType(valueType model.ValueType, owner string) error {
	for _, field := range valueType.Fields {
		err := this.source.PublishValueType(field.Type, owner)
		if err != nil {
			return err
		}
	}

	//cascade valuetype changes to devicetypes
	deviceTypes, err := this.db.ListDeviceTypesUsingValueType(valueType.Id)
	if err != nil {
		return err
	}
	for _, dt := range deviceTypes {
		for serviceIndex, service := range dt.Services {
			for assignmentIndex, assignment := range service.Input {
				if assignment.Type.Id == valueType.Id {
					dt.Services[serviceIndex].Input[assignmentIndex].Type = valueType
				}
			}
			for assignmentIndex, assignment := range service.Output {
				if assignment.Type.Id == valueType.Id {
					dt.Services[serviceIndex].Output[assignmentIndex].Type = valueType
				}
			}
		}
		err = this.db.SetDeviceType(dt)
		if err != nil {
			return err
		}
	}

	//cascade valuetype changes to other valueTypes
	valuetypes, err := this.db.ListValueTypesUsingValueType(valueType.Id)
	if err != nil {
		return err
	}
	for _, vt := range valuetypes {
		for fieldIndex, field := range vt.Fields {
			if field.Type.Id == valueType.Id {
				vt.Fields[fieldIndex].Type = valueType
			}
		}
		err = this.db.SetValueType(vt)
		if err != nil {
			return err
		}
	}

	return this.db.SetValueType(valueType)
}

func (this *Controller) DeleteValueType(id string) error {
	//TODO: cascade valuetype changes to using devicetypes and other using valueTypes (maybe not?)
	return this.db.RemoveValueType(id)
}

func (this *Controller) publishMissingValueTypesOfDeviceType(deviceType model.DeviceType, owner string) error {
	for _, service := range deviceType.Services {
		for _, assignment := range service.Input {
			err := this.publishMissingValueType(assignment.Type, owner)
			if err != nil {
				return err
			}
		}
		for _, assignment := range service.Output {
			err := this.publishMissingValueType(assignment.Type, owner)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *Controller) publishMissingValueType(valueType model.ValueType, owner string) error {
	_, exists, err := this.db.GetValueType(valueType.Id)
	if err != nil {
		return err
	}
	if !exists {
		return this.source.PublishValueType(valueType, owner)
	}
	return nil
}
