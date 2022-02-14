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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"
)

var protocol1id = uuid.NewV4().String()
var protocol1name = uuid.NewV4().String()
var protocol2id = uuid.NewV4().String()
var protocol2name = uuid.NewV4().String()

func TestProtocolQuery(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	/*
		err = InitTopic(conf.ZookeeperUrl, conf.ProtocolTopic)
		if err != nil {
			t.Error(err)
			return
		}
	*/
	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishProtocol(model.Protocol{Id: protocol1id, Name: protocol1name}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 20; i++ {
		err = producer.PublishProtocol(model.Protocol{Id: uuid.NewV4().String(), Name: uuid.NewV4().String()}, userid)
		if err != nil {
			t.Error(err)
			return
		}
	}
	time.Sleep(5 * time.Second)

	t.Run("unexisting", func(t *testing.T) {
		testProtocolReadNotFound(t, conf, uuid.NewV4().String())
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

func testProtocolRead(t *testing.T, conf config.Config, expectedDt ...model.Protocol) {
	expected := model.Protocol{Id: protocol1id, Name: protocol1name}
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
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := model.Protocol{}
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
	result := []model.Protocol{}
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
	result := []model.Protocol{}
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
	result := []model.Protocol{}
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
	defaultresult := []model.Protocol{}
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
	ascresult := []model.Protocol{}
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
	descresult := []model.Protocol{}
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
