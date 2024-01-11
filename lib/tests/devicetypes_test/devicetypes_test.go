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

package devicetypes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testenv"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"
)

var devicetype1id = "urn:infai:ses:device-type:2cc43032-207e-494e-8de4-94784cd4961d"
var devicetype1name = uuid.NewString()
var devicetype2id = uuid.NewString()
var devicetype2name = uuid.NewString()

func TestDeviceTypeSubAspectValidation(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t, func(c *config.Config) {
		c.AllowNoneLeafAspectNodesInDeviceTypes = true
	})
	if err != nil {
		t.Error(err)
		return
	}
	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishAspect(models.Aspect{
		Id:   "parent_2",
		Name: "parent_2",
		SubAspects: []models.Aspect{
			{
				Id:   "aid_2",
				Name: "aid_2",
				SubAspects: []models.Aspect{
					{
						Id:   "child_2",
						Name: "child_2",
					},
				},
			},
		},
	}, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishProtocol(models.Protocol{
		Id:      "p",
		Name:    "p",
		Handler: "p",
		ProtocolSegments: []models.ProtocolSegment{
			{
				Id:   "ps",
				Name: "ps",
			},
		},
	}, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(5 * time.Second)

	body, err := json.Marshal(models.DeviceType{
		Id:            "dt",
		Name:          "test",
		Description:   "",
		ServiceGroups: nil,
		Services: []models.Service{
			{
				Id:          "s",
				LocalId:     "sid",
				Name:        "s",
				Description: "",
				Interaction: models.REQUEST,
				ProtocolId:  "p",
				Inputs: []models.Content{
					{
						Id: "i",
						ContentVariable: models.ContentVariable{
							Id:                   "v",
							Name:                 "val",
							IsVoid:               false,
							Type:                 models.String,
							SubContentVariables:  nil,
							CharacteristicId:     "",
							Value:                nil,
							SerializationOptions: nil,
							UnitReference:        "",
							FunctionId:           "",
							AspectId:             "aid_2",
						},
						Serialization:     "json",
						ProtocolSegmentId: "ps",
					},
				},
				Outputs:         nil,
				Attributes:      nil,
				ServiceGroupKey: "",
			},
		},
		DeviceClassId: "",
		Attributes:    nil,
	})

	endpoint := "http://localhost:" + conf.ServerPort + "/device-types?dry-run=true"
	req, err := http.NewRequest("PUT", endpoint, bytes.NewReader(body))
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
}

func TestServiceQuery(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}
	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishDeviceType(models.DeviceType{Id: devicetype1id, Name: devicetype1name, Services: []models.Service{{Id: "service_42", Name: "foo"}}}, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("testServiceRead", func(t *testing.T) {
		testServiceRead(t, conf, models.Service{Id: "service_42", Name: "foo"})
	})
}

func TestSubContentVarUpdate(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	dt := models.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []models.Service{{
			Id:   "service1",
			Name: "serviceName",
			Outputs: []models.Content{{
				Id: "content",
				ContentVariable: models.ContentVariable{
					Id:   "main",
					Name: "main",
					Type: models.Structure,
					SubContentVariables: []models.ContentVariable{{
						Id:   "sub",
						Name: "sub",
						Type: models.String,
					}},
				},
				Serialization:     "json",
				ProtocolSegmentId: "payload",
			}},
		}},
	}

	err = producer.PublishDeviceType(dt, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("after create", testDeviceTypeReadV2(conf, dt))

	dtUpdated := models.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []models.Service{{
			Id:   "service1",
			Name: "serviceName",
			Outputs: []models.Content{{
				Id: "content",
				ContentVariable: models.ContentVariable{
					Id:   "main",
					Name: "main",
					Type: models.Structure,
					SubContentVariables: []models.ContentVariable{{
						Id:   "sub2",
						Name: "sub2",
						Type: models.Integer,
					}},
				},
				Serialization:     "json",
				ProtocolSegmentId: "payload",
			}},
		}},
	}

	err = producer.PublishDeviceType(dtUpdated, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("after update", testDeviceTypeReadV2(conf, dtUpdated))

	dtSubVarDeleted := models.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []models.Service{{
			Id:   "service1",
			Name: "serviceName",
			Outputs: []models.Content{{
				Id: "content",
				ContentVariable: models.ContentVariable{
					Id:   "main",
					Name: "main",
					Type: models.Structure,
				},
				Serialization:     "json",
				ProtocolSegmentId: "payload",
			}},
		}},
	}

	err = producer.PublishDeviceType(dtSubVarDeleted, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("after sub delete", testDeviceTypeReadV2(conf, dtSubVarDeleted))
}

func TestDeviceTypeQuery(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishDeviceType(models.DeviceType{Id: devicetype1id, Name: devicetype1name}, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 20; i++ {
		err = producer.PublishDeviceType(models.DeviceType{Id: uuid.NewString(), Name: uuid.NewString()}, testenv.Userid)
		if err != nil {
			t.Error(err)
			return
		}
	}
	time.Sleep(10 * time.Second)

	t.Run("unexisting", func(t *testing.T) {
		testDeviceTypeReadNotFound(t, conf, uuid.NewString())
	})
	t.Run("testDeviceTypeRead", func(t *testing.T) {
		testDeviceTypeRead(t, conf)
	})
	t.Run("testDeviceTypeList", func(t *testing.T) {
		testDeviceTypeList(t, conf)
	})
	t.Run("testDeviceTypeListLimit10", func(t *testing.T) {
		testDeviceTypeListLimit10(t, conf)
	})
	t.Run("testDeviceTypeListLimit10Offset20", func(t *testing.T) {
		testDeviceTypeListLimit10Offset20(t, conf)
	})
	t.Run("testDeviceTypeListSort", func(t *testing.T) {
		testDeviceTypeListSort(t, conf)
	})

	err = producer.PublishDeviceType(models.DeviceType{Id: devicetype1id, Name: devicetype1name, Services: []models.Service{{Id: "service_42", Name: "foo"}}}, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("testServiceRead", func(t *testing.T) {
		testServiceRead(t, conf, models.Service{Id: "service_42", Name: "foo"})
	})
}

func TestDeviceTypeWithServiceGroups(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	dt := models.DeviceType{Id: devicetype1id, Name: devicetype1name, ServiceGroups: []models.ServiceGroup{
		{
			Key:         "test",
			Name:        "test group",
			Description: "foo",
		},
	}, Services: []models.Service{
		{
			Id:              "s1",
			LocalId:         "s1",
			Name:            "n1",
			Interaction:     models.REQUEST,
			ProtocolId:      "p1",
			ServiceGroupKey: "test",
			Outputs: []models.Content{{
				ContentVariable: models.ContentVariable{
					FunctionId: "f1",
					AspectId:   "a1",
				},
			}},
		},
		{
			Id:              "s2",
			LocalId:         "s2",
			Name:            "n2",
			Interaction:     models.REQUEST,
			ProtocolId:      "p1",
			ServiceGroupKey: "",
			Outputs: []models.Content{{
				ContentVariable: models.ContentVariable{
					FunctionId: "f1",
					AspectId:   "a1",
				},
			}},
		},
	}}

	t.Run("validation of service-groups", func(t *testing.T) {
		err = controller.ValidateServiceGroups(dt.ServiceGroups, dt.Services)
		if err != nil {
			t.Error(err)
			return
		}
	})

	err = producer.PublishDeviceType(dt, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(5 * time.Second)

	t.Run("testDeviceTypeRead", testDeviceTypeReadV2(conf, dt))

}

func TestDeviceTypeWithAttribute(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	dt := models.DeviceType{Id: devicetype1id, Name: devicetype1name, Attributes: []models.Attribute{{Key: "foo", Value: "bar"}}}

	err = producer.PublishDeviceType(dt, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	t.Run("testDeviceTypeRead", testDeviceTypeReadV2(conf, dt))

}

func TestServiceWithAttribute(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := testenv.CreateTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	dt := models.DeviceType{Id: devicetype1id, Name: devicetype1name, Attributes: []models.Attribute{{Key: "foo", Value: "bar"}}, Services: []models.Service{
		{
			Id:          "sid1",
			LocalId:     "lsid1",
			Name:        "s",
			Description: "s",
			Interaction: models.EVENT,
			ProtocolId:  "pid1",
			Inputs:      nil,
			Attributes:  []models.Attribute{{Key: "batz", Value: "blub"}},
			Outputs: []models.Content{{
				ContentVariable: models.ContentVariable{
					FunctionId: "fid1",
					AspectId:   "aid1",
				},
			}},
		},
	}}

	err = producer.PublishDeviceType(dt, testenv.Userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	t.Run("testDeviceTypeRead", testDeviceTypeReadV2(conf, dt))

}

func testDeviceTypeRead(t *testing.T, conf config.Config, expectedDt ...models.DeviceType) {
	expected := models.DeviceType{Id: devicetype1id, Name: devicetype1name}
	if len(expectedDt) > 0 {
		expected = expectedDt[0]
	}
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types/" + url.PathEscape(expected.Id)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != expected.Name {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceTypeReadV2(conf config.Config, expected models.DeviceType) func(t *testing.T) {
	return func(t *testing.T) {
		endpoint := "http://localhost:" + conf.ServerPort + "/device-types/" + url.PathEscape(expected.Id)
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", testenv.Userjwt)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
			return
		}
		result := models.DeviceType{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(result, expected) {
			actual, _ := json.Marshal(result)
			expectedStr, _ := json.Marshal(expected)
			t.Error("unexpected result", string(actual), string(expectedStr))
			return
		}
	}
}

func testServiceRead(t *testing.T, conf config.Config, expected models.Service) {
	endpoint := "http://localhost:" + conf.ServerPort + "/services/" + url.PathEscape(expected.Id)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := models.Service{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != expected.Name {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceTypeList(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 21 {
		t.Error("unexpected result", len(result), result)
		return
	}
}

func testDeviceTypeListLimit10(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types?limit=10"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 10 {
		t.Error("unexpected result", len(result), result)
		return
	}
}

func testDeviceTypeListLimit10Offset20(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types?limit=10&offset=20"
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Error("unexpected result", len(result), result)
		return
	}
}

func testDeviceTypeListSort(t *testing.T, config config.Config) {
	defaultendpoint := "http://localhost:" + config.ServerPort + "/device-types?sort=name"
	req, err := http.NewRequest("GET", defaultendpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", defaultendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	defaultresult := []models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&defaultresult)
	if err != nil {
		t.Error(err)
	}
	if len(defaultresult) != 21 {
		t.Error("unexpected result", len(defaultresult))
		return
	}
	ascendpoint := "http://localhost:" + config.ServerPort + "/device-types?sort=name.asc"
	req, err = http.NewRequest("GET", ascendpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", ascendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	ascresult := []models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&ascresult)
	if err != nil {
		t.Error(err)
	}
	if len(ascresult) != 21 {
		t.Error("unexpected result", ascresult)
		return
	}
	if !reflect.DeepEqual(defaultresult, ascresult) {
		t.Error("unexpected result", defaultresult, ascresult)
		return
	}

	descendpoint := "http://localhost:" + config.ServerPort + "/device-types?sort=name.desc"
	req, err = http.NewRequest("GET", descendpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", descendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	descresult := []models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&descresult)
	if err != nil {
		t.Error(err)
	}
	if len(ascresult) != 21 {
		t.Error("unexpected result", descresult)
		return
	}

	for i := 0; i < 21; i++ {
		if descresult[i].Id != ascresult[20-i].Id {
			t.Error("unexpected sorting result", i, descresult[i].Id, ascresult[20-i].Id)
			return
		}
	}
}

func testDeviceTypeReadNotFound(t *testing.T, conf config.Config, id string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types/" + url.PathEscape(id)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", testenv.Userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusNotFound {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
}
