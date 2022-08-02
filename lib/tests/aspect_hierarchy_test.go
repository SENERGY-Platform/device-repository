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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestSubAspectMoveSNRGY2202(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctrl, err := createMongoTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}
	initialStr := `{
   "id":"urn:infai:ses:aspect:60f8f6a6-c0e0-4fc3-9c11-bf2ccef1ed11",
   "name":"HVAC",
   "sub_aspects":[
      {
         "id":"urn:infai:ses:aspect:1d69d3b6-d16f-430c-bf58-f23b26cd84c4",
         "name":"Heating",
         "sub_aspects":[
            {
               "id":"urn:infai:ses:aspect:ac049c69-8e88-4306-817d-fb1fa10885e5",
               "name":"Seat",
               "sub_aspects":[
                  {
                     "id":"urn:infai:ses:aspect:5b00764e-2470-400c-ae0c-d84ac10dd32d",
                     "name":"Driver",
                     "sub_aspects":null
                  },
                  {
                     "id":"urn:infai:ses:aspect:7186cbff-ca46-4ea0-a892-0765c0633cb1",
                     "name":"Passenger",
                     "sub_aspects":null
                  }
               ]
            },
            {
               "id":"urn:infai:ses:aspect:78c065bf-3d73-461a-99c7-2be05168e6f3",
               "name":"Wiper",
               "sub_aspects":null
            },
            {
               "id":"urn:infai:ses:aspect:97ce06b8-76cd-4b04-8f43-45cd499e923b",
               "name":"Defroster",
               "sub_aspects":[
                  {
                     "id":"urn:infai:ses:aspect:a098cb37-3759-4570-bc6b-d8b067012701",
                     "name":"Front",
                     "sub_aspects":null
                  },
                  {
                     "id":"urn:infai:ses:aspect:b73e4438-044b-407e-9555-443474ef91d9",
                     "name":"Back",
                     "sub_aspects":null
                  }
               ]
            },
            {
               "id":"urn:infai:ses:aspect:4714a814-dbb9-4132-8a2e-fa85660dccec",
               "name":"Mirror",
               "sub_aspects":null
            },
            {
               "id":"urn:infai:ses:aspect:a83cd1a7-12c2-460f-90d2-051c1880c138",
               "name":"Steering Wheel",
               "sub_aspects":null
            }
         ]
      },
      {
         "id":"urn:infai:ses:aspect:3288e184-dd53-48bb-810b-146a44088199",
         "name":"Cooling",
         "sub_aspects":[
            
         ]
      },
      {
         "id":"urn:infai:ses:aspect:efabc932-ba9f-4ca1-b935-06830879053b",
         "name":"Vent",
         "sub_aspects":[
            {
               "id":"urn:infai:ses:aspect:6e7c50b0-9c46-4f8a-ad28-b0a4199585b1",
               "name":"Driver",
               "sub_aspects":null
            },
            {
               "id":"urn:infai:ses:aspect:f4275e0f-513c-441d-9dc1-e81a3d7b5a26",
               "name":"Passenger",
               "sub_aspects":null
            }
         ]
      }
   ]
}`
	changeStr := `{
   "id":"urn:infai:ses:aspect:60f8f6a6-c0e0-4fc3-9c11-bf2ccef1ed11",
   "name":"HVAC",
   "sub_aspects":[
      {
         "id":"urn:infai:ses:aspect:1d69d3b6-d16f-430c-bf58-f23b26cd84c4",
         "name":"Heating",
         "sub_aspects":[
            {
               "id":"urn:infai:ses:aspect:78c065bf-3d73-461a-99c7-2be05168e6f3",
               "name":"Wiper",
               "sub_aspects":null
            },
            {
               "id":"urn:infai:ses:aspect:97ce06b8-76cd-4b04-8f43-45cd499e923b",
               "name":"Defroster",
               "sub_aspects":[
                  {
                     "id":"urn:infai:ses:aspect:a098cb37-3759-4570-bc6b-d8b067012701",
                     "name":"Front",
                     "sub_aspects":null
                  },
                  {
                     "id":"urn:infai:ses:aspect:b73e4438-044b-407e-9555-443474ef91d9",
                     "name":"Back",
                     "sub_aspects":null
                  }
               ]
            },
            {
               "id":"urn:infai:ses:aspect:4714a814-dbb9-4132-8a2e-fa85660dccec",
               "name":"Mirror",
               "sub_aspects":null
            },
            {
               "id":"urn:infai:ses:aspect:a83cd1a7-12c2-460f-90d2-051c1880c138",
               "name":"Steering Wheel",
               "sub_aspects":null
            }
         ]
      },
      {
         "id":"urn:infai:ses:aspect:3288e184-dd53-48bb-810b-146a44088199",
         "name":"Cooling",
         "sub_aspects":[
            
         ]
      },
      {
         "id":"urn:infai:ses:aspect:efabc932-ba9f-4ca1-b935-06830879053b",
         "name":"Vent",
         "sub_aspects":[
            {
               "id":"urn:infai:ses:aspect:6e7c50b0-9c46-4f8a-ad28-b0a4199585b1",
               "name":"Driver",
               "sub_aspects":null
            },
            {
               "id":"urn:infai:ses:aspect:f4275e0f-513c-441d-9dc1-e81a3d7b5a26",
               "name":"Passenger",
               "sub_aspects":null
            }
         ]
      },
      {
         "id":"urn:infai:ses:aspect:ac049c69-8e88-4306-817d-fb1fa10885e5",
         "name":"Seat",
         "sub_aspects":[
            {
               "id":"urn:infai:ses:aspect:5b00764e-2470-400c-ae0c-d84ac10dd32d",
               "name":"Driver",
               "sub_aspects":null
            },
            {
               "id":"urn:infai:ses:aspect:7186cbff-ca46-4ea0-a892-0765c0633cb1",
               "name":"Passenger",
               "sub_aspects":null
            }
         ]
      }
   ]
}`

	usedSubAspect := "urn:infai:ses:aspect:5b00764e-2470-400c-ae0c-d84ac10dd32d"

	initialAspect := model.Aspect{}
	err = json.Unmarshal([]byte(initialStr), &initialAspect)
	if err != nil {
		t.Error(err)
		return
	}

	err = ctrl.SetAspect(initialAspect, userid)
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
							AspectId:         usedSubAspect,
						},
					},
				},
			},
		},
	}
	err = ctrl.SetDeviceType(dt, userid)
	if err != nil {
		t.Error(err)
		return
	}

	changedAspect := model.Aspect{}
	err = json.Unmarshal([]byte(changeStr), &changedAspect)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("validate", func(t *testing.T) {
		err, _ = ctrl.ValidateAspect(changedAspect)
		if err != nil {
			t.Error(err)
			return
		}
	})

}

