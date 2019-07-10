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

package tests

import (
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

type VariableValidationTest struct {
	Error       bool           `json:"error"`        //if true -> expect validation error
	ExpectedMsg string         `json:"expected_msg"` //if != "" -> checks error messages
	Variable    model.Variable `json:"variable"`
}

func TestVariableValidation(t *testing.T) {
	file, error := os.Open("resources/variables.json")
	if error != nil {
		t.Fatal(error)
	}
	tests := map[string]VariableValidationTest{}
	err := json.NewDecoder(RemoveComment(file)).Decode(&tests)
	if err != nil {
		t.Fatal(err)
	}
	for testname, test := range tests {
		t.Run(testname, test.Run)
	}
}

func (this VariableValidationTest) Run(t *testing.T) {
	err, _ := controller.ValidateVariable(this.Variable)
	if this.Error {
		if err == nil {
			t.Fatal("expected error")
		}
		if this.ExpectedMsg != "" && err.Error() != this.ExpectedMsg {
			t.Fatal("unexpected eror msg:", err.Error(), "!=", this.ExpectedMsg)
		}
	} else {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func RemoveComment(in io.Reader) (out io.Reader) {
	buffer, err := ioutil.ReadAll(in)
	if err != nil {
		panic(err)
	}
	str := string(buffer)
	str = regexp.MustCompile(`(?im)^\s+\/\/.*$`).ReplaceAllString(str, "")
	str = regexp.MustCompile(`(?im)\/\/[^"\[\]]+$`).ReplaceAllString(str, "")
	return strings.NewReader(str)
}
