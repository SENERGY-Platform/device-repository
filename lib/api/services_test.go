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
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestServiceQuery(t *testing.T) {
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
				Id:  service1id,
				Url: service1uri,
				Protocol: model.Protocol{
					Id:                 protocol1id,
					ProtocolHandlerUrl: protocol1url,
				},
			},
			{
				Id:  service2id,
				Url: service2uri,
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

	t.Run("testServiceRead", func(t *testing.T) {
		testServiceRead(t, conf)
	})
}

func testServiceRead(t *testing.T, config config.Config) {
	endpoint := "http://localhost:" + config.ServerPort + "/services/" + url.PathEscape(service1id)
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
	result := model.Service{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Url != service1uri || result.Id != service1id {
		t.Error("unexpected result", result)
		return
	}
}
