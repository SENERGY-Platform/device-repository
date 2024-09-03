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
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/lib/database/testdb"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"sync"
	"testing"
	"time"
)

type aspectNodeProviderMock struct {
	result []models.AspectNode
}

func (this *aspectNodeProviderMock) ListAspectNodesByIdList(ctx context.Context, ids []string) ([]models.AspectNode, error) {
	return this.result, nil
}

func TestFilterGenericDuplicateCriteria(t *testing.T) {
	result, err := controller.DeviceGroupFilterGenericDuplicateCriteria(models.DeviceGroup{}, &aspectNodeProviderMock{result: []models.AspectNode{}})
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(result, models.DeviceGroup{Criteria: []models.DeviceGroupFilterCriteria{}, CriteriaShort: []string{}}) {
		t.Errorf("%#v", result)
	}

	result, err = controller.DeviceGroupFilterGenericDuplicateCriteria(models.DeviceGroup{
		Id:        "id",
		Name:      "name",
		Image:     "image",
		DeviceIds: []string{"did1", "did2"},
		Attributes: []models.Attribute{{
			Key:    "attr1",
			Value:  "attrv1",
			Origin: "o1",
		}, {
			Key:    "attr2",
			Value:  "attrv2",
			Origin: "o2",
		}},
		Criteria: []models.DeviceGroupFilterCriteria{
			{
				Interaction:   "keep",
				FunctionId:    "keep",
				AspectId:      "keep",
				DeviceClassId: "keep",
			},
			{
				Interaction:   "keep2",
				FunctionId:    "keep2",
				DeviceClassId: "keep2",
			},
			{
				Interaction: "i1",
				FunctionId:  "f1",
				AspectId:    "a1",
			},
			{
				Interaction: "i1",
				FunctionId:  "f1",
			},
			{
				Interaction: "i2",
				FunctionId:  "f2",
			},
			{
				Interaction:   "i3",
				FunctionId:    "f3",
				DeviceClassId: "dc3",
			},
			//should never happen, because the function id should be different for criteria with aspects and criteria with device-classes
			{
				Interaction: "i3",
				FunctionId:  "f3",
				AspectId:    "a3",
			},
		},
		CriteriaShort: nil,
	}, &aspectNodeProviderMock{result: []models.AspectNode{}})
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(result, models.DeviceGroup{
		Id:        "id",
		Name:      "name",
		Image:     "image",
		DeviceIds: []string{"did1", "did2"},
		Attributes: []models.Attribute{{
			Key:    "attr1",
			Value:  "attrv1",
			Origin: "o1",
		}, {
			Key:    "attr2",
			Value:  "attrv2",
			Origin: "o2",
		}},
		Criteria: []models.DeviceGroupFilterCriteria{
			{
				Interaction:   "keep",
				FunctionId:    "keep",
				AspectId:      "keep",
				DeviceClassId: "keep",
			},
			{
				Interaction:   "keep2",
				FunctionId:    "keep2",
				DeviceClassId: "keep2",
			},
			{
				Interaction: "i1",
				FunctionId:  "f1",
				AspectId:    "a1",
			},
			{
				Interaction: "i2",
				FunctionId:  "f2",
			},
			{
				Interaction:   "i3",
				FunctionId:    "f3",
				DeviceClassId: "dc3",
			},
			//should never happen, because the function id should be different for criteria with aspects and criteria with device-classes
			{
				Interaction: "i3",
				FunctionId:  "f3",
				AspectId:    "a3",
			},
		},
		CriteriaShort: []string{"keep_keep_keep_keep", "keep2__keep2_keep2", "f1_a1__i1", "f2___i2", "f3__dc3_i3", "f3_a3__i3"},
	}) {
		t.Errorf("%#v", result)
	}

	result, err = controller.DeviceGroupFilterGenericDuplicateCriteria(models.DeviceGroup{
		Id:        "id",
		Name:      "name",
		Image:     "image",
		DeviceIds: []string{"did1", "did2"},
		Attributes: []models.Attribute{{
			Key:    "attr1",
			Value:  "attrv1",
			Origin: "o1",
		}, {
			Key:    "attr2",
			Value:  "attrv2",
			Origin: "o2",
		}},
		Criteria: []models.DeviceGroupFilterCriteria{
			{ //keep
				Interaction:   "i1",
				FunctionId:    "f1",
				AspectId:      "a1",
				DeviceClassId: "dc1",
			},
			{ //keep
				Interaction: "i2",
				FunctionId:  "f2",
				AspectId:    "a2",
			},
			{ //removed
				Interaction: "i2",
				FunctionId:  "f2",
			},
			{ //keep
				Interaction: "i3",
				FunctionId:  "f3",
			},
			{ //remove
				Interaction: "i4",
				FunctionId:  "f4",
				AspectId:    "base",
			},
			{ //keep
				Interaction: "i4",
				FunctionId:  "f4",
				AspectId:    "child1",
			},
			{ //keep
				Interaction: "i4",
				FunctionId:  "f4",
				AspectId:    "child2",
			},
			{ //remove
				Interaction: "i5",
				FunctionId:  "f5",
				AspectId:    "",
			},
			{ //keep
				Interaction: "i5",
				FunctionId:  "f5",
				AspectId:    "child1",
			},
			{ //keep
				Interaction: "i5",
				FunctionId:  "f5",
				AspectId:    "child2",
			},

			{ //remove
				Interaction: "i6",
				FunctionId:  "f6",
				AspectId:    "",
			},
			{ //keep
				Interaction: "i6",
				FunctionId:  "f6",
				AspectId:    "base",
			},
		},
		CriteriaShort: nil,
	}, &aspectNodeProviderMock{result: []models.AspectNode{
		{
			Id:            "a1",
			DescendentIds: []string{},
		},
		{
			Id:            "a2",
			DescendentIds: []string{},
		},
		{
			Id:            "base",
			DescendentIds: []string{"child1", "child2"},
		},
	}})
	if err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(result, models.DeviceGroup{
		Id:        "id",
		Name:      "name",
		Image:     "image",
		DeviceIds: []string{"did1", "did2"},
		Attributes: []models.Attribute{{
			Key:    "attr1",
			Value:  "attrv1",
			Origin: "o1",
		}, {
			Key:    "attr2",
			Value:  "attrv2",
			Origin: "o2",
		}},
		Criteria: []models.DeviceGroupFilterCriteria{
			{ //keep
				Interaction:   "i1",
				FunctionId:    "f1",
				AspectId:      "a1",
				DeviceClassId: "dc1",
			},
			{ //keep
				Interaction: "i2",
				FunctionId:  "f2",
				AspectId:    "a2",
			},
			{ //keep
				Interaction: "i3",
				FunctionId:  "f3",
			},
			{ //keep
				Interaction: "i4",
				FunctionId:  "f4",
				AspectId:    "child1",
			},
			{ //keep
				Interaction: "i4",
				FunctionId:  "f4",
				AspectId:    "child2",
			},
			{ //keep
				Interaction: "i5",
				FunctionId:  "f5",
				AspectId:    "child1",
			},
			{ //keep
				Interaction: "i5",
				FunctionId:  "f5",
				AspectId:    "child2",
			},
			{ //keep
				Interaction: "i6",
				FunctionId:  "f6",
				AspectId:    "base",
			},
		},
		CriteriaShort: []string{"f1_a1_dc1_i1", "f2_a2__i2", "f3___i3", "f4_child1__i4", "f4_child2__i4", "f5_child1__i5", "f5_child2__i5", "f6_base__i6"},
	}) {
		t.Errorf("%#v", result)
	}
}

