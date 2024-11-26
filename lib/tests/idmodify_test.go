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
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	"github.com/SENERGY-Platform/models/go/models"
	"sync"
	"testing"
	"time"
)

func TestModifiedDevice(t *testing.T) {
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

	sgKey := "a8ee3b1c-4cda-4f0d-9f55-4ef4882ce0af"

	err = producer.PublishDeviceType(models.DeviceType{Id: devicetype1id, Name: devicetype1name, ServiceGroups: []models.ServiceGroup{
		{
			Key:         sgKey,
			Name:        "sg1",
			Description: "",
		},
		{
			Key:         "a8ee3b1c-4cda-4f0d-9f55-4ef4882ce0aa",
			Name:        "sg2",
			Description: "",
		},
	}, Services: []models.Service{
		{
			Id:              "service1",
			LocalId:         "service1",
			Name:            "service1",
			ServiceGroupKey: "",
		},
		{
			Id:              "service2",
			LocalId:         "service2",
			Name:            "service2",
			ServiceGroupKey: sgKey,
		},
		{
			Id:              "service3",
			LocalId:         "service3",
			Name:            "service3",
			ServiceGroupKey: "a8ee3b1c-4cda-4f0d-9f55-4ef4882ce0aa",
		},
	}}, userid)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(10 * time.Second)

	d1 := models.Device{
		Id:           device1id,
		LocalId:      device1lid,
		Name:         device1name,
		DeviceTypeId: devicetype1id,
		OwnerId:      userid,
	}

	err = producer.PublishDevice(d1, userid)
	if err != nil {
		t.Error(err)
		return
	}

	d2 := models.Device{
		Id:      device3id,
		LocalId: device3lid,
		Name:    device3name,
		Attributes: []models.Attribute{
			{
				Key:   controller.DisplayNameAttributeName,
				Value: "foo",
			},
		},
		DeviceTypeId: devicetype1id,
		OwnerId:      userid,
	}

	err = producer.PublishDevice(d2, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	idModifier := idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {sgKey}})
	modifiedNameSuffix := " sg1"
	sgKeyUnknown := sgKey + "unknown"
	modifiedNameSuffixUnknown := " " + sgKeyUnknown
	idModifierUnknown := idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {sgKeyUnknown}})

	d1Modified := models.Device{
		Id:           device1id + idModifier,
		LocalId:      device1lid,
		Name:         device1name + modifiedNameSuffix,
		DeviceTypeId: devicetype1id + idModifier,
		OwnerId:      userid,
	}

	d2Modified := models.Device{
		Id:      device3id + idModifier,
		LocalId: device3lid,
		Name:    device3name + modifiedNameSuffix,
		Attributes: []models.Attribute{
			{
				Key:   controller.DisplayNameAttributeName,
				Value: "foo" + modifiedNameSuffix,
			},
		},
		DeviceTypeId: devicetype1id + idModifier,
		OwnerId:      userid,
	}

	d1ModifiedUnknown := models.Device{
		Id:           device1id + idModifierUnknown,
		LocalId:      device1lid,
		Name:         device1name + modifiedNameSuffixUnknown,
		DeviceTypeId: devicetype1id + idModifierUnknown,
		OwnerId:      userid,
	}
	d2ModifiedUnknown := models.Device{
		Id:      device3id + idModifierUnknown,
		LocalId: device3lid,
		Name:    device3name + modifiedNameSuffixUnknown,
		Attributes: []models.Attribute{
			{
				Key:   controller.DisplayNameAttributeName,
				Value: "foo" + modifiedNameSuffixUnknown,
			},
		},
		DeviceTypeId: devicetype1id + idModifierUnknown,
		OwnerId:      userid,
	}

	dtModified := models.DeviceType{
		Id:   devicetype1id + idModifier,
		Name: devicetype1name + modifiedNameSuffix,
		ServiceGroups: []models.ServiceGroup{
			{
				Key:         sgKey,
				Name:        "sg1",
				Description: "",
			},
			{
				Key:         "a8ee3b1c-4cda-4f0d-9f55-4ef4882ce0aa",
				Name:        "sg2",
				Description: "",
			},
		},
		Services: []models.Service{
			{
				Id:              "service1",
				LocalId:         "service1",
				Name:            "service1",
				ServiceGroupKey: "",
			},
			{
				Id:              "service2",
				LocalId:         "service2",
				Name:            "service2",
				ServiceGroupKey: sgKey,
			},
		},
	}

	dtModifiedUnknown := models.DeviceType{
		Id:   devicetype1id + idModifierUnknown,
		Name: devicetype1name,
		ServiceGroups: []models.ServiceGroup{
			{
				Key:         sgKey,
				Name:        "sg1",
				Description: "",
			},
			{
				Key:         "a8ee3b1c-4cda-4f0d-9f55-4ef4882ce0aa",
				Name:        "sg2",
				Description: "",
			},
		},
		Services: []models.Service{
			{
				Id:              "service1",
				LocalId:         "service1",
				Name:            "service1",
				ServiceGroupKey: "",
			},
		},
	}

	t.Run("testDeviceRead", func(t *testing.T) {
		testDeviceRead(t, conf, false, d1Modified, d2Modified, d1ModifiedUnknown, d2ModifiedUnknown)
	})

	t.Run("testDeviceTypeRead", func(t *testing.T) {
		testDeviceTypeRead(t, conf, dtModified, dtModifiedUnknown)
	})

	t.Run("extended modified device 1", func(t *testing.T) {
		actual, err, _ := client.NewClient("http://localhost:"+conf.ServerPort).ReadExtendedDevice(d1Modified.Id, userjwt, client.READ, true)
		if err != nil {
			t.Error(err)
			return
		}
		if actual.Name != d1Modified.Name {
			t.Errorf("got %s, want %s", actual.Name, d1Modified.Name)
			return
		}
		if actual.DeviceTypeId != d1Modified.DeviceTypeId {
			t.Errorf("got %s, want %s", actual.DeviceTypeId, d1Modified.DeviceTypeId)
			return
		}
		if actual.DisplayName != d1Modified.Name {
			t.Errorf("got %s, want %s", actual.DisplayName, d1Modified.Name)
			return
		}
		if actual.DeviceTypeName != dtModified.Name {
			t.Errorf("got %s, want %s", actual.DeviceTypeName, dtModified.Name)
			return
		}
		if actual.DeviceTypeId != d1Modified.DeviceTypeId {
			t.Errorf("\na=%s\ne=%s", actual.DeviceTypeId, d1Modified.DeviceTypeId)
			return
		}
		if actual.DeviceType.Id != d1Modified.DeviceTypeId {
			t.Errorf("\na=%s\ne=%s", actual.DeviceType.Id, d1Modified.DeviceTypeId)
			return
		}
		if actual.DeviceType.Id != dtModified.Id {
			t.Errorf("\na=%s\ne=%s", actual.DeviceType.Id, dtModified.Id)
			return
		}
		if actual.DeviceType.Name != dtModified.Name {
			t.Errorf("got %s, want %s", actual.DeviceType.Name, dtModified.Name)
			return
		}
	})

	t.Run("extended modified device 2", func(t *testing.T) {
		actual, err, _ := client.NewClient("http://localhost:"+conf.ServerPort).ReadExtendedDevice(d2Modified.Id, userjwt, client.READ, true)
		if err != nil {
			t.Error(err)
			return
		}
		if actual.Name != d2Modified.Name {
			t.Errorf("got %s, want %s", actual.Name, d2Modified.Name)
			return
		}
		if actual.DeviceTypeId != d2Modified.DeviceTypeId {
			t.Errorf("got %s, want %s", actual.DeviceTypeId, d2Modified.DeviceTypeId)
			return
		}
		if actual.DisplayName != d2Modified.Attributes[0].Value {
			t.Errorf("got %s, want %s", actual.DisplayName, d2Modified.Attributes[0].Value)
			return
		}
		if actual.DeviceTypeName != dtModified.Name {
			t.Errorf("got %s, want %s", actual.DeviceTypeName, dtModified.Name)
			return
		}
		if actual.DeviceType.Id != d2Modified.DeviceTypeId {
			t.Errorf("got %s, want %s", actual.DeviceTypeId, d2Modified.DeviceTypeId)
			return
		}
		if actual.DeviceType.Name != dtModified.Name {
			t.Errorf("got %s, want %s", actual.DeviceType.Name, dtModified.Name)
			return
		}
	})
}
