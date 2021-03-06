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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
)

func ValidateVariable(variable model.ContentVariable, serialization model.Serialization) (err error, code int) {
	if variable.Id == "" {
		return errors.New("missing content variable id"), http.StatusBadRequest
	}
	if variable.Name == "" {
		return errors.New("missing content variable name"), http.StatusBadRequest
	}
	if variable.Type == "" {
		return errors.New("missing content variable type for " + variable.Name), http.StatusBadRequest
	}
	switch variable.Type {
	case model.String:
		if len(variable.SubContentVariables) > 0 {
			return errors.New("strings can not have sub content variables for " + variable.Name), http.StatusBadRequest
		}
	case model.Integer:
		if len(variable.SubContentVariables) > 0 {
			return errors.New("integers can not have sub content variables for " + variable.Name), http.StatusBadRequest
		}
	case model.Float:
		if len(variable.SubContentVariables) > 0 {
			return errors.New("floats can not have sub content variables for " + variable.Name), http.StatusBadRequest
		}
	case model.Boolean:
		if len(variable.SubContentVariables) > 0 {
			return errors.New("booleans can not have sub content variables for " + variable.Name), http.StatusBadRequest
		}
	case model.List:
		err, code = ValidateListSubVariables(variable.SubContentVariables, serialization)
		if err != nil {
			return err, code
		}
	case model.Structure:
		err, code = ValidateStructureSubVariables(variable.SubContentVariables, serialization)
		if err != nil {
			return err, code
		}
	default:
		return errors.New("unknown content value type: " + string(variable.Type) + " in " + variable.Name), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func validateSerializationOptions(options []string, serialization model.Serialization) error {
	log.Println("TODO:")
	debug.PrintStack()
	return nil
}

func ValidateListSubVariables(variables []model.ContentVariable, serialization model.Serialization) (err error, code int) {
	if len(variables) == 0 {
		return errors.New("lists expect sub content variables"), http.StatusBadRequest
	}
	if variables[0].Name == "*" {
		if len(variables) != 1 {
			return errors.New("lists with name placeholder '*' have a variable length -> only one sub variable may be defined"), http.StatusBadRequest
		}
		return ValidateVariable(variables[0], serialization)
	}
	nameIndex := map[string]bool{}
	for _, variable := range variables {
		_, err = strconv.Atoi(variable.Name)
		if err != nil {
			return errors.New("name of list variable should be a number (if list is variable in length is may be defined with one element and the placeholder '*' as name)"), http.StatusBadRequest
		}
		nameIndex[variable.Name] = true
		err, code = ValidateVariable(variable, serialization)
		if err != nil {
			return err, code
		}
	}
	for i := 0; i < len(variables); i++ {
		if !nameIndex[strconv.Itoa(i)] {
			return errors.New("missing index name '" + strconv.Itoa(i) + "' in list content variable"), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

func ValidateStructureSubVariables(variables []model.ContentVariable, serialization model.Serialization) (err error, code int) {
	if len(variables) == 0 {
		return errors.New("structures expect sub content variables"), http.StatusBadRequest
	}
	if variables[0].Name == "*" {
		if len(variables) != 1 {
			return errors.New("structures with name placeholder '*' work as maps of variable length -> only one sub content variable may be defined"), http.StatusBadRequest
		}
	}
	nameIndex := map[string]bool{}
	for _, variable := range variables {
		if _, exists := nameIndex[variable.Name]; exists {
			return errors.New("structure sub content variable reuses name '" + variable.Name + "'"), http.StatusBadRequest
		}
		nameIndex[variable.Name] = true
		err, code = ValidateVariable(variable, serialization)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}
