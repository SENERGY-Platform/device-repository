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
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func (this *Controller) ValidateVariable(variable models.ContentVariable, serialization models.Serialization, options model.ValidationOptions) (err error, code int) {
	if variable.Id == "" {
		return errors.New("missing content variable id"), http.StatusBadRequest
	}

	if variable.OmitEmpty {
		switch v := variable.Value.(type) {
		case string:
			if v != "" {
				return fmt.Errorf("validation error: %v has a value (%v) set while omit_empty is true", variable.Name, variable.Value), http.StatusBadRequest
			}
		case float64:
			if v != 0 {
				return fmt.Errorf("validation error: %v has a value (%v) set while omit_empty is true", variable.Name, variable.Value), http.StatusBadRequest
			}
		}
	}

	if !variable.IsVoid {
		if variable.Name == "" {
			return errors.New("missing content variable name"), http.StatusBadRequest
		}
		if variable.Type == "" {
			return errors.New("missing content variable type for " + variable.Name), http.StatusBadRequest
		}

		err, code = ValidateVariableName(variable.Name)
		if err != nil {
			return err, code
		}

		switch variable.Type {
		case models.String:
			if len(variable.SubContentVariables) > 0 {
				return errors.New("strings can not have sub content variables for " + variable.Name), http.StatusBadRequest
			}
		case models.Integer:
			if len(variable.SubContentVariables) > 0 {
				return errors.New("integers can not have sub content variables for " + variable.Name), http.StatusBadRequest
			}
		case models.Float:
			if len(variable.SubContentVariables) > 0 {
				return errors.New("floats can not have sub content variables for " + variable.Name), http.StatusBadRequest
			}
		case models.Boolean:
			if len(variable.SubContentVariables) > 0 {
				return errors.New("booleans can not have sub content variables for " + variable.Name), http.StatusBadRequest
			}
		case models.List:
			err, code = this.ValidateListSubVariables(variable.SubContentVariables, serialization, options)
			if err != nil {
				return err, code
			}
		case models.Structure:
			err, code = this.ValidateStructureSubVariables(variable.SubContentVariables, serialization, options)
			if err != nil {
				return err, code
			}
		default:
			return errors.New("unknown content value type: " + string(variable.Type) + " in " + variable.Name), http.StatusBadRequest
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	if variable.AspectId != "" && this != nil {
		aspectNode, exists, err := this.db.GetAspectNode(ctx, variable.AspectId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown aspect id:" + variable.AspectId), http.StatusBadRequest
		}
		if !options.CheckAllowNoneLeafAspectNodesInDeviceTypes(this.config) && len(aspectNode.DescendentIds) > 0 {
			return errors.New("only leaf aspects are allowed in device-types " + variable.Name), http.StatusBadRequest
		}
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	if variable.FunctionId != "" && this != nil {
		function, exists, err := this.db.GetFunction(ctx, variable.FunctionId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown function id:" + variable.FunctionId), http.StatusBadRequest
		}
		if variable.CharacteristicId != "" && function.ConceptId != "" {
			concept, exists, err := this.db.GetConceptWithoutCharacteristics(ctx, function.ConceptId)
			if err != nil {
				return err, http.StatusInternalServerError
			}
			if exists && !contains(concept.CharacteristicIds, variable.CharacteristicId) {
				return errors.New(fmt.Sprintf("variable characteristicId does not match variable function: %v, %v, %v, %v, %v",
					variable.Id, variable.Name, function.Id, function.ConceptId, variable.CharacteristicId)), http.StatusInternalServerError
			}
		}
	}

	return nil, http.StatusOK
}

func ValidateName(name string) (err error, code int) {
	pattern := `^[A-Za-z_][A-Za-z0-9-_]*$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(name) {
		return errors.New("invalid name:" + name), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func ValidateVariableName(name string) (err error, code int) {
	//may be placeholder for map key or list index
	if name == "*" {
		return nil, http.StatusOK
	}

	//may be a number as index of an array
	if _, err = strconv.Atoi(name); err == nil {
		return nil, http.StatusOK
	}

	return ValidateName(name)
}

func (this *Controller) ValidateListSubVariables(variables []models.ContentVariable, serialization models.Serialization, options model.ValidationOptions) (err error, code int) {
	if len(variables) == 0 {
		return errors.New("lists expect sub content variables"), http.StatusBadRequest
	}
	if variables[0].Name == "*" {
		if len(variables) != 1 {
			return errors.New("lists with name placeholder '*' have a variable length -> only one sub variable may be defined"), http.StatusBadRequest
		}
		return this.ValidateVariable(variables[0], serialization, options)
	}
	nameIndex := map[string]bool{}
	for _, variable := range variables {
		_, err = strconv.Atoi(variable.Name)
		if err != nil {
			return errors.New("name of list variable should be a number (if list is variable in length is may be defined with one element and the placeholder '*' as name)"), http.StatusBadRequest
		}
		nameIndex[variable.Name] = true
		err, code = this.ValidateVariable(variable, serialization, options)
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

func (this *Controller) ValidateStructureSubVariables(variables []models.ContentVariable, serialization models.Serialization, options model.ValidationOptions) (err error, code int) {
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
		err, code = this.ValidateVariable(variable, serialization, options)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}
