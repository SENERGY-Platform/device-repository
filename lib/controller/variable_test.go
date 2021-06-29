/*
 * Copyright 2020 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"testing"
)

func TestValidateVariable1(t *testing.T) {
	err, _ := ValidateVariable(model.ContentVariable{
		Id:   "foo",
		Name: "n",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:   "bar",
				Name: "*",
				Type: model.Integer,
			},
		},
	}, model.JSON)
	if err != nil {
		t.Error(err)
	}
}

func TestValidateVariable2(t *testing.T) {
	err, _ := ValidateVariable(model.ContentVariable{
		Id:   "foo",
		Name: "n",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:   "bar",
				Name: "0",
				Type: model.Integer,
			},
		},
	}, model.JSON)
	if err != nil {
		t.Error(err)
	}
}

func TestValidateVariable3(t *testing.T) {
	err, _ := ValidateVariable(model.ContentVariable{
		Id:   "foo",
		Name: "n",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:   "bar",
				Name: "*",
				Type: model.Integer,
			}, {
				Id:   "batz",
				Name: "0",
				Type: model.Integer,
			},
		},
	}, model.JSON)
	if err == nil {
		t.Error("expected error")
	}
}

func TestValidateVariable4(t *testing.T) {
	err, _ := ValidateVariable(model.ContentVariable{
		Id:   "foo.bar",
		Name: "foo.bar",
		Type: model.String,
	}, model.JSON)
	if err == nil {
		t.Error("expected error")
	}
}

func TestValidateVariable5(t *testing.T) {
	err, _ := ValidateVariable(model.ContentVariable{
		Id:   "foobar",
		Name: "foobar",
		Type: model.String,
	}, model.JSON)
	if err != nil {
		t.Error(err)
	}
}

func TestValidateVariable6(t *testing.T) {
	err, _ := ValidateVariable(model.ContentVariable{
		Id:   "foo_bar",
		Name: "foo_bar",
		Type: model.String,
	}, model.JSON)
	if err != nil {
		t.Error(err)
	}
}