func TestDeviceGroupsValidation(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := config.Load("../../config.json")
	if err != nil {
		log.Println("ERROR: unable to load config: ", err)
		t.Error(err)
		return
	}
	conf.FatalErrHandler = t.Fatal
	conf.MongoReplSet = false
	conf, err = docker.NewEnv(ctx, wg, conf)
	if err != nil {
		log.Println("ERROR: unable to create docker env", err)
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)
	dbMock, err := mongo.New(conf)
	if err != nil {
		t.Error(err)
		return
	}
	ctrl, err := controller.New(config.Config{}, dbMock, nil)
	if err != nil {
		t.Error(err)
		return
	}
	err = ctrl.SetAspect(models.Aspect{
		Id:   "parent",
		Name: "parent",
		SubAspects: []models.Aspect{
			{
				Id:   "aid",
				Name: "aid",
				SubAspects: []models.Aspect{
					{
						Id:   "child",
						Name: "child",
					},
				},
			},
		},
	}, "")
	if err != nil {
		t.Error(err)
		return
	}
	err = ctrl.SetDeviceType(models.DeviceType{
		Id:            "dt1",
		DeviceClassId: "dcid",
		Services: []models.Service{{
			Id:          "s1id",
			Interaction: models.REQUEST,
			Outputs: []models.Content{
				{
					ContentVariable: models.ContentVariable{
						FunctionId: "fid",
						AspectId:   "aid",
					},
				},
			},
		}},
	}, "")
	if err != nil {
		t.Error(err)
		return
	}
	err = ctrl.SetDevice(models.Device{
		Id:           "did",
		DeviceTypeId: "dt1",
	}, "")
	if err != nil {
		t.Error(err)
		return
	}
	t.Run("minimal ok", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
	}, http.StatusOK, false))

	t.Run("ok with image", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:    "id",
		Name:  "name",
		Image: "imageUrl",
	}, http.StatusOK, false))

	t.Run("missing id", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Name: "name",
	}, http.StatusBadRequest, true))

	t.Run("missing name", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Name: "id",
	}, http.StatusBadRequest, true))

	t.Run("ok with one dc criteria and no device", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "dcid",
			Interaction:   models.REQUEST,
		}},
	}, http.StatusOK, false))

	t.Run("ok with one dc criteria and one device", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "dcid",
			Interaction:   models.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusOK, false))

	t.Run("ok with one aspect criteria and no device", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "aid",
			Interaction: models.REQUEST,
		}},
	}, http.StatusOK, false))

	t.Run("ok with one parent aspect criteria and no device", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "parent",
			Interaction: models.REQUEST,
		}},
	}, http.StatusOK, false))

	t.Run("ok with one aspect criteria and one device", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "aid",
			Interaction: models.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusOK, false))

	t.Run("ok with one parent aspect criteria and one device", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "parent",
			Interaction: models.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusOK, false))

	t.Run("not ok with one child aspect criteria and one device", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "child",
			Interaction: models.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusBadRequest, true))

	t.Run("device uses blocked interaction", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "aid",
			Interaction: models.EVENT,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusBadRequest, true))

	t.Run("wrong aspect", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "aid_unknown",
			Interaction: models.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusBadRequest, true))

	t.Run("wrong function", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:  "fid_unknown",
			AspectId:    "aid",
			Interaction: models.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusBadRequest, true))

	t.Run("wrong device-class", testDeviceGroupValidation(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
			Interaction:   models.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusBadRequest, true))
}

