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

var hub1id = uuid.NewV4().String()
var hub1name = uuid.NewV4().String()
var hub1hash = uuid.NewV4().String()

func TestHubQuery(t *testing.T) {
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

	err = producer.PublishHub(model.GatewayFlat{Id: hub1id, Name: hub1name, Hash: hub1hash, Devices: []string{device1id}}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)

	t.Run("read", func(t *testing.T) {
		testHubRead(t, conf)
	})
	t.Run("readName", func(t *testing.T) {
		testHubReadName(t, conf)
	})
	t.Run("readHash", func(t *testing.T) {
		testHubReadHash(t, conf)
	})
	t.Run("readDevices", func(t *testing.T) {
		testHubReadDevices(t, conf)
	})
	t.Run("readDevicesAsId", func(t *testing.T) {
		testHubReadDevicesAs(t, conf, "id", device1id)
	})
	t.Run("readDevicesAsUri", func(t *testing.T) {
		testHubReadDevicesAs(t, conf, "uri", device1uri)
	})
	t.Run("readDevicesAsUrl", func(t *testing.T) {
		testHubReadDevicesAs(t, conf, "url", device1uri)
	})
}

func testHubRead(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(hub1id)
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
	result := model.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != hub1name || result.Hash != hub1hash || result.Id != hub1id || !reflect.DeepEqual(result.Devices, []string{device1uri}) {
		t.Error("unexpected result", result)
		return
	}
}

func testHubReadName(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(hub1id) + "/name"
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
	result := ""
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result != hub1name {
		t.Error("unexpected result", result)
		return
	}
}

func testHubReadHash(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(hub1id) + "/hash"
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
	result := ""
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result != hub1hash {
		t.Error("unexpected result", result)
		return
	}
}

func testHubReadDevices(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(hub1id) + "/devices"
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
	result := []string{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, []string{device1uri}) {
		t.Error("unexpected result", result)
		return
	}
}

func testHubReadDevicesAs(t *testing.T, conf config.Config, as string, asResult string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(hub1id) + "/devices?as=" + as
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
	result := []string{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, []string{asResult}) {
		t.Error("unexpected result", result)
		return
	}
}

func TestHubDeviceSideEffects(t *testing.T) {
	t.Skip("not implemented")
}

func TestHubCommand(t *testing.T) {
	t.Skip("not implemented")
}
