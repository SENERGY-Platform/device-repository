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
	"reflect"
	"testing"
)

func testDeviceGroup(port string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("delete unknown", func(t *testing.T) {
			resp, err := helper.Jwtdelete(userjwt, "http://localhost:"+port+"/device-groups/unknown?wait=true")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
				b, _ := io.ReadAll(resp.Body)
				t.Fatal(resp.Status, resp.StatusCode, string(b))
			}
		})

		t.Run("tests", func(t *testing.T) {
			resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-groups?wait=true", models.DeviceGroup{
				Name: "dg1",
			})
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(resp.Body)
				t.Fatal(resp.Status, resp.StatusCode, string(b))
			}

			deviceGroup := models.DeviceGroup{}
			err = json.NewDecoder(resp.Body).Decode(&deviceGroup)
			if err != nil {
				t.Fatal(err)
			}

			if deviceGroup.Id == "" {
				t.Fatal(deviceGroup)
			}

			resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/device-groups/"+url.PathEscape(deviceGroup.Id))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(resp.Body)
				t.Fatal(resp.Status, resp.StatusCode, string(b))
			}

			result := models.DeviceGroup{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				t.Fatal(err)
			}

			if result.Name != "dg1" {
				t.Fatal(result)
			}

			resp, err = helper.Jwtput(userjwt, "http://localhost:"+port+"/device-groups/"+url.PathEscape(deviceGroup.Id)+"?wait=true", models.DeviceGroup{
				Id:   deviceGroup.Id,
				Name: "dg2",
			})
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(resp.Body)
				t.Fatal(resp.Status, resp.StatusCode, string(b))
			}

			deviceGroup = models.DeviceGroup{}
			err = json.NewDecoder(resp.Body).Decode(&deviceGroup)
			if err != nil {
				t.Fatal(err)
			}

			if deviceGroup.Id == "" {
				t.Fatal(deviceGroup)
			}

			resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/device-groups/"+url.PathEscape(deviceGroup.Id))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(resp.Body)
				t.Fatal(resp.Status, resp.StatusCode, string(b))
			}

			result = models.DeviceGroup{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				t.Fatal(err)
			}

			if result.Name != "dg2" {
				t.Fatal(result)
			}
		})
	}
}

func TestDeviceGroupShortCriteria(t *testing.T) {
	dg := models.DeviceGroup{Criteria: []models.DeviceGroupFilterCriteria{
		{
			FunctionId:  "f1",
			AspectId:    "a1",
			Interaction: models.EVENT,
		},
		{
			FunctionId:  "f1",
			AspectId:    "a2",
			Interaction: models.EVENT,
		},
		{
			FunctionId:    "f2",
			DeviceClassId: "dc1",
			Interaction:   models.REQUEST,
		},
	}}
	dg.SetShortCriteria()

	expected := models.DeviceGroup{
		Criteria: []models.DeviceGroupFilterCriteria{
			{
				FunctionId:  "f1",
				AspectId:    "a1",
				Interaction: models.EVENT,
			},
			{
				FunctionId:  "f1",
				AspectId:    "a2",
				Interaction: models.EVENT,
			},
			{
				FunctionId:    "f2",
				DeviceClassId: "dc1",
				Interaction:   models.REQUEST,
			},
		},
		CriteriaShort: []string{"f1_a1__event", "f1_a2__event", "f2__dc1_request"},
	}

	if !reflect.DeepEqual(dg, expected) {
		t.Error(dg, expected)
	}
}