func testDeviceGroupValidation(ctrl *controller.Controller, group models.DeviceGroup, expectedStatusCode int, expectError bool) func(t *testing.T) {
	return func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error(r, "\n", string(debug.Stack()))
			}
		}()
		err, code := ctrl.ValidateDeviceGroup(group)
		if (err != nil) != expectError {
			t.Error(expectError, err)
			return
		}
		if code != expectedStatusCode {
			t.Error(expectedStatusCode, code)
			return
		}
	}
}

func TestDeviceGroupsDeviceFilter(t *testing.T) {
	conf := config.Config{DeviceGroupTopic: "device-group", DeviceTopic: "devices"}
	db := testdb.NewTestDB(conf)
	ctrl, err := controller.New(conf, db, nil)
	if err != nil {
		t.Error(err)
		return
	}

	err = db.SetRights(conf.DeviceTopic, "d1", model.ResourceRights{UserRights: map[string]model.Right{userid: {Read: true, Write: true, Execute: true, Administrate: true}}})
	if err != nil {
		t.Error(err)
		return
	}
	err = db.SetRights(conf.DeviceTopic, "d2", model.ResourceRights{UserRights: map[string]model.Right{userid: {Read: true, Write: true, Execute: true, Administrate: true}}})
	if err != nil {
		t.Error(err)
		return
	}
	err = db.SetRights(conf.DeviceTopic, "d3", model.ResourceRights{UserRights: map[string]model.Right{userid: {Read: true, Write: true, Execute: true, Administrate: true}}})
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("empty", testDeviceGroupsDeviceFilter(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
	}, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
	}))

	t.Run("empty 2", testDeviceGroupsDeviceFilter(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
	}, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
	}))

	t.Run("empty 3", testDeviceGroupsDeviceFilter(ctrl, models.DeviceGroup{
		Id:       "id",
		Name:     "name",
		Criteria: []models.DeviceGroupFilterCriteria{},
	}, models.DeviceGroup{
		Id:       "id",
		Name:     "name",
		Criteria: []models.DeviceGroupFilterCriteria{},
	}))

	t.Run("full access", testDeviceGroupsDeviceFilter(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
		DeviceIds: []string{"d1", "d2", "d3"},
	}, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
		DeviceIds: []string{"d1", "d2", "d3"},
	}))
	t.Run("one access missing", testDeviceGroupsDeviceFilter(ctrl, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
		DeviceIds: []string{"d1", "d2", "unknown", "d3"},
	}, models.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
		DeviceIds: []string{"d1", "d2", "d3"},
	}))
}

