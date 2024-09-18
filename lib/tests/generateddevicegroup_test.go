/*
 * Copyright 2024 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestGeneratedDeviceGroups(t *testing.T) {
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
		Id:   devicetype1id,
		Name: devicetype1name,
		Services: []models.Service{
			{
				Id:          "sid",
				LocalId:     "sid",
				Name:        "service",
				Interaction: models.EVENT_AND_REQUEST,
				ProtocolId:  "pid",
				Outputs: []models.Content{
					{
						Id: "c",
						ContentVariable: models.ContentVariable{
							Id:               "cv",
							Name:             "cv",
							Type:             "string",
							CharacteristicId: "cid",
							FunctionId:       "fid",
							AspectId:         "aid",
						},
						Serialization:     models.JSON,
						ProtocolSegmentId: "foo",
					},
				},
			},
		},
	}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	d1 := models.Device{
		Id:           device1id,
		LocalId:      device1lid,
		Name:         "a d1",
		DeviceTypeId: devicetype1id,
		OwnerId:      userid,
	}

	err = producer.PublishDevice(d1, userid)
	if err != nil {
		t.Error(err)
		return
	}

	d2 := models.Device{
		Id:           device2id,
		LocalId:      device2lid,
		Name:         "b d2",
		DeviceTypeId: devicetype1id,
		OwnerId:      userid,
	}

	err = producer.PublishDevice(d2, userid)
	if err != nil {
		t.Error(err)
		return
	}

	d3 := models.Device{
		Id:      device3id,
		LocalId: device3lid,
		Name:    "a d3",
		Attributes: []models.Attribute{
			{Key: "foo", Value: "bar"},
			{Key: "bar", Value: "batz"},
		},
		DeviceTypeId: devicetype1id,
		OwnerId:      userid,
	}

	err = producer.PublishDevice(d3, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	c := client.NewClient("http://localhost:" + conf.ServerPort)

	var dg1 models.DeviceGroup
	t.Run("check dg1", func(t *testing.T) {
		dg1, err, _ = c.ReadDeviceGroup(generatedDeviceGroup1Id, userjwt, false)
		if err != nil {
			t.Error(err)
			return
		}
		if dg1.AutoGeneratedByDevice != device1id {
			t.Error(dg1.AutoGeneratedByDevice)
			return
		}
		if dg1.Name != d1.Name+"_group" {
			t.Error(dg1.Name)
			return
		}
		if !reflect.DeepEqual(dg1.DeviceIds, []string{device1id}) {
			t.Error(dg1.DeviceIds)
			return
		}
		if len(dg1.Criteria) == 0 {
			t.Errorf("%#v", dg1.Criteria)
			return
		}
	})

	var dg2 models.DeviceGroup
	t.Run("check dg2", func(t *testing.T) {
		dg2, err, _ = c.ReadDeviceGroup(generatedDeviceGroup2Id, userjwt, false)
		if err != nil {
			t.Error(err)
			return
		}
		if dg2.AutoGeneratedByDevice != device2id {
			t.Error(dg2.AutoGeneratedByDevice)
			return
		}
		if dg2.Name != d2.Name+"_group" {
			t.Error(dg2.Name)
			return
		}
		if !reflect.DeepEqual(dg2.DeviceIds, []string{device2id}) {
			t.Error(dg2.DeviceIds)
			return
		}
	})

	var dg3 models.DeviceGroup
	t.Run("check dg3", func(t *testing.T) {
		dg3, err, _ = c.ReadDeviceGroup(generatedDeviceGroup3Id, userjwt, false)
		if err != nil {
			t.Error(err)
			return
		}
		if dg3.AutoGeneratedByDevice != device3id {
			t.Error(dg3.AutoGeneratedByDevice)
			return
		}
		if dg3.Name != d3.Name+"_group" {
			t.Error(dg3.Name)
			return
		}
		if !reflect.DeepEqual(dg3.DeviceIds, []string{device3id}) {
			t.Error(dg3.DeviceIds)
			return
		}
	})
	t.Run("validate dg1 delete", func(t *testing.T) {
		err, code := c.ValidateDeviceGroupDelete(userjwt, generatedDeviceGroup1Id)
		if err == nil {
			t.Error("expected error")
			return
		}
		if code != http.StatusBadRequest {
			t.Error(code)
			return
		}
	})
	t.Run("validate removed dg1 AutoGeneratedByDevice", func(t *testing.T) {
		dg := models.DeviceGroup{
			Id:                    dg1.Id,
			Name:                  dg1.Name,
			Image:                 dg1.Image,
			Criteria:              dg1.Criteria,
			DeviceIds:             dg1.DeviceIds,
			CriteriaShort:         dg1.CriteriaShort,
			Attributes:            dg1.Attributes,
			AutoGeneratedByDevice: "",
		}
		err, code := c.ValidateDeviceGroup(userjwt, dg)
		if err == nil {
			t.Error("expected error")
			return
		}
		if code != http.StatusBadRequest {
			t.Error(code)
			return
		}
	})
	t.Run("validate changed dg1 AutoGeneratedByDevice", func(t *testing.T) {
		dg := models.DeviceGroup{
			Id:                    dg1.Id,
			Name:                  dg1.Name,
			Image:                 dg1.Image,
			Criteria:              dg1.Criteria,
			DeviceIds:             dg1.DeviceIds,
			CriteriaShort:         dg1.CriteriaShort,
			Attributes:            dg1.Attributes,
			AutoGeneratedByDevice: device2id,
		}
		err, code := c.ValidateDeviceGroup(userjwt, dg)
		if err == nil {
			t.Error("expected error")
			return
		}
		if code != http.StatusBadRequest {
			t.Error(code)
			return
		}
	})
	t.Run("validate unrelated dg", func(t *testing.T) {
		dg := models.DeviceGroup{
			Id:            "foobar",
			Name:          dg1.Name,
			Image:         dg1.Image,
			Criteria:      dg1.Criteria,
			DeviceIds:     dg1.DeviceIds,
			CriteriaShort: dg1.CriteriaShort,
			Attributes:    dg1.Attributes,
		}
		err, code := c.ValidateDeviceGroup(userjwt, dg)
		if err != nil {
			t.Error(err)
			return
		}
		if code != http.StatusOK {
			t.Error(code)
			return
		}
	})
	t.Run("delete d1", func(t *testing.T) {
		err = producer.PublishDeviceDelete(d1.Id, userid)
		if err != nil {
			t.Error(err)
			return
		}
	})
	t.Run("check dg1", func(t *testing.T) {
		time.Sleep(1 * time.Second)
		_, err, code := c.ReadDeviceGroup(generatedDeviceGroup1Id, userjwt, false)
		if err == nil {
			t.Error("expected error")
			return
		}
		if code != http.StatusNotFound {
			t.Error(code)
			return
		}
	})
	t.Run("validate add d3 to dg2", func(t *testing.T) {
		dg := models.DeviceGroup{
			Id:                    dg2.Id,
			Name:                  dg2.Name,
			Image:                 dg2.Image,
			Criteria:              dg2.Criteria,
			DeviceIds:             []string{device2id, device3id},
			CriteriaShort:         dg2.CriteriaShort,
			Attributes:            dg2.Attributes,
			AutoGeneratedByDevice: dg2.AutoGeneratedByDevice,
		}
		err, code := c.ValidateDeviceGroup(userjwt, dg)
		if err != nil {
			t.Error(err)
			return
		}
		if code != http.StatusOK {
			t.Error(code)
			return
		}
	})
	t.Run("add d3 to dg2", func(t *testing.T) {
		err = producer.PublishDeviceGroup(models.DeviceGroup{
			Id:                    dg2.Id,
			Name:                  dg2.Name,
			Image:                 dg2.Image,
			Criteria:              dg2.Criteria,
			DeviceIds:             []string{device2id, device3id},
			CriteriaShort:         dg2.CriteriaShort,
			Attributes:            dg2.Attributes,
			AutoGeneratedByDevice: dg2.AutoGeneratedByDevice,
		}, userid)
		if err != nil {
			t.Error(err)
			return
		}
	})
	t.Run("validate dg2 delete", func(t *testing.T) {
		err, code := c.ValidateDeviceGroupDelete(userjwt, generatedDeviceGroup2Id)
		if err == nil {
			t.Error("expected error")
			return
		}
		if code != http.StatusBadRequest {
			t.Error(code)
			return
		}
	})
	t.Run("delete d2", func(t *testing.T) {
		err = producer.PublishDeviceDelete(d2.Id, userid)
		if err != nil {
			t.Error(err)
			return
		}
	})
	t.Run("check dg2", func(t *testing.T) {
		time.Sleep(1 * time.Second)
		dg, err, _ := c.ReadDeviceGroup(generatedDeviceGroup2Id, userjwt, false)
		if err != nil {
			t.Error(err)
			return
		}
		if dg.AutoGeneratedByDevice != "" {
			t.Error(dg.AutoGeneratedByDevice)
			return
		}
		if dg.Name != d2.Name+"_group" {
			t.Error(dg.Name)
			return
		}
		if !reflect.DeepEqual(dg.DeviceIds, []string{device3id}) {
			t.Error(dg.DeviceIds)
			return
		}
	})
	t.Run("validate dg2 delete", func(t *testing.T) {
		err, code := c.ValidateDeviceGroupDelete(userjwt, generatedDeviceGroup2Id)
		if err != nil {
			t.Error(err)
			return
		}
		if code != http.StatusOK {
			t.Error(code)
			return
		}
	})
	t.Run("delete dg2", func(t *testing.T) {
		err = producer.PublishDeviceGroupDelete(dg2.Id, userid)
		if err != nil {
			t.Error(err)
			return
		}
	})
	t.Run("check dg2", func(t *testing.T) {
		time.Sleep(1 * time.Second)
		_, err, code := c.ReadDeviceGroup(generatedDeviceGroup2Id, userjwt, false)
		if err == nil {
			t.Error("expected error")
			return
		}
		if code != http.StatusNotFound {
			t.Error(code)
			return
		}
	})
	t.Run("check dg3", func(t *testing.T) {
		dg, err, _ := c.ReadDeviceGroup(generatedDeviceGroup3Id, userjwt, false)
		if err != nil {
			t.Error(err)
			return
		}
		if dg.AutoGeneratedByDevice != device3id {
			t.Error(dg.AutoGeneratedByDevice)
			return
		}
		if dg.Name != d3.Name+"_group" {
			t.Error(dg.Name)
			return
		}
		if !reflect.DeepEqual(dg.DeviceIds, []string{device3id}) {
			t.Error(dg.DeviceIds)
			return
		}
	})
}
