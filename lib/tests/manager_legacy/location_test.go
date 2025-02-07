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

func testLocation(port string) func(t *testing.T) {
	return func(t *testing.T) {
		resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+port+"/locations?wait=true", models.Location{
			Name: "foo",
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		Location := models.Location{}
		err = json.NewDecoder(resp.Body).Decode(&Location)
		if err != nil {
			t.Fatal(err)
		}

		if Location.Id == "" {
			t.Fatal(Location)
		}

		result := models.Location{}
		resp, err = helper.Jwtget(adminjwt, "http://localhost:"+port+"/locations/"+url.PathEscape(Location.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Log("http://localhost:" + port + "/locations/" + url.PathEscape(Location.Id))
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		result = models.Location{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		if result.Name != "foo" {
			t.Fatal(result)
		}

		resp, err = helper.Jwtdelete(adminjwt, "http://localhost:"+port+"/locations/"+url.PathEscape(Location.Id)+"?wait=true")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		resp, err = helper.Jwtget(adminjwt, "http://localhost:"+port+"/locations/"+url.PathEscape(Location.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Fatal(resp.Status, resp.StatusCode)
		}
	}
}
