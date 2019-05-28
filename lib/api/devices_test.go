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
	"testing"
	"time"
)

var devicetype1id = uuid.NewV4().String()
var devicetype1name = uuid.NewV4().String()
var device1id = uuid.NewV4().String()
var device1name = uuid.NewV4().String()
var device1uri = uuid.NewV4().String()
var device2id = uuid.NewV4().String()
var device2name = uuid.NewV4().String()
var device2uri = uuid.NewV4().String()

func TestDeviceQuery(t *testing.T) {
	t.Parallel()
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
	err = producer.PublishDevice(model.DeviceInstance{Id: device1id, Name: device1name, Url: device1uri, DeviceType: devicetype1id}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 20; i++ {
		err = producer.PublishDevice(model.DeviceInstance{Id: uuid.NewV4().String(), Name: uuid.NewV4().String(), Url: uuid.NewV4().String(), DeviceType: devicetype1id}, userid)
		if err != nil {
			t.Error(err)
			return
		}
	}
	time.Sleep(3 * time.Second)

	t.Run("testHeartbeat", func(t *testing.T) {
		testHeartbeat(t, conf)
	})
	t.Run("testDeviceRead", func(t *testing.T) {
		testDeviceRead(t, conf)
	})
	t.Run("testDeviceList", func(t *testing.T) {
		testDeviceList(t, conf)
	})
	t.Run("testDeviceListLimit10", func(t *testing.T) {
		testDeviceListLimit10(t, conf)
	})
	t.Run("testDeviceListLimit10Offset20", func(t *testing.T) {
		testDeviceListLimit10Offset20(t, conf)
	})
	t.Run("testDeviceListSort", func(t *testing.T) {
		testDeviceListSort(t, conf)
	})
}

func testHeartbeat(t *testing.T, configuration config.Config) {
	resp, err := userjwt.Get("http://localhost:" + configuration.ServerPort)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("no heart beat")
		return
	}
}

func testDeviceRead(t *testing.T, configuration config.Config) {
	endpoint := "http://localhost:" + configuration.ServerPort + "/devices/" + url.PathEscape(device1id)
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
	result := model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != device1name || result.Url != device1uri {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceList(t *testing.T, configuration config.Config) {
	endpoint := "http://localhost:" + configuration.ServerPort + "/devices"
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
	result := []model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 21 {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceListLimit10(t *testing.T, configuration config.Config) {
	endpoint := "http://localhost:" + configuration.ServerPort + "/devices?limit=10"
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
	result := []model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 10 {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceListLimit10Offset20(t *testing.T, configuration config.Config) {
	endpoint := "http://localhost:" + configuration.ServerPort + "/devices?limit=10&offset=20"
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
	result := []model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceListSort(t *testing.T, configuration config.Config) {
	ascendpoint := "http://localhost:" + configuration.ServerPort + "/devices?sort=name.asc"
	resp, err := userjwt.Get(ascendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", ascendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	ascresult := []model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&ascresult)
	if err != nil {
		t.Error(err)
	}
	if len(ascresult) != 21 {
		t.Error("unexpected result", ascresult)
		return
	}

	descendpoint := "http://localhost:" + configuration.ServerPort + "/devices?sort=name.desc"
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
	descresult := []model.DeviceInstance{}
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

func TestDeviceControl(t *testing.T) {
	t.Skip("not implemented")
	t.Parallel()
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

	t.Run("testDeviceCreate", func(t *testing.T) {
		testDeviceCreate(t, conf)
	})
	t.Run("testDeviceUpdate", func(t *testing.T) {
		testDeviceUpdate(t, conf)
	})
	t.Run("testDeviceDelete", func(t *testing.T) {
		testDeviceDelete(t, conf)
	})
}

func testDeviceCreate(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}

func testDeviceUpdate(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}

func testDeviceDelete(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}