func testDeviceGroupsDeviceFilter(ctrl *controller.Controller, group models.DeviceGroup, expectedResult models.DeviceGroup) func(t *testing.T) {
	return func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error(r, "\n", string(debug.Stack()))
			}
		}()
		result, err, code := ctrl.FilterDevicesOfGroupByAccess(userjwt, group)
		if err != nil {
			t.Error(err)
			return
		}
		if code != http.StatusOK {
			t.Error(code)
			return
		}
		if !reflect.DeepEqual(result, expectedResult) {
			t.Errorf("\n%#v\n%#v\n", result, expectedResult)
			return
		}
	}
}

const devicegroup1id = "dg1id"
const devicegroup1name = "dg1name"

func TestDeviceGroupsIntegration(t *testing.T) {
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

	err = producer.PublishDeviceType(models.DeviceType{
		Id:            devicetype1id,
		Name:          devicetype1name,
		DeviceClassId: "dcid",
		Services: []models.Service{{
			Id:          "s1id",
			Interaction: models.REQUEST,
			Outputs: []models.Content{
				{ContentVariable: models.ContentVariable{
					FunctionId: "fid",
					AspectId:   "aid",
				}},
			},
		}},
	}, userid)
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

	err = producer.PublishDevice(d1, userid)
	if err != nil {
		t.Error(err)
		return
	}

	dg1 := models.DeviceGroup{
		Id:   devicegroup1id,
		Name: devicegroup1name,
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "dcid",
			Interaction:   models.REQUEST,
		}},
		DeviceIds: []string{device1id},
	}

	err = producer.PublishDeviceGroup(dg1, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	t.Run("not existing", func(t *testing.T) {
		testDeviceGroupReadNotFound(t, conf, "foobar")
	})
	t.Run("testDeviceRead", func(t *testing.T) {
		testDeviceGroupRead(t, conf, dg1)
	})
}

func TestDeviceGroupsAttributes(t *testing.T) {
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

	dg1 := models.DeviceGroup{
		Id:   devicegroup1id,
		Name: devicegroup1name,
		Criteria: []models.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "dcid",
			Interaction:   models.REQUEST,
		}},
		Attributes: []models.Attribute{
			{
				Key:    "a1",
				Value:  "v1",
				Origin: "test",
			},
			{
				Key:   "a2",
				Value: "v2",
			},
		},
	}

	err = producer.PublishDeviceGroup(dg1, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	t.Run("not existing", func(t *testing.T) {
		testDeviceGroupReadNotFound(t, conf, "foobar")
	})
	t.Run("testDeviceRead", func(t *testing.T) {
		testDeviceGroupRead(t, conf, dg1)
	})
}

func testDeviceGroupReadNotFound(t *testing.T, conf config.Config, id string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-groups/" + url.PathEscape(id)
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

func testDeviceGroupRead(t *testing.T, conf config.Config, expectedDeviceGroupss ...models.DeviceGroup) {
	for _, expected := range expectedDeviceGroupss {
		endpoint := "http://localhost:" + conf.ServerPort + "/device-groups/" + url.PathEscape(expected.Id)
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

		result := models.DeviceGroup{}
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
