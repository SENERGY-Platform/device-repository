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

var valuetype1id = uuid.NewV4().String()
var valuetype1name = uuid.NewV4().String()
var valuetype2id = uuid.NewV4().String()
var valuetype2name = uuid.NewV4().String()

func TestValueTypeQuery(t *testing.T) {
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
				Input: []model.TypeAssignment{
					{
						Id:   uuid.NewV4().String(),
						Name: uuid.NewV4().String(),
						Type: model.ValueType{
							Id:   valuetype1id,
							Name: valuetype1name,
							Fields: []model.FieldType{
								{
									Id:   uuid.NewV4().String(),
									Name: uuid.NewV4().String(),
									Type: model.ValueType{
										Id:   valuetype2id,
										Name: valuetype2name,
										Fields: []model.FieldType{
											{
												Id:   uuid.NewV4().String(),
												Name: uuid.NewV4().String(),
												Type: model.ValueType{
													Id:   uuid.NewV4().String(),
													Name: uuid.NewV4().String(),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(3 * time.Second)

	t.Run("valueTypeRead", func(t *testing.T) {
		testValueTypeRead(t, conf)
	})
	t.Run("valueTypeList", func(t *testing.T) {
		testValueTypeList(t, conf)
	})
	t.Run("valueTypeListLimit2", func(t *testing.T) {
		testValueTypeListLimit2(t, conf)
	})
	t.Run("valueTypeListLimit2Offset2", func(t *testing.T) {
		testValueTypeListLimit2Offset2(t, conf)
	})
	t.Run("valueTypeListSort", func(t *testing.T) {
		testValueTypeListSort(t, conf)
	})
}

func testValueTypeRead(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/value-types/" + url.PathEscape(valuetype1id)
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
	result := model.ValueType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != valuetype1name {
		t.Error("unexpected result", result)
		return
	}
	endpoint = "http://localhost:" + conf.ServerPort + "/value-types/" + url.PathEscape(valuetype2id)
	resp, err = userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result = model.ValueType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != valuetype2name {
		t.Error("unexpected result", result)
		return
	}
}

func testValueTypeList(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/value-types"
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
	result := []model.ValueType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 3 {
		t.Error("unexpected result", result)
		return
	}
}

func testValueTypeListLimit2(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/value-types?limit=2"
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
	result := []model.ValueType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 2 {
		t.Error("unexpected result", result)
		return
	}
}

func testValueTypeListLimit2Offset2(t *testing.T, conf config.Config) {
	endpoint := "http://localhost:" + conf.ServerPort + "/value-types?limit=2&offset=2"
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
	result := []model.ValueType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 {
		t.Error("unexpected result", result)
		return
	}
}

func testValueTypeListSort(t *testing.T, config config.Config) {
	defaultendpoint := "http://localhost:" + config.ServerPort + "/value-types?sort=name"
	resp, err := userjwt.Get(defaultendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", defaultendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	defaultresult := []model.ValueType{}
	err = json.NewDecoder(resp.Body).Decode(&defaultresult)
	if err != nil {
		t.Error(err)
	}
	if len(defaultresult) != 3 {
		t.Error("unexpected result", len(defaultresult))
		return
	}
	ascendpoint := "http://localhost:" + config.ServerPort + "/value-types?sort=name.asc"
	resp, err = userjwt.Get(ascendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpected response", ascendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	ascresult := []model.ValueType{}
	err = json.NewDecoder(resp.Body).Decode(&ascresult)
	if err != nil {
		t.Error(err)
	}
	if len(ascresult) != 3 {
		t.Error("unexpected result", ascresult)
		return
	}
	if !reflect.DeepEqual(defaultresult, ascresult) {
		t.Error("unexpected result", defaultresult, ascresult)
		return
	}

	descendpoint := "http://localhost:" + config.ServerPort + "/value-types?sort=name.desc"
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
	descresult := []model.ValueType{}
	err = json.NewDecoder(resp.Body).Decode(&descresult)
	if err != nil {
		t.Error(err)
	}
	if len(ascresult) != 3 {
		t.Error("unexpected result", descresult)
		return
	}

	for i := 0; i < 3; i++ {
		if descresult[i].Id != ascresult[2-i].Id {
			t.Error("unexpected sorting result", i, descresult[i].Id, ascresult[2-i].Id)
			return
		}
	}
}

func TestValueTypeControl(t *testing.T) {
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

	t.Run("testValueTypeCreate", func(t *testing.T) {
		testValueTypeCreate(t, conf)
	})
	t.Run("testValueTypeUpdate", func(t *testing.T) {
		testValueTypeUpdate(t, conf)
	})
	t.Run("testValueTypeDelete", func(t *testing.T) {
		testValueTypeDelete(t, conf)
	})
}

func testValueTypeCreate(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}

func testValueTypeUpdate(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}

func testValueTypeDelete(t *testing.T, conf config.Config) {
	t.Skip("not implemented")
}

func TestValueTypeUpdateCascade(t *testing.T) {
	t.Skip("not implemented")

}
