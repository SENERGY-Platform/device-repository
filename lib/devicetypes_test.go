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

package lib

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"
)

var devicetype1id = "urn:infai:ses:device-type:2cc43032-207e-494e-8de4-94784cd4961d"
var devicetype1name = uuid.NewV4().String()
var devicetype2id = uuid.NewV4().String()
var devicetype2name = uuid.NewV4().String()

func TestServiceQuery(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	producer, err := NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishDeviceType(model.DeviceType{Id: devicetype1id, Name: devicetype1name, Services: []model.Service{{Id: "service_42", Name: "foo"}}}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("testServiceRead", func(t *testing.T) {
		testServiceRead(t, conf, model.Service{Id: "service_42", Name: "foo"})
	})
}

func TestSubContentVarUpdate(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	producer, err := NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	dt := model.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []model.Service{{
			Id:   "service1",
			Name: "serviceName",
			Outputs: []model.Content{{
				Id: "content",
				ContentVariable: model.ContentVariable{
					Id:   "main",
					Name: "main",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{{
						Id:   "sub",
						Name: "sub",
						Type: model.String,
					}},
				},
				Serialization:     "json",
				ProtocolSegmentId: "payload",
			}},
		}},
	}

	err = producer.PublishDeviceType(dt, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("after create", testDeviceTypeReadV2(conf, dt))

	dtUpdated := model.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []model.Service{{
			Id:   "service1",
			Name: "serviceName",
			Outputs: []model.Content{{
				Id: "content",
				ContentVariable: model.ContentVariable{
					Id:   "main",
					Name: "main",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{{
						Id:   "sub2",
						Name: "sub2",
						Type: model.Integer,
					}},
				},
				Serialization:     "json",
				ProtocolSegmentId: "payload",
			}},
		}},
	}

	err = producer.PublishDeviceType(dtUpdated, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("after update", testDeviceTypeReadV2(conf, dtUpdated))

	dtSubVarDeleted := model.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []model.Service{{
			Id:   "service1",
			Name: "serviceName",
			Outputs: []model.Content{{
				Id: "content",
				ContentVariable: model.ContentVariable{
					Id:   "main",
					Name: "main",
					Type: model.Structure,
				},
				Serialization:     "json",
				ProtocolSegmentId: "payload",
			}},
		}},
	}

	err = producer.PublishDeviceType(dtSubVarDeleted, userid)
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
	conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	producer, err := NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishDeviceType(model.DeviceType{Id: devicetype1id, Name: devicetype1name}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 20; i++ {
		err = producer.PublishDeviceType(model.DeviceType{Id: uuid.NewV4().String(), Name: uuid.NewV4().String()}, userid)
		if err != nil {
			t.Error(err)
			return
		}
	}
	time.Sleep(10 * time.Second)

	t.Run("unexisting", func(t *testing.T) {
		testDeviceTypeReadNotFound(t, conf, uuid.NewV4().String())
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

	err = producer.PublishDeviceType(model.DeviceType{Id: devicetype1id, Name: devicetype1name, Services: []model.Service{{Id: "service_42", Name: "foo"}}}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("testServiceRead", func(t *testing.T) {
		testServiceRead(t, conf, model.Service{Id: "service_42", Name: "foo"})
	})
}

func TestDeviceTypeWithAttribute(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	producer, err := NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	dt := model.DeviceType{Id: devicetype1id, Name: devicetype1name, Attributes: []model.Attribute{{Key: "foo", Value: "bar"}}}

	err = producer.PublishDeviceType(dt, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(5 * time.Second)

	t.Run("testDeviceTypeRead", testDeviceTypeReadV2(conf, dt))

}

func TestServiceWithAttribute(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	producer, err := NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	dt := model.DeviceType{Id: devicetype1id, Name: devicetype1name, Attributes: []model.Attribute{{Key: "foo", Value: "bar"}}, Services: []model.Service{
		{
			Id:          "sid1",
			LocalId:     "lsid1",
			Name:        "s",
			Description: "s",
			Interaction: model.EVENT,
			AspectIds:   []string{"aid1"},
			ProtocolId:  "pid1",
			Inputs:      nil,
			Outputs:     nil,
			FunctionIds: []string{"fid1"},
			Attributes:  []model.Attribute{{Key: "batz", Value: "blub"}},
		},
	}}

	err = producer.PublishDeviceType(dt, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	t.Run("testDeviceTypeRead", testDeviceTypeReadV2(conf, dt))

}

func testDeviceTypeRead(t *testing.T, conf config.Config, expectedDt ...model.DeviceType) {
	expected := model.DeviceType{Id: devicetype1id, Name: devicetype1name}
	if len(expectedDt) > 0 {
		expected = expectedDt[0]
	}
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types/" + url.PathEscape(expected.Id)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := model.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != expected.Name {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceTypeReadV2(conf config.Config, expected model.DeviceType) func(t *testing.T) {
	return func(t *testing.T) {
		endpoint := "http://localhost:" + conf.ServerPort + "/device-types/" + url.PathEscape(expected.Id)
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", userjwt)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := ioutil.ReadAll(resp.Body)
			t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
			return
		}
		result := model.DeviceType{}
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

func testServiceRead(t *testing.T, conf config.Config, expected model.Service) {
	endpoint := "http://localhost:" + conf.ServerPort + "/services/" + url.PathEscape(expected.Id)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := model.Service{}
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
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []model.DeviceType{}
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
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []model.DeviceType{}
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
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []model.DeviceType{}
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
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", defaultendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	defaultresult := []model.DeviceType{}
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
	req.Header.Set("Authorization", userjwt)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", ascendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	ascresult := []model.DeviceType{}
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
	req.Header.Set("Authorization", userjwt)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", descendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	descresult := []model.DeviceType{}
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
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusNotFound {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
}
