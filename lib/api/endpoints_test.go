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

var service1id = uuid.NewV4().String()
var service1uri = uuid.NewV4().String()
var service2id = uuid.NewV4().String()
var service2uri = uuid.NewV4().String()
var service3id = uuid.NewV4().String()
var service3uri = uuid.NewV4().String()

var protocol1id = uuid.NewV4().String()
var protocol1url = uuid.NewV4().String()

func TestEndpointsQuery(t *testing.T) {
	closer, conf, producer, err := createTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	if true {
		defer closer()
	}
	err = producer.PublishDeviceType(model.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []model.Service{
			{
				Id:             service1id,
				Url:            service1uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
			{
				Id:             service2id,
				Url:            service2uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
		},
	}, userid)
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

	t.Run("testEndpointReadDevice", func(t *testing.T) {
		testEndpointReadDevice(t, conf)
	})
	t.Run("testEndpointReadService", func(t *testing.T) {
		testEndpointReadService(t, conf)
	})
	t.Run("testEndpointReadEndpoint", func(t *testing.T) {
		testEndpointReadEndpoint(t, conf)
	})
	t.Run("testEndpointReadIn", func(t *testing.T) {
		testEndpointReadIn(t, conf)
	})
	t.Run("testEndpointReadOut", func(t *testing.T) {
		testEndpointReadOut(t, conf)
	})
}

func testEndpointReadDevice(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?device=" + url.QueryEscape(device1id)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 2 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if (result.Endpoint != device1uri+"/"+service1uri && result.Endpoint != device1uri+"/"+service2uri) || result.Device != device1id || result.ProtocolHandler != protocol1url {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointReadService(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?service=" + url.QueryEscape(service1id)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 21 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Service != service1id || result.ProtocolHandler != protocol1url {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointReadEndpoint(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?endpoint=" + url.QueryEscape(device1uri+"/"+service1uri)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != device1uri+"/"+service1uri || result.Device != device1id || result.ProtocolHandler != protocol1url || result.Service != service1id {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointReadIn(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?protocol=" + url.QueryEscape(protocol1url) + "&endpoint=" + url.QueryEscape(device1uri+"/"+service1uri)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != device1uri+"/"+service1uri || result.Device != device1id || result.ProtocolHandler != protocol1url || result.Service != service1id {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointReadOut(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?service=" + url.QueryEscape(service1id) + "&device=" + url.QueryEscape(device1id)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != device1uri+"/"+service1uri || result.Device != device1id || result.ProtocolHandler != protocol1url || result.Service != service1id {
			t.Error("unexpected result", result)
			return
		}
	}
}

func TestEndpointsUpdateByDeviceUpdate(t *testing.T) {
	closer, conf, producer, err := createTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	if true {
		defer closer()
	}
	err = producer.PublishDeviceType(model.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []model.Service{
			{
				Id:             service1id,
				Url:            service1uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
			{
				Id:             service2id,
				Url:            service2uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
		},
	}, userid)
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
	for i := 0; i < 19; i++ {
		err = producer.PublishDevice(model.DeviceInstance{Id: uuid.NewV4().String(), Name: uuid.NewV4().String(), Url: uuid.NewV4().String(), DeviceType: devicetype1id}, userid)
		if err != nil {
			t.Error(err)
			return
		}
	}
	time.Sleep(3 * time.Second)

	err = producer.PublishDevice(model.DeviceInstance{Id: device2id, Name: device2name, Url: "changed" + device2uri, DeviceType: devicetype1id}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("unchangedReadDevice", func(t *testing.T) {
		testEndpointReadDevice(t, conf)
	})
	t.Run("unchangedReadService", func(t *testing.T) {
		testEndpointReadService(t, conf)
	})
	t.Run("unchangedReadEndpoint", func(t *testing.T) {
		testEndpointReadEndpoint(t, conf)
	})
	t.Run("unchangedReadIn", func(t *testing.T) {
		testEndpointReadIn(t, conf)
	})
	t.Run("unchangedReadOut", func(t *testing.T) {
		testEndpointReadOut(t, conf)
	})

	t.Run("changedReadDevice", func(t *testing.T) {
		testEndpointUpdateByDeviceUpdateDeviceReadChanged(t, conf)
	})
	t.Run("changedReadEndpoint", func(t *testing.T) {
		testEndpointUpdateByDeviceUpdateEndpointReadChanged(t, conf)
	})
	t.Run("changedReadIn", func(t *testing.T) {
		testEndpointUpdateByDeviceUpdateInReadChanged(t, conf)
	})
	t.Run("changedReadOut", func(t *testing.T) {
		testEndpointUpdateByDeviceUpdateOutReadChanged(t, conf)
	})
}

func testEndpointUpdateByDeviceUpdateDeviceReadChanged(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?device=" + url.QueryEscape(device2id)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 2 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if (result.Endpoint != "changed"+device2uri+"/"+service1uri && result.Endpoint != "changed"+device2uri+"/"+service2uri) || result.Device != device2id || result.ProtocolHandler != protocol1url {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointUpdateByDeviceUpdateEndpointReadChanged(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?endpoint=" + url.QueryEscape("changed"+device2uri+"/"+service1uri)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != "changed"+device2uri+"/"+service1uri || result.Device != device2id || result.ProtocolHandler != protocol1url || result.Service != service1id {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointUpdateByDeviceUpdateInReadChanged(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?protocol=" + url.QueryEscape(protocol1url) + "&endpoint=" + url.QueryEscape("changed"+device2uri+"/"+service1uri)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != "changed"+device2uri+"/"+service1uri || result.Device != device2id || result.ProtocolHandler != protocol1url || result.Service != service1id {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointUpdateByDeviceUpdateOutReadChanged(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?service=" + url.QueryEscape(service1id) + "&device=" + url.QueryEscape(device2id)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != "changed"+device2uri+"/"+service1uri || result.Device != device2id || result.ProtocolHandler != protocol1url || result.Service != service1id {
			t.Error("unexpected result", result)
			return
		}
	}
}

func TestEndpointsUpdateByEndpointFormatUpdate(t *testing.T) {
	closer, conf, producer, err := createTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	if true {
		defer closer()
	}
	err = producer.PublishDeviceType(model.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []model.Service{
			{
				Id:             service1id,
				Url:            service1uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
			{
				Id:             service2id,
				Url:            service2uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	err = producer.PublishDeviceType(model.DeviceType{
		Id:   devicetype2id,
		Name: devicetype2name,
		Services: []model.Service{
			{
				Id:             service3id,
				Url:            service3uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
		},
	}, userid)
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
	err = producer.PublishDevice(model.DeviceInstance{Id: device2id, Name: device2name, Url: device2uri, DeviceType: devicetype2id}, userid)
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
	for i := 0; i < 20; i++ {
		err = producer.PublishDevice(model.DeviceInstance{Id: uuid.NewV4().String(), Name: uuid.NewV4().String(), Url: uuid.NewV4().String(), DeviceType: devicetype2id}, userid)
		if err != nil {
			t.Error(err)
			return
		}
	}
	time.Sleep(3 * time.Second)

	err = producer.PublishDeviceType(model.DeviceType{
		Id:   devicetype2id,
		Name: devicetype2name,
		Services: []model.Service{
			{
				Id:             service3id,
				Url:            service3uri,
				EndpointFormat: "{{device_uri}}/changed{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)
	t.Run("unchangedReadDevice", func(t *testing.T) {
		testEndpointReadDevice(t, conf)
	})
	t.Run("unchangedReadService", func(t *testing.T) {
		testEndpointReadService(t, conf)
	})
	t.Run("unchangedReadEndpoint", func(t *testing.T) {
		testEndpointReadEndpoint(t, conf)
	})
	t.Run("unchangedReadIn", func(t *testing.T) {
		testEndpointReadIn(t, conf)
	})
	t.Run("unchangedReadOut", func(t *testing.T) {
		testEndpointReadOut(t, conf)
	})
	t.Run("changedReadDevice", func(t *testing.T) {
		testEndpointUpdateByDeviceTypeUpdateDeviceReadChanged(t, conf)
	})
	t.Run("changedReadEndpoint", func(t *testing.T) {
		testEndpointUpdateByDeviceTypeUpdateEndpointReadChanged(t, conf)
	})
	t.Run("changedReadIn", func(t *testing.T) {
		testEndpointUpdateByDeviceTypeUpdateInReadChanged(t, conf)
	})
	t.Run("changedReadOut", func(t *testing.T) {
		testEndpointUpdateByDeviceTypeUpdateOutReadChanged(t, conf)
	})
}

func TestEndpointsUpdateByServiceUriUpdate(t *testing.T) {
	closer, conf, producer, err := createTestEnv()
	if err != nil {
		t.Fatal(err)
	}
	if true {
		defer closer()
	}
	err = producer.PublishDeviceType(model.DeviceType{
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []model.Service{
			{
				Id:             service1id,
				Url:            service1uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
			{
				Id:             service2id,
				Url:            service2uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	err = producer.PublishDeviceType(model.DeviceType{
		Id:   devicetype2id,
		Name: devicetype2name,
		Services: []model.Service{
			{
				Id:             service3id,
				Url:            service3uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
		},
	}, userid)
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
	err = producer.PublishDevice(model.DeviceInstance{Id: device2id, Name: device2name, Url: device2uri, DeviceType: devicetype2id}, userid)
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
	for i := 0; i < 20; i++ {
		err = producer.PublishDevice(model.DeviceInstance{Id: uuid.NewV4().String(), Name: uuid.NewV4().String(), Url: uuid.NewV4().String(), DeviceType: devicetype2id}, userid)
		if err != nil {
			t.Error(err)
			return
		}
	}
	time.Sleep(3 * time.Second)

	err = producer.PublishDeviceType(model.DeviceType{
		Id:   devicetype2id,
		Name: devicetype2name,
		Services: []model.Service{
			{
				Id:             service3id,
				Url:            "changed" + service3uri,
				EndpointFormat: "{{device_uri}}/{{service_uri}}",
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)
	t.Run("unchangedReadDevice", func(t *testing.T) {
		testEndpointReadDevice(t, conf)
	})
	t.Run("unchangedReadService", func(t *testing.T) {
		testEndpointReadService(t, conf)
	})
	t.Run("unchangedReadEndpoint", func(t *testing.T) {
		testEndpointReadEndpoint(t, conf)
	})
	t.Run("unchangedReadIn", func(t *testing.T) {
		testEndpointReadIn(t, conf)
	})
	t.Run("unchangedReadOut", func(t *testing.T) {
		testEndpointReadOut(t, conf)
	})
	t.Run("changedReadDevice", func(t *testing.T) {
		testEndpointUpdateByDeviceTypeUpdateDeviceReadChanged(t, conf)
	})
	t.Run("changedReadEndpoint", func(t *testing.T) {
		testEndpointUpdateByDeviceTypeUpdateEndpointReadChanged(t, conf)
	})
	t.Run("changedReadIn", func(t *testing.T) {
		testEndpointUpdateByDeviceTypeUpdateInReadChanged(t, conf)
	})
	t.Run("changedReadOut", func(t *testing.T) {
		testEndpointUpdateByDeviceTypeUpdateOutReadChanged(t, conf)
	})
}

func testEndpointUpdateByDeviceTypeUpdateDeviceReadChanged(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?device=" + url.QueryEscape(device2id)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != device2uri+"/"+"changed"+service3uri || result.Device != device2id || result.ProtocolHandler != protocol1url {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointUpdateByDeviceTypeUpdateEndpointReadChanged(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?endpoint=" + url.QueryEscape(device2uri+"/"+"changed"+service3uri)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != device2uri+"/"+"changed"+service3uri || result.Device != device2id || result.ProtocolHandler != protocol1url || result.Service != service3id {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointUpdateByDeviceTypeUpdateInReadChanged(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?protocol=" + url.QueryEscape(protocol1url) + "&endpoint=" + url.QueryEscape(device2uri+"/"+"changed"+service3uri)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != device2uri+"/"+"changed"+service3uri || result.Device != device2id || result.ProtocolHandler != protocol1url || result.Service != service3id {
			t.Error("unexpected result", result)
			return
		}
	}
}

func testEndpointUpdateByDeviceTypeUpdateOutReadChanged(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/endpoints?service=" + url.QueryEscape(service3id) + "&device=" + url.QueryEscape(device2id)
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
	results := []model.Endpoint{}
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Error("unexpected result", results)
		return
	}
	for _, result := range results {
		if result.Endpoint != device2uri+"/"+"changed"+service3uri || result.Device != device2id || result.ProtocolHandler != protocol1url || result.Service != service3id {
			t.Error("unexpected result", result)
			return
		}
	}
}
