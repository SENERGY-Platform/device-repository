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

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"strings"
	"time"
)

func (this *Controller) ValidateService(service models.Service, protocolCache *map[string]models.Protocol, options model.ValidationOptions) (error, int) {
	if service.Id == "" {
		return errors.New("missing service id"), http.StatusBadRequest
	}
	if service.Name == "" {
		return errors.New("missing service name"), http.StatusBadRequest
	}
	if service.LocalId == "" {
		return errors.New("missing service local id"), http.StatusBadRequest
	}
	if service.ProtocolId == "" {
		return errors.New("missing service protocol id"), http.StatusBadRequest
	}

	var protocol models.Protocol
	var ok bool
	var err error
	if protocol, ok = (*protocolCache)[service.ProtocolId]; !ok {
		ctx, _ := getTimeoutContext()
		protocol, ok, err = this.db.GetProtocol(ctx, service.ProtocolId)
		if err != nil {
			return err, http.StatusBadRequest
		}
		if !ok {
			return errors.New("unknown protocol"), http.StatusBadRequest
		}
		(*protocolCache)[service.ProtocolId] = protocol
	}

	if contains(protocol.Constraints, model.SenergyConnectorLocalIdConstraint) {
		if strings.ContainsAny(service.LocalId, "+#/") {
			return errors.New("service local id may not contain any +#/"), http.StatusBadRequest
		}
	}

	knownContentNames := map[string]bool{}
	for _, content := range service.Inputs {
		if _, ok := knownContentNames[content.ContentVariable.Name]; ok {
			return errors.New("reused input content name: " + content.ContentVariable.Name), http.StatusBadRequest
		} else {
			knownContentNames[content.ContentVariable.Name] = true
		}
	}
	knownContentNames = map[string]bool{}
	for _, content := range service.Outputs {
		if _, ok := knownContentNames[content.ContentVariable.Name]; ok {
			return errors.New("reused output content name: " + content.ContentVariable.Name), http.StatusBadRequest
		} else {
			knownContentNames[content.ContentVariable.Name] = true
		}
	}

	for _, content := range service.Inputs {
		err, code := this.ValidateContent(content, protocol, options)
		if err != nil {
			return err, code
		}
		err = validateFunctionTypeUse(content.ContentVariable, true)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}
	for _, content := range service.Outputs {
		err, code := this.ValidateContent(content, protocol, options)
		if err != nil {
			return err, code
		}
		err = validateFunctionTypeUse(content.ContentVariable, false)
		if err != nil {
			return err, http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

func validateFunctionTypeUse(variable models.ContentVariable, isInput bool) (err error) {
	if variable.FunctionId != "" && strings.HasPrefix(variable.FunctionId, model.URN_PREFIX) {
		isCtrlFun := isControllingFunction(variable.FunctionId)
		if isCtrlFun != isInput {
			if isCtrlFun {
				return errors.New("use controlling function " + variable.FunctionId + " in output variable " + variable.Name + " (" + variable.Id + ")")
			} else {
				return errors.New("use measuring function " + variable.FunctionId + " in input variable " + variable.Name + " (" + variable.Id + ")")
			}
		}
	}
	return nil
}

func isControllingFunction(functionId string) bool {
	if strings.HasPrefix(functionId, "urn:infai:ses:controlling-function:") {
		return true
	}
	return false
}

func (this *Controller) GetService(id string) (result models.Service, err error, code int) {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	dts, err := this.db.GetDeviceTypesByServiceId(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if len(dts) == 0 {
		return result, errors.New("not found"), http.StatusNotFound
	}
	for _, dt := range dts {
		for _, service := range dt.Services {
			if service.Id == id {
				return service, nil, 200
			}
		}
	}
	return result, errors.New("not found"), http.StatusNotFound
}
