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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/testutils/mocks"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
	"reflect"
	"runtime/debug"
	"testing"
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

	t.Run("ok with blocked interaction", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:                 "id",
		Name:               "name",
		BlockedInteraction: model.EVENT,
	}, http.StatusOK, false))

	t.Run("unknown interaction", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:                 "id",
		Name:               "name",
		BlockedInteraction: "foo",
	}, http.StatusBadRequest, true))

	t.Run("missing id", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Name: "name",
	}, http.StatusBadRequest, true))

	t.Run("missing name", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Name: "id",
	}, http.StatusBadRequest, true))

	t.Run("ok with one dc criteria and no device", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "dcid",
			},
		}},
	}, http.StatusOK, false))

	t.Run("ok with one dc criteria and one device", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "dcid",
			},
			Selection: []model.Selection{{
				DeviceId:   "did",
				ServiceIds: []string{"s1id"},
			}},
		}},
	}, http.StatusOK, false))

	t.Run("ok with one aspect criteria and no device", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId: "fid",
				AspectId:   "aid",
			},
		}},
	}, http.StatusOK, false))

	t.Run("ok with one aspect criteria and one device", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId: "fid",
				AspectId:   "aid",
			},
			Selection: []model.Selection{{
				DeviceId:   "did",
				ServiceIds: []string{"s1id"},
			}},
		}},
	}, http.StatusOK, false))

	t.Run("device uses blocked interaction", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:                 "id",
		Name:               "name",
		BlockedInteraction: model.REQUEST,
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId: "fid",
				AspectId:   "aid",
			},
			Selection: []model.Selection{{
				DeviceId:   "did",
				ServiceIds: []string{"s1id"},
			}},
		}},
	}, http.StatusBadRequest, true))

	t.Run("asymetric device usage", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{
			{
				Criteria: model.FilterCriteria{
					FunctionId: "fid",
					AspectId:   "aid",
				},
				Selection: []model.Selection{{
					DeviceId:   "did",
					ServiceIds: []string{"s1id"},
				}},
			},
			{
				Criteria: model.FilterCriteria{
					FunctionId: "fid",
					AspectId:   "aid",
				},
				Selection: []model.Selection{},
			},
		},
	}, http.StatusBadRequest, true))

	t.Run("wrong aspect", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId: "fid",
				AspectId:   "aid_unknown",
			},
			Selection: []model.Selection{{
				DeviceId:   "did",
				ServiceIds: []string{"s1id"},
			}},
		}},
	}, http.StatusBadRequest, true))

	t.Run("wrong function", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId: "fid_unknown",
				AspectId:   "aid",
			},
			Selection: []model.Selection{{
				DeviceId:   "did",
				ServiceIds: []string{"s1id"},
			}},
		}},
	}, http.StatusBadRequest, true))

	t.Run("wrong device-class", testDeviceGroupValidation(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "unknown",
			},
			Selection: []model.Selection{{
				DeviceId:   "did",
				ServiceIds: []string{"s1id"},
			}},
		}},
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

	sec.Set(conf.DeviceGroupTopic, "d1", true)
	sec.Set(conf.DeviceGroupTopic, "d2", true)
	sec.Set(conf.DeviceGroupTopic, "d3", true)

	t.Run("empty", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "unknown",
			},
			Selection: []model.Selection{},
		}},
	}, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "unknown",
			},
			Selection: []model.Selection{},
		}},
	}))

	t.Run("empty 2", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "unknown",
			},
		}},
	}, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "unknown",
			},
			Selection: []model.Selection{},
		}},
	}))

	t.Run("empty 3", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:      "id",
		Name:    "name",
		Devices: []model.DeviceGroupMapping{},
	}, model.DeviceGroup{
		Id:      "id",
		Name:    "name",
		Devices: []model.DeviceGroupMapping{},
	}))

	t.Run("full access", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "unknown",
			},
			Selection: []model.Selection{
				{DeviceId: "d1", ServiceIds: []string{"s1id"}},
				{DeviceId: "d2"},
				{DeviceId: "d3"},
			},
		}},
	}, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "unknown",
			},
			Selection: []model.Selection{
				{DeviceId: "d1", ServiceIds: []string{"s1id"}},
				{DeviceId: "d2"},
				{DeviceId: "d3"},
			},
		}},
	}))
	t.Run("one access missing", testDeviceGroupsDeviceFilter(ctrl, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "unknown",
			},
			Selection: []model.Selection{
				{DeviceId: "d1", ServiceIds: []string{"s1id"}},
				{DeviceId: "d2"},
				{DeviceId: "unknown", ServiceIds: []string{"s1id"}},
				{DeviceId: "d3"},
			},
		}},
	}, model.DeviceGroup{
		Id:   "id",
		Name: "name",
		Devices: []model.DeviceGroupMapping{{
			Criteria: model.FilterCriteria{
				FunctionId:    "fid",
				DeviceClassId: "unknown",
			},
			Selection: []model.Selection{
				{DeviceId: "d1", ServiceIds: []string{"s1id"}},
				{DeviceId: "d2"},
				{DeviceId: "d3"},
			},
		}},
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

func TestDeviceGroups(t *testing.T) {
	t.Error("not implemented")
}
