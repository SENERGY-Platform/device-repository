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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/tests/manager_legacy/helper"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func testHubOwner(t *testing.T, conf config.Config) {
	//replace sleep with wait=true query parameter
	tempSleepAfterEdit := helper.SleepAfterEdit
	helper.SleepAfterEdit = 0
	defer func() {
		helper.SleepAfterEdit = tempSleepAfterEdit
	}()

	hub1 := models.Hub{}
	t.Run("create hubs with implicit owner", func(t *testing.T) {
		hub := models.Hub{
			Name: "hub1",
		}
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+conf.ServerPort+"/hubs?wait=true", hub)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&hub1)
		if err != nil {
			t.Error(err)
			return
		}
		hub.Id = hub1.Id
		hub.OwnerId = userjwtUser
		hub.DeviceIds = []string{}
		hub.DeviceLocalIds = []string{}
		hub1.DeviceIds = []string{}
		hub1.DeviceLocalIds = []string{}
		if !reflect.DeepEqual(hub1, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub1, hub)
			return
		}
	})

	t.Run("check hub1 after create", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+url.PathEscape(hub1.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		hub := models.Hub{}
		err = json.NewDecoder(resp.Body).Decode(&hub)
		if err != nil {
			t.Error(err)
			return
		}
		hub.DeviceIds = []string{}
		hub.DeviceLocalIds = []string{}
		hub1.DeviceIds = []string{}
		hub1.DeviceLocalIds = []string{}
		if !reflect.DeepEqual(hub1, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub1, hub)
			return
		}
	})

	hub2 := models.Hub{}
	t.Run("create hub with explicit owner", func(t *testing.T) {
		hub := models.Hub{
			Name:    "hub1",
			OwnerId: userjwtUser,
		}
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+conf.ServerPort+"/hubs?wait=true", hub)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&hub2)
		if err != nil {
			t.Error(err)
			return
		}
		hub.Id = hub2.Id
		hub.OwnerId = userjwtUser
		hub.DeviceIds = []string{}
		hub.DeviceLocalIds = []string{}
		hub2.DeviceIds = []string{}
		hub2.DeviceLocalIds = []string{}
		if !reflect.DeepEqual(hub2, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub2, hub)
			return
		}
	})

	t.Run("check hub2 after create", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+url.PathEscape(hub2.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		hub := models.Hub{}
		err = json.NewDecoder(resp.Body).Decode(&hub)
		if err != nil {
			t.Error(err)
			return
		}
		hub.DeviceIds = []string{}
		hub.DeviceLocalIds = []string{}
		hub2.DeviceIds = []string{}
		hub2.DeviceLocalIds = []string{}
		if !reflect.DeepEqual(hub2, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub2, hub)
			return
		}
	})

	t.Run("update hub with implicit owner", func(t *testing.T) {
		hub := hub1
		hub.Name = "hub1 update1"
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+url.PathEscape(hub.Id)+"?wait=true", hub)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&hub1)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(hub1, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub1, hub)
			return
		}
	})

	t.Run("check hub after update", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+url.PathEscape(hub1.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		hub := models.Hub{}
		err = json.NewDecoder(resp.Body).Decode(&hub)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(hub1, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub1, hub)
			return
		}
	})

	t.Run("update hub with explicit owner", func(t *testing.T) {
		hub := hub1
		hub.Name = "hub1 update2"
		hub.OwnerId = userjwtUser
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+url.PathEscape(hub.Id)+"?wait=true", hub)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&hub1)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(hub1, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub1, hub)
			return
		}
	})

	t.Run("check hub after update", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+url.PathEscape(hub1.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		hub := models.Hub{}
		err = json.NewDecoder(resp.Body).Decode(&hub)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(hub1, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub1, hub)
			return
		}
	})

	t.Run("try create hub with foreign owner", func(t *testing.T) {
		resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+conf.ServerPort+"/hubs?wait=true", models.Hub{
			Name:    "hub1",
			OwnerId: userjwtUser,
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
		hub := hub1
		hub.OwnerId = userid
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+url.PathEscape(hub.Id)+"?wait=true", hub)
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
		_, err, _ := client.New(conf.PermissionsV2Url).SetPermission(client.InternalAdminToken, conf.HubTopic, hub1.Id, client.ResourcePermissions{
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
		hub := hub1
		hub.OwnerId = userid
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+url.PathEscape(hub.Id)+"?wait=true", hub)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		err = json.NewDecoder(resp.Body).Decode(&hub1)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(hub1, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub1, hub)
			return
		}
	})

	t.Run("check hub after update", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+conf.ServerPort+"/hubs/"+url.PathEscape(hub1.Id)+"?wait=true")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			temp, _ := io.ReadAll(resp.Body)
			t.Errorf("%v %v", resp.Status, string(temp))
			return
		}
		hub := models.Hub{}
		err = json.NewDecoder(resp.Body).Decode(&hub)
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(hub1, hub) {
			t.Errorf("ERROR: \n%#v\n!=\n%#v\n", hub1, hub)
			return
		}
	})
}

