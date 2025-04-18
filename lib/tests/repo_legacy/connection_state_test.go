/*
 * Copyright 2025 InfAI (CC SES)
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

package repo_legacy

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/google/uuid"
	"reflect"
	"sync"
	"testing"
)

func TestConnectionStateHandling(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	c := client.NewClient("http://localhost:"+conf.ServerPort, nil)

	_, err, _ = c.SetDeviceType(AdminToken, models.DeviceType{Id: devicetype1id, Name: devicetype1name}, client.DeviceTypeUpdateOptions{})
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

	_, err, _ = c.SetDevice(userjwt, d1, client.DeviceUpdateOptions{})
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

	_, err, _ = c.SetDevice(userjwt, d2, client.DeviceUpdateOptions{})
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

	_, err, _ = c.SetDevice(userjwt, d3, client.DeviceUpdateOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	h1 := models.Hub{
		Id:             "urn:infai:ses:hub:" + uuid.NewString(),
		Name:           "a h1",
		DeviceLocalIds: []string{d1.LocalId},
		DeviceIds:      []string{d1.Id},
		OwnerId:        userid,
	}

	hx1 := models.ExtendedHub{
		Hub:             h1,
		ConnectionState: models.ConnectionStateUnknown,
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
	}

	_, err, _ = c.SetHub(userjwt, h1)
	if err != nil {
		t.Error(err)
		return
	}

	h2 := models.Hub{
		Id:             "urn:infai:ses:hub:" + uuid.NewString(),
		Name:           "b h2",
		DeviceLocalIds: []string{d2.LocalId},
		DeviceIds:      []string{d2.Id},
		OwnerId:        userid,
	}

	hx2 := models.ExtendedHub{
		Hub:             h2,
		ConnectionState: models.ConnectionStateUnknown,
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
	}

	_, err, _ = c.SetHub(userjwt, h2)
	if err != nil {
		t.Error(err)
		return
	}

	h3 := models.Hub{
		Id:             "urn:infai:ses:hub:" + uuid.NewString(),
		Name:           "a h3",
		DeviceLocalIds: []string{d3.LocalId},
		DeviceIds:      []string{d3.Id},
		OwnerId:        userid,
	}

	hx3 := models.ExtendedHub{
		Hub:             h3,
		ConnectionState: models.ConnectionStateUnknown,
		Shared:          false,
		Permissions: models.Permissions{
			Read:         true,
			Write:        true,
			Execute:      true,
			Administrate: true,
		},
	}

	_, err, _ = c.SetHub(userjwt, h3)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("check with unknown connection state", func(t *testing.T) {
		t.Run("check extended device list", func(t *testing.T) {
			t.Run("list all", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3, dx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list online", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list offline", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateOffline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list unknown", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateUnknown})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3, dx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'd1'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Search: "d1"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'a d'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Search: "a d"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search online 'a'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Search: "a", ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
		})
		t.Run("check extended device d1", func(t *testing.T) {
			result, err, _ := c.ReadExtendedDevice(d1.Id, userjwt, models.Read, false)
			if err != nil {
				t.Error(err)
				return
			}
			expected := dx1
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
		t.Run("check extended hub list", func(t *testing.T) {
			t.Run("filter hub by device local id", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{LocalDeviceId: d1.LocalId})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("filter hub by device local id and owner", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{LocalDeviceId: d3.LocalId, OwnerId: d3.OwnerId})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list all", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx3, hx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list online", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list offline", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{ConnectionState: client.ConnectionStateOffline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list unknown", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{ConnectionState: client.ConnectionStateUnknown})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx3, hx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'h1'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{Search: "h1"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'a h'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{Search: "a h"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search online 'a'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{Search: "a", ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
		})
		t.Run("check extended hub h", func(t *testing.T) {
			result, err, _ := c.ReadExtendedHub(h1.Id, userjwt, models.Read)
			if err != nil {
				t.Error(err)
				return
			}
			expected := hx1
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
	})

	t.Run("set connection state", func(t *testing.T) {
		t.Run("set unknown1 online", func(t *testing.T) {
			dx1.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetDeviceConnectionState(AdminToken, "unknown1", true)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set unknown2 offline", func(t *testing.T) {
			dx1.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetDeviceConnectionState(AdminToken, "unknown2", true)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set unknown1 hub online", func(t *testing.T) {
			dx1.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetDeviceConnectionState(AdminToken, "unknown1", true)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set unknown2 hub offline", func(t *testing.T) {
			dx1.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetDeviceConnectionState(AdminToken, "unknown2", true)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set d1 online", func(t *testing.T) {
			dx1.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetDeviceConnectionState(AdminToken, d1.Id, true)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set d2 online", func(t *testing.T) {
			dx2.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetDeviceConnectionState(AdminToken, d2.Id, true)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set d3 offline", func(t *testing.T) {
			dx3.ConnectionState = models.ConnectionStateOffline
			err, _ = c.SetDeviceConnectionState(AdminToken, d3.Id, false)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set h1 online", func(t *testing.T) {
			hx1.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetHubConnectionState(AdminToken, h1.Id, true)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set h2 online", func(t *testing.T) {
			hx2.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetHubConnectionState(AdminToken, h2.Id, true)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set h3 offline", func(t *testing.T) {
			hx3.ConnectionState = models.ConnectionStateOffline
			err, _ = c.SetHubConnectionState(AdminToken, h3.Id, false)
			if err != nil {
				t.Error(err)
				return
			}
		})
	})

	t.Run("check with set connection state", func(t *testing.T) {
		t.Run("check extended device list", func(t *testing.T) {
			t.Run("list all", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3, dx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list online", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list offline", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateOffline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list unknown", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateUnknown})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'd1'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Search: "d1"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'a d'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Search: "a d"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search online 'a'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Search: "a", ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
		})
		t.Run("check extended device d1", func(t *testing.T) {
			result, err, _ := c.ReadExtendedDevice(d1.Id, userjwt, models.Read, false)
			if err != nil {
				t.Error(err)
				return
			}
			expected := dx1
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
		t.Run("check extended hub list", func(t *testing.T) {
			t.Run("list all", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx3, hx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list online", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list offline", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{ConnectionState: client.ConnectionStateOffline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list unknown", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{ConnectionState: client.ConnectionStateUnknown})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'h1'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{Search: "h1"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'a h'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{Search: "a h"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search online 'a'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{Search: "a", ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
		})
		t.Run("check extended hub h", func(t *testing.T) {
			result, err, _ := c.ReadExtendedHub(h1.Id, userjwt, models.Read)
			if err != nil {
				t.Error(err)
				return
			}
			expected := hx1
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
	})

	t.Run("update connection state", func(t *testing.T) {
		t.Run("set d1 online", func(t *testing.T) {
			dx1.ConnectionState = models.ConnectionStateOffline
			err, _ = c.SetDeviceConnectionState(AdminToken, d1.Id, false)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set d2 online", func(t *testing.T) {
			dx2.ConnectionState = models.ConnectionStateOffline
			err, _ = c.SetDeviceConnectionState(AdminToken, d2.Id, false)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set d3 offline", func(t *testing.T) {
			dx3.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetDeviceConnectionState(AdminToken, d3.Id, true)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set h1 online", func(t *testing.T) {
			hx1.ConnectionState = models.ConnectionStateOffline
			err, _ = c.SetHubConnectionState(AdminToken, h1.Id, false)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set h2 online", func(t *testing.T) {
			hx2.ConnectionState = models.ConnectionStateOffline
			err, _ = c.SetHubConnectionState(AdminToken, h2.Id, false)
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("set h3 offline", func(t *testing.T) {
			hx3.ConnectionState = models.ConnectionStateOnline
			err, _ = c.SetHubConnectionState(AdminToken, h3.Id, true)
			if err != nil {
				t.Error(err)
				return
			}
		})
	})

	t.Run("check with updated connection state", func(t *testing.T) {
		t.Run("check extended device list", func(t *testing.T) {
			t.Run("list all", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3, dx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("\na=%#v\ne=%#v\n", result, expected)
				}
			})
			t.Run("list online", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list offline", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateOffline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list unknown", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateUnknown})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'd1'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Search: "d1"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'a d'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Search: "a d"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search online 'a'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(userjwt, client.ExtendedDeviceListOptions{Search: "a", ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
		})
		t.Run("check extended device d1", func(t *testing.T) {
			result, err, _ := c.ReadExtendedDevice(d1.Id, userjwt, models.Read, false)
			if err != nil {
				t.Error(err)
				return
			}
			expected := dx1
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
		t.Run("check extended hub list", func(t *testing.T) {
			t.Run("list all", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx3, hx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list online", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list offline", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{ConnectionState: client.ConnectionStateOffline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list unknown", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{ConnectionState: client.ConnectionStateUnknown})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'h1'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{Search: "h1"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'a h'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{Search: "a h"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search online 'a'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(userjwt, client.HubListOptions{Search: "a", ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
		})
		t.Run("check extended hub h", func(t *testing.T) {
			result, err, _ := c.ReadExtendedHub(h1.Id, userjwt, models.Read)
			if err != nil {
				t.Error(err)
				return
			}
			expected := hx1
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
	})

	t.Run("check as admin", func(t *testing.T) {
		t.Run("check extended device list", func(t *testing.T) {
			t.Run("list all", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3, dx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list online", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list offline", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateOffline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list unknown", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{ConnectionState: client.ConnectionStateUnknown})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'd1'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{Search: "d1"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'a d'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{Search: "a d"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search online 'a'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{Search: "a", ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("pagination", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{Limit: 1, Offset: 1})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx3} //by sort: 3 before 2
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("sort display_name asc", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{SortBy: "display_name"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx1, dx3, dx2} //by sort: 3 before 2
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("sort display_name desc", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedDevices(adminjwt, client.ExtendedDeviceListOptions{SortBy: "display_name.desc"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedDevice{dx2, dx3, dx1} //by sort: 3 before 2
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
		})
		t.Run("check extended device d1", func(t *testing.T) {
			result, err, _ := c.ReadExtendedDevice(d1.Id, adminjwt, models.Read, false)
			if err != nil {
				t.Error(err)
				return
			}
			expected := dx1
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
		t.Run("check extended hub list", func(t *testing.T) {
			t.Run("list all", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(adminjwt, client.HubListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx3, hx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list online", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(adminjwt, client.HubListOptions{ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list offline", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(adminjwt, client.HubListOptions{ConnectionState: client.ConnectionStateOffline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx2}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("list unknown", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(adminjwt, client.HubListOptions{ConnectionState: client.ConnectionStateUnknown})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 0 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'h1'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(adminjwt, client.HubListOptions{Search: "h1"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search 'a h'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(adminjwt, client.HubListOptions{Search: "a h"})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 2 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx1, hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
			t.Run("search online 'a'", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(adminjwt, client.HubListOptions{Search: "a", ConnectionState: client.ConnectionStateOnline})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 1 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx3}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})

			t.Run("pagination", func(t *testing.T) {
				result, total, err, _ := c.ListExtendedHubs(adminjwt, client.HubListOptions{Limit: 1, Offset: 1})
				if err != nil {
					t.Error(err)
					return
				}
				if total != 3 {
					t.Error(total)
					return
				}
				expected := []models.ExtendedHub{hx3} //by sort: 3 before 2
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("%#v\n", result)
				}
			})
		})
		t.Run("check extended hub h", func(t *testing.T) {
			result, err, _ := c.ReadExtendedHub(h1.Id, adminjwt, models.Read)
			if err != nil {
				t.Error(err)
				return
			}
			expected := hx1
			if !reflect.DeepEqual(result, expected) {
				t.Errorf("%#v\n", result)
			}
		})
	})
}
