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
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"
)

//TODO
func TestAspectFunctions(t *testing.T) {
	t.Skip("TODO")
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishDeviceType(model.DeviceType{}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)
}

func TestDeviceTypeFilterCriteria(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishAspect(model.Aspect{
		Id:   "parent",
		Name: "parent",
		SubAspects: []model.Aspect{
			{
				Id:   "aid",
				Name: "aid",
				SubAspects: []model.Aspect{
					{
						Id:   "child",
						Name: "child",
					},
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	dt := model.DeviceType{
		Id:            "dt",
		Name:          "dt",
		ServiceGroups: nil,
		Services: []model.Service{
			{
				Id:          "sid",
				LocalId:     "s",
				Name:        "s",
				Interaction: model.EVENT_AND_REQUEST,
				ProtocolId:  "pid",
				Outputs: []model.Content{
					{
						ContentVariable: model.ContentVariable{
							Id:               "vid",
							Name:             "v",
							CharacteristicId: "cid",
							FunctionId:       "fid",
							AspectId:         "aid",
						},
					},
				},
			},
		},
	}
	err = producer.PublishDeviceType(dt, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	t.Run("matching", testGetRequest(userjwt, conf, "/device-types?filter="+url.QueryEscape(`[{"function_id":"fid","aspect_id":"aid"}]`), []model.DeviceType{dt}))
	t.Run("parent", testGetRequest(userjwt, conf, "/device-types?filter="+url.QueryEscape(`[{"function_id":"fid","aspect_id":"parent"}]`), []model.DeviceType{dt}))
	t.Run("child", testGetRequest(userjwt, conf, "/device-types?filter="+url.QueryEscape(`[{"function_id":"fid","aspect_id":"child"}]`), []model.DeviceType{}))
}

func testGetRequest(token string, conf config.Config, path string, expected interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		req, err := http.NewRequest("GET", "http://localhost:"+conf.ServerPort+path, nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := ioutil.ReadAll(resp.Body)
			t.Error("unexpected response", path, resp.Status, resp.StatusCode, string(b))
			return
		}

		var result interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
		}
		expectedNormalized := normalize(expected)
		if !reflect.DeepEqual(expectedNormalized, result) {
			t.Error("unexpected result", expectedNormalized, result)
			return
		}
	}
}

func normalize(expected interface{}) (result interface{}) {
	temp, _ := json.Marshal(expected)
	json.Unmarshal(temp, &result)
	return
}
