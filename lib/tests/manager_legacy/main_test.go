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
	"context"
	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/device-repository/lib/tests/manager_legacy/helper"
	"github.com/SENERGY-Platform/models/go/models"
	"strconv"
	"sync"
	"testing"
	"time"
)

const adminjwt = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ"
const userjwt = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiIzMmE1OTljZC0zNDgxLTQzYWUtYWY0NC04YTVmNjU4NzYxZTUiLCJleHAiOjE1NjI5MjAwMDUsIm5iZiI6MCwiaWF0IjoxNTYyOTE2NDA1LCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJlYmJhZDkyNy00YzM5LTRkMTItODY5MC04OWIwNjdkZDRjZTciLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiNTVlMzA4N2UtZjljNi00MmQ2LWE0MmEtMGZiMjcxNWE4OTkyIiwiYXV0aF90aW1lIjoxNTYyOTE2NDA0LCJzZXNzaW9uX3N0YXRlIjoiYmU5MDQ2MmYtOGE3Yy00NWU4LTg1MjAtMGRlYzViZWI1ZWZlIiwiYWNyIjoiMSIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1bWFfYXV0aG9yaXphdGlvbiIsInVzZXIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1bWFfYXV0aG9yaXphdGlvbiIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJpbmdvIn0.pggKYb3V0VxFINWBqpFE_t14MKhSM7bhw8YqrYBRvOzh8ft7zu_-bOvLOYbJBwo0GU1D68U2d_eerkYEIt-mc0dNtdFasy5DG_GtvnWA4nsbf0BVsYKSZcRiDK4d4qbHu9NMjBdEwSkP9KDGEtou0yHtOnVzB1eHHNm_uSUO-O_kz2LWsXOPK2sbL1LTiCKS0XToJPdlaNczDMZB0nXR3sHbyi3Lwk-Va2ATS6Kke5M1KmFMowK-Y0jK2urt8GnCBIXvZMT6gUW9-dvlv4w_lAuVXQ9hFg_r0sBnoWzZOUR_xlrz2T-syjrZzmXlAkJrcD8KWPH-lCs0jD9pdiROhQ"
const userjwtUser = "ebbad927-4c39-4d12-8690-89b067dd4ce7"

const userid = "dd69ea0d-f553-4336-80f3-7f4567f85c7b"

const Userid = "testOwner"
const Userjwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`

const AdminTokenUser = "admin"
const AdminToken = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJhZG1pbiIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwibm9uY2UiOiI5MmM0M2M5NS03NWIwLTQ2Y2YtODBhZS00NWRkOTczYjRiN2YiLCJhdXRoX3RpbWUiOjE1NDY1MDcwMDksInNlc3Npb25fc3RhdGUiOiI1ZGY5MjhmNC0wOGYwLTRlYjktOWI2MC0zYTBhZTIyZWZjNzMiLCJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbIioiXSwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbInVzZXIiLCJhZG1pbiJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LXJlYWxtIiwidmlldy1pZGVudGl0eS1wcm92aWRlcnMiLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidXNlciIsImFkbWluIl19.ggcFFFEsjwdfSzEFzmZt_m6W4IiSQub2FRhZVfWttDI`

