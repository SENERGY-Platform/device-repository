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
	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/source/util"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testenv"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/donewait"
	"github.com/SENERGY-Platform/service-commons/pkg/kafka"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

//TODO: test permissions-v2 integration

var device1id = "urn:infai:ses:device:1"
var generatedDeviceGroup1Id = "urn:infai:ses:device-group:1"
var device1lid = "lid1"
var device1name = uuid.NewString()
var device2id = "urn:infai:ses:device:2"
var generatedDeviceGroup2Id = "urn:infai:ses:device-group:2"
var device2lid = "lid2"
var device2name = uuid.NewString()
var device3id = "urn:infai:ses:device:3"
var generatedDeviceGroup3Id = "urn:infai:ses:device-group:3"
var device3lid = "lid3"
var device3name = uuid.NewString()

func TestDeviceNameValidation(t *testing.T) {
	err := controller.ValidateDeviceName(models.Device{Name: "foo"})
	if err != nil {
		t.Error(err)
		return
	}
	err = controller.ValidateDeviceName(models.Device{Name: "", Attributes: []models.Attribute{{Key: "shared/nickname", Origin: "shared", Value: "bar"}}})
	if err != nil {
		t.Error(err)
		return
	}

	err = controller.ValidateDeviceName(models.Device{})
	if err == nil {
		t.Error("missing error")
		return
	}
}

