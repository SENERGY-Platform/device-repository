/*
 * Copyright 2025 InfAI (CC SES)
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

package repo_legacy

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
)

var protocol1id = uuid.NewString()
var protocol1name = uuid.NewString()
var protocol2id = uuid.NewString()
var protocol2name = uuid.NewString()

func TestDeviceTypeProtocolFilter(t *testing.T) {
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

	protocolIds := []string{protocol1id, protocol2id, "foobarprotocol"}
	t.Run("create protocols", func(t *testing.T) {
		for _, id := range protocolIds {
			_, err, _ = c.SetProtocol(AdminToken, models.Protocol{Id: id, Name: protocol1name})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	deviceTypes := []models.DeviceType{
		{
			Id: "dt1",
			Services: []models.Service{
				{
					Id:         "dt1s1",
					ProtocolId: protocolIds[0],
				},
			},
		},
		{
			Id: "dt2",
			Services: []models.Service{
				{
					Id:         "dt2s1",
					ProtocolId: protocolIds[1],
				},
			},
		},
		{
			Id: "dt3",
			Services: []models.Service{
				{
					Id:         "dt3s1",
					ProtocolId: protocolIds[2],
				},
			},
		},
		{
			Id: "dt4",
			Services: []models.Service{
				{
					Id:         "dt4s1",
					ProtocolId: protocolIds[0],
				},
				{
					Id:         "dt4s2",
					ProtocolId: protocolIds[1],
				},
			},
		},
	}

	t.Run("create device-types", func(t *testing.T) {
		for _, dt := range deviceTypes {
			_, err, _ = c.SetDeviceType(AdminToken, dt, client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("find device-types by protocolIds", func(t *testing.T) {
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		t.Run("0", func(t *testing.T) {
			list, _, err, _ := c.ListDeviceTypesV3(userjwt, client.DeviceTypeListOptions{
				ProtocolIds: []string{protocolIds[0]},
			})
			if err != nil {
				t.Error(err)
				return
			}
			ids := []string{}
			for _, d := range list {
				ids = append(ids, d.Id)
			}
			expected := []string{deviceTypes[0].Id, deviceTypes[3].Id}
			if !reflect.DeepEqual(ids, expected) {
				t.Errorf("\na=%#v\ne=%#v\n", ids, expected)
				return
			}
		})
		t.Run("1", func(t *testing.T) {
			list, _, err, _ := c.ListDeviceTypesV3(userjwt, client.DeviceTypeListOptions{
				ProtocolIds: []string{protocolIds[1]},
			})
			if err != nil {
				t.Error(err)
				return
			}
			ids := []string{}
			for _, d := range list {
				ids = append(ids, d.Id)
			}
			expected := []string{deviceTypes[1].Id, deviceTypes[3].Id}
			if !reflect.DeepEqual(ids, expected) {
				t.Errorf("\na=%#v\ne=%#v\n", ids, expected)
				return
			}
		})
		t.Run("2", func(t *testing.T) {
			list, _, err, _ := c.ListDeviceTypesV3(userjwt, client.DeviceTypeListOptions{
				ProtocolIds: []string{protocolIds[2]},
			})
			if err != nil {
				t.Error(err)
				return
			}
			ids := []string{}
			for _, d := range list {
				ids = append(ids, d.Id)
			}
			expected := []string{deviceTypes[2].Id}
			if !reflect.DeepEqual(ids, expected) {
				t.Errorf("\na=%#v\ne=%#v\n", ids, expected)
				return
			}
		})
		t.Run("0,1", func(t *testing.T) {
			list, _, err, _ := c.ListDeviceTypesV3(userjwt, client.DeviceTypeListOptions{
				ProtocolIds: []string{protocolIds[0], protocolIds[1]},
			})
			if err != nil {
				t.Error(err)
				return
			}
			ids := []string{}
			for _, d := range list {
				ids = append(ids, d.Id)
			}
			expected := []string{deviceTypes[0].Id, deviceTypes[1].Id, deviceTypes[3].Id}
			if !reflect.DeepEqual(ids, expected) {
				t.Errorf("\na=%#v\ne=%#v\n", ids, expected)
				return
			}
		})
		t.Run("0,2", func(t *testing.T) {
			list, total, err, _ := c.ListDeviceTypesV3(userjwt, client.DeviceTypeListOptions{
				ProtocolIds: []string{protocolIds[0], protocolIds[2]},
			})
			if err != nil {
				t.Error(err)
				return
			}
			if total != 3 {
				t.Errorf("\na=%#v\ne=%#v\n", total, 3)
			}
			ids := []string{}
			for _, d := range list {
				ids = append(ids, d.Id)
			}
			expected := []string{deviceTypes[0].Id, deviceTypes[2].Id, deviceTypes[3].Id}
			if !reflect.DeepEqual(ids, expected) {
				t.Errorf("\na=%#v\ne=%#v\n", ids, expected)
				return
			}
		})
	})
}

func TestProtocolQuery(t *testing.T) {
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

	_, err, _ = c.SetProtocol(AdminToken, models.Protocol{Id: protocol1id, Name: protocol1name})
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 20; i++ {
		_, err, _ = c.SetProtocol(AdminToken, models.Protocol{Id: uuid.NewString(), Name: uuid.NewString()})
		if err != nil {
			t.Error(err)
			return
		}
	}

	t.Run("unexisting", func(t *testing.T) {
		testProtocolReadNotFound(t, conf, uuid.NewString())
	})
	t.Run("testProtocolRead", func(t *testing.T) {
		testProtocolRead(t, conf)
	})
	t.Run("testProtocolList", func(t *testing.T) {
		testProtocolList(t, conf)
	})
	t.Run("testProtocolListLimit10", func(t *testing.T) {
		testProtocolListLimit10(t, conf)
	})
	t.Run("testProtocolListLimit10Offset20", func(t *testing.T) {
		testProtocolListLimit10Offset20(t, conf)
	})
	t.Run("testProtocolListSort", func(t *testing.T) {
		testProtocolListSort(t, conf)
	})
}

func testProtocolRead(t *testing.T, conf config.Config, expectedDt ...models.Protocol) {
	expected := models.Protocol{Id: protocol1id, Name: protocol1name}
	if len(expectedDt) > 0 {
		expected = expectedDt[0]
	}
	endpoint := "http://localhost:" + conf.ServerPort + "/protocols/" + url.PathEscape(expected.Id)
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
	result := models.Protocol{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != expected.Name {
		t.Error("unexpected result", result)
		return
	}
}

func testProtocolList(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/protocols"
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
	result := []models.Protocol{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 21 {
		t.Error("unexpected result", len(result), result)
		return
	}
}

func testProtocolListLimit10(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/protocols?limit=10"
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
	result := []models.Protocol{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 10 {
		t.Error("unexpected result", result)
		return
	}
}

func testProtocolListLimit10Offset20(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/protocols?limit=10&offset=20"
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
	result := []models.Protocol{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Error("unexpected result", result)
		return
	}
}

func testProtocolListSort(t *testing.T, config config.Config) {
	defaultendpoint := "http://localhost:" + config.ServerPort + "/protocols?sort=name"
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
	defaultresult := []models.Protocol{}
	err = json.NewDecoder(resp.Body).Decode(&defaultresult)
	if err != nil {
		t.Error(err)
	}
	if len(defaultresult) != 21 {
		t.Error("unexpected result", len(defaultresult))
		return
	}
	ascendpoint := "http://localhost:" + config.ServerPort + "/protocols?sort=name.asc"
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
	ascresult := []models.Protocol{}
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

	descendpoint := "http://localhost:" + config.ServerPort + "/protocols?sort=name.desc"
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
	descresult := []models.Protocol{}
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

func testProtocolReadNotFound(t *testing.T, conf config.Config, id string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/protocols/" + url.PathEscape(id)
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
