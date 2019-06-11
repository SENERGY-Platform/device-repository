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
	"bytes"
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

	t.Run("head", func(t *testing.T) {
		testHubHead(t, conf)
	})
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
	t.Run("deviceWithHubRef", func(t *testing.T) {
		testDeviceWithHubRef(t, conf, device1id, hub1id)
	})
}

func testDeviceWithHubRef(t *testing.T, conf config.Config, deviceId string, hubId string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/devices/" + url.PathEscape(deviceId)
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Gateway != hubId {
		t.Error("unexpected result", result.Gateway, hubId)
		return
	}
}

func testHubHead(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(hub1id)
	resp, err := head(endpoint, string(userjwt))
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode)
		return
	}
	endpoint = "http://localhost:" + conf.ServerPort + "/hubs/foobar"
	resp, err = head(endpoint, string(userjwt))
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode)
		return
	}
}

func testHubRead(t *testing.T, conf config.Config, expectedHub ...model.Hub) {
	expected := model.Hub{Id: hub1id, Name: hub1name, Hash: hub1hash, Devices: []string{device1uri}}
	if len(expectedHub) > 0 {
		expected = expectedHub[0]
	}
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(expected.Id)
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := model.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != expected.Name || result.Hash != expected.Hash || result.Id != expected.Id || !reflect.DeepEqual(result.Devices, expected.Devices) {
		t.Error("unexpected result", result, expected)
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
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
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
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
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
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
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
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []string{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(result, []string{asResult}) {
		t.Error("unexpected result", result, asResult)
		return
	}
}

func TestHubDeviceSideEffects(t *testing.T) {
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
	err = producer.PublishDevice(model.DeviceInstance{Id: device2id, Name: device2name, Url: device2uri, DeviceType: devicetype1id}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)

	err = producer.PublishHub(model.GatewayFlat{Id: hub1id, Name: hub1name, Hash: hub1hash, Devices: []string{device1id, device2id}}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)

	err = producer.PublishDeviceDelete(device2id)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("hubEmpty", func(t *testing.T) {
		testHubEmpty(t, conf)
	})
	t.Run("deviceWithoutHub", func(t *testing.T) {
		testDeviceWithHubRef(t, conf, device1id, "")
	})
}

func testHubEmpty(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(hub1id)
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := model.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != hub1name || result.Hash != "" || result.Id != hub1id || (result.Devices != nil && !reflect.DeepEqual(result.Devices, []string{})) {
		b, _ := json.Marshal(result)
		t.Error("unexpected result", string(b), "'"+hub1name+"'", "'"+hub1id+"'")
		return
	}
}

func TestHubControl(t *testing.T) {
	closer, conf, _, err := createTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	if true {
		defer closer()
	}

	t.Run("testHubCreate", func(t *testing.T) {
		testHubCreate(t, conf)
	})
	t.Run("testHubUpdate", func(t *testing.T) {
		testHubUpdate(t, conf)
	})
	t.Run("testHubDelete", func(t *testing.T) {
		testHubDelete(t, conf)
	})
}

func testHubCreate(t *testing.T, conf config.Config) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(model.Hub{Name: hub1name, Hash: hub1hash})
	if err != nil {
		t.Error(err)
		return
	}
	url := "http://localhost:" + conf.ServerPort + "/hubs"
	resp, err := userjwt.Post(url, "application/json", b)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", url, resp.Status, resp.StatusCode, string(b))
		return
	}
	hub := model.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&hub)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)
	t.Run("testHubRead", func(t *testing.T) {
		testHubRead(t, conf, model.Hub{Id: hub.Id, Name: hub1name, Hash: hub1hash})
	})
}

func testHubUpdate(t *testing.T, conf config.Config) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(model.Hub{Name: hub1id, Hash: hub1hash})
	if err != nil {
		t.Error(err)
		return
	}
	url := "http://localhost:" + conf.ServerPort + "/hubs"
	resp, err := userjwt.Post(url, "application/json", b)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", url, resp.Status, resp.StatusCode, string(b))
		return
	}
	hub := model.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&hub)
	if err != nil {
		t.Error(err)
		return
	}
	b = new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(model.Hub{Id: hub.Id, Name: "foobar", Hash: "hash"})
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(2 * time.Second)
	url = "http://localhost:" + conf.ServerPort + "/hubs/" + hub.Id
	resp, err = jwtput(userjwt, url, "application/json", b)
	if err != nil {
		t.Error(err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", url, resp.Status, resp.StatusCode, string(b))
		return
	}
	time.Sleep(2 * time.Second)
	t.Run("testDeviceRead", func(t *testing.T) {
		testHubRead(t, conf, model.Hub{Id: hub.Id, Name: "foobar", Hash: "hash"})
	})
}

func testHubDelete(t *testing.T, conf config.Config) {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(model.Hub{Name: device3name})
	if err != nil {
		t.Error(err)
		return
	}
	url := "http://localhost:" + conf.ServerPort + "/hubs"
	resp, err := userjwt.Post(url, "application/json", b)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", url, resp.Status, resp.StatusCode, string(b))
		return
	}
	hub3 := model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&hub3)
	if err != nil {
		t.Error(err)
		return
	}
	b = new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(model.Hub{Name: device4name})
	if err != nil {
		t.Error(err)
		return
	}
	url = "http://localhost:" + conf.ServerPort + "/hubs"
	resp, err = userjwt.Post(url, "application/json", b)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", url, resp.Status, resp.StatusCode, string(b))
		return
	}
	hub4 := model.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&hub4)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(2 * time.Second)
	resp, err = jwtdelete(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+hub3.Id)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", url, resp.Status, resp.StatusCode, string(b))
		return
	}
	time.Sleep(2 * time.Second)
	t.Run("noUnexpectedDelete", func(t *testing.T) {
		testHubRead(t, conf, hub4)
	})
	t.Run("expectedDelete", func(t *testing.T) {
		testHubReadNotFound(t, conf, hub3.Id)
	})
}

func testHubReadNotFound(t *testing.T, conf config.Config, id string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(id)
	resp, err := userjwt.Get(endpoint)
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
