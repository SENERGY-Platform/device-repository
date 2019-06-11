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

func TestDeviceUriQuery(t *testing.T) {
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

	t.Run("testDeviceUriRead", func(t *testing.T) {
		testDeviceUriRead(t, conf)
	})
	t.Run("testDeviceUriHeadRead200", func(t *testing.T) {
		testDeviceUriHeadRead(t, conf, device1uri, http.StatusOK)
	})
	t.Run("testDeviceUriHeadRead404", func(t *testing.T) {
		testDeviceUriHeadRead(t, conf, "nope", http.StatusNotFound)
	})
	t.Run("testDeviceUriList", func(t *testing.T) {
		testDeviceUriList(t, conf)
	})
	t.Run("testDeviceUriListLimit10", func(t *testing.T) {
		testDeviceUriListLimit10(t, conf)
	})
	t.Run("testDeviceUriListLimit10Offset20", func(t *testing.T) {
		testDeviceUriListLimit10Offset20(t, conf)
	})
	t.Run("testDeviceUriListSort", func(t *testing.T) {
		testDeviceUriListSort(t, conf)
	})
}

func testDeviceUriRead(t *testing.T, configuration config.Config) {
	endpoint := "http://localhost:" + configuration.ServerPort + "/device-uris/" + url.PathEscape(device1uri)
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
	if result.Name != device1name || result.Url != device1uri || result.Id != device1id {
		t.Error("unexpected result", result)
		return
	}
}

func testDeviceUriHeadRead(t *testing.T, configuration config.Config, uri string, status int) {
	endpoint := "http://localhost:" + configuration.ServerPort + "/device-uris/" + url.PathEscape(uri)
	resp, err := head(endpoint, string(userjwt))
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != status {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
}

func testDeviceUriList(t *testing.T, configuration config.Config) {
	endpoint := "http://localhost:" + configuration.ServerPort + "/device-uris"
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
	result := []model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 21 {
		t.Error("unexpected result", len(result), result)
		return
	}
}

func testDeviceUriListLimit10(t *testing.T, configuration config.Config) {
	endpoint := "http://localhost:" + configuration.ServerPort + "/device-uris?limit=10"
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

func testDeviceUriListLimit10Offset20(t *testing.T, configuration config.Config) {
	endpoint := "http://localhost:" + configuration.ServerPort + "/device-uris?limit=10&offset=20"
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

func testDeviceUriListSort(t *testing.T, configuration config.Config) {
	ascendpoint := "http://localhost:" + configuration.ServerPort + "/device-uris?sort=name.asc"
	resp, err := userjwt.Get(ascendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", ascendpoint, resp.Status, resp.StatusCode, string(b))
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

	descendpoint := "http://localhost:" + configuration.ServerPort + "/device-uris?sort=name.desc"
	resp, err = userjwt.Get(descendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", descendpoint, resp.Status, resp.StatusCode, string(b))
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

func TestDeviceUriControl(t *testing.T) {
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

	t.Run("testDeviceUriCreate", func(t *testing.T) {
		testDeviceUriCreate(t, conf)
	})
	t.Run("testDeviceUriUpdate", func(t *testing.T) {
		testDeviceUriUpdate(t, conf)
	})
	t.Run("testDeviceUriDelete", func(t *testing.T) {
		testDeviceUriDelete(t, conf)
	})
}

func testDeviceUriCreate(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}

func testDeviceUriUpdate(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}

func testDeviceUriDelete(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}