func testHub(t *testing.T, port string) {

	protocol := models.Protocol{}
	t.Run("protocol", func(t *testing.T) {
		resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+port+"/protocols?wait=true", models.Protocol{
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
	t.Run("dt", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true", models.DeviceType{
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
			t.Error(resp.Status, resp.StatusCode, string(b))
			return
		}
		err = json.NewDecoder(resp.Body).Decode(&dt)
		if err != nil {
			t.Error(err)
			return
		}

		if dt.Id == "" {
			t.Error(dt)
			return
		}
	})

	device1 := models.Device{}
	t.Run("device1", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
			Name:         "d1",
			DeviceTypeId: dt.Id,
			LocalId:      "hublid1",
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Error(resp.Status, resp.StatusCode)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(&device1)
		if err != nil {
			t.Error(err)
			return
		}

		if device1.Id == "" {
			t.Error(device1)
			return
		}
	})

	device2 := models.Device{}
	t.Run("device2", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
			Name:         "d2",
			DeviceTypeId: dt.Id,
			LocalId:      "hublid2",
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Error(resp.Status, resp.StatusCode)
			return
		}
		err = json.NewDecoder(resp.Body).Decode(&device2)
		if err != nil {
			t.Error(err)
			return
		}

		if device2.Id == "" {
			t.Error(device2)
			return
		}
		time.Sleep(time.Second)
	})

	t.Run("invalid 1", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/hubs?wait=true", models.Hub{})
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		//expect validation error
		if resp.StatusCode == http.StatusOK {
			t.Error(resp.Status, resp.StatusCode)
			return
		}
	})

	t.Run("invalid 2", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/hubs?wait=true", models.Hub{
			Name:           "h1",
			DeviceLocalIds: []string{"unknown"},
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		//expect validation error
		if resp.StatusCode == http.StatusOK {
			t.Error(resp.Status, resp.StatusCode)
			return
		}
	})

	t.Run("invalid 3", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/hubs?wait=true", models.Hub{
			Name:      "h1",
			DeviceIds: []string{"unknown"},
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		//expect validation error
		if resp.StatusCode == http.StatusOK {
			resultHub := models.Hub{}
			json.NewDecoder(resp.Body).Decode(&resultHub)
			t.Errorf("%v %#v", resp.Status, resultHub)
			return
		}
	})

	hub := models.Hub{}
	t.Run("create 1", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/hubs?wait=true", models.Hub{
			Name:           "h1",
			Hash:           "foobar",
			DeviceLocalIds: []string{device1.LocalId},
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

		err = json.NewDecoder(resp.Body).Decode(&hub)
		if err != nil {
			t.Error(err)
			return
		}

		if hub.Id == "" {
			t.Error(hub)
			return
		}
		if hub.Name != "h1" || hub.Hash != "foobar" || !reflect.DeepEqual(hub.DeviceLocalIds, []string{device1.LocalId}) || !reflect.DeepEqual(hub.DeviceIds, []string{device1.Id}) {
			t.Error(hub)
			return
		}
	})

	hub2 := models.Hub{}
	t.Run("create 2", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/hubs?wait=true", models.Hub{
			Name:      "h2",
			Hash:      "foobar",
			DeviceIds: []string{device2.Id},
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

		err = json.NewDecoder(resp.Body).Decode(&hub2)
		if err != nil {
			t.Error(err)
			return
		}

		if hub2.Id == "" {
			t.Error(hub)
			return
		}
		if hub2.Name != "h2" || hub2.Hash != "foobar" || !reflect.DeepEqual(hub2.DeviceLocalIds, []string{device2.LocalId}) || !reflect.DeepEqual(hub2.DeviceIds, []string{device2.Id}) {
			t.Error(hub2)
			return
		}
	})

	t.Run("name 1", func(t *testing.T) {
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(hub.Id)+"/name?wait=true", "h1_changed")
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
	})

	t.Run("name 2", func(t *testing.T) {
		resp, err := helper.Jwtput(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(hub2.Id)+"/name?wait=true", "h2_changed")
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

		result := models.Hub{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
			return
		}

		if result.Name != "h2_changed" || result.Hash != "foobar" || !reflect.DeepEqual(result.DeviceLocalIds, []string{device2.LocalId}) || !reflect.DeepEqual(result.DeviceIds, []string{device2.Id}) {
			t.Error(result)
			return
		}
	})

	t.Run("get", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(hub.Id))
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

		result := models.Hub{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
			return
		}

		if result.Name != "h1_changed" || result.Hash != "foobar" || !reflect.DeepEqual(result.DeviceLocalIds, []string{device1.LocalId}) || !reflect.DeepEqual(result.DeviceIds, []string{device1.Id}) {
			t.Error(result)
			return
		}
	})

	t.Run("get 2", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(hub2.Id))
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

		result := models.Hub{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Error(err)
			return
		}

		if result.Name != "h2_changed" || result.Hash != "foobar" || !reflect.DeepEqual(result.DeviceLocalIds, []string{device2.LocalId}) || !reflect.DeepEqual(result.DeviceIds, []string{device2.Id}) {
			t.Error(result)
			return
		}
	})

	t.Run("delete", func(t *testing.T) {
		resp, err := helper.Jwtdelete(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(hub.Id)+"?wait=true")
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
	})

	t.Run("check delete", func(t *testing.T) {
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(hub.Id))
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()

		//expect 404 error
		if resp.StatusCode != http.StatusNotFound {
			t.Error(resp.Status, resp.StatusCode)
			return
		}
	})

}

