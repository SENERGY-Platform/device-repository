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

package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/models/go/models"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
)

var hub1id = "urn:infai:ses:device:1"
var hub1name = "hub1"

func TestHubs(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	c := client.NewClient("http://localhost:"+conf.ServerPort, nil)

	_, err, _ = c.SetDeviceType(AdminToken, models.DeviceType{Id: devicetype1id, Name: devicetype1name}, client.DeviceTypeUpdateOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	d1 := models.Device{
		Id:           device1id,
		LocalId:      device1lid,
		Name:         devicetype1name,
		DeviceTypeId: devicetype1id,
	}

	_, err, _ = c.SetDevice(userjwt, d1, client.DeviceUpdateOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	d2 := models.Device{
		Id:           device2id,
		LocalId:      device2lid,
		Name:         devicetype2name,
		DeviceTypeId: devicetype2id,
	}

	_, err, _ = c.SetDevice(userjwt, d2, client.DeviceUpdateOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	h1 := models.Hub{
		Id:             hub1id,
		Name:           hub1name,
		Hash:           "hash1",
		DeviceLocalIds: []string{device1lid, device2lid},
		DeviceIds:      []string{device1id, device2id},
	}
	expectedHub := models.Hub{
		Id:             hub1id,
		Name:           hub1name,
		Hash:           "hash1",
		DeviceLocalIds: []string{device1lid, device2lid},
		DeviceIds:      []string{device1id, device2id},
		OwnerId:        userid,
	}

	_, err, _ = c.SetHub(userjwt, h1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("not existing", func(t *testing.T) {
		testHubReadNotFound(t, conf, "foobar")
	})
	t.Run("testHubRead", func(t *testing.T) {
		testHubRead(t, conf, expectedHub)
	})
	t.Run("testHubDeviceRead", func(t *testing.T) {
		testHubDeviceRead(t, conf, h1, d1, d2)
	})
}

func testHubReadNotFound(t *testing.T, conf config.Config, id string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(id)
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
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
}

func testHubRead(t *testing.T, conf config.Config, expectedHubs ...models.Hub) {
	for _, expected := range expectedHubs {
		endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(expected.Id)
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
			b, _ := io.ReadAll(resp.Body)
			t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
			return
		}

		result := models.Hub{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(expected, result) {
			t.Errorf("unexpected result\n%#v\n%#v\n", expected, result)
			return
		}
	}
}

func testHubDeviceRead(t *testing.T, conf config.Config, hub models.Hub, expectedDevices ...models.Device) {
	endpoint := "http://localhost:" + conf.ServerPort + "/hubs/" + url.PathEscape(hub.Id) + "/devices"
	call := func(endpoint string) ([]string, error) {
		var result []string
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return result, err
		}
		req.Header.Set("Authorization", userjwt)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return result, err
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
			return result, errors.New(fmt.Sprint("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b)))
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		return result, err
	}
	localids, err := call(endpoint)
	if err != nil {
		t.Error(err)
		return
	}

	ids, err := call(endpoint + "?as=id")
	if err != nil {
		t.Error(err)
		return
	}

	for i, expected := range expectedDevices {
		if expected.Id != ids[i] {
			t.Error("expected.Id != ids[i] => ", expected.Id, ids[i])
			return
		}
		if expected.LocalId != localids[i] {
			t.Error("expected.LocalId != localids[i] => ", expected.LocalId, localids[i])
			return
		}
	}
}
