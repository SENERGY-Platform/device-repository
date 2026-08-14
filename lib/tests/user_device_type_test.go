/*
 * Copyright 2026 InfAI (CC SES)
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
	"slices"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
)

func TestUserDeviceTypes(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const user1jwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`

	const user2jwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJzZWNvbmRPd25lciIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwibm9uY2UiOiI5MmM0M2M5NS03NWIwLTQ2Y2YtODBhZS00NWRkOTczYjRiN2YiLCJhdXRoX3RpbWUiOjE1NDY1MDcwMDksInNlc3Npb25fc3RhdGUiOiI1ZGY5MjhmNC0wOGYwLTRlYjktOWI2MC0zYTBhZTIyZWZjNzMiLCJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbIioiXSwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbInVzZXIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJtYXN0ZXItcmVhbG0iOnsicm9sZXMiOlsidmlldy1yZWFsbSIsInZpZXctaWRlbnRpdHktcHJvdmlkZXJzIiwibWFuYWdlLWlkZW50aXR5LXByb3ZpZGVycyIsImltcGVyc29uYXRpb24iLCJjcmVhdGUtY2xpZW50IiwibWFuYWdlLXVzZXJzIiwicXVlcnktcmVhbG1zIiwidmlldy1hdXRob3JpemF0aW9uIiwicXVlcnktY2xpZW50cyIsInF1ZXJ5LXVzZXJzIiwibWFuYWdlLWV2ZW50cyIsIm1hbmFnZS1yZWFsbSIsInZpZXctZXZlbnRzIiwidmlldy11c2VycyIsInZpZXctY2xpZW50cyIsIm1hbmFnZS1hdXRob3JpemF0aW9uIiwibWFuYWdlLWNsaWVudHMiLCJxdWVyeS1ncm91cHMiXX0sImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInJvbGVzIjpbInVzZXIiXX0.cq8YeUuR0jSsXCEzp634fTzNbGkq_B8KbVrwBPgceJ4`

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.SyncLockDuration = time.Second.String()
	config.Debug = true

	config, err = docker.NewEnv(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	c := client.NewClient("http://localhost:"+config.ServerPort, nil)

	var protocol models.Protocol

	t.Run("create protocol", func(t *testing.T) {
		protocol, err, _ = c.SetProtocol(client.InternalAdminToken, models.Protocol{
			Name:    "p1",
			Handler: "p1",
			ProtocolSegments: []models.ProtocolSegment{{
				Id:   "ps1",
				Name: "ps1",
			}},
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	var deviceTypes []models.DeviceType
	t.Run("create device-types", func(t *testing.T) {
		for i := range 10 {
			deviceType, err, _ := c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
				Name: "dt_" + strconv.Itoa(i),
				Services: []models.Service{
					{
						LocalId:     "s" + strconv.Itoa(i),
						Name:        "s" + strconv.Itoa(i),
						Interaction: models.REQUEST,
						ProtocolId:  protocol.Id,
					},
				},
			}, client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			deviceTypes = append(deviceTypes, deviceType)
		}
	})

	expectedDtsForUser := map[string][]string{}
	t.Run("create devices", func(t *testing.T) {
		for i := range 10 {
			token := ""
			switch i % 3 {
			case 0:
				continue
			case 1:
				token = user1jwt
			case 2:
				token = user2jwt
			}
			_, err, _ = c.CreateDevice(token, models.Device{
				LocalId:      "d" + strconv.Itoa(i),
				Name:         "d" + strconv.Itoa(i),
				DeviceTypeId: deviceTypes[i].Id,
			})
			if err != nil {
				t.Error(err)
				return
			}
			expectedDtsForUser[token] = append(expectedDtsForUser[token], deviceTypes[i].Id)
			expectedDtsForUser[client.InternalAdminToken] = append(expectedDtsForUser[client.InternalAdminToken], deviceTypes[i].Id)
		}
	})

	t.Run("check expectedDtsForUser", func(t *testing.T) {
		if len(expectedDtsForUser[user1jwt]) != 3 {
			t.Error(len(expectedDtsForUser[user1jwt]))
		}
		if len(expectedDtsForUser[user2jwt]) != 3 {
			t.Error(len(expectedDtsForUser[user2jwt]))
		}
		if len(expectedDtsForUser[client.InternalAdminToken]) != 6 {
			t.Error(len(expectedDtsForUser[client.InternalAdminToken]))
		}
	})

	t.Run("check user device-types", func(t *testing.T) {
		tokenList := []string{user1jwt, user2jwt, client.InternalAdminToken}
		for i, token := range tokenList {
			list, total, err, _ := c.ListDeviceTypesUsedByUser(token, model.DeviceTypeListOptions{})
			if err != nil {
				t.Error(i, err)
				return
			}
			if len(list) != len(expectedDtsForUser[token]) {
				t.Error(i, len(list))
				return
			}
			if int(total) != len(expectedDtsForUser[token]) {
				t.Error(i, total)
				return
			}
			dtIds := []string{}
			for _, dt := range list {
				dtIds = append(dtIds, dt.Id)
			}
			slices.Sort(dtIds)
			expecpected := expectedDtsForUser[token]
			slices.Sort(expecpected)
			if !slices.Equal(dtIds, expecpected) {
				t.Errorf("%v\n%#v\n%#v\n", i, dtIds, expecpected)
				return
			}
		}

	})

	t.Run("check user1 filtered device-types", func(t *testing.T) {
		list, total, err, _ := c.ListDeviceTypesUsedByUser(user1jwt, model.DeviceTypeListOptions{Ids: []string{expectedDtsForUser[user1jwt][0], expectedDtsForUser[user2jwt][0], "unknown"}})
		if err != nil {
			t.Error(err)
			return
		}
		if len(list) != 1 {
			t.Error(len(list))
			return
		}
		if total != 1 {
			t.Error(total)
			return
		}
		if list[0].Id != expectedDtsForUser[user1jwt][0] {
			t.Error(list)
			return
		}
	})

}
