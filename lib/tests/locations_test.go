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
	"bytes"
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/models/go/models"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"testing"
)

func TestLocations(t *testing.T) {
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
	deviceManagerUrl := "http://localhost:" + conf.ServerPort //manager has been replaced/integrated into device-repository

	locations := []models.Location{}
	userLocations := []models.Location{}
	secondOwnerLocations := []models.Location{}

	controller.DisableFeaturesForTestEnv = false

	t.Run("create Userjwt locations", func(t *testing.T) {
		names := []string{"a1", "b1", "c2", "d2", "e3", "f3", "g4"}
		for _, name := range names {
			location := models.Location{
				Name: name,
			}
			buf, err := json.Marshal(location)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := http.NewRequest(http.MethodPost, deviceManagerUrl+"/locations", bytes.NewReader(buf))
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Authorization", Userjwt)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			if resp.StatusCode >= 300 {
				payload, _ := io.ReadAll(resp.Body)
				t.Error(resp.StatusCode, string(payload))
				return
			}
			result := models.Location{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				t.Error(err)
				return
			}
			locations = append(locations, result)
			userLocations = append(userLocations, result)
		}
	})

	t.Run("create SecondOwnerToken locations", func(t *testing.T) {
		names := []string{"h1", "i1", "j2", "k2", "l3", "m3", "n4"}
		for _, name := range names {
			location := models.Location{
				Name: name,
			}
			buf, err := json.Marshal(location)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := http.NewRequest(http.MethodPost, deviceManagerUrl+"/locations", bytes.NewReader(buf))
			if err != nil {
				t.Error(err)
				return
			}
			req.Header.Set("Authorization", SecondOwnerToken)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			if resp.StatusCode >= 300 {
				t.Error(resp.StatusCode)
				return
			}
			result := models.Location{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			if err != nil {
				t.Error(err)
				return
			}
			locations = append(locations, result)
			secondOwnerLocations = append(secondOwnerLocations, result)
		}
	})

	t.Run("check locations as admin", func(t *testing.T) {
		if len(locations) == 0 {
			t.Error("no locations")
			return
		}
		for _, expected := range locations {
			actual, err, _ := c.GetLocation(expected.Id, AdminToken)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("\ne=%#v\na=%#v\n", expected, actual)
				return
			}
		}
	})

	t.Run("check locations as Userjwt", func(t *testing.T) {
		for _, expected := range userLocations {
			actual, err, _ := c.GetLocation(expected.Id, Userjwt)
			if err != nil {
				t.Error(expected.Name, ":", err)
				return
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("name=%v\ne=%#v\na=%#v\n", expected.Name, expected, actual)
				return
			}
		}
		for _, location := range secondOwnerLocations {
			_, err, _ := c.GetLocation(location.Id, Userjwt)
			if err == nil {
				t.Error("expected error")
				return
			}
		}
	})

	t.Run("check locations as secondOwner", func(t *testing.T) {
		for _, expected := range secondOwnerLocations {
			actual, err, _ := c.GetLocation(expected.Id, SecondOwnerToken)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("\ne=%#v\na=%#v\n", expected, actual)
				return
			}
		}
		for _, location := range userLocations {
			_, err, _ := c.GetLocation(location.Id, SecondOwnerToken)
			if err == nil {
				t.Error("expected error")
				return
			}
		}
	})

	t.Run("list locations as admin", func(t *testing.T) {
		actual, total, err, _ := c.ListLocations(AdminToken, client.LocationListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != int64(len(locations)) {
			t.Errorf("\na=%#v\ne=%#v\n", total, len(locations))
		}
		if !reflect.DeepEqual(actual, locations) {
			t.Errorf("\na=%#v\ne=%#v\n", actual, locations)
		}
	})

	t.Run("list locations as user", func(t *testing.T) {
		actual, total, err, _ := c.ListLocations(Userjwt, client.LocationListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != int64(len(userLocations)) {
			t.Errorf("\na=%#v\ne=%#v\n", total, len(userLocations))
		}
		if !reflect.DeepEqual(actual, userLocations) {
			t.Errorf("\na=%#v\ne=%#v\n", actual, userLocations)
		}
	})

	t.Run("search 2 as user", func(t *testing.T) {
		actual, total, err, _ := c.ListLocations(Userjwt, client.LocationListOptions{Search: "2"})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.Location{
			userLocations[2],
			userLocations[3],
		}
		if total != 2 {
			t.Errorf("\na=%#v\ne=%#v\n", total, 2)
			return
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\na=%#v\ne=%#v\n", actual, expected)
			return
		}
	})

	t.Run("limit offset", func(t *testing.T) {
		actual, total, err, _ := c.ListLocations(Userjwt, client.LocationListOptions{Limit: 3, Offset: 2})
		if err != nil {
			t.Error(err)
			return
		}
		expected := []models.Location{
			userLocations[2],
			userLocations[3],
			userLocations[4],
		}
		if total != int64(len(userLocations)) {
			t.Errorf("\na=%#v\ne=%#v\n", total, int64(len(userLocations)))
			return
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("\na=%#v\ne=%#v\n", actual, expected)
			return
		}
	})

	t.Run("try update", func(t *testing.T) {
		location := userLocations[0]
		location.Description = "new description"
		buf, err := json.Marshal(location)
		if err != nil {
			t.Error(err)
			return
		}
		req, err := http.NewRequest(http.MethodPut, deviceManagerUrl+"/locations/"+url.PathEscape(location.Id), bytes.NewReader(buf))
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", Userjwt)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode >= 300 {
			t.Error(resp.StatusCode)
			return
		}
	})

	t.Run("try delete", func(t *testing.T) {
		location := userLocations[0]
		req, err := http.NewRequest(http.MethodDelete, deviceManagerUrl+"/locations/"+url.PathEscape(location.Id), nil)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", Userjwt)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode >= 300 {
			t.Error(resp.StatusCode)
			return
		}
	})

	t.Run("try read of deleted location", func(t *testing.T) {
		location := userLocations[0]
		_, err, _ := c.GetLocation(location.Id, Userjwt)
		if err == nil {
			t.Error("expected error")
			return
		}
	})
}
