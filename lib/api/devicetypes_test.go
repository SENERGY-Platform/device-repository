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

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

var devicetype1id = uuid.NewV4().String()
var devicetype1name = uuid.NewV4().String()
var devicetype2id = uuid.NewV4().String()
var devicetype2name = uuid.NewV4().String()

func TestDeviceTypeQuery(t *testing.T) {
	closer, conf, producer, err := createTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	if true {
		defer closer()
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
	time.Sleep(3 * time.Second)

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
}

func testDeviceTypeRead(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types/" + url.PathEscape(devicetype1id)
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := model.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != devicetype1name {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceTypeList(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types"
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []model.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 21 {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceTypeListLimit10(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types?limit=10"
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []model.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 10 {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceTypeListLimit10Offset20(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-types?limit=10&offset=20"
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []model.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceTypeListSort(t *testing.T, config config.Config) {
	defaultendpoint := "http://localhost:" + config.ServerPort + "/device-types?sort=name"
	resp, err := userjwt.Get(defaultendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", defaultendpoint, resp.Status, resp.StatusCode, string(b))
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
	resp, err = userjwt.Get(ascendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", ascendpoint, resp.Status, resp.StatusCode, string(b))
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
	resp, err = userjwt.Get(descendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", descendpoint, resp.Status, resp.StatusCode, string(b))
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

func TestDeviceTypeControl(t *testing.T) {
	t.Skip("not implemented")
	closer, conf, producer, err := createTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	if true {
		defer closer()
	}
	err = producer.PublishDeviceType(model.DeviceType{Id: devicetype1id, Name: devicetype1name}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)

	t.Run("testDeviceTypeCreate", func(t *testing.T) {
		testDeviceTypeCreate(t, conf)
	})
	t.Run("testDeviceTypeUpdate", func(t *testing.T) {
		testDeviceTypeUpdate(t, conf)
	})
	t.Run("testDeviceTypeDelete", func(t *testing.T) {
		testDeviceTypeDelete(t, conf)
	})
}

func testDeviceTypeCreate(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}

func testDeviceTypeUpdate(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}

func testDeviceTypeDelete(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}
