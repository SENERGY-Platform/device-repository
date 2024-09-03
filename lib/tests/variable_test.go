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

package tests

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"log"
	"sync"
	"testing"
)

func TestVariableValidation(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.Load("../../config.json")
	if err != nil {
		log.Println("ERROR: unable to load config: ", err)
		t.Error(err)
		return
	}
	conf.FatalErrHandler = t.Fatal
	conf.MongoReplSet = false
	conf.Debug = true

	_, ip, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	conf.MongoUrl = "mongodb://" + ip + ":27017"

	db, err := database.New(conf)
	if err != nil {
		log.Println("ERROR: unable to connect to database", err)
		t.Error(err)
		return
	}
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		<-ctx.Done()
		db.Disconnect()
		if wg != nil {
			wg.Done()
		}
	}()

	ctrl, err := controller.New(conf, db, nil)
	if err != nil {
		db.Disconnect()
		log.Println("ERROR: unable to start control", err)
		t.Error(err)
		return
	}

	err = ctrl.SetConcept(models.Concept{
		Id:                   "concept",
		Name:                 "concept",
		CharacteristicIds:    []string{"c1", "c2"},
		BaseCharacteristicId: "c1",
	}, "")
	if err != nil {
		t.Error(err)
		return
	}

	err = ctrl.SetFunction(models.Function{
		Id:        "f1",
		Name:      "f1",
		ConceptId: "",
	}, "")
	if err != nil {
		t.Error(err)
		return
	}

	err = ctrl.SetFunction(models.Function{
		Id:        "f2",
		Name:      "f2",
		ConceptId: "concept",
	}, "")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("simple no characteristic & no function", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                models.String,
		SubContentVariables: nil,
		CharacteristicId:    "",
		FunctionId:          "",
	}))

	t.Run("simple no characteristic with function f1", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                models.String,
		SubContentVariables: nil,
		CharacteristicId:    "",
		FunctionId:          "f1",
	}))

	t.Run("simple with unknown characteristic with function f1", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                models.String,
		SubContentVariables: nil,
		CharacteristicId:    "foo",
		FunctionId:          "f1",
	}))

	t.Run("simple with characteristic c1 with function f1", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                models.String,
		SubContentVariables: nil,
		CharacteristicId:    "c1",
		FunctionId:          "f1",
	}))

	t.Run("simple with characteristic c1 with function f2", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                models.String,
		SubContentVariables: nil,
		CharacteristicId:    "c1",
		FunctionId:          "f2",
	}))

	t.Run("simple no characteristic with function f2", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                models.String,
		SubContentVariables: nil,
		CharacteristicId:    "",
		FunctionId:          "f2",
	}))

	t.Run("simple no function", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                models.String,
		SubContentVariables: nil,
		CharacteristicId:    "foo",
		FunctionId:          "",
	}))

	t.Run("simple with unknown characteristic with function f2", testValidateVariable(ctrl, true, models.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                models.String,
		SubContentVariables: nil,
		CharacteristicId:    "foo",
		FunctionId:          "f2",
	}))

	t.Run("struct no characteristic & no function", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:               "v",
				Name:             "v",
				Type:             models.String,
				CharacteristicId: "",
				FunctionId:       "",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct no characteristic with function f1", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct with unknown characteristic with function f1", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct with characteristic c1 with function f1", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "c1",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct with characteristic c1 with function f2", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "c1",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct no characteristic with function f2", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct no function", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct with unknown characteristic with function f2", testValidateVariable(ctrl, true, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list no characteristic & no function", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubContentVariables: []models.ContentVariable{
			{
				Id:               "v",
				Name:             "0",
				Type:             models.String,
				CharacteristicId: "",
				FunctionId:       "",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list no characteristic with function f1", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list with unknown characteristic with function f1", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list with characteristic c1 with function f1", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "c1",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list with characteristic c1 with function f2", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "c1",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list no characteristic with function f2", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list no function", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list with unknown characteristic with function f2", testValidateVariable(ctrl, true, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubContentVariables: []models.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                models.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct duplicate sub variable name", testValidateVariable(ctrl, true, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:   "v",
				Name: "v",
				Type: models.String,
			},
			{
				Id:   "v2",
				Name: "v",
				Type: models.Integer,
			},
		},
	}))

	t.Run("list duplicate sub variable name", testValidateVariable(ctrl, true, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.List,
		SubContentVariables: []models.ContentVariable{
			{
				Id:   "v",
				Name: "0",
				Type: models.String,
			},
			{
				Id:   "v2",
				Name: "0",
				Type: models.String,
			},
		},
	}))

	t.Run("omit empty value ''", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:        "v",
				Name:      "v",
				Type:      models.String,
				Value:     "",
				OmitEmpty: true,
			},
		},
	}))

	t.Run("omit empty value 'foo'", testValidateVariable(ctrl, true, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:        "v",
				Name:      "v",
				Type:      models.String,
				Value:     "foo",
				OmitEmpty: true,
			},
		},
	}))

	t.Run("omit empty value 0", testValidateVariable(ctrl, false, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:        "v",
				Name:      "v",
				Type:      models.Float,
				Value:     0.0,
				OmitEmpty: true,
			},
		},
	}))

	t.Run("omit empty value 42", testValidateVariable(ctrl, true, models.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: models.Structure,
		SubContentVariables: []models.ContentVariable{
			{
				Id:        "v",
				Name:      "v",
				Type:      models.Float,
				Value:     4.2,
				OmitEmpty: true,
			},
		},
	}))
}

func testValidateVariable(ctrl *controller.Controller, expectError bool, variable models.ContentVariable) func(t *testing.T) {
	return func(t *testing.T) {
		err, _ := ctrl.ValidateVariable(variable, "json", model.ValidationOptions{})
		if (err != nil) != expectError {
			t.Error(expectError, err)
		}
	}
}