func TestDeviceDeviceTypeFilter(t *testing.T) {
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

	err = producer.PublishDeviceType(models.DeviceType{Id: devicetype1id, Name: devicetype1name}, userid)
	if err != nil {
		t.Error(err)
		return
	}

	d1 := models.Device{
		Id:           device1id,
		LocalId:      device1lid,
		Name:         "a d1",
		DeviceTypeId: devicetype1id,
		OwnerId:      userid,
	}
	dx1 := models.ExtendedDevice{
		Device:          d1,
		ConnectionState: models.ConnectionStateUnknown,
		DisplayName:     d1.Name,
		DeviceTypeName:  devicetype1name,
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
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

	dx2 := models.ExtendedDevice{
		Device:          d2,
		ConnectionState: models.ConnectionStateUnknown,
		DisplayName:     d2.Name,
		DeviceTypeName:  devicetype1name,
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
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
		DeviceTypeId: devicetype2id,
		OwnerId:      userid,
	}

	dx3 := models.ExtendedDevice{
		Device:          d3,
		ConnectionState: models.ConnectionStateUnknown,
		DisplayName:     d3.Name,
		DeviceTypeName:  "",
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
	}

	err = producer.PublishDevice(d3, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	c := client.NewClient("http://localhost:" + conf.ServerPort)
	t.Run("list none", func(t *testing.T) {
		result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{DeviceTypeIds: []string{}})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.Device{}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
	t.Run("list none extended", func(t *testing.T) {
		result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{DeviceTypeIds: []string{}})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.ExtendedDevice{}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
	t.Run("list all", func(t *testing.T) {
		result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.Device{d1, d3, d2}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
	t.Run("list all extended", func(t *testing.T) {
		result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.ExtendedDevice{dx1, dx3, dx2}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
	t.Run("list dt1", func(t *testing.T) {
		result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{DeviceTypeIds: []string{devicetype1id}})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.Device{d1, d2}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
	t.Run("list dt1 extended", func(t *testing.T) {
		result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{DeviceTypeIds: []string{devicetype1id}})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.ExtendedDevice{dx1, dx2}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
	t.Run("list dt2", func(t *testing.T) {
		result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{DeviceTypeIds: []string{devicetype2id}})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.Device{d3}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
	t.Run("list dt2 extended", func(t *testing.T) {
		result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{DeviceTypeIds: []string{devicetype2id}})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.ExtendedDevice{dx3}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
	t.Run("list dt1+dt2", func(t *testing.T) {
		result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{DeviceTypeIds: []string{devicetype1id, devicetype2id}})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.Device{d1, d3, d2}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
	t.Run("list dt1+dt2 extended", func(t *testing.T) {
		result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{DeviceTypeIds: []string{devicetype1id, devicetype2id}})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.ExtendedDevice{dx1, dx3, dx2}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("%#v\n", result)
		}
	})
}

func TestDeviceQuery(t *testing.T) {
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

	err = producer.PublishDeviceType(models.DeviceType{Id: devicetype1id, Name: devicetype1name}, userid)
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
	dx1 := models.ExtendedDevice{
		Device:          d1,
		ConnectionState: models.ConnectionStateUnknown,
		DisplayName:     d1.Name,
		DeviceTypeName:  devicetype1name,
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
	}

	dx1wdt := models.ExtendedDevice{
		Device:          d1,
		ConnectionState: models.ConnectionStateUnknown,
		DisplayName:     d1.Name,
		DeviceTypeName:  devicetype1name,
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
		DeviceType: &models.DeviceType{Id: devicetype1id, Name: devicetype1name},
	}

	err = producer.PublishDevice(d1, userid)
	if err != nil {
		t.Error(err)
		return
	}

	d2 := models.Device{
		Id:           device2id,
		LocalId:      device2lid,
		Name:         device2name,
		DeviceTypeId: devicetype1id,
		OwnerId:      userid,
	}

	dx2 := models.ExtendedDevice{
		Device:          d2,
		ConnectionState: models.ConnectionStateUnknown,
		DisplayName:     d2.Name,
		DeviceTypeName:  devicetype1name,
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
	}

	err = producer.PublishDevice(d2, userid)
	if err != nil {
		t.Error(err)
		return
	}

	d3 := models.Device{
		Id:      device3id,
		LocalId: device3lid,
		Name:    device3name,
		Attributes: []models.Attribute{
			{Key: "foo", Value: "bar"},
			{Key: "bar", Value: "batz"},
		},
		DeviceTypeId: devicetype1id,
		OwnerId:      userid,
	}

	dx3 := models.ExtendedDevice{
		Device:          d3,
		ConnectionState: models.ConnectionStateUnknown,
		DisplayName:     d3.Name,
		DeviceTypeName:  devicetype1name,
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
	}

	err = producer.PublishDevice(d3, userid)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Second)

	t.Run("check fulldt query param", func(t *testing.T) {
		c := client.NewClient("http://localhost:" + conf.ServerPort)
		t.Run("ReadExtendedDevice", func(t *testing.T) {
			result, err, _ := c.ReadExtendedDevice(d1.Id, userjwt, models.Read, true)
			if err != nil {
				t.Error(err)
				return
			}
			expected := dx1wdt
			if !reflect.DeepEqual(result, expected) {
				if result.DeviceType != nil {
					t.Errorf("%#v\n", *result.DeviceType)
				}
				t.Errorf("%#v\n", result)
			}
		})
		t.Run("ReadExtendedDeviceByLocalId", func(t *testing.T) {
			result, err, _ := c.ReadExtendedDeviceByLocalId(userid, d1.LocalId, userjwt, models.Read, true)
			if err != nil {
				t.Error(err)
				return
			}
			expected := dx1wdt
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})

		t.Run("ListExtendedDevices", func(t *testing.T) {
			result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{
				Ids:    []string{d1.Id},
				FullDt: true,
			})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.ExtendedDevice{dx1wdt}
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
	})

	t.Run("not existing", func(t *testing.T) {
		testDeviceReadNotFound(t, conf, false, "foobar")
	})
	t.Run("not existing localId", func(t *testing.T) {
		testDeviceReadNotFound(t, conf, true, "foobar")
	})
	t.Run("testDeviceRead", func(t *testing.T) {
		testDeviceRead(t, conf, false, d1, d2, d3)
	})
	t.Run("testDeviceRead localid", func(t *testing.T) {
		testDeviceRead(t, conf, true, d1, d2, d3)
	})

	t.Run("test list devices", func(t *testing.T) {
		c := client.NewClient("http://localhost:" + conf.ServerPort)
		t.Run("list none", func(t *testing.T) {
			result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{Ids: []string{}})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.Device{}
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
		t.Run("list all", func(t *testing.T) {
			result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{SortBy: "localid"})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.Device{d1, d2, d3}
			slices.SortFunc(expected, func(a, b models.Device) int {
				return strings.Compare(a.Id, b.Id)
			})
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\n%#v\n%#v\n", result, expected)
			}
		})
		t.Run("list limit/offset", func(t *testing.T) {
			result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{Limit: 1, Offset: 1, SortBy: "localid"})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.Device{d1, d2, d3}
			slices.SortFunc(expected, func(a, b models.Device) int {
				return strings.Compare(a.Id, b.Id)
			})
			expected = expected[1:2]
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\n%#v\n%#v\n", result, expected)
			}
		})
		t.Run("list ids all", func(t *testing.T) {
			result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{SortBy: "localid", Ids: []string{d1.Id, d2.Id, d3.Id}})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.Device{d1, d2, d3}
			slices.SortFunc(expected, func(a, b models.Device) int {
				return strings.Compare(a.Id, b.Id)
			})
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\n%#v\n%#v\n", result, expected)
			}
		})

		t.Run("list ids d1, d3", func(t *testing.T) {
			result, err, _ := c.ListDevices(userjwt, client.DeviceListOptions{SortBy: "localid", Ids: []string{d1.Id, d3.Id}})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.Device{d1, d3}
			slices.SortFunc(expected, func(a, b models.Device) int {
				return strings.Compare(a.Id, b.Id)
			})
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\n%#v\n%#v\n", result, expected)
			}
		})
	})

	t.Run("test list extended-devices", func(t *testing.T) {
		c := client.NewClient("http://localhost:" + conf.ServerPort)
		t.Run("list none", func(t *testing.T) {
			result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Ids: []string{}})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.ExtendedDevice{}
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
		t.Run("list all", func(t *testing.T) {
			result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{SortBy: "localid"})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.ExtendedDevice{dx1, dx2, dx3}
			slices.SortFunc(expected, func(a, b models.ExtendedDevice) int {
				return strings.Compare(a.LocalId, b.LocalId)
			})
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\n%#v\n%#v\n", result, expected)
			}
		})
		t.Run("list limit/offset", func(t *testing.T) {
			result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Limit: 1, Offset: 1, SortBy: "localid"})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.ExtendedDevice{dx1, dx2, dx3}
			slices.SortFunc(expected, func(a, b models.ExtendedDevice) int {
				return strings.Compare(a.LocalId, b.LocalId)
			})
			expected = expected[1:2]
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\n%#v\n%#v\n", result, expected)
			}
		})
		t.Run("list ids all", func(t *testing.T) {
			result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{SortBy: "localid", Ids: []string{d1.Id, d2.Id, d3.Id}})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.ExtendedDevice{dx1, dx2, dx3}
			slices.SortFunc(expected, func(a, b models.ExtendedDevice) int {
				return strings.Compare(a.LocalId, b.LocalId)
			})
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\n%#v\n%#v\n", result, expected)
			}
		})

		t.Run("list ids d1, d3", func(t *testing.T) {
			result, _, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{SortBy: "localid", Ids: []string{d1.Id, d3.Id}})
			if err != nil {
				t.Error(err)
				return
			}
			expected := []models.ExtendedDevice{dx1, dx3}
			slices.SortFunc(expected, func(a, b models.ExtendedDevice) int {
				return strings.Compare(a.LocalId, b.LocalId)
			})
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("\n%#v\n%#v\n", result, expected)
			}
		})
	})
}