func testHubAssertions(t *testing.T, port string) {
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

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
		Name:         "d3",
		DeviceTypeId: dt.Id,
		LocalId:      "lid3",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.Status, resp.StatusCode)
	}

	d3 := models.Device{}
	err = json.NewDecoder(resp.Body).Decode(&d3)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
		Name:         "d4",
		DeviceTypeId: dt.Id,
		LocalId:      "lid4",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.Status, resp.StatusCode)
	}

	d4 := models.Device{}
	err = json.NewDecoder(resp.Body).Decode(&d4)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
		Name:         "d5",
		DeviceTypeId: dt.Id,
		LocalId:      "lid5",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.Status, resp.StatusCode)
	}

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/hubs?wait=true", models.Hub{
		Name:           "h2",
		Hash:           "foobar",
		DeviceLocalIds: []string{"lid3", "lid4", "lid5"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(payload))
	}

	hub := models.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&hub)
	if err != nil {
		t.Fatal(err)
	}

	if hub.Id == "" {
		t.Fatal(hub)
	}

	// update hub on device local id change

	resp, err = helper.Jwtput(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(d3.Id)+"?wait=true", models.Device{
		Id:           d3.Id,
		Name:         "d3",
		DeviceTypeId: dt.Id,
		LocalId:      "lid3_changed",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.Status, resp.StatusCode)
	}

	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(hub.Id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	result := models.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Name != hub.Name || result.Hash != "" || !reflect.DeepEqual(result.DeviceLocalIds, []string{"lid3_changed", "lid4", "lid5"}) {
		t.Fatalf("%#v", result)
	}

	// update hub on device delete

	resp, err = helper.Jwtdelete(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(d4.Id)+"?wait=true")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.Status, resp.StatusCode)
	}

	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(hub.Id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	result = models.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Name != hub.Name || result.Hash != "" || !reflect.DeepEqual(result.DeviceLocalIds, []string{"lid3_changed", "lid5"}) {
		t.Fatal(result)
	}

	// only one hub may have device

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/hubs?wait=true", models.Hub{
		Name:           "h3",
		Hash:           "foobar",
		DeviceLocalIds: []string{"lid5"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.Status, resp.StatusCode)
	}

	newHub := models.Hub{}
	err = json.NewDecoder(resp.Body).Decode(&newHub)
	if err != nil {
		t.Fatal(err)
	}

	if newHub.Id == "" {
		t.Fatal(newHub)
	}

	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(hub.Id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	err = json.NewDecoder(resp.Body).Decode(&hub)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/hubs/"+url.PathEscape(newHub.Id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	err = json.NewDecoder(resp.Body).Decode(&newHub)
	if err != nil {
		t.Fatal(err)
	}

	if len(hub.DeviceLocalIds) != 1 || len(newHub.DeviceLocalIds) != 1 {
		t.Fatal(hub, newHub)
	}
}