const SecendOwnerTokenUser = "secondOwner"
const SecondOwnerToken = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJzZWNvbmRPd25lciIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwibm9uY2UiOiI5MmM0M2M5NS03NWIwLTQ2Y2YtODBhZS00NWRkOTczYjRiN2YiLCJhdXRoX3RpbWUiOjE1NDY1MDcwMDksInNlc3Npb25fc3RhdGUiOiI1ZGY5MjhmNC0wOGYwLTRlYjktOWI2MC0zYTBhZTIyZWZjNzMiLCJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbIioiXSwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbInVzZXIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJtYXN0ZXItcmVhbG0iOnsicm9sZXMiOlsidmlldy1yZWFsbSIsInZpZXctaWRlbnRpdHktcHJvdmlkZXJzIiwibWFuYWdlLWlkZW50aXR5LXByb3ZpZGVycyIsImltcGVyc29uYXRpb24iLCJjcmVhdGUtY2xpZW50IiwibWFuYWdlLXVzZXJzIiwicXVlcnktcmVhbG1zIiwidmlldy1hdXRob3JpemF0aW9uIiwicXVlcnktY2xpZW50cyIsInF1ZXJ5LXVzZXJzIiwibWFuYWdlLWV2ZW50cyIsIm1hbmFnZS1yZWFsbSIsInZpZXctZXZlbnRzIiwidmlldy11c2VycyIsInZpZXctY2xpZW50cyIsIm1hbmFnZS1hdXRob3JpemF0aW9uIiwibWFuYWdlLWNsaWVudHMiLCJxdWVyeS1ncm91cHMiXX0sImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInJvbGVzIjpbInVzZXIiXX0.cq8YeUuR0jSsXCEzp634fTzNbGkq_B8KbVrwBPgceJ4`

var userIdToUserToken = map[string]string{
	Userid:               Userjwt,
	SecendOwnerTokenUser: SecondOwnerToken,
	AdminTokenUser:       AdminToken,
}

func TestWithDocker(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := configuration.Load("./../../../config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}

	port, err := helper.GetFreePort()
	if err != nil {
		t.Fatal(err)
	}
	conf.ServerPort = strconv.Itoa(port)

	conf, err = docker.NewEnv(ctx, wg, conf)
	if err != nil {
		t.Error(err)
		return
	}

	err = lib.Start(ctx, wg, conf)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	tests(t, conf, false)
}

const a1Id = models.URN_PREFIX + "aspect:a1"
const f1Id = models.URN_PREFIX + "controlling-function:f1"

func tests(t *testing.T, conf configuration.Config, mock bool) {
	t.Run("id helper", func(t *testing.T) {
		testHelperId(t, conf.ServerPort)
	})
	t.Run("aspects", testAspect(conf.ServerPort))
	t.Run("functions", testFunction(conf.ServerPort))
	t.Run("deviceclasses", testDeviceClass(conf.ServerPort))
	t.Run("locations", testLocation(conf.ServerPort))

	t.Run("testDeviceType", func(t *testing.T) {
		testDeviceType(t, conf.ServerPort)
	})

	t.Run("testDeviceTypeWithDistinctAttributes", func(t *testing.T) {
		if mock {
			t.Skip("the mocks are not implemented for this test")
		}
		testDeviceTypeWithDistinctAttributes(t, conf.ServerPort)
	})

	t.Run("testDeviceTypeWithServiceGroups", func(t *testing.T) {
		testDeviceTypeWithServiceGroups(t, conf.ServerPort)
	})

	t.Run("testDevice", func(t *testing.T) {
		testDevice(t, conf.ServerPort)
	})

	t.Run("testDeviceAttributes", func(t *testing.T) {
		testDeviceAttributes(t, conf.ServerPort)
	})

	t.Run("testLocalDevice", func(t *testing.T) {
		if mock {
			t.Skip("missing endpoint in permissions search mock")
		}
		testLocalDevice(t, conf.ServerPort)
	})

	t.Run("test device owner", func(t *testing.T) {
		if mock {
			t.Skip("the mocks are not implemented for this test")
		}
		testDeviceOwner(t, conf)
	})

	t.Run("testHub", func(t *testing.T) {
		if mock {
			t.Skip("the mocks are not implemented for this test (completeHub() wont find devices)")
		}
		testHub(t, conf.ServerPort)
	})

	t.Run("testHubAssertions", func(t *testing.T) {
		if mock {
			t.Skip("the mocks are not implemented for this test (completeHub() wont find devices)")
		}
		testHubAssertions(t, conf.ServerPort)
	})

	t.Run("test hub owner", func(t *testing.T) {
		if mock {
			t.Skip("the mocks are not implemented for this test")
		}
		testHubOwner(t, conf)
	})

	t.Run("testConcepts", func(t *testing.T) {
		testConcepts(t, conf)
	})

	t.Run("testCharacteristics", func(t *testing.T) {
		testCharacteristics(t, conf)
	})

	t.Run("testDeviceGroup", testDeviceGroup(conf.ServerPort))

	t.Run("testDeviceBatchDelete", testDeviceBatchDelete(conf.ServerPort))
}