func testDeviceRead(t *testing.T, conf config.Config, asLocalId bool, expectedDevices ...models.Device) {
	for _, expected := range expectedDevices {
		endpoint := "http://localhost:" + conf.ServerPort + "/devices/"
		if asLocalId {
			endpoint = endpoint + url.PathEscape(expected.LocalId) + "?as=local_id"
		} else {
			endpoint = endpoint + url.PathEscape(expected.Id)
		}
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
			b, _ := io.ReadAll(resp.Body)
			t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
			return
		}

		result := models.Device{}
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

func testDeviceReadNotFound(t *testing.T, conf config.Config, asLocalId bool, id string) {
	endpoint := "http://localhost:" + conf.ServerPort + "/devices/" + url.PathEscape(id)
	if asLocalId {
		endpoint = endpoint + "?as=local_id"
	}
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
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
}

func TestDeviceLocalIdOwnerConstraintLocalPermissions(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := config.Load("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	conf.FatalErrHandler = t.Fatal
	conf.MongoReplSet = false
	conf.Debug = true
	conf.LocalIdUniqueForOwner = true
	whPort, err := docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	conf.ServerPort = strconv.Itoa(whPort)

	_, ip, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	conf.MongoUrl = "mongodb://" + ip + ":27017"

	_, zkIp, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	conf.KafkaUrl, err = docker.Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}

	err = util.InitTopic(conf.KafkaUrl,
		"concepts",
		"device-groups",
		"aspects",
		"characteristics",
		"processmodel",
		"device-types",
		"hubs",
		"devices",
		"device-classes",
		"functions",
		"protocols",
		"import-types",
		"locations",
		"smart_service_releases",
		"gateway_log",
		"device_log")
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)
	t.Run("test", testDeviceLocalIdOwnerConstraint(ctx, wg, conf))
}

func TestDeviceLocalIdOwnerConstraintPermissionsSearch(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := config.Load("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	conf.FatalErrHandler = t.Fatal
	conf.MongoReplSet = false
	conf.Debug = true
	conf.LocalIdUniqueForOwner = true
	conf, err = docker.NewEnv(ctx, wg, conf)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)
	t.Run("test", testDeviceLocalIdOwnerConstraint(ctx, wg, conf))
}

