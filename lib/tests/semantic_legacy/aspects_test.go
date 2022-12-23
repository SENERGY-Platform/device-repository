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

package semantic_legacy

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/semantic_legacy/producer"
	"github.com/SENERGY-Platform/models/go/models"
	"sync"
	"testing"
	"time"
)

func TestAspects(t *testing.T) {
	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()
	conf, ctrl, prod, err := NewPartialMockEnv(ctx, wg, conf, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("produce aspect", testProduceAspect(prod))
	t.Run("read aspect", testAspectRead(ctrl))
	t.Run("delete aspect", testAspectDelete(prod))
	t.Run("produce device-type with aspect", testProduceDeviceTypeForAspectTest(prod))
	t.Run("read aspect measuring-functions", testReadAspectMeasuringFunctions(ctrl))
}

func TestAspects2(t *testing.T) {
	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()
	conf, ctrl, prod, err := NewPartialMockEnv(ctx, wg, conf, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("test_2_ProduceDeviceTypeforAspectTest", test_2_ProduceDeviceTypeforAspectTest(prod))
	time.Sleep(2 * time.Second)
	t.Run("test_2_ReadAspectsWithMeasuringFunctions", test_2_ReadAspectsWithMeasuringFunctions(ctrl))
}

func testProduceAspect(producer *producer.Producer) func(t *testing.T) {
	return func(t *testing.T) {
		aspect := models.Aspect{}
		aspect.Id = "urn:infai:ses:aspect:eb4a4449-01a1-4434-9dcc-064b3955abcf"
		aspect.Name = "Air"
		err := producer.PublishAspect(aspect, "sdfdsfsf")
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testAspectRead(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		res, err, code := con.GetAspects()
		if err != nil {
			t.Fatal(res, err, code)
		} else {
			//t.Log(res)
		}
		if res[0].Id != "urn:infai:ses:aspect:eb4a4449-01a1-4434-9dcc-064b3955abcf" {
			t.Fatal("error id", res[0].Id)
		}
		if res[0].Name != "Air" {
			t.Fatal("error Name")
		}
	}
}

func testAspectDelete(producer *producer.Producer) func(t *testing.T) {
	return func(t *testing.T) {
		err := producer.PublishAspectDelete("urn:infai:ses:aspect:eb4a4449-01a1-4434-9dcc-064b3955abcf", "sdfdsfsf")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testProduceDeviceTypeForAspectTest(producer *producer.Producer) func(t *testing.T) {
	return func(t *testing.T) {
		devicetype := models.DeviceType{}
		devicetype.Id = "urn:infai:ses:devicetype:1e1e-AspectTest"
		devicetype.Name = "Philips Hue Color"
		devicetype.DeviceClassId = "urn:infai:ses:deviceclass:2e2e-AspectTest"
		devicetype.Description = "description"
		devicetype.Services = []models.Service{}
		devicetype.Services = append(devicetype.Services, models.Service{
			Id:         "urn:infai:ses:service:3e3e-AspectTest",
			LocalId:    "localId",
			Name:       "setBrightness",
			ProtocolId: "urn:infai:ses:protocol:asdasda",
			Inputs: []models.Content{
				{
					ContentVariable: models.ContentVariable{
						SubContentVariables: []models.ContentVariable{
							{
								AspectId:   "urn:infai:ses:aspect:4e4e-AspectTest",
								FunctionId: "urn:infai:ses:controlling-function:5e5e1-AspectTest",
							},
							{
								AspectId:   "urn:infai:ses:aspect:4e4e-AspectTest",
								FunctionId: "urn:infai:ses:controlling-function:5e5e2-AspectTest",
							},
						},
					},
					Serialization:     models.JSON,
					ProtocolSegmentId: "",
				},
			},
		})
		devicetype.Services = append(devicetype.Services, models.Service{
			Id:         "urn:infai:ses:service:3f3f-AspectTest",
			LocalId:    "localId",
			Name:       "getBrightness",
			ProtocolId: "urn:infai:ses:protocol:asdasda",
			Outputs: []models.Content{
				{
					ContentVariable: models.ContentVariable{
						SubContentVariables: []models.ContentVariable{
							{
								AspectId:   "urn:infai:ses:aspect:4e4e-AspectTest",
								FunctionId: "urn:infai:ses:measuring-function:5e5e3-AspectTest",
							},
							{
								AspectId:   "urn:infai:ses:aspect:4e4e-AspectTest",
								FunctionId: "urn:infai:ses:measuring-function:5e5e4-AspectTest",
							},
						},
					},
					Serialization:     models.JSON,
					ProtocolSegmentId: "",
				},
			},
		})

		err := producer.PublishDeviceType(devicetype, "sdfdsfsf")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishAspect(models.Aspect{Id: "urn:infai:ses:aspect:4e4e-AspectTest", Name: "Lighting"}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishFunction(models.Function{Id: "urn:infai:ses:controlling-function:5e5e1-AspectTest", Name: "brightnessAdjustment1", ConceptId: "urn:infai:ses:concept:1a1a1a", RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishFunction(models.Function{Id: "urn:infai:ses:controlling-function:5e5e2-AspectTest", Name: "brightnessAdjustment2", ConceptId: "urn:infai:ses:concept:1a1a1a", RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishFunction(models.Function{Id: "urn:infai:ses:measuring-function:5e5e3-AspectTest", Name: "brightnessFunction4", ConceptId: "urn:infai:ses:concept:1a1a1a", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishFunction(models.Function{Id: "urn:infai:ses:measuring-function:5e5e4-AspectTest", Name: "brightnessFunction2", ConceptId: "urn:infai:ses:concept:1a1a1a", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Second)
	}
}

func testReadAspectMeasuringFunctions(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		res, err, code := con.GetAspectNodesMeasuringFunctions("urn:infai:ses:aspect:4e4e-AspectTest", false, false)
		if err != nil {
			t.Fatal(res, err, code)
		} else {
			t.Log(res)
		}

		if res[0].Id != "urn:infai:ses:measuring-function:5e5e3-AspectTest" {
			t.Fatal("error id", res[0].Id)
		}
		if res[0].Name != "brightnessFunction4" {
			t.Fatal("error Name")
		}
		if res[0].ConceptId != "urn:infai:ses:concept:1a1a1a" {
			t.Fatal("error ConceptId")
		}
		if res[0].RdfType != model.SES_ONTOLOGY_MEASURING_FUNCTION {
			t.Fatal("wrong RdfType")
		}

		if res[1].Id != "urn:infai:ses:measuring-function:5e5e4-AspectTest" {
			t.Fatal("error id", res[1].Id)
		}
		if res[1].Name != "brightnessFunction2" {
			t.Fatal("error Name", res[1].Name)
		}
		if res[1].ConceptId != "urn:infai:ses:concept:1a1a1a" {
			t.Fatal("error ConceptId", res[1].ConceptId)
		}
		if res[1].RdfType != model.SES_ONTOLOGY_MEASURING_FUNCTION {
			t.Fatal("wrong RdfType")
		}

	}
}

func test_2_ProduceDeviceTypeforAspectTest(producer *producer.Producer) func(t *testing.T) {
	return func(t *testing.T) {
		devicetype := models.DeviceType{}
		devicetype.Id = "urn:infai:ses:devicetype:08-01-20"
		devicetype.Name = "Philips Hue Color"
		devicetype.DeviceClassId = "urn:infai:ses:deviceclass:08-01-20"
		devicetype.Description = "description"
		devicetype.Services = []models.Service{}

		devicetype.Services = append(devicetype.Services, models.Service{
			Id:         "urn:infai:ses:service:08-01-20_1",
			LocalId:    "localId",
			Name:       "setBrightness",
			ProtocolId: "urn:infai:ses:protocol:asdasda",
			Inputs: []models.Content{
				{
					ContentVariable: models.ContentVariable{
						SubContentVariables: []models.ContentVariable{
							{
								AspectId:   "urn:infai:ses:aspect:08-01-20_1",
								FunctionId: "urn:infai:ses:controlling-function:08-01-20_1",
							},
							{
								AspectId:   "urn:infai:ses:aspect:08-01-20_1",
								FunctionId: "urn:infai:ses:controlling-function:08-01-20_2",
							},
						},
					},
					Serialization:     models.JSON,
					ProtocolSegmentId: "",
				},
			},
		})
		devicetype.Services = append(devicetype.Services, models.Service{
			Id:         "urn:infai:ses:service:08-01-20_2",
			LocalId:    "localId",
			Name:       "setBrightness",
			ProtocolId: "urn:infai:ses:protocol:asdasda",
			Outputs: []models.Content{
				{
					ContentVariable: models.ContentVariable{
						SubContentVariables: []models.ContentVariable{
							{
								AspectId:   "urn:infai:ses:aspect:08-01-20_2",
								FunctionId: "urn:infai:ses:measuring-function:08-01-20_3",
							},
							{
								AspectId:   "urn:infai:ses:aspect:08-01-20_2",
								FunctionId: "urn:infai:ses:measuring-function:08-01-20_4",
							},
						},
					},
					Serialization:     models.JSON,
					ProtocolSegmentId: "",
				},
			},
		})

		err := producer.PublishDeviceType(devicetype, "sdfdsfsf")

		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishAspect(models.Aspect{Id: "urn:infai:ses:aspect:08-01-20_1", Name: "aspect1"}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishAspect(models.Aspect{Id: "urn:infai:ses:aspect:08-01-20_2", Name: "aspect2"}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishFunction(models.Function{Id: "urn:infai:ses:controlling-function:08-01-20_1", Name: "func1", ConceptId: "urn:infai:ses:concept:1a1a1a", RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishFunction(models.Function{Id: "urn:infai:ses:controlling-function:08-01-20_2", Name: "func2", ConceptId: "urn:infai:ses:concept:1a1a1a", RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishFunction(models.Function{Id: "urn:infai:ses:measuring-function:08-01-20_3", Name: "func3", ConceptId: "urn:infai:ses:concept:1a1a1a", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
		err = producer.PublishFunction(models.Function{Id: "urn:infai:ses:measuring-function:08-01-20_4", Name: "func4", ConceptId: "urn:infai:ses:concept:1a1a1a", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION}, "sdfsdfdsds")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func test_2_ReadAspectsWithMeasuringFunctions(con *controller.Controller) func(t *testing.T) {
	return func(t *testing.T) {
		res, err, code := con.GetAspectsWithMeasuringFunction(false, false)
		if err != nil {
			t.Fatal(res, err, code)
		} else {
			t.Log(res)
		}
		if len(res) == 0 {
			t.Fatal(res)
		}
		if res[0].Id != "urn:infai:ses:aspect:08-01-20_2" {
			t.Fatal("error id", res[0].Id)
		}
		if res[0].Name != "aspect2" {
			t.Fatal("error Name")
		}
	}
}
