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

package semantic_legacy

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/tests/repo_legacy/testenv"
	"github.com/SENERGY-Platform/models/go/models"
	"sync"
	"testing"
)

func TestCharacteristics(t *testing.T) {
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

	t.Run("testProduceValidCharacteristic1", testProduceValidCharacteristic1(conf))
	t.Run("testProduceValidCharacteristic2", testProduceValidCharacteristic2(conf))
	t.Run("testReadCharacteristic1", testReadCharacteristic1(ctrl))
	t.Run("testReadAllCharacteristic", testReadAllCharacteristic(ctrl))
	t.Run("testDeleteCharacteristic1", testDeleteCharacteristic1(conf))
}

func TestSpecialCaseCharacteristics(t *testing.T) {
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

	t.Run("produce characteristic with special characters", testProduceValidCharacteristic3(conf))
	t.Run("read characteristic with special characters", testReadCharacteristic3(ctrl))
}

func testProduceValidCharacteristic1(conf configuration.Config) func(t *testing.T) {
	return func(t *testing.T) {
		characteristic := models.Characteristic{}
		characteristic.Id = "urn:ses:infai:characteristic:1d1e1f"
		characteristic.Name = "struct1"
		characteristic.DisplayUnit = "°C"
		characteristic.Type = models.Structure
		characteristic.SubCharacteristics = []models.Characteristic{{
			Id:                 "urn:infai:ses:characteristic:2r2r2r",
			MinValue:           -2,
			MaxValue:           3,
			Value:              2.2,
			Type:               models.Float,
			Name:               "charFloat",
			SubCharacteristics: nil,
		}, {
			Id:       "urn:infai:ses:characteristic:3t3t3t",
			MinValue: nil,
			MaxValue: nil,
			Value:    nil,
			Type:     models.Structure,
			Name:     "charInnerStructure1",
			SubCharacteristics: []models.Characteristic{
				{
					Id:                 "urn:infai:ses:characteristic:4z4z4z",
					MinValue:           nil,
					MaxValue:           nil,
					Value:              true,
					Type:               models.Boolean,
					Name:               "charBoolean",
					SubCharacteristics: nil}},
		}}
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetCharacteristic(testenv.AdminToken, characteristic)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testProduceValidCharacteristic2(conf configuration.Config) func(t *testing.T) {
	return func(t *testing.T) {
		characteristic := models.Characteristic{}
		characteristic.Id = "urn:ses:infai:characteristic:4711111-20.03.2020"
		characteristic.Name = "bool"
		characteristic.DisplayUnit = "°F"
		characteristic.Type = models.Boolean
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetCharacteristic(testenv.AdminToken, characteristic)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testProduceValidCharacteristic3(conf configuration.Config) func(t *testing.T) {
	return func(t *testing.T) {
		characteristic := models.Characteristic{}
		characteristic.Id = "urn:ses:infai:characteristic:1111111-30.03.2021"
		characteristic.Name = "µg/m³"
		characteristic.Type = models.Boolean
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetCharacteristic(testenv.AdminToken, characteristic)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testReadCharacteristic1(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		characteristic, err, _ := con.GetCharacteristic("urn:ses:infai:characteristic:1d1e1f")
		if err == nil {
			if characteristic.Id != "urn:ses:infai:characteristic:1d1e1f" {
				t.Fatal("wrong id", characteristic.Id)
			}
			if characteristic.Name != "struct1" {
				t.Fatal("wrong name")
			}
			if characteristic.DisplayUnit != "°C" {
				t.Fatal("wrong display unit:", characteristic.DisplayUnit)
			}
			if characteristic.Type != models.Structure {
				t.Fatal("wrong Type")
			}
			if characteristic.Value != nil {
				t.Fatal("wrong Value")
			}
			if characteristic.MaxValue != nil {
				t.Fatal("wrong MaxValue")
			}
			if characteristic.MinValue != nil {
				t.Fatal("wrong MinValue")
			}
			///////// index -> 0
			if characteristic.SubCharacteristics[0].Id != "urn:infai:ses:characteristic:2r2r2r" {
				t.Fatal("wrong id")
			}
			if characteristic.SubCharacteristics[0].Name != "charFloat" {
				t.Fatal("wrong id")
			}
			if characteristic.SubCharacteristics[0].Type != models.Float {
				t.Fatal("wrong id")
			}
			if characteristic.SubCharacteristics[0].Value != 2.2 {
				t.Fatal("wrong Value")
			}
			if characteristic.SubCharacteristics[0].MaxValue != 3.0 {
				t.Fatal("wrong MaxValue")
			}
			if characteristic.SubCharacteristics[0].MinValue != -2.0 {
				t.Fatal("wrong MinValue")
			}
			if characteristic.SubCharacteristics[0].SubCharacteristics != nil {
				t.Fatal("wrong SubCharacteristics")
			}
			///////// index -> 1
			if characteristic.SubCharacteristics[1].Id != "urn:infai:ses:characteristic:3t3t3t" {
				t.Fatal("wrong id")
			}
			if characteristic.SubCharacteristics[1].Name != "charInnerStructure1" {
				t.Fatal("wrong id")
			}
			if characteristic.SubCharacteristics[1].Type != models.Structure {
				t.Fatal("wrong id")
			}
			if characteristic.SubCharacteristics[1].Value != nil {
				t.Fatal("wrong Value")
			}
			if characteristic.SubCharacteristics[1].MaxValue != nil {
				t.Fatal("wrong MaxValue")
			}
			if characteristic.SubCharacteristics[1].MinValue != nil {
				t.Fatal("wrong MinValue")
			}
			///////// index -> 1 -> 0
			if characteristic.SubCharacteristics[1].SubCharacteristics[0].Id != "urn:infai:ses:characteristic:4z4z4z" {
				t.Fatal("wrong id")
			}
			if characteristic.SubCharacteristics[1].SubCharacteristics[0].Name != "charBoolean" {
				t.Fatal("wrong id")
			}
			if characteristic.SubCharacteristics[1].SubCharacteristics[0].Type != models.Boolean {
				t.Fatal("wrong id")
			}
			if characteristic.SubCharacteristics[1].SubCharacteristics[0].Value != true {
				t.Fatal("wrong Value")
			}
			if characteristic.SubCharacteristics[1].SubCharacteristics[0].MaxValue != nil {
				t.Fatal("wrong MaxValue")
			}
			if characteristic.SubCharacteristics[1].SubCharacteristics[0].MinValue != nil {
				t.Fatal("wrong MinValue")
			}
			t.Log(characteristic)
		} else {
			t.Fatal(err)
		}
	}
}

func testReadCharacteristic3(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		characteristic, err, _ := con.GetCharacteristic("urn:ses:infai:characteristic:1111111-30.03.2021")
		if err == nil {
			if characteristic.Id != "urn:ses:infai:characteristic:1111111-30.03.2021" {
				t.Fatal("wrong id", characteristic.Id)
			}
			if characteristic.Name != "µg/m³" {
				t.Fatal("wrong name", characteristic.Name)
			}
		} else {
			t.Fatal(err)
		}
	}
}

func testReadAllCharacteristic(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		characteristics, err, _ := con.GetCharacteristics(true)

		if err == nil {
			if len(characteristics) != 3 {
				t.Fatal("wrong number of elements")
			}
			t.Logf("%+v\n", characteristics)
			t.Log(len(characteristics))
		} else {
			t.Fatal(err)
		}
	}
}

func testDeleteCharacteristic1(conf configuration.Config) func(t *testing.T) {
	return func(t *testing.T) {
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		err, _ := c.DeleteCharacteristic(testenv.AdminToken, "urn:ses:infai:characteristic:1d1e1f")
		if err != nil {
			t.Fatal(err)
		}
	}
}
