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

package tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/api"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/tests/manager_legacy/helper"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"testing"
	"time"
)

func testDeviceOwner(t *testing.T, conf configuration.Config) {
	//replace sleep with wait=true query parameter
	tempSleepAfterEdit := helper.SleepAfterEdit
	helper.SleepAfterEdit = 0
	defer func() {
		helper.SleepAfterEdit = tempSleepAfterEdit
	}()
	protocol := models.Protocol{}
	t.Run("create protocol", func(t *testing.T) {
		resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+conf.ServerPort+"/protocols?wait=true", models.Protocol{
			Name:             "p2",
			Handler:          "ph1",
			ProtocolSegments: []models.ProtocolSegment{{Name: "ps2"}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Error(resp.Status, resp.StatusCode, string(b))
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&protocol)
		if err != nil {
			t.Error(err)
			return
		}
	})

	dt := models.DeviceType{}
	t.Run("create device-type", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+conf.ServerPort+"/device-types?wait=true", models.DeviceType{
			Name:          "foo",
			DeviceClassId: "dc1",
			Services: []models.Service{
				{
					Name:    "s1name",
					LocalId: "lid1",
					Inputs: []models.Content{
						{
							ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
							Serialization:     "json",
							ContentVariable: models.ContentVariable{
								Name:       "v1name",
								Type:       models.String,
								FunctionId: f1Id,
								AspectId:   a1Id,
							},
						},
					},
					ProtocolId: protocol.Id,
				},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}
		err = json.NewDecoder(resp.Body).Decode(&dt)
		if err != nil {
			t.Error(err)
			return
		}
	})

	device1 := models.Device{}
	t.Run("create device with implicit owner", func(t *testing.T) {
		device := models.Device{
			Name:         "device1",
			LocalId:      uuid.New().String(),
			DeviceTypeId: dt.Id,
		}
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+conf.ServerPort+"/devices?wait=true", device)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&device1)
		if err != nil {
			t.Error(err)
			return
		}
		device.Id = device1.Id
		device.OwnerId = userjwtUser
		if !reflect.DeepEqual(device1, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device1, device)
			return
		}
	})

	t.Run("check device1 after create", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/devices/"+url.PathEscape(device1.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		device := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&device)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(device1, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device1, device)
			return
		}
	})

	device2 := models.Device{}
	t.Run("create device with explicit owner", func(t *testing.T) {
		device := models.Device{
			Name:         "device1",
			LocalId:      uuid.New().String(),
			DeviceTypeId: dt.Id,
			OwnerId:      userjwtUser,
		}
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+conf.ServerPort+"/devices?wait=true", device)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&device2)
		if err != nil {
			t.Error(err)
			return
		}
		device.Id = device2.Id
		device.OwnerId = userjwtUser
		if !reflect.DeepEqual(device2, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device2, device)
			return
		}
	})

	t.Run("check device2 after create", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/devices/"+url.PathEscape(device2.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		device := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&device)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(device2, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device2, device)
			return
		}
	})

	t.Run("update device with implicit owner", func(t *testing.T) {
		device := device1
		device.Name = "device1 update1"
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+conf.ServerPort+"/devices/"+url.PathEscape(device.Id)+"?wait=true", device)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&device1)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(device1, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device1, device)
			return
		}
	})

	t.Run("check device after update", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/devices/"+url.PathEscape(device1.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		device := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&device)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(device1, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device1, device)
			return
		}
	})

	t.Run("update device with explicit owner", func(t *testing.T) {
		device := device1
		device.Name = "device1 update2"
		device.OwnerId = userjwtUser
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+conf.ServerPort+"/devices/"+url.PathEscape(device.Id)+"?wait=true", device)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&device1)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(device1, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device1, device)
			return
		}
	})

	t.Run("check device after update", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/devices/"+url.PathEscape(device1.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		device := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&device)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(device1, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device1, device)
			return
		}
	})

	t.Run("try create device with foreign owner", func(t *testing.T) {
		resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+conf.ServerPort+"/devices?wait=true", models.Device{
			Name:         "device1",
			LocalId:      uuid.New().String(),
			DeviceTypeId: dt.Id,
			OwnerId:      userjwtUser,
		})
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode == http.StatusOK {
			t.Error("expect error")
			return
		}
	})

	t.Run("try change owner to none admin user", func(t *testing.T) {
		device := device1
		device.OwnerId = userid
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+conf.ServerPort+"/devices/"+url.PathEscape(device.Id)+"?wait=true", device)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode == http.StatusOK {
			t.Error("expect error")
			return
		}
	})

	t.Run("give user admin rights", func(t *testing.T) {
		_, err, _ := client.New(conf.PermissionsV2Url).SetPermission(client.InternalAdminToken, conf.DeviceTopic, device1.Id, client.ResourcePermissions{
			UserPermissions: map[string]client.PermissionsMap{
				userjwtUser: {Read: true, Write: true, Execute: true, Administrate: true},
				userid:      {Read: true, Write: true, Execute: true, Administrate: true},
			},
			RolePermissions: map[string]client.PermissionsMap{
				"admin": {Read: true, Write: true, Execute: true, Administrate: true},
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(5 * time.Second)
	})

	t.Run("change owner to other admin user", func(t *testing.T) {
		device := device1
		device.OwnerId = userid
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+conf.ServerPort+"/devices/"+url.PathEscape(device.Id)+"?wait=true", device)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		err = json.NewDecoder(resp.Body).Decode(&device1)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(device1, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device1, device)
			return
		}
	})

	t.Run("check device after update", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/devices/"+url.PathEscape(device1.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		device := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&device)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(device1, device) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", device1, device)
			return
		}
	})
}

