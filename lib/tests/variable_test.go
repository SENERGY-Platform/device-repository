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
	"github.com/ory/dockertest/v3"
	"log"
	"sync"
	"testing"
)

func TestVariableValidation(t *testing.T) {
	conf, err := config.Load("../../config.json")
	if err != nil {
		log.Println("ERROR: unable to load config: ", err)
		t.Error(err)
		return
	}
	conf.FatalErrHandler = t.Fatal
	conf.MongoReplSet = false
	conf.Debug = true

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Println("Could not connect to docker:", err)
		t.Error(err)
		return
	}

	_, ip, err := docker.MongoDB(pool, ctx, wg)
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

	ctrl, err := controller.New(conf, db, nil, nil)
	if err != nil {
		db.Disconnect()
		log.Println("ERROR: unable to start control", err)
		t.Error(err)
		return
	}

	err = ctrl.SetConcept(model.Concept{
		Id:                   "concept",
		Name:                 "concept",
		CharacteristicIds:    []string{"c1", "c2"},
		BaseCharacteristicId: "c1",
	}, "")
	if err != nil {
		t.Error(err)
		return
	}

	err = ctrl.SetFunction(model.Function{
		Id:        "f1",
		Name:      "f1",
		ConceptId: "",
	}, "")
	if err != nil {
		t.Error(err)
		return
	}

	err = ctrl.SetFunction(model.Function{
		Id:        "f2",
		Name:      "f2",
		ConceptId: "concept",
	}, "")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("simple no characteristic & no function", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                model.String,
		SubContentVariables: nil,
		CharacteristicId:    "",
		FunctionId:          "",
	}))

	t.Run("simple no characteristic with function f1", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                model.String,
		SubContentVariables: nil,
		CharacteristicId:    "",
		FunctionId:          "f1",
	}))

	t.Run("simple with unknown characteristic with function f1", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                model.String,
		SubContentVariables: nil,
		CharacteristicId:    "foo",
		FunctionId:          "f1",
	}))

	t.Run("simple with characteristic c1 with function f1", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                model.String,
		SubContentVariables: nil,
		CharacteristicId:    "c1",
		FunctionId:          "f1",
	}))

	t.Run("simple with characteristic c1 with function f2", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                model.String,
		SubContentVariables: nil,
		CharacteristicId:    "c1",
		FunctionId:          "f2",
	}))

	t.Run("simple no characteristic with function f2", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                model.String,
		SubContentVariables: nil,
		CharacteristicId:    "",
		FunctionId:          "f2",
	}))

	t.Run("simple no function", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                model.String,
		SubContentVariables: nil,
		CharacteristicId:    "foo",
		FunctionId:          "",
	}))

	t.Run("simple with unknown characteristic with function f2", testValidateVariable(ctrl, true, model.ContentVariable{
		Id:                  "v",
		Name:                "v",
		Type:                model.String,
		SubContentVariables: nil,
		CharacteristicId:    "foo",
		FunctionId:          "f2",
	}))

	t.Run("struct no characteristic & no function", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:               "v",
				Name:             "v",
				Type:             model.String,
				CharacteristicId: "",
				FunctionId:       "",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct no characteristic with function f1", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct with unknown characteristic with function f1", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct with characteristic c1 with function f1", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "c1",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct with characteristic c1 with function f2", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "c1",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct no characteristic with function f2", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct no function", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("struct with unknown characteristic with function f2", testValidateVariable(ctrl, true, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "v",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list no characteristic & no function", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:               "v",
				Name:             "0",
				Type:             model.String,
				CharacteristicId: "",
				FunctionId:       "",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list no characteristic with function f1", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list with unknown characteristic with function f1", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list with characteristic c1 with function f1", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "c1",
				FunctionId:          "f1",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list with characteristic c1 with function f2", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "c1",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list no characteristic with function f2", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list no function", testValidateVariable(ctrl, false, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))

	t.Run("list with unknown characteristic with function f2", testValidateVariable(ctrl, true, model.ContentVariable{
		Id:   "root",
		Name: "root",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:                  "v",
				Name:                "0",
				Type:                model.String,
				SubContentVariables: nil,
				CharacteristicId:    "foo",
				FunctionId:          "f2",
			},
		},
		CharacteristicId: "",
		FunctionId:       "",
	}))
}

func testValidateVariable(ctrl *controller.Controller, expectError bool, variable model.ContentVariable) func(t *testing.T) {
	return func(t *testing.T) {
		err, _ := ctrl.ValidateVariable(variable, "json")
		if (err != nil) != expectError {
			t.Error(expectError, err)
		}
	}
}
