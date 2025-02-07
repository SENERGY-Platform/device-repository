/*
 * Copyright 2021 InfAI (CC SES)
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

func testDeviceBatchDelete(port string) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+port+"/protocols?wait=true", models.Protocol{
			Name:             "p3",
			Handler:          "ph3",
			ProtocolSegments: []models.ProtocolSegment{{Name: "ps3"}},
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
			Name:         "delete_d1",
			DeviceTypeId: dt.Id,
			LocalId:      "dlid1",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatal(resp.Status, resp.StatusCode)
		}

		d1 := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&d1)
		if err != nil {
			t.Fatal(err)
		}

		if d1.Id == "" {
			t.Fatal(d1)
		}

		resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
			Name:         "delete_d2",
			DeviceTypeId: dt.Id,
			LocalId:      "dlid2",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatal(resp.Status, resp.StatusCode)
		}

		d2 := models.Device{}
		err = json.NewDecoder(resp.Body).Decode(&d2)
		if err != nil {
			t.Fatal(err)
		}

		if d2.Id == "" {
			t.Fatal(d2)
		}

		resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/devices?wait=true", models.Device{
			Name:         "delete_d3",
			DeviceTypeId: dt.Id,
			LocalId:      "dlid3",
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

		if d3.Id == "" {
			t.Fatal(d3)
		}

		resp, err = helper.JwtDeleteWithBody(userjwt, "http://localhost:"+port+"/devices", []string{d1.Id, d2.Id})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatal(resp.Status, resp.StatusCode)
		}

		resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(d1.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		//expect 404 error
		if resp.StatusCode != http.StatusNotFound {
			t.Fatal(resp.Status, resp.StatusCode)
		}

		resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(d2.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		//expect 404 error
		if resp.StatusCode != http.StatusNotFound {
			t.Fatal(resp.Status, resp.StatusCode)
		}

		resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/devices/"+url.PathEscape(d3.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatal(resp.Status, resp.StatusCode)
		}

	}
}