func testAspectEditValidation(t *testing.T, config config.Config, aspect model.Aspect, expectedCode int) error {
	t.Helper()
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(aspect)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", "http://localhost:"+config.ServerPort+"/aspects/"+url.PathEscape(aspect.Id)+"?dry-run=true", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", userjwt)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != expectedCode {
		temp, _ := io.ReadAll(resp.Body)
		t.Log(string(temp))
		return errors.New(resp.Status)
	}
	return nil
}

func TestSubAspectMovePartial(t *testing.T) {
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
		Id:   "a1",
		Name: "a1",
		SubAspects: []model.Aspect{
			{
				Id:   "a1.1",
				Name: "a1.1",
				SubAspects: []model.Aspect{
					{
						Id:   "a1.1.1",
						Name: "a1.1.1",
					},
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishAspect(model.Aspect{
		Id:   "a2",
		Name: "a2",
		SubAspects: []model.Aspect{
			{
				Id:   "a2.1",
				Name: "a2.1",
				SubAspects: []model.Aspect{
					{
						Id:   "a2.1.1",
						Name: "a2.1.1",
					},
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	t.Run("aspect-nodes before move", testGetRequest(userjwt, conf, "/aspect-nodes", []model.AspectNode{
		{
			Id:            "a1",
			Name:          "a1",
			RootId:        "a1",
			AncestorIds:   []string{},
			ChildIds:      []string{"a1.1"},
			DescendentIds: []string{"a1.1", "a1.1.1"},
		},
		{
			Id:            "a1.1",
			Name:          "a1.1",
			RootId:        "a1",
			ParentId:      "a1",
			AncestorIds:   []string{"a1"},
			ChildIds:      []string{"a1.1.1"},
			DescendentIds: []string{"a1.1.1"},
		},
		{
			Id:            "a1.1.1",
			Name:          "a1.1.1",
			RootId:        "a1",
			ParentId:      "a1.1",
			AncestorIds:   []string{"a1", "a1.1"},
			ChildIds:      []string{},
			DescendentIds: []string{},
		},

		{
			Id:            "a2",
			Name:          "a2",
			RootId:        "a2",
			AncestorIds:   []string{},
			ChildIds:      []string{"a2.1"},
			DescendentIds: []string{"a2.1", "a2.1.1"},
		},
		{
			Id:            "a2.1",
			Name:          "a2.1",
			RootId:        "a2",
			ParentId:      "a2",
			AncestorIds:   []string{"a2"},
			ChildIds:      []string{"a2.1.1"},
			DescendentIds: []string{"a2.1.1"},
		},
		{
			Id:            "a2.1.1",
			Name:          "a2.1.1",
			RootId:        "a2",
			ParentId:      "a2.1",
			AncestorIds:   []string{"a2", "a2.1"},
			ChildIds:      []string{},
			DescendentIds: []string{},
		},
	}))

	t.Run("aspect-nodes move", func(t *testing.T) {
		err = producer.PublishAspect(model.Aspect{
			Id:   "a1",
			Name: "a1",
			SubAspects: []model.Aspect{
				{
					Id:   "a1.1",
					Name: "a1.1",
					SubAspects: []model.Aspect{
						{
							Id:   "a1.1.1",
							Name: "a1.1.1",
						},
					},
				},
				{
					Id:   "a2.1",
					Name: "a2.1",
					SubAspects: []model.Aspect{
						{
							Id:   "a2.1.1",
							Name: "a2.1.1",
						},
					},
				},
			},
		}, userid)
		if err != nil {
			t.Error(err)
			return
		}
	})

	time.Sleep(2 * time.Second)

	t.Run("aspect-nodes after move", testGetRequest(userjwt, conf, "/aspect-nodes", []model.AspectNode{
		{
			Id:            "a1",
			Name:          "a1",
			RootId:        "a1",
			AncestorIds:   []string{},
			ChildIds:      []string{"a1.1", "a2.1"},
			DescendentIds: []string{"a1.1", "a1.1.1", "a2.1", "a2.1.1"},
		},
		{
			Id:            "a1.1",
			Name:          "a1.1",
			RootId:        "a1",
			ParentId:      "a1",
			AncestorIds:   []string{"a1"},
			ChildIds:      []string{"a1.1.1"},
			DescendentIds: []string{"a1.1.1"},
		},
		{
			Id:            "a1.1.1",
			Name:          "a1.1.1",
			RootId:        "a1",
			ParentId:      "a1.1",
			AncestorIds:   []string{"a1", "a1.1"},
			ChildIds:      []string{},
			DescendentIds: []string{},
		},

		{
			Id:            "a2",
			Name:          "a2",
			RootId:        "a2",
			AncestorIds:   []string{},
			ChildIds:      []string{},
			DescendentIds: []string{},
		},
		{
			Id:            "a2.1",
			Name:          "a2.1",
			RootId:        "a1",
			ParentId:      "a1",
			AncestorIds:   []string{"a1"},
			ChildIds:      []string{"a2.1.1"},
			DescendentIds: []string{"a2.1.1"},
		},
		{
			Id:            "a2.1.1",
			Name:          "a2.1.1",
			RootId:        "a1",
			ParentId:      "a2.1",
			AncestorIds:   []string{"a1", "a2.1"},
			ChildIds:      []string{},
			DescendentIds: []string{},
		},
	}))
}

func TestSubAspectMoveRoot(t *testing.T) {
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
		Id:   "a1",
		Name: "a1",
		SubAspects: []model.Aspect{
			{
				Id:   "a1.1",
				Name: "a1.1",
				SubAspects: []model.Aspect{
					{
						Id:   "a1.1.1",
						Name: "a1.1.1",
					},
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	err = producer.PublishAspect(model.Aspect{
		Id:   "a2",
		Name: "a2",
		SubAspects: []model.Aspect{
			{
				Id:   "a2.1",
				Name: "a2.1",
				SubAspects: []model.Aspect{
					{
						Id:   "a2.1.1",
						Name: "a2.1.1",
					},
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	t.Run("aspect-nodes before move", testGetRequest(userjwt, conf, "/aspect-nodes", []model.AspectNode{
		{
			Id:            "a1",
			Name:          "a1",
			RootId:        "a1",
			AncestorIds:   []string{},
			ChildIds:      []string{"a1.1"},
			DescendentIds: []string{"a1.1", "a1.1.1"},
		},
		{
			Id:            "a1.1",
			Name:          "a1.1",
			RootId:        "a1",
			ParentId:      "a1",
			AncestorIds:   []string{"a1"},
			ChildIds:      []string{"a1.1.1"},
			DescendentIds: []string{"a1.1.1"},
		},
		{
			Id:            "a1.1.1",
			Name:          "a1.1.1",
			RootId:        "a1",
			ParentId:      "a1.1",
			AncestorIds:   []string{"a1", "a1.1"},
			ChildIds:      []string{},
			DescendentIds: []string{},
		},

		{
			Id:            "a2",
			Name:          "a2",
			RootId:        "a2",
			AncestorIds:   []string{},
			ChildIds:      []string{"a2.1"},
			DescendentIds: []string{"a2.1", "a2.1.1"},
		},
		{
			Id:            "a2.1",
			Name:          "a2.1",
			RootId:        "a2",
			ParentId:      "a2",
			AncestorIds:   []string{"a2"},
			ChildIds:      []string{"a2.1.1"},
			DescendentIds: []string{"a2.1.1"},
		},
		{
			Id:            "a2.1.1",
			Name:          "a2.1.1",
			RootId:        "a2",
			ParentId:      "a2.1",
			AncestorIds:   []string{"a2", "a2.1"},
			ChildIds:      []string{},
			DescendentIds: []string{},
		},
	}))

	t.Run("aspect-nodes move", func(t *testing.T) {
		err = producer.PublishAspect(model.Aspect{
			Id:   "a1",
			Name: "a1",
			SubAspects: []model.Aspect{
				{
					Id:   "a1.1",
					Name: "a1.1",
					SubAspects: []model.Aspect{
						{
							Id:   "a1.1.1",
							Name: "a1.1.1",
						},
					},
				},
				{
					Id:   "a2",
					Name: "a2",
					SubAspects: []model.Aspect{
						{
							Id:   "a2.1",
							Name: "a2.1",
							SubAspects: []model.Aspect{
								{
									Id:   "a2.1.1",
									Name: "a2.1.1",
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
	})

	time.Sleep(2 * time.Second)

	t.Run("aspect-nodes after move", testGetRequest(userjwt, conf, "/aspect-nodes", []model.AspectNode{
		{
			Id:            "a1",
			Name:          "a1",
			RootId:        "a1",
			AncestorIds:   []string{},
			ChildIds:      []string{"a1.1", "a2"},
			DescendentIds: []string{"a1.1", "a1.1.1", "a2", "a2.1", "a2.1.1"},
		},
		{
			Id:            "a1.1",
			Name:          "a1.1",
			RootId:        "a1",
			ParentId:      "a1",
			AncestorIds:   []string{"a1"},
			ChildIds:      []string{"a1.1.1"},
			DescendentIds: []string{"a1.1.1"},
		},
		{
			Id:            "a1.1.1",
			Name:          "a1.1.1",
			RootId:        "a1",
			ParentId:      "a1.1",
			AncestorIds:   []string{"a1", "a1.1"},
			ChildIds:      []string{},
			DescendentIds: []string{},
		},

		{
			Id:            "a2",
			Name:          "a2",
			RootId:        "a1",
			ParentId:      "a1",
			AncestorIds:   []string{"a1"},
			ChildIds:      []string{"a2.1"},
			DescendentIds: []string{"a2.1", "a2.1.1"},
		},
		{
			Id:            "a2.1",
			Name:          "a2.1",
			RootId:        "a1",
			ParentId:      "a2",
			AncestorIds:   []string{"a1", "a2"},
			ChildIds:      []string{"a2.1.1"},
			DescendentIds: []string{"a2.1.1"},
		},
		{
			Id:            "a2.1.1",
			Name:          "a2.1.1",
			RootId:        "a1",
			ParentId:      "a2.1",
			AncestorIds:   []string{"a1", "a2", "a2.1"},
			ChildIds:      []string{},
			DescendentIds: []string{},
		},
	}))
}

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
