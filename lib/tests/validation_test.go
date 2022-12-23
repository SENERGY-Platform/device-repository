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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	"github.com/SENERGY-Platform/models/go/models"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestProtocolConstraintInValidation(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}
	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishProtocol(models.Protocol{
		Id:      "p1",
		Name:    "p1",
		Handler: "p1",
		ProtocolSegments: []models.ProtocolSegment{{
			Id:   "segment",
			Name: "segment",
		}},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishProtocol(models.Protocol{
		Id:      "p2",
		Name:    "p2",
		Handler: "p2",
		ProtocolSegments: []models.ProtocolSegment{{
			Id:   "segment",
			Name: "segment",
		}},
		Constraints: []string{model.SenergyConnectorLocalIdConstraint},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishDeviceType(models.DeviceType{
		Id:          "dt1",
		Name:        "dt1",
		Description: "",
		Services: []models.Service{{
			Id:          "dt1s1",
			LocalId:     "dt1s1",
			Name:        "dt1s1",
			Description: "",
			Interaction: models.REQUEST,
			ProtocolId:  "p1",
		}},
		DeviceClassId: "dc1",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishDeviceType(models.DeviceType{
		Id:          "dt2",
		Name:        "dt2",
		Description: "",
		Services: []models.Service{{
			Id:          "dt2s1",
			LocalId:     "dt2s1",
			Name:        "dt2s1",
			Description: "",
			Interaction: models.REQUEST,
			ProtocolId:  "p2",
		}},
		DeviceClassId: "dc1",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	t.Run("device-type unconstrained ok", testRequest(conf, "PUT", "/device-types?dry-run=true", models.DeviceType{
		Id:          "dt",
		Name:        "dt",
		Description: "",
		Services: []models.Service{{
			Id:          "dts1",
			LocalId:     "foo/bar",
			Name:        "dts1",
			Description: "",
			Interaction: models.REQUEST,
			ProtocolId:  "p1",
		}},
		DeviceClassId: "dc1",
	}, http.StatusOK, nil))

	t.Run("device-type constrained error", testRequest(conf, "PUT", "/device-types?dry-run=true", models.DeviceType{
		Id:          "dt",
		Name:        "dt",
		Description: "",
		Services: []models.Service{{
			Id:          "dts1",
			LocalId:     "foo/bar",
			Name:        "dts1",
			Description: "",
			Interaction: models.REQUEST,
			ProtocolId:  "p2",
		}},
		DeviceClassId: "dc1",
	}, http.StatusBadRequest, nil))

	t.Run("device-type constrained ok", testRequest(conf, "PUT", "/device-types?dry-run=true", models.DeviceType{
		Id:          "dt",
		Name:        "dt",
		Description: "",
		Services: []models.Service{{
			Id:          "dts1",
			LocalId:     "foobar",
			Name:        "dts1",
			Description: "",
			Interaction: models.REQUEST,
			ProtocolId:  "p2",
		}},
		DeviceClassId: "dc1",
	}, http.StatusOK, nil))

	t.Run("device unconstrained ok", testRequest(conf, "PUT", "/devices?dry-run=true", models.Device{
		Id:           "d",
		Name:         "d",
		LocalId:      "foo/bar",
		DeviceTypeId: "dt1",
	}, http.StatusOK, nil))

	t.Run("device constrained error", testRequest(conf, "PUT", "/devices?dry-run=true", models.Device{
		Id:           "d",
		Name:         "d",
		LocalId:      "foo/bar",
		DeviceTypeId: "dt2",
	}, http.StatusBadRequest, nil))

	t.Run("device constrained ok", testRequest(conf, "PUT", "/devices?dry-run=true", models.Device{
		Id:           "d",
		Name:         "d",
		LocalId:      "foobar",
		DeviceTypeId: "dt2",
	}, http.StatusOK, nil))
}

func TestDeleteValidations(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}
	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishDeviceClass(models.DeviceClass{
		Id:   "used_device_class",
		Name: "used_device_class",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	err = producer.PublishDeviceClass(models.DeviceClass{
		Id:   "unused_device_class",
		Name: "unused_device_class",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishCharacteristic(models.Characteristic{
		Id:   "used_characteristic",
		Name: "used_characteristic",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishCharacteristic(models.Characteristic{
		Id:   "used_characteristic_2",
		Name: "used_characteristic_2",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishCharacteristic(models.Characteristic{
		Id:   "unused_characteristic",
		Name: "unused_characteristic",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishFunction(models.Function{
		Id:        model.CONTROLLING_FUNCTION_PREFIX + "used_function",
		Name:      "used_function",
		ConceptId: "used_concept",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishFunction(models.Function{
		Id:        model.MEASURING_FUNCTION_PREFIX + "used_function_2",
		Name:      "used_function_2",
		ConceptId: "used_concept",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishFunction(models.Function{
		Id:        model.MEASURING_FUNCTION_PREFIX + "unused_function_2",
		Name:      "unused_function_2",
		ConceptId: "used_concept_2",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishFunction(models.Function{
		Id:        model.CONTROLLING_FUNCTION_PREFIX + "unused_function",
		Name:      "unused_function",
		ConceptId: "used_concept_2",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishConcept(models.Concept{
		Id:   "used_concept",
		Name: "used_concept",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishConcept(models.Concept{
		Id:   "used_concept_2",
		Name: "used_concept_2",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishConcept(models.Concept{
		Id:   "unused_concept",
		Name: "unused_concept",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishAspect(models.Aspect{
		Id:   model.URN_PREFIX + "used_root_aspect",
		Name: "used_root_aspect",
		SubAspects: []models.Aspect{
			{
				Id:   model.URN_PREFIX + "sub1",
				Name: "sub1",
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishAspect(models.Aspect{
		Id:   model.URN_PREFIX + "root_aspect",
		Name: "root_aspect",
		SubAspects: []models.Aspect{
			{
				Id:   model.URN_PREFIX + "used_aspect",
				Name: "used_aspect",
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishAspect(models.Aspect{
		Id:   model.URN_PREFIX + "unused_root_aspect",
		Name: "unused_root_aspect",
		SubAspects: []models.Aspect{
			{
				Id:   model.URN_PREFIX + "unused_used_aspect",
				Name: "unused_used_aspect",
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishDeviceType(models.DeviceType{Id: devicetype1id, Name: devicetype1name,
		DeviceClassId: "used_device_class",
		Services: []models.Service{
			{
				Id:          "s1",
				LocalId:     "s1",
				Name:        "s1",
				Interaction: models.EVENT_AND_REQUEST,
				ProtocolId:  "pid",
				Inputs: []models.Content{
					{
						Id: "input",
						ContentVariable: models.ContentVariable{
							Id:               "c1",
							Name:             "c1",
							Type:             "string",
							CharacteristicId: "used_characteristic",
							FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "used_function",
							AspectId:         model.URN_PREFIX + "used_aspect",
						},
						Serialization:     "json",
						ProtocolSegmentId: "s",
					},
				},
				Outputs: []models.Content{
					{
						Id: "output",
						ContentVariable: models.ContentVariable{
							Id:               "c2",
							Name:             "c2",
							Type:             "string",
							CharacteristicId: "used_characteristic_2",
							FunctionId:       model.MEASURING_FUNCTION_PREFIX + "used_function_2",
							AspectId:         model.URN_PREFIX + "used_root_aspect",
						},
						Serialization:     "json",
						ProtocolSegmentId: "s2",
					},
				},
			},
		}}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(10 * time.Second)

	t.Run("used_device_class", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"device-classes",
			"used_device_class",
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("unused_device_class", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"device-classes",
			"unused_device_class",
			http.StatusOK)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("used_characteristic", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"characteristics",
			"used_characteristic",
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("used_characteristic_2", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"characteristics",
			"used_characteristic_2",
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("unused_characteristic", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"characteristics",
			"unused_characteristic",
			http.StatusOK)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("used_function", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"functions",
			model.CONTROLLING_FUNCTION_PREFIX+"used_function",
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("used_function_2", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"functions",
			model.MEASURING_FUNCTION_PREFIX+"used_function_2",
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("unused_function_2", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"functions",
			model.MEASURING_FUNCTION_PREFIX+"unused_function_2",
			http.StatusOK)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("unused_function", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"functions",
			model.CONTROLLING_FUNCTION_PREFIX+"unused_function",
			http.StatusOK)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("used_concept", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"concepts",
			"used_concept",
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("used_concept_2", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"concepts",
			"used_concept_2",
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("unused_concept", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"concepts",
			"unused_concept",
			http.StatusOK)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("used_root_aspect", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"aspects",
			model.URN_PREFIX+"used_root_aspect",
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("root_aspect", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"aspects",
			model.URN_PREFIX+"root_aspect",
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("unused_root_aspect", func(t *testing.T) {
		err = testDeleteValidation(
			t,
			conf,
			"aspects",
			model.URN_PREFIX+"unused_root_aspect",
			http.StatusOK)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("sub1", func(t *testing.T) {
		err = testAspectValidation(
			t,
			conf,
			models.Aspect{
				Id:   model.URN_PREFIX + "used_root_aspect",
				Name: "used_root_aspect",
			},
			http.StatusOK)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("used_aspect", func(t *testing.T) {
		err = testAspectValidation(
			t,
			conf,
			models.Aspect{
				Id:   model.URN_PREFIX + "root_aspect",
				Name: "root_aspect",
			},
			http.StatusBadRequest)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("unused_used_aspect", func(t *testing.T) {
		err = testAspectValidation(
			t,
			conf,
			models.Aspect{
				Id:   model.URN_PREFIX + "unused_root_aspect",
				Name: "unused_root_aspect",
			},
			http.StatusOK)
		if err != nil {
			t.Error(err)
			return
		}
	})
}

func testDeleteValidation(t *testing.T, config config.Config, resource string, id string, expectedCode int) error {
	t.Helper()
	req, err := http.NewRequest("DELETE", "http://localhost:"+config.ServerPort+"/"+resource+"/"+id+"?dry-run=true", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != expectedCode {
		temp, _ := io.ReadAll(resp.Body)
		t.Log(string(temp))
		return errors.New(resp.Status)
	}
	return nil
}

func testAspectValidation(t *testing.T, config config.Config, aspect models.Aspect, expectedCode int) error {
	t.Helper()
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(aspect)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", "http://localhost:"+config.ServerPort+"/aspects?dry-run=true", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != expectedCode {
		temp, _ := io.ReadAll(resp.Body)
		t.Log(string(temp))
		return errors.New(resp.Status)
	}
	return nil
}
