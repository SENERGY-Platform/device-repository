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
	"github.com/SENERGY-Platform/device-repository/lib/tests/manager_legacy/helper"
	"github.com/SENERGY-Platform/models/go/models"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func testLocalDevice(t *testing.T, port string) {
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

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/local-devices?wait=true", models.Device{
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

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/local-devices?wait=true", models.Device{
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

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/local-devices?wait=true", models.Device{
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

	device := models.Device{}
	err = json.NewDecoder(resp.Body).Decode(&device)
	if err != nil {
		t.Fatal(err)
	}

	if device.Id == "" {
		t.Fatal(device)
	}

	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/local-devices/"+url.PathEscape(device.LocalId))
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

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/local-devices?wait=true", models.Device{
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

	//update
	resp, err = helper.Jwtput(userjwt, "http://localhost:"+port+"/local-devices/"+url.PathEscape(device.LocalId)+"?wait=true", models.Device{
		Name:         "updated_device_name",
		DeviceTypeId: dt.Id,
		LocalId:      device.LocalId,
		Id:           device.Id,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/local-devices/"+url.PathEscape(device.LocalId))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	result = models.Device{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Name != "updated_device_name" || result.LocalId != "lid1" || result.DeviceTypeId != dt.Id {
		t.Fatal(result)
	}

	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/local-devices?ids="+url.QueryEscape(device.LocalId+",unknown,foo"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	list := []models.Device{}
	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 1 || list[0].Name != "updated_device_name" || list[0].LocalId != "lid1" || list[0].DeviceTypeId != dt.Id {
		t.Fatal(list)
	}

	//delete
	resp, err = helper.Jwtdelete(userjwt, "http://localhost:"+port+"/local-devices/"+url.PathEscape(device.LocalId)+"?wait=true")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/local-devices/"+url.PathEscape(device.LocalId))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	//expect 404 error
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal(resp.Status, resp.StatusCode)
	}
}
