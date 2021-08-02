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

var device1id = "urn:infai:ses:device:1"
var device1lid = "lid1"
var device1name = uuid.NewV4().String()
var device2id = "urn:infai:ses:device:2"
var device2lid = "lid2"
var device2name = uuid.NewV4().String()
var device3id = "urn:infai:ses:device:3"
var device3lid = "lid3"
var device3name = uuid.NewV4().String()

func TestDeviceQuery(t *testing.T) {
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
	time.Sleep(10 * time.Second)

	d1 := model.Device{
		Id:           device1id,
		LocalId:      device1lid,
		Name:         device1name,
		DeviceTypeId: devicetype1id,
	}

	err = producer.PublishDevice(d1, userid)
	if err != nil {
		t.Error(err)
		return
	}

	d2 := model.Device{
		Id:           device2id,
		LocalId:      device2lid,
		Name:         device2name,
		DeviceTypeId: devicetype2id,
	}

	err = producer.PublishDevice(d2, userid)
	if err != nil {
		t.Error(err)
		return
	}

	d3 := model.Device{
		Id:      device3id,
		LocalId: device3lid,
		Name:    device3name,
		Attributes: []model.Attribute{
			{Key: "foo", Value: "bar"},
			{Key: "bar", Value: "batz"},
		},
		DeviceTypeId: devicetype2id,
	}

	err = producer.PublishDevice(d3, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	t.Run("not existing", func(t *testing.T) {
		testDeviceReadNotFound(t, conf, false, "foobar")
	})
	t.Run("not existing localId", func(t *testing.T) {
		testDeviceReadNotFound(t, conf, true, "foobar")
	})
	t.Run("testDeviceRead", func(t *testing.T) {
		testDeviceRead(t, conf, false, d1, d2, d3)
	})
	t.Run("testDeviceRead localid", func(t *testing.T) {
		testDeviceRead(t, conf, true, d1, d2, d3)
	})
}

func testDeviceRead(t *testing.T, conf config.Config, asLocalId bool, expectedDevices ...model.Device) {
	for _, expected := range expectedDevices {
		endpoint := "http://localhost:" + conf.ServerPort + "/devices/"
		if asLocalId {
			endpoint = endpoint + url.PathEscape(expected.LocalId) + "?as=local_id"
		} else {
			endpoint = endpoint + url.PathEscape(expected.Id)
		}
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

		result := model.Device{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(expected, result) {
			t.Error("unexpected result", expected, result)
			return
		}
	}

}

func testDeviceReadNotFound(t *testing.T, conf config.Config, asLocalId bool, id string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/devices/" + url.PathEscape(id)
	if asLocalId {
		endpoint = endpoint + "?as=local_id"
	}
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