func testDeviceLocalIdOwnerConstraint(ctx context.Context, wg *sync.WaitGroup, conf config.Config) func(t *testing.T) {
	return func(t *testing.T) {
		err := lib.Start(ctx, wg, conf)
		if err != nil {
			t.Error(err)
			return
		}
		time.Sleep(1 * time.Second)
		producer, err := testutils.NewPublisher(conf)
		if err != nil {
			t.Error(err)
			return
		}

		err = donewait.StartDoneWaitListener(ctx, kafka.Config{
			KafkaUrl:    conf.KafkaUrl,
			StartOffset: kafka.LastOffset,
			Debug:       conf.Debug,
		}, []string{conf.DoneTopic}, nil)
		if err != nil {
			t.Error(err)
			return
		}

		t.Run("create device-type", func(t *testing.T) {
			doneTimeout, _ := context.WithTimeout(ctx, 10*time.Second)
			wait := donewait.AsyncWait(doneTimeout, donewait.DoneMsg{
				ResourceKind: conf.DeviceTypeTopic,
				ResourceId:   devicetype1id,
				Command:      "PUT",
			}, nil)

			err = producer.PublishDeviceType(models.DeviceType{Id: devicetype1id, Name: devicetype1name}, userid)
			if err != nil {
				t.Error(err)
				return
			}

			err = wait()
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("create devices", func(t *testing.T) {
			doneTimeout, _ := context.WithTimeout(ctx, 10*time.Second)
			wait := donewait.AsyncWait(doneTimeout, donewait.DoneMsg{
				ResourceKind: conf.DeviceTopic,
				ResourceId:   testenv.SecendOwnerTokenUser,
				Command:      "PUT",
			}, nil)
			for i := range 10 {
				err = producer.PublishDevice(models.Device{
					Id:           testenv.TestTokenUser + "/" + strconv.Itoa(i),
					LocalId:      strconv.Itoa(i),
					Name:         testenv.TestTokenUser + "/" + strconv.Itoa(i),
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.TestTokenUser,
				}, testenv.TestTokenUser)
				if err != nil {
					t.Error(err)
					return
				}
				err = producer.PublishDevice(models.Device{
					Id:           testenv.SecendOwnerTokenUser + "/" + strconv.Itoa(i),
					LocalId:      strconv.Itoa(i),
					Name:         testenv.SecendOwnerTokenUser + "/" + strconv.Itoa(i),
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.SecendOwnerTokenUser,
				}, testenv.SecendOwnerTokenUser)
				if err != nil {
					t.Error(err)
					return
				}
			}
			err = producer.PublishDevice(models.Device{
				Id:           testenv.TestTokenUser,
				LocalId:      testenv.TestTokenUser,
				Name:         testenv.TestTokenUser,
				DeviceTypeId: devicetype1id,
				OwnerId:      testenv.TestTokenUser,
			}, testenv.TestTokenUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = producer.PublishDevice(models.Device{
				Id:           testenv.SecendOwnerTokenUser,
				LocalId:      testenv.SecendOwnerTokenUser,
				Name:         testenv.SecendOwnerTokenUser,
				DeviceTypeId: devicetype1id,
				OwnerId:      testenv.SecendOwnerTokenUser,
			}, testenv.TestTokenUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = wait()
			if err != nil {
				t.Error(err)
				return
			}

		})

		t.Run("validates", func(t *testing.T) {
			c := client.NewClient("http://localhost:" + conf.ServerPort)
			t.Run("user may add new device with new local-id", func(t *testing.T) {
				err, _ = c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser + "/20",
					LocalId:      "20",
					Name:         testenv.TestTokenUser + "/20",
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.TestTokenUser,
				})
				if err != nil {
					t.Error(err)
					return
				}
			})
			t.Run("user may update device with new local-id", func(t *testing.T) {
				err, _ = c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser + "/1",
					LocalId:      "20",
					Name:         testenv.TestTokenUser + "/1",
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.TestTokenUser,
				})
				if err != nil {
					t.Error(err)
					return
				}
			})
			t.Run("user may update device with local-id existing for other owner", func(t *testing.T) {
				err, _ = c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser + "/1",
					LocalId:      testenv.SecendOwnerTokenUser,
					Name:         testenv.TestTokenUser + "/1",
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.TestTokenUser,
				})
				if err != nil {
					t.Error(err)
					return
				}
			})
			t.Run("user may update device", func(t *testing.T) {
				err, _ = c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser + "/1",
					LocalId:      "1",
					Name:         "updated name",
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.TestTokenUser,
				})
				if err != nil {
					t.Error(err)
					return
				}
			})
			t.Run("user may not add new device with existing local-id", func(t *testing.T) {
				err, code := c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser + "/20",
					LocalId:      "1",
					Name:         testenv.TestTokenUser + "/20",
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.TestTokenUser,
				})
				if err == nil {
					t.Error(err, code)
					return
				}
			})
			t.Run("user may not update device with existing local-id", func(t *testing.T) {
				err, _ = c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser + "/1",
					LocalId:      "2",
					Name:         testenv.TestTokenUser + "/1",
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.TestTokenUser,
				})
				if err == nil {
					t.Error(err)
					return
				}
			})
			t.Run("user may not update owner to none admin", func(t *testing.T) {
				err, _ = c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser + "/1",
					LocalId:      "20",
					Name:         testenv.TestTokenUser + "/1",
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.SecendOwnerTokenUser,
				})
				if err == nil {
					t.Error(err)
					return
				}
			})
			t.Run("user may update owner to admin with not existing local id", func(t *testing.T) {
				doneTimeout, _ := context.WithTimeout(ctx, 10*time.Second)
				wait := donewait.AsyncWait(doneTimeout, donewait.DoneMsg{
					ResourceKind: conf.DeviceTopic,
					ResourceId:   testenv.TestTokenUser + "/1",
					Command:      "RIGHTS",
				}, nil)
				err = producer.PublishDeviceRights(testenv.TestTokenUser+"/1", testenv.TestTokenUser, model.ResourceRights{
					UserRights: map[string]model.Right{
						testenv.TestTokenUser:        {Read: true, Write: true, Execute: true, Administrate: true},
						testenv.SecendOwnerTokenUser: {Read: true, Write: true, Execute: true, Administrate: true},
					},
				})
				if err != nil {
					t.Error(err)
					return
				}
				err = wait()
				if err != nil {
					t.Error(err)
					return
				}
				time.Sleep(time.Second)
				err, _ = c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser + "/1",
					LocalId:      "20",
					Name:         testenv.TestTokenUser + "/1",
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.SecendOwnerTokenUser,
				})
				if err != nil {
					t.Error(err)
					return
				}
			})

			t.Run("user may update owner to admin with not existing local id (unchanged)", func(t *testing.T) {
				doneTimeout, _ := context.WithTimeout(ctx, 10*time.Second)
				wait := donewait.AsyncWait(doneTimeout, donewait.DoneMsg{
					ResourceKind: conf.DeviceTopic,
					ResourceId:   testenv.TestTokenUser,
					Command:      "RIGHTS",
				}, nil)
				err = producer.PublishDeviceRights(testenv.TestTokenUser, testenv.TestTokenUser, model.ResourceRights{
					UserRights: map[string]model.Right{
						testenv.TestTokenUser:        {Read: true, Write: true, Execute: true, Administrate: true},
						testenv.SecendOwnerTokenUser: {Read: true, Write: true, Execute: true, Administrate: true},
					},
				})
				if err != nil {
					t.Error(err)
					return
				}
				err = wait()
				if err != nil {
					t.Error(err)
					return
				}
				time.Sleep(time.Second)
				err, _ = c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser,
					LocalId:      testenv.TestTokenUser,
					Name:         testenv.TestTokenUser,
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.SecendOwnerTokenUser,
				})
				if err != nil {
					t.Error(err)
					return
				}
			})

			t.Run("user may not update owner to admin with existing local id", func(t *testing.T) {
				doneTimeout, _ := context.WithTimeout(ctx, 10*time.Second)
				wait := donewait.AsyncWait(doneTimeout, donewait.DoneMsg{
					ResourceKind: conf.DeviceTopic,
					ResourceId:   testenv.TestTokenUser + "/1",
					Command:      "RIGHTS",
				}, nil)
				err = producer.PublishDeviceRights(testenv.TestTokenUser+"/1", testenv.TestTokenUser, model.ResourceRights{
					UserRights: map[string]model.Right{
						testenv.TestTokenUser:        {Read: true, Write: true, Execute: true, Administrate: true},
						testenv.SecendOwnerTokenUser: {Read: true, Write: true, Execute: true, Administrate: true},
					},
				})
				if err != nil {
					t.Error(err)
					return
				}
				err = wait()
				if err != nil {
					t.Error(err)
					return
				}
				err, _ = c.ValidateDevice(testenv.TestToken, models.Device{
					Id:           testenv.TestTokenUser + "/1",
					LocalId:      "1",
					Name:         testenv.TestTokenUser + "/1",
					DeviceTypeId: devicetype1id,
					OwnerId:      testenv.SecendOwnerTokenUser,
				})
				if err == nil {
					t.Error(err)
					return
				}
			})

		})

	}
}
