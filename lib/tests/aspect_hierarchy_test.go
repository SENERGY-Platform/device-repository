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

func TestAspectFunctions(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}
	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishFunction(model.Function{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishFunction(model.Function{
		Id:        "fid_2",
		Name:      "fid_2",
		ConceptId: "concept_2",
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishAspect(model.Aspect{
		Id:   "parent_2",
		Name: "parent_2",
		SubAspects: []model.Aspect{
			{
				Id:   "aid_2",
				Name: "aid_2",
				SubAspects: []model.Aspect{
					{
						Id:   "child_2",
						Name: "child_2",
					},
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	aspect := model.Aspect{
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
	}
	err = producer.PublishAspect(aspect, userid)
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

	//defaults to ancestors=false&descendants=true
	t.Run("aspects", testGetRequest(userjwt, conf, "/aspects?function=measuring-function", []model.Aspect{aspect}))

	//find only exact match
	t.Run("aspects_ff", testGetRequest(userjwt, conf, "/aspects?function=measuring-function&ancestors=false&descendants=false", []model.Aspect{}))

	//find aid either as
	//	descendent of parent --> returns parent as root
	//	or as ancestor of child --> returns parent as root
	t.Run("aspects_ft", testGetRequest(userjwt, conf, "/aspects?function=measuring-function&ancestors=false&descendants=true", []model.Aspect{aspect}))
	t.Run("aspects_tf", testGetRequest(userjwt, conf, "/aspects?function=measuring-function&ancestors=true&descendants=false", []model.Aspect{aspect}))
	t.Run("aspects_tt", testGetRequest(userjwt, conf, "/aspects?function=measuring-function&ancestors=true&descendants=true", []model.Aspect{aspect}))

	//defaults to ancestors=false&descendants=true
	t.Run("aspect-nodes", testGetRequest(userjwt, conf, "/aspect-nodes?function=measuring-function", []model.AspectNode{
		{
			Id:            "aid",
			Name:          "aid",
			RootId:        "parent",
			ParentId:      "parent",
			ChildIds:      []string{"child"},
			AncestorIds:   []string{"parent"},
			DescendentIds: []string{"child"},
		},
		{
			Id:            "parent",
			Name:          "parent",
			RootId:        "parent",
			ParentId:      "",
			ChildIds:      []string{"aid"},
			AncestorIds:   []string{},
			DescendentIds: []string{"aid", "child"},
		},
	}))
	//find only exact match
	t.Run("aspect-nodes_ff", testGetRequest(userjwt, conf, "/aspect-nodes?function=measuring-function&ancestors=false&descendants=false", []model.AspectNode{
		{
			Id:            "aid",
			Name:          "aid",
			RootId:        "parent",
			ParentId:      "parent",
			ChildIds:      []string{"child"},
			AncestorIds:   []string{"parent"},
			DescendentIds: []string{"child"},
		},
	}))

	t.Run("aspect-nodes_ft", testGetRequest(userjwt, conf, "/aspect-nodes?function=measuring-function&ancestors=false&descendants=true", []model.AspectNode{
		{
			Id:            "aid",
			Name:          "aid",
			RootId:        "parent",
			ParentId:      "parent",
			ChildIds:      []string{"child"},
			AncestorIds:   []string{"parent"},
			DescendentIds: []string{"child"},
		},
		{
			Id:            "parent",
			Name:          "parent",
			RootId:        "parent",
			ParentId:      "",
			ChildIds:      []string{"aid"},
			AncestorIds:   []string{},
			DescendentIds: []string{"aid", "child"},
		},
	}))

	t.Run("aspect-nodes_tf", testGetRequest(userjwt, conf, "/aspect-nodes?function=measuring-function&ancestors=true&descendants=false", []model.AspectNode{
		{
			Id:            "aid",
			Name:          "aid",
			RootId:        "parent",
			ParentId:      "parent",
			ChildIds:      []string{"child"},
			AncestorIds:   []string{"parent"},
			DescendentIds: []string{"child"},
		},
		{
			Id:            "child",
			Name:          "child",
			RootId:        "parent",
			ParentId:      "aid",
			ChildIds:      []string{},
			AncestorIds:   []string{"aid", "parent"},
			DescendentIds: []string{},
		},
	}))

	t.Run("aspect-nodes_tt", testGetRequest(userjwt, conf, "/aspect-nodes?function=measuring-function&ancestors=true&descendants=true", []model.AspectNode{
		{
			Id:            "aid",
			Name:          "aid",
			RootId:        "parent",
			ParentId:      "parent",
			ChildIds:      []string{"child"},
			AncestorIds:   []string{"parent"},
			DescendentIds: []string{"child"},
		},
		{
			Id:            "child",
			Name:          "child",
			RootId:        "parent",
			ParentId:      "aid",
			ChildIds:      []string{},
			AncestorIds:   []string{"aid", "parent"},
			DescendentIds: []string{},
		},
		{
			Id:            "parent",
			Name:          "parent",
			RootId:        "parent",
			ParentId:      "",
			ChildIds:      []string{"aid"},
			AncestorIds:   []string{},
			DescendentIds: []string{"aid", "child"},
		},
	}))

	t.Run("aspect-nodes_measuring-functions_aid", testGetRequest(userjwt, conf, "/aspect-nodes/aid/measuring-functions", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))
	t.Run("aspect-nodes_measuring-functions_aid_ff", testGetRequest(userjwt, conf, "/aspect-nodes/aid/measuring-functions?ancestors=false&descendants=false", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))
	t.Run("aspect-nodes_measuring-functions_aid_ft", testGetRequest(userjwt, conf, "/aspect-nodes/aid/measuring-functions?ancestors=false&descendants=true", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))
	t.Run("aspect-nodes_measuring-functions_aid_tf", testGetRequest(userjwt, conf, "/aspect-nodes/aid/measuring-functions?ancestors=true&descendants=false", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))
	t.Run("aspect-nodes_measuring-functions_aid_tt", testGetRequest(userjwt, conf, "/aspect-nodes/aid/measuring-functions?ancestors=true&descendants=true", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))

	t.Run("aspect-nodes_measuring-functions_parent", testGetRequest(userjwt, conf, "/aspect-nodes/parent/measuring-functions", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))
	t.Run("aspect-nodes_measuring-functions_parent_ff", testGetRequest(userjwt, conf, "/aspect-nodes/parent/measuring-functions?ancestors=false&descendants=false", []model.Function{}))
	t.Run("aspect-nodes_measuring-functions_parent_ft", testGetRequest(userjwt, conf, "/aspect-nodes/parent/measuring-functions?ancestors=false&descendants=true", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))
	t.Run("aspect-nodes_measuring-functions_parent_tf", testGetRequest(userjwt, conf, "/aspect-nodes/parent/measuring-functions?ancestors=true&descendants=false", []model.Function{}))
	t.Run("aspect-nodes_measuring-functions_parent_tt", testGetRequest(userjwt, conf, "/aspect-nodes/parent/measuring-functions?ancestors=true&descendants=true", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))

	t.Run("aspect-nodes_measuring-functions_child", testGetRequest(userjwt, conf, "/aspect-nodes/child/measuring-functions", []model.Function{}))
	t.Run("aspect-nodes_measuring-functions_child_ff", testGetRequest(userjwt, conf, "/aspect-nodes/child/measuring-functions?ancestors=false&descendants=false", []model.Function{}))
	t.Run("aspect-nodes_measuring-functions_child_ft", testGetRequest(userjwt, conf, "/aspect-nodes/child/measuring-functions?ancestors=false&descendants=true", []model.Function{}))
	t.Run("aspect-nodes_measuring-functions_child_tf", testGetRequest(userjwt, conf, "/aspect-nodes/child/measuring-functions?ancestors=true&descendants=false", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))
	t.Run("aspect-nodes_measuring-functions_child_tt", testGetRequest(userjwt, conf, "/aspect-nodes/child/measuring-functions?ancestors=true&descendants=true", []model.Function{{
		Id:        "fid",
		Name:      "fid",
		ConceptId: "concept_1",
	}}))

}

func TestDeviceTypeFilterCriteria(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
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
			eJson, _ := json.Marshal(expectedNormalized)
			rJson, _ := json.Marshal(result)
			t.Error("unexpected result", expectedNormalized, result, "\n", string(eJson), "\n", string(rJson))
			return
		}
	}
}

func normalize(expected interface{}) (result interface{}) {
	temp, _ := json.Marshal(expected)
	json.Unmarshal(temp, &result)
	return
}
