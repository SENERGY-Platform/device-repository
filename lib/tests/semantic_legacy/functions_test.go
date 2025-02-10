/*
 *
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *
 */

package semantic_legacy

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/repo_legacy/testenv"
	"github.com/SENERGY-Platform/models/go/models"
	"sync"
	"testing"
)

func TestFunction(t *testing.T) {
	conf, err := configuration.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()
	conf, ctrl, err := NewPartialMockEnv(ctx, wg, conf, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("testProduceFunctions", testProduceFunctions(conf))
	t.Run("testUpdateFunctions", testUpdateFunctionsDisplayName(conf))
	t.Run("testReadControllingFunction", testReadControllingFunction(ctrl))
	t.Run("testReadMeasuringFunction", testReadMeasuringFunction(ctrl))
	t.Run("list functions", func(t *testing.T) {
		t.Run("list all", func(t *testing.T) {
			_, total, err, _ := ctrl.ListFunctions(model.FunctionListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 6 {
				t.Error(total)
				return
			}
		})
		t.Run("find humidity", func(t *testing.T) {
			list, total, err, _ := ctrl.ListFunctions(model.FunctionListOptions{Search: "humidity"})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 1 {
				t.Error(total)
				return
			}
			if list[0].Id != "urn:infai:ses:measuring-function:467" {
				t.Error(list[0].Id)
				return
			}
		})
		t.Run("list controlling", func(t *testing.T) {
			list, total, err, _ := ctrl.ListFunctions(model.FunctionListOptions{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 3 {
				t.Error(total)
				return
			}
			for _, function := range list {
				if function.RdfType != model.SES_ONTOLOGY_CONTROLLING_FUNCTION {
					t.Error(function)
					return
				}
			}
		})
		t.Run("list measuring", func(t *testing.T) {
			list, total, err, _ := ctrl.ListFunctions(model.FunctionListOptions{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 3 {
				t.Error(total)
				return
			}
			for _, function := range list {
				if function.RdfType != model.SES_ONTOLOGY_MEASURING_FUNCTION {
					t.Error(function)
					return
				}
			}
		})
	})
	t.Run("testFunctionDelete", testFunctionDelete(conf))
}

func testProduceFunctions(conf configuration.Config) func(t *testing.T) {
	return func(t *testing.T) {
		confunction1 := models.Function{}
		confunction1.Id = "urn:infai:ses:controlling-function:333"
		confunction1.Name = "setOnFunction"
		confunction1.DisplayName = "foo"
		confunction1.Description = "Turn the device on"

		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetFunction(testenv.AdminToken, confunction1)
		if err != nil {
			t.Fatal(err)
		}

		confunction2 := models.Function{}
		confunction2.Id = "urn:infai:ses:controlling-function:2222"
		confunction2.Name = "setOffFunction"
		confunction2.DisplayName = "off-function"
		confunction2.ConceptId = ""

		_, err, _ = c.SetFunction(testenv.AdminToken, confunction2)
		if err != nil {
			t.Fatal(err)
		}

		confunction3 := models.Function{}
		confunction3.Id = "urn:infai:ses:controlling-function:5467567"
		confunction3.Name = "setColorFunction"
		confunction3.DisplayName = "ctrl display name"
		confunction3.ConceptId = "urn:infai:ses:concept:efffsdfd-01a1-4434-9dcc-064b3955000f"

		_, err, _ = c.SetFunction(testenv.AdminToken, confunction3)
		if err != nil {
			t.Fatal(err)
		}

		measfunction1 := models.Function{}
		measfunction1.Id = "urn:infai:ses:measuring-function:23"
		measfunction1.Name = "getOnOffFunction"
		measfunction1.DisplayName = "bar"

		_, err, _ = c.SetFunction(testenv.AdminToken, measfunction1)
		if err != nil {
			t.Fatal(err)
		}

		measfunction2 := models.Function{}
		measfunction2.Id = "urn:infai:ses:measuring-function:321"
		measfunction2.Name = "getTemperatureFunction"
		measfunction2.ConceptId = "urn:infai:ses:concept:efffsdfd-aaaa-bbbb-ccc-0000"
		measfunction2.DisplayName = "batz"

		_, err, _ = c.SetFunction(testenv.AdminToken, measfunction2)
		if err != nil {
			t.Fatal(err)
		}

		measfunction3 := models.Function{}
		measfunction3.Id = "urn:infai:ses:measuring-function:467"
		measfunction3.Name = "getHumidityFunction"
		measfunction3.DisplayName = "hum_display"

		_, err, _ = c.SetFunction(testenv.AdminToken, measfunction3)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testUpdateFunctionsDisplayName(conf configuration.Config) func(t *testing.T) {
	return func(t *testing.T) {
		confunction1 := models.Function{}
		confunction1.Id = "urn:infai:ses:controlling-function:333"
		confunction1.Name = "setOnFunction"
		confunction1.DisplayName = "foo 2"
		confunction1.Description = "Turn the device on"

		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetFunction(testenv.AdminToken, confunction1)
		if err != nil {
			t.Fatal(err)
		}

		confunction2 := models.Function{}
		confunction2.Id = "urn:infai:ses:controlling-function:2222"
		confunction2.Name = "setOffFunction"
		confunction2.DisplayName = "off-function 2"
		confunction2.ConceptId = "2"

		_, err, _ = c.SetFunction(testenv.AdminToken, confunction2)
		if err != nil {
			t.Fatal(err)
		}

		confunction3 := models.Function{}
		confunction3.Id = "urn:infai:ses:controlling-function:5467567"
		confunction3.Name = "setColorFunction"
		confunction3.DisplayName = "ctrl display name 2"
		confunction3.ConceptId = "urn:infai:ses:concept:efffsdfd-01a1-4434-9dcc-064b3955000f"

		_, err, _ = c.SetFunction(testenv.AdminToken, confunction3)
		if err != nil {
			t.Fatal(err)
		}

		measfunction1 := models.Function{}
		measfunction1.Id = "urn:infai:ses:measuring-function:23"
		measfunction1.Name = "getOnOffFunction"
		measfunction1.DisplayName = "bar 2"

		_, err, _ = c.SetFunction(testenv.AdminToken, measfunction1)
		if err != nil {
			t.Fatal(err)
		}

		measfunction2 := models.Function{}
		measfunction2.Id = "urn:infai:ses:measuring-function:321"
		measfunction2.Name = "getTemperatureFunction"
		measfunction2.ConceptId = "urn:infai:ses:concept:efffsdfd-aaaa-bbbb-ccc-0000"
		measfunction2.DisplayName = "batz 2"

		_, err, _ = c.SetFunction(testenv.AdminToken, measfunction2)
		if err != nil {
			t.Fatal(err)
		}

		measfunction3 := models.Function{}
		measfunction3.Id = "urn:infai:ses:measuring-function:467"
		measfunction3.Name = "getHumidityFunction"
		measfunction3.DisplayName = "hum_display 2"

		_, err, _ = c.SetFunction(testenv.AdminToken, measfunction3)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testReadControllingFunction(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		res, err, code := con.GetFunctionsByType(model.SES_ONTOLOGY_CONTROLLING_FUNCTION)
		if err != nil {
			t.Fatal(res, err, code)
		} else {
			t.Log(res)
		}
		if len(res) < 3 {
			t.Error(res)
			return
		}
		if res[0].Id != "urn:infai:ses:controlling-function:2222" {
			t.Fatal("error id")
		}
		if res[0].Name != "setOffFunction" {
			t.Fatal("error Name")
		}
		if res[0].DisplayName != "off-function 2" {
			t.Fatal("error Display Name", res[0])
		}
		if res[0].ConceptId != "2" {
			t.Fatal("error ConceptId", res[0].ConceptId)
		}

		if res[1].Id != "urn:infai:ses:controlling-function:333" {
			t.Fatal("error id")
		}
		if res[1].Name != "setOnFunction" {
			t.Fatal("error Name")
		}
		if res[1].ConceptId != "" {
			t.Fatal("error ConceptId", res[1].ConceptId)
		}
		if res[1].Description != "Turn the device on" {
			t.Fatal("error Description")
		}

		if res[2].Id != "urn:infai:ses:controlling-function:5467567" {
			t.Fatal("error id", res[2].Id)
		}
		if res[2].Name != "setColorFunction" {
			t.Fatal("error Name")
		}
		if res[2].DisplayName != "ctrl display name 2" {
			t.Fatal("error DisplayName", res[2].DisplayName)
		}
		if res[2].ConceptId != "urn:infai:ses:concept:efffsdfd-01a1-4434-9dcc-064b3955000f" {
			t.Fatal("error ConceptId", res[2].ConceptId)
		}
	}
}

func testReadMeasuringFunction(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		res, err, code := con.GetFunctionsByType(model.SES_ONTOLOGY_MEASURING_FUNCTION)
		if err != nil {
			t.Fatal(res, err, code)
		} else {
			t.Log(res)
		}
		if len(res) < 3 {
			t.Error(res)
			return
		}

		if res[0].Id != "urn:infai:ses:measuring-function:23" {
			t.Fatal("error id")
		}
		if res[0].Name != "getOnOffFunction" {
			t.Fatal("error Name")
		}
		if res[0].ConceptId != "" {
			t.Fatal("error ConceptId")
		}

		if res[1].Id != "urn:infai:ses:measuring-function:321" {
			t.Fatal("error id")
		}
		if res[1].Name != "getTemperatureFunction" {
			t.Fatal("error Name")
		}
		if res[1].ConceptId != "urn:infai:ses:concept:efffsdfd-aaaa-bbbb-ccc-0000" {
			t.Fatal("error ConceptId")
		}

		if res[2].Id != "urn:infai:ses:measuring-function:467" {
			t.Fatal("error id", res[0].Id)
		}
		if res[2].Name != "getHumidityFunction" {
			t.Fatal("error Name")
		}
		if res[2].DisplayName != "hum_display 2" {
			t.Fatal("error Name", res[0].DisplayName)
		}
		if res[2].ConceptId != "" {
			t.Fatal("error ConceptId")
		}
	}
}

func testFunctionDelete(conf configuration.Config) func(t *testing.T) {
	return func(t *testing.T) {
		funcids := [6]string{
			"urn:infai:ses:controlling-function:333",
			"urn:infai:ses:controlling-function:2222",
			"urn:infai:ses:controlling-function:5467567",
			"urn:infai:ses:measuring-function:23",
			"urn:infai:ses:measuring-function:321",
			"urn:infai:ses:measuring-function:467"}

		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		for _, funcid := range funcids {
			err, _ := c.DeleteFunction(testenv.AdminToken, funcid)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