func testDeviceAttributes(t *testing.T, port string) {
	resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+port+"/protocols?wait=true", models.Protocol{
		Name:             "p2",
		Handler:          "ph1",
		ProtocolSegments: []models.ProtocolSegment{{Name: "ps2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	protocol := models.Protocol{}
	err = json.NewDecoder(resp.Body).Decode(&protocol)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true", models.DeviceType{
		Name:          "foo",
		DeviceClassId: "dc1",
		Services: []models.Service{
			{
				Name:    "s1name",
				LocalId: "lid1",
				Inputs: []models.Content{
					{
						ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
						Serialization:     "json",
						ContentVariable: models.ContentVariable{
							Name:       "v1name",
							Type:       models.String,
							FunctionId: f1Id,
							AspectId:   a1Id,
						},
					},
				},
				ProtocolId: protocol.Id,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	dt := models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&dt)
	if err != nil {
		t.Fatal(err)
	}

	if dt.Id == "" {
		t.Fatal(dt)
	}

	device, err := initDevice(port, dt)
	if err != nil {
		t.Fatal(err)
	}
	deviceId := device.Id

	t.Run("normal attr init", tryDeviceAttributeUpdate(port, dt.Id, deviceId, device.LocalId, "", []models.Attribute{
		{
			Key:    "a1",
			Value:  "va1",
			Origin: "",
		},
		{
			Key:    "a2",
			Value:  "va2",
			Origin: "test1",
		},
		{
			Key:    "a3",
			Value:  "va3",
			Origin: "test1",
		},
		{
			Key:    "a4",
			Value:  "va4",
			Origin: "test2",
		},
		{
			Key:    "a5",
			Value:  "va5",
			Origin: "test2",
		},
	}, []models.Attribute{
		{
			Key:    "a1",
			Value:  "va1",
			Origin: "",
		},
		{
			Key:    "a2",
			Value:  "va2",
			Origin: "test1",
		},
		{
			Key:    "a3",
			Value:  "va3",
			Origin: "test1",
		},
		{
			Key:    "a4",
			Value:  "va4",
			Origin: "test2",
		},
		{
			Key:    "a5",
			Value:  "va5",
			Origin: "test2",
		},
	}))

	t.Run("normal attr update", tryDeviceAttributeUpdate(port, dt.Id, deviceId, device.LocalId, "", []models.Attribute{
		{
			Key:    "a12",
			Value:  "va12",
			Origin: "",
		},
		{
			Key:    "a22",
			Value:  "va22",
			Origin: "test1",
		},
		{
			Key:    "a32",
			Value:  "va32",
			Origin: "test1",
		},
		{
			Key:    "a42",
			Value:  "va42",
			Origin: "test2",
		},
		{
			Key:    "a52",
			Value:  "va52",
			Origin: "test2",
		},
	}, []models.Attribute{
		{
			Key:    "a12",
			Value:  "va12",
			Origin: "",
		},
		{
			Key:    "a22",
			Value:  "va22",
			Origin: "test1",
		},
		{
			Key:    "a32",
			Value:  "va32",
			Origin: "test1",
		},
		{
			Key:    "a42",
			Value:  "va42",
			Origin: "test2",
		},
		{
			Key:    "a52",
			Value:  "va52",
			Origin: "test2",
		},
	}))

	t.Run("origin attr update", tryDeviceAttributeUpdate(port, dt.Id, deviceId, device.LocalId, "test1", []models.Attribute{
		{
			Key:    "a13",
			Value:  "va13",
			Origin: "",
		},
		{
			Key:    "a23",
			Value:  "va23",
			Origin: "test1",
		},
		{
			Key:    "a33",
			Value:  "va33",
			Origin: "test1",
		},
		{
			Key:    "a43",
			Value:  "va43",
			Origin: "test2",
		},
		{
			Key:    "a53",
			Value:  "va53",
			Origin: "test2",
		},
	}, []models.Attribute{
		{
			Key:    "a12",
			Value:  "va12",
			Origin: "",
		},
		{
			Key:    "a23",
			Value:  "va23",
			Origin: "test1",
		},
		{
			Key:    "a33",
			Value:  "va33",
			Origin: "test1",
		},
		{
			Key:    "a42",
			Value:  "va42",
			Origin: "test2",
		},
		{
			Key:    "a52",
			Value:  "va52",
			Origin: "test2",
		},
	}))

	t.Run("origin list create", tryDeviceAttributeUpdate(port, dt.Id, deviceId, device.LocalId, "shared,test3", []models.Attribute{
		{
			Key:    "a13",
			Value:  "foo",
			Origin: "",
		},
		{
			Key:    "a23",
			Value:  "bar",
			Origin: "test1",
		},
		{
			Key:    "a43",
			Value:  "42",
			Origin: "test2",
		},
		{
			Key:    "shared/val1",
			Value:  "s42",
			Origin: "shared",
		},
		{
			Key:    "shared/val2",
			Value:  "s42",
			Origin: "shared",
		},
		{
			Key:    "test3/val1",
			Value:  "t42",
			Origin: "test3",
		},
		{
			Key:    "test3/val2",
			Value:  "t42",
			Origin: "test3",
		},
	}, []models.Attribute{
		{
			Key:    "a12",
			Value:  "va12",
			Origin: "",
		},
		{
			Key:    "a23",
			Value:  "va23",
			Origin: "test1",
		},
		{
			Key:    "a33",
			Value:  "va33",
			Origin: "test1",
		},
		{
			Key:    "a42",
			Value:  "va42",
			Origin: "test2",
		},
		{
			Key:    "a52",
			Value:  "va52",
			Origin: "test2",
		},
		{
			Key:    "shared/val1",
			Value:  "s42",
			Origin: "shared",
		},
		{
			Key:    "shared/val2",
			Value:  "s42",
			Origin: "shared",
		},
		{
			Key:    "test3/val1",
			Value:  "t42",
			Origin: "test3",
		},
		{
			Key:    "test3/val2",
			Value:  "t42",
			Origin: "test3",
		},
	}))

	t.Run("origin list update", tryDeviceAttributeUpdate(port, dt.Id, deviceId, device.LocalId, "shared,test3", []models.Attribute{
		{
			Key:    "a13",
			Value:  "foo",
			Origin: "",
		},
		{
			Key:    "a23",
			Value:  "bar",
			Origin: "test1",
		},
		{
			Key:    "a43",
			Value:  "42",
			Origin: "test2",
		},
		{
			Key:    "shared/val1",
			Value:  "s42u",
			Origin: "shared",
		},
		{
			Key:    "test3/val3",
			Value:  "t42u",
			Origin: "test3",
		},
	}, []models.Attribute{
		{
			Key:    "a12",
			Value:  "va12",
			Origin: "",
		},
		{
			Key:    "a23",
			Value:  "va23",
			Origin: "test1",
		},
		{
			Key:    "a33",
			Value:  "va33",
			Origin: "test1",
		},
		{
			Key:    "a42",
			Value:  "va42",
			Origin: "test2",
		},
		{
			Key:    "a52",
			Value:  "va52",
			Origin: "test2",
		},
		{
			Key:    "shared/val1",
			Value:  "s42u",
			Origin: "shared",
		},
		{
			Key:    "test3/val3",
			Value:  "t42u",
			Origin: "test3",
		},
	}))

	t.Run("reset attributes", tryDeviceAttributeUpdate(port, dt.Id, deviceId, device.LocalId, "", []models.Attribute{}, []models.Attribute{}))

	t.Run("initial display-name", tryDeviceDisplayNameUpdate(port, deviceId, "display-name-1", models.Device{
		Id:           deviceId,
		LocalId:      device.LocalId,
		Name:         device.Name,
		DeviceTypeId: device.DeviceTypeId,
		Attributes:   []models.Attribute{{Key: api.DisplayNameAttributeKey, Value: "display-name-1", Origin: api.DisplayNameAttributeOrigin}},
	}))

	t.Run("update display-name", tryDeviceDisplayNameUpdate(port, deviceId, "display-name-2", models.Device{
		Id:           deviceId,
		LocalId:      device.LocalId,
		Name:         device.Name,
		DeviceTypeId: device.DeviceTypeId,
		Attributes:   []models.Attribute{{Key: api.DisplayNameAttributeKey, Value: "display-name-2", Origin: api.DisplayNameAttributeOrigin}},
	}))

	t.Run("reset attributes with some existing", tryDeviceAttributeUpdate(port, dt.Id, deviceId, device.LocalId, "", []models.Attribute{
		{
			Key:    "shared/val1",
			Value:  "s42u",
			Origin: "shared",
		},
		{
			Key:    "test3/val3",
			Value:  "t42u",
			Origin: "test3",
		},
	}, []models.Attribute{
		{
			Key:    "shared/val1",
			Value:  "s42u",
			Origin: "shared",
		},
		{
			Key:    "test3/val3",
			Value:  "t42u",
			Origin: "test3",
		},
	}))

	t.Run("initial display-name with existing attributes", tryDeviceDisplayNameUpdate(port, deviceId, "display-name-3", models.Device{
		Id:           deviceId,
		LocalId:      device.LocalId,
		Name:         device.Name,
		DeviceTypeId: device.DeviceTypeId,
		Attributes: []models.Attribute{
			{
				Key:    "shared/val1",
				Value:  "s42u",
				Origin: "shared",
			},
			{
				Key:    "test3/val3",
				Value:  "t42u",
				Origin: "test3",
			},
			{
				Key:    api.DisplayNameAttributeKey,
				Value:  "display-name-3",
				Origin: api.DisplayNameAttributeOrigin,
			},
		},
	}))

	t.Run("update display-name with existing attributes", tryDeviceDisplayNameUpdate(port, deviceId, "display-name-4", models.Device{
		Id:           deviceId,
		LocalId:      device.LocalId,
		Name:         device.Name,
		DeviceTypeId: device.DeviceTypeId,
		Attributes: []models.Attribute{
			{
				Key:    "shared/val1",
				Value:  "s42u",
				Origin: "shared",
			},
			{
				Key:    "test3/val3",
				Value:  "t42u",
				Origin: "test3",
			},
			{
				Key:    api.DisplayNameAttributeKey,
				Value:  "display-name-4",
				Origin: api.DisplayNameAttributeOrigin,
			},
		},
	}))
}

func testDevice(t *testing.T, port string) {
	resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+port+"/protocols?wait=true", models.Protocol{
		Name:             "p2",
		Handler:          "ph1",
		ProtocolSegments: []models.ProtocolSegment{{Name: "ps2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	protocol := models.Protocol{}
	err = json.NewDecoder(resp.Body).Decode(&protocol)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true", models.DeviceType{
		Name:          "foo",
		DeviceClassId: "dc1",
		Services: []models.Service{
			{
				Name:    "s1name",
				LocalId: "lid1",
				Inputs: []models.Content{
					{
						ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
						Serialization:     "json",
						ContentVariable: models.ContentVariable{
							Name:       "v1name",
							Type:       models.String,
							FunctionId: f1Id,
							AspectId:   a1Id,
						},
					},
				},

				ProtocolId: protocol.Id,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	dt := models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&dt)
	if err != nil {
		t.Fatal(err)
	}

	if dt.Id == "" {
		t.Fatal(dt)
	}

	t.Run("missing dt id", func(t *testing.T) {
		resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
			Name:    "d1",
			LocalId: "lid1",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		//expect validation error
		if resp.StatusCode == http.StatusOK {
			t.Fatal(resp.Status, resp.StatusCode)
		}
	})

	t.Run("missing local id", func(t *testing.T) {
		resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
			Name:         "d1",
			DeviceTypeId: dt.Id,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		//expect validation error
		if resp.StatusCode == http.StatusOK {
			t.Fatal(resp.Status, resp.StatusCode)
		}
	})

	device := models.Device{}
	t.Run("create", func(t *testing.T) {
		resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
			Name:         "d1",
			DeviceTypeId: dt.Id,
			LocalId:      "lid1",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}
		err = json.NewDecoder(resp.Body).Decode(&device)
		if err != nil {
			t.Fatal(err)
		}

		if device.Id == "" {
			t.Fatal(device)
		}
		time.Sleep(time.Second)
	})

	t.Run("get", func(t *testing.T) {
		resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(device.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		result := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		if result.Name != "d1" || result.LocalId != "lid1" || result.DeviceTypeId != dt.Id {
			t.Fatal(result)
		}
	})

	t.Run("local id duplicate", func(t *testing.T) {
		resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
			Name:         "reused_local_id",
			DeviceTypeId: dt.Id,
			LocalId:      "lid1",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		//expect validation error
		if resp.StatusCode == http.StatusOK {
			t.Fatal("device.local_id should be validated for global uniqueness: ", resp.Status, resp.StatusCode)
		}
	})

	t.Run("delete", func(t *testing.T) {
		resp, err = helper.Jwtdelete(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(device.Id)+"?wait=true")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}
	})

	t.Run("read after update", func(t *testing.T) {
		time.Sleep(time.Second)
		resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(device.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		//expect 404 error
		if resp.StatusCode != http.StatusNotFound {
			t.Fatal(resp.Status, resp.StatusCode)
		}
	})

}

func tryDeviceDisplayNameUpdate(port string, deviceId string, displayName string, expectedDevice models.Device) func(t *testing.T) {
	return func(t *testing.T) {
		expectedDevice.OwnerId = userjwtUser
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(deviceId)+"/display_name?wait=true", displayName)
		if err != nil {
			t.Fatal(err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
			return
		}

		device := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&device)
		if err != nil {
			t.Fatal(err)
		}

		if device.Id == "" {
			t.Fatal(device)
		}
		if !reflect.DeepEqual(device, expectedDevice) {
			t.Errorf("\n%#v\n%#v\n", device, expectedDevice)
			return
		}

		//time.Sleep(5 * time.Second)

		resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(device.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		result := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(device, expectedDevice) {
			t.Errorf("\n%#v\n%#v\n", device, expectedDevice)
			return
		}

	}
}

func tryDeviceAttributeUpdate(port string, dtId string, deviceId string, localDeviceId string, origin string, attributes []models.Attribute, expected []models.Attribute) func(t *testing.T) {
	return func(t *testing.T) {
		sort.Slice(attributes, func(i, j int) bool {
			return attributes[i].Key < attributes[j].Key
		})

		endpoint := "http://localhost:" + port + "/devices/" + url.PathEscape(deviceId)
		if origin != "" {
			endpoint = endpoint + "?" + url.Values{api.UpdateOnlySameOriginAttributesKey: {origin}, "wait": {"true"}}.Encode()
		} else {
			endpoint = endpoint + "?" + url.Values{"wait": {"true"}}.Encode()
		}
		resp, err := helper.Jwtput(userjwt, endpoint, models.Device{
			Id:           deviceId,
			Name:         "d1",
			LocalId:      localDeviceId,
			DeviceTypeId: dtId,
			Attributes:   attributes,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(endpoint, resp.Status, resp.StatusCode, string(b))
			return
		}

		device := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&device)
		if err != nil {
			t.Fatal(err)
		}

		if device.Id == "" {
			t.Fatal(device)
		}
		if !reflect.DeepEqual(device.Attributes, expected) {
			a, _ := json.Marshal(device.Attributes)
			e, _ := json.Marshal(expected)
			t.Error("\n", string(a), "\n", string(e))
			return
		}

		//time.Sleep(5 * time.Second)

		resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(device.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		result := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(result.Attributes, expected) {
			t.Error(device, expected)
			return
		}
	}
}

func initDevice(port string, dt models.DeviceType) (models.Device, error) {
	device := models.Device{
		Name:         "d1",
		LocalId:      uuid.New().String(),
		DeviceTypeId: dt.Id,
	}
	resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", device)
	if err != nil {
		return models.Device{}, err
	}
	if resp.StatusCode != http.StatusOK {
		temp, _ := io.ReadAll(resp.Body)
		return models.Device{}, fmt.Errorf("%v %v", resp.Status, string(temp))
	}
	result := models.Device{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return models.Device{}, err
	}
	device.Id = result.Id
	device.OwnerId = userjwtUser
	if !reflect.DeepEqual(result, device) {
		log.Printf("ERROR: \n%#v\n!=\n%#v\n", result, device)
		return models.Device{}, errors.New("returned device != expected device")
	} else {
		log.Printf("DEBUG:created device: \n%#v\n", result)
	}
	return result, err
}
