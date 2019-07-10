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
	"net/http"
	"strconv"
)

func ValidateVariable(variable model.Variable) (err error, code int) {
	if variable.Id == "" {
		return errors.New("missing variable id"), http.StatusBadRequest
	}
	if variable.Name == "" {
		return errors.New("missing variable name"), http.StatusBadRequest
	}
	if variable.Type == "" {
		return errors.New("missing variable type"), http.StatusBadRequest
	}
	err, code = ValidateProperty(variable.Property, variable.Type)
	if err != nil {
		return err, code
	}
	switch variable.Type {
	case model.String:
		if len(variable.SubVariables) > 0 {
			return errors.New("strings can not have sub variables"), http.StatusBadRequest
		}
	case model.Integer:
		if len(variable.SubVariables) > 0 {
			return errors.New("integers can not have sub variables"), http.StatusBadRequest
		}
	case model.Float:
		if len(variable.SubVariables) > 0 {
			return errors.New("floats can not have sub variables"), http.StatusBadRequest
		}
	case model.Boolean:
		if len(variable.SubVariables) > 0 {
			return errors.New("booleans can not have sub variables"), http.StatusBadRequest
		}
	case model.List:
		err, code = ValidateListSubVariables(variable.SubVariables)
		if err != nil {
			return err, code
		}
	case model.Structure:
		err, code = ValidateStructureSubVariables(variable.SubVariables)
		if err != nil {
			return err, code
		}
	default:
		return errors.New("unknown variable type: " + string(variable.Type)), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func ValidateListSubVariables(variables []model.Variable) (err error, code int) {
	if len(variables) == 0 {
		return errors.New("lists expect sub variables"), http.StatusBadRequest
	}
	if variables[0].Name == "*" {
		if len(variables) != 1 {
			return errors.New("lists with name placeholder '*' have a variable length -> only one sub variable may be defined"), http.StatusBadRequest
		}
	}
	nameIndex := map[string]bool{}
	for _, variable := range variables {
		_, err = strconv.Atoi(variable.Name)
		if err != nil {
			return errors.New("name of list variable should be a number (if list is variable in length is may be defined with one element and the placeholder '*' as name)"), http.StatusBadRequest
		}
		nameIndex[variable.Name] = true
		err, code = ValidateVariable(variable)
		if err != nil {
			return err, code
		}
	}
	for i := 0; i < len(variables); i++ {
		if !nameIndex[strconv.Itoa(i)] {
			return errors.New("missing index name '" + strconv.Itoa(i) + "' in list variable"), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

func ValidateStructureSubVariables(variables []model.Variable) (err error, code int) {
	if len(variables) == 0 {
		return errors.New("structures expect sub variables"), http.StatusBadRequest
	}
	if variables[0].Name == "*" {
		if len(variables) != 1 {
			return errors.New("structures with name placeholder '*' work as maps of variable length -> only one sub variable may be defined"), http.StatusBadRequest
		}
	}
	nameIndex := map[string]bool{}
	for _, variable := range variables {
		if _, exists := nameIndex[variable.Name]; exists {
			return errors.New("structure sub variable reuses name '" + variable.Name + "'"), http.StatusBadRequest
		}
		nameIndex[variable.Name] = true
		err, code = ValidateVariable(variable)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}

func ValidateProperty(property model.Property, variableType model.VariableType) (error, int) {
	zero := model.Property{}
	if property == zero {
		return nil, http.StatusOK
	}
	if property.Id == "" {
		return errors.New("missing property id"), http.StatusBadRequest
	}
	if property.Value != nil {
		switch variableType {
		case model.String:
			if _, ok := property.Value.(string); !ok {
				return errors.New("expect string as property value if variable has type " + string(model.String)), http.StatusBadRequest
			}
		case model.Integer:
			if _, ok := property.Value.(float64); !ok {
				return errors.New("expect number as property value if variable has type " + string(model.Integer)), http.StatusBadRequest
			}
		case model.Float:
			if _, ok := property.Value.(float64); !ok {
				return errors.New("expect number as property value if variable has type " + string(model.Float)), http.StatusBadRequest
			}
		case model.Boolean:
			if _, ok := property.Value.(bool); !ok {
				return errors.New("expect bool as property value if variable has type " + string(model.Boolean)), http.StatusBadRequest
			}
		case model.List:
			return errors.New("lists dont support literal values in property"), http.StatusBadRequest
		case model.Structure:
			return errors.New("structures dont support literal values in property"), http.StatusBadRequest
		default:
			return errors.New("unknown variable type: " + string(variableType)), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}
