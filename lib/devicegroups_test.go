/*
 * Copyright 2020 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/testutils/mocks"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"sync"
	"testing"
	"time"
)

func TestDeviceGroupsValidation(t *testing.T) {
	dbMock := mocks.NewDatabase()
	dbMock.SetDeviceType(nil, model.DeviceType{
		Id:            "dt1",
		DeviceClassId: "dcid",
		Services: []model.Service{{
			Id:          "s1id",
			Interaction: model.REQUEST,
			AspectIds:   []string{"aid"},
			FunctionIds: []string{"fid"},
		}},
	})

	dbMock.SetDevice(nil, model.Device{
		Id:           "did",
		DeviceTypeId: "dt1",
	})
	ctrl, err := controller.New(config.Config{}, dbMock, nil, nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Run("minimal ok", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
	}, http.StatusOK, false))

	t.Run("ok with image", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:    "id",
		Name:  "name",
		Image: "imageUrl",
	}, http.StatusOK, false))

	t.Run("missing id", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Name: "name",
	}, http.StatusBadRequest, true))

	t.Run("missing name", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Name: "id",
	}, http.StatusBadRequest, true))

	t.Run("ok with one dc criteria and no device", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "dcid",
			Interaction:   model.REQUEST,
		}},
	}, http.StatusOK, false))

	t.Run("ok with one dc criteria and one device", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "dcid",
			Interaction:   model.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusOK, false))

	t.Run("ok with one aspect criteria and no device", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "aid",
			Interaction: model.REQUEST,
		}},
	}, http.StatusOK, false))

	t.Run("ok with one aspect criteria and one device", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "aid",
			Interaction: model.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusOK, false))

	t.Run("device uses blocked interaction", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "aid",
			Interaction: model.EVENT,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusBadRequest, true))

	t.Run("wrong aspect", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:  "fid",
			AspectId:    "aid_unknown",
			Interaction: model.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusBadRequest, true))

	t.Run("wrong function", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:  "fid_unknown",
			AspectId:    "aid",
			Interaction: model.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusBadRequest, true))

	t.Run("wrong device-class", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
			Interaction:   model.REQUEST,
		}},
		DeviceIds: []string{"did"},
	}, http.StatusBadRequest, true))
}

func testDeviceGroupValidation(ctrl *controller.Controller, group model.DeviceGroup, expectedStatusCode int, expectError bool) func(t *testing.T) {
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
	conf := config.Config{DeviceGroupTopic: "device-group"}
	sec := mocks.NewSecurity()
	ctrl, err := controller.New(conf, nil, sec, nil)
	if err != nil {
		t.Error(err)
		return
	}

	sec.Set(conf.DeviceTopic, "d1", true)
	sec.Set(conf.DeviceTopic, "d2", true)
	sec.Set(conf.DeviceTopic, "d3", true)

	t.Run("empty", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
	}, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
	}))

	t.Run("empty 2", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
	}, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
	}))

	t.Run("empty 3", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:       "id",
		Name:     "name",
		Criteria: []model.DeviceGroupFilterCriteria{},
	}, model.DeviceGroup{
		Id:       "id",
		Name:     "name",
		Criteria: []model.DeviceGroupFilterCriteria{},
	}))

	t.Run("full access", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
		DeviceIds: []string{"d1", "d2", "d3"},
	}, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
		DeviceIds: []string{"d1", "d2", "d3"},
	}))
	t.Run("one access missing", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
		DeviceIds: []string{"d1", "d2", "unknown", "d3"},
	}, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "unknown",
		}},
		DeviceIds: []string{"d1", "d2", "d3"},
	}))
}

func testDeviceGroupsDeviceFilter(ctrl *controller.Controller, group model.DeviceGroup, expectedResult model.DeviceGroup) func(t *testing.T) {
	return func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error(r, "\n", string(debug.Stack()))
			}
		}()
		result, err, code := ctrl.FilterDevicesOfGroupByAccess(jwt_http_router.Jwt{Impersonate: userjwt}, group)
		if err != nil {
			t.Error(err)
			return
		}
		if code != http.StatusOK {
			t.Error(code)
			return
		}
		if !reflect.DeepEqual(result, expectedResult) {
			t.Error(result, expectedResult)
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

	err = producer.PublishDeviceType(model.DeviceType{
		Id:            devicetype1id,
		Name:          devicetype1name,
		DeviceClassId: "dcid",
		Services: []model.Service{{
			Id:          "s1id",
			Interaction: model.REQUEST,
			AspectIds:   []string{"aid"},
			FunctionIds: []string{"fid"},
		}},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	d1 := model.Device{
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

	dg1 := model.DeviceGroup{
		Id:   devicegroup1id,
		Name: devicegroup1name,
		Criteria: []model.DeviceGroupFilterCriteria{{
			FunctionId:    "fid",
			DeviceClassId: "dcid",
			Interaction:   model.REQUEST,
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

func testDeviceGroupReadNotFound(t *testing.T, conf config.Config, id string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/device-groups/" + url.PathEscape(id)
	resp, err := userjwt.Get(endpoint)
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

func testDeviceGroupRead(t *testing.T, conf config.Config, expectedDeviceGroupss ...model.DeviceGroup) {
	for _, expected := range expectedDeviceGroupss {
		endpoint := "http://localhost:" + conf.ServerPort + "/device-groups/" + url.PathEscape(expected.Id)
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

		result := model.DeviceGroup{}
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
