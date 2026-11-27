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
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib"
	"github.com/SENERGY-Platform/device-repository/v3/lib/client"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	permclient "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	permmodel "github.com/SENERGY-Platform/permissions-v2/pkg/model"
)

type cacheTestResponse struct {
	status int
	header http.Header
	body   string
}

// TestCacheHeaders runs against the real stack: ETag, If-None-Match and Cache-Control of the endpoints in the positive list
func TestCacheHeaders(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const user1jwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`
	const user1id = "testOwner"

	const user2jwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJzZWNvbmRPd25lciIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwibm9uY2UiOiI5MmM0M2M5NS03NWIwLTQ2Y2YtODBhZS00NWRkOTczYjRiN2YiLCJhdXRoX3RpbWUiOjE1NDY1MDcwMDksInNlc3Npb25fc3RhdGUiOiI1ZGY5MjhmNC0wOGYwLTRlYjktOWI2MC0zYTBhZTIyZWZjNzMiLCJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbIioiXSwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbInVzZXIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJtYXN0ZXItcmVhbG0iOnsicm9sZXMiOlsidmlldy1yZWFsbSIsInZpZXctaWRlbnRpdHktcHJvdmlkZXJzIiwibWFuYWdlLWlkZW50aXR5LXByb3ZpZGVycyIsImltcGVyc29uYXRpb24iLCJjcmVhdGUtY2xpZW50IiwibWFuYWdlLXVzZXJzIiwicXVlcnktcmVhbG1zIiwidmlldy1hdXRob3JpemF0aW9uIiwicXVlcnktY2xpZW50cyIsInF1ZXJ5LXVzZXJzIiwibWFuYWdlLWV2ZW50cyIsIm1hbmFnZS1yZWFsbSIsInZpZXctZXZlbnRzIiwidmlldy11c2VycyIsInZpZXctY2xpZW50cyIsIm1hbmFnZS1hdXRob3JpemF0aW9uIiwibWFuYWdlLWNsaWVudHMiLCJxdWVyeS1ncm91cHMiXX0sImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInJvbGVzIjpbInVzZXIiXX0.cq8YeUuR0jSsXCEzp634fTzNbGkq_B8KbVrwBPgceJ4`
	const user2id = "secondOwner"

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
	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	baseUrl := "http://localhost:" + config.ServerPort
	c := client.NewClient(baseUrl, client.Options{})
	permClient := permclient.New(config.PermissionsV2Url)

	// do sends a request and checks what holds for every response: no Vary header
	do := func(t *testing.T, method string, path string, token string, headers map[string]string) cacheTestResponse {
		t.Helper()
		req, err := http.NewRequest(method, baseUrl+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if token != "" {
			req.Header.Set("Authorization", token)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := http.DefaultTransport.RoundTrip(req) // no automatic handling of cache headers or redirects
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if vary := resp.Header.Get("Vary"); vary != "" {
			t.Errorf("%v %v: unexpected Vary header %q", method, path, vary)
		}
		return cacheTestResponse{status: resp.StatusCode, header: resp.Header, body: string(body)}
	}
	get := func(t *testing.T, path string, token string, ifNoneMatch string) cacheTestResponse {
		t.Helper()
		headers := map[string]string{}
		if ifNoneMatch != "" {
			headers["If-None-Match"] = ifNoneMatch
		}
		return do(t, http.MethodGet, path, token, headers)
	}
	// getOk expects 200 with the cache headers and returns the ETag
	getOk := func(t *testing.T, path string, token string) (etag string, resp cacheTestResponse) {
		t.Helper()
		resp = get(t, path, token, "")
		if resp.status != http.StatusOK {
			t.Fatalf("%v: %v %v", path, resp.status, resp.body)
		}
		etag = resp.header.Get("ETag")
		if !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) || len(etag) != 34 {
			t.Fatalf("%v: unexpected ETag %q", path, etag)
		}
		if resp.header.Get("Cache-Control") != "private, no-cache" {
			t.Fatalf("%v: unexpected Cache-Control %q", path, resp.header.Get("Cache-Control"))
		}
		return etag, resp
	}
	expect304 := func(t *testing.T, path string, token string, etag string) {
		t.Helper()
		resp := get(t, path, token, etag)
		if resp.status != http.StatusNotModified {
			t.Errorf("%v: expected 304, got %v %v", path, resp.status, resp.body)
			return
		}
		if resp.body != "" {
			t.Errorf("%v: expected empty 304 body, got %q", path, resp.body)
		}
		if resp.header.Get("ETag") != etag {
			t.Errorf("%v: unexpected ETag in 304: %q", path, resp.header.Get("ETag"))
		}
		if resp.header.Get("Cache-Control") != "private, no-cache" {
			t.Errorf("%v: unexpected Cache-Control in 304: %q", path, resp.header.Get("Cache-Control"))
		}
		if resp.header.Get("Date") == "" {
			t.Errorf("%v: missing Date in 304", path)
		}
	}
	// expectChanged expects that the old ETag no longer matches and returns the new one
	expectChanged := func(t *testing.T, path string, token string, oldEtag string) (newEtag string) {
		t.Helper()
		resp := get(t, path, token, oldEtag)
		if resp.status != http.StatusOK {
			t.Fatalf("%v: expected 200, got %v %v", path, resp.status, resp.body)
		}
		newEtag = resp.header.Get("ETag")
		if newEtag == "" || newEtag == oldEtag {
			t.Fatalf("%v: expected a new ETag, old %q new %q", path, oldEtag, newEtag)
		}
		return newEtag
	}
	// the timestamps have millisecond resolution
	pause := func() { time.Sleep(10 * time.Millisecond) }

	const (
		aspectId          = models.URN_PREFIX + "aspect:air"
		aspectClassId     = models.URN_PREFIX + "aspect-class:ac1"
		characteristicId  = models.URN_PREFIX + "characteristic:c1"
		conceptId         = models.URN_PREFIX + "concept:co1"
		measuringFunction = models.URN_PREFIX + "measuring-function:getTemperature"
		controllingFunc   = models.URN_PREFIX + "controlling-function:setTemperature"
		deviceClassId     = models.URN_PREFIX + "device-class:dc1"
	)

	var dt1, dt2 models.DeviceType
	t.Run("create metadata", func(t *testing.T) {
		_, err, _ := c.SetAspectClass(context.Background(), client.InternalAdminToken, models.AspectClass{Id: aspectClassId, Name: "ac1"})
		if err != nil {
			t.Fatal(err)
		}
		_, err, _ = c.SetAspect(context.Background(), client.InternalAdminToken, models.Aspect{Id: aspectId, Name: "air"})
		if err != nil {
			t.Fatal(err)
		}
		_, err, _ = c.SetCharacteristic(context.Background(), client.InternalAdminToken, models.Characteristic{Id: characteristicId, Name: "c1", Type: models.String})
		if err != nil {
			t.Fatal(err)
		}
		_, err, _ = c.SetConcept(context.Background(), client.InternalAdminToken, models.Concept{Id: conceptId, Name: "co1", CharacteristicIds: []string{characteristicId}, BaseCharacteristicId: characteristicId})
		if err != nil {
			t.Fatal(err)
		}
		_, err, _ = c.SetFunction(context.Background(), client.InternalAdminToken, models.Function{Id: measuringFunction, Name: "getTemperature", DisplayName: "getTemperature", RdfType: models.SES_ONTOLOGY_MEASURING_FUNCTION})
		if err != nil {
			t.Fatal(err)
		}
		_, err, _ = c.SetFunction(context.Background(), client.InternalAdminToken, models.Function{Id: controllingFunc, Name: "setTemperature", DisplayName: "setTemperature", RdfType: models.SES_ONTOLOGY_CONTROLLING_FUNCTION})
		if err != nil {
			t.Fatal(err)
		}
		_, err, _ = c.SetDeviceClass(context.Background(), client.InternalAdminToken, models.DeviceClass{Id: deviceClassId, Name: "dc1"})
		if err != nil {
			t.Fatal(err)
		}
		_, err, _ = c.SetProtocol(context.Background(), client.InternalAdminToken, models.Protocol{
			Id: "p1", Name: "p1", Handler: "p1", ProtocolSegments: []models.ProtocolSegment{{Id: "ps1", Name: "ps1"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		newDeviceType := func(id string) models.DeviceType {
			return models.DeviceType{
				Id:            id,
				Name:          id,
				DeviceClassId: deviceClassId,
				Services: []models.Service{{
					Id: id + "_s1", LocalId: id + "_s1", Name: "s1", Interaction: models.EVENT_AND_REQUEST, ProtocolId: "p1",
					Outputs: []models.Content{{
						Id: id + "_c1", Serialization: models.JSON, ProtocolSegmentId: "ps1",
						ContentVariable: models.ContentVariable{Id: id + "_cv1", Name: "temperature", Type: models.String, FunctionId: measuringFunction, AspectIds: []string{aspectId}},
					}},
				}},
			}
		}
		dt1, err, _ = c.SetDeviceType(context.Background(), client.InternalAdminToken, newDeviceType("dt1"), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Fatal(err)
		}
		dt2, err, _ = c.SetDeviceType(context.Background(), client.InternalAdminToken, newDeviceType("dt2"), client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Fatal(err)
		}
	})
	if t.Failed() {
		return
	}

	groupA := func() []string {
		return []string{
			"/v2/aspects", "/aspects", "/aspects/" + aspectId,
			"/aspects/" + aspectId + "/measuring-functions", "/aspects/" + aspectId + "/controlling-functions",
			"/aspect-classes", "/aspect-classes/" + aspectClassId,
			"/v2/aspect-nodes", "/aspect-nodes", "/aspect-nodes/" + aspectId,
			"/aspect-nodes/" + aspectId + "/measuring-functions", "/aspect-nodes/" + aspectId + "/controlling-functions",
			"/v2/concepts", "/v2/concepts-with-characteristics", "/concepts/" + conceptId,
			"/v2/characteristics", "/characteristics", "/characteristics/" + characteristicId,
			"/functions", "/functions/" + measuringFunction, "/measuring-functions", "/controlling-functions",
			"/v2/device-classes", "/device-classes", "/device-classes/" + deviceClassId,
			"/device-classes/" + deviceClassId + "/functions", "/device-classes/" + deviceClassId + "/controlling-functions",
			"/v3/device-types", "/device-types", "/device-types/" + dt1.Id,
			"/protocols", "/protocols/p1", "/services/" + dt1.Services[0].Id,
		}
	}

	t.Run("group A: the fixed set names collections the database layer writes", func(t *testing.T) {
		// GetCacheTimestamps names the collections itself; a typo would silently list a timestamp that is always 0
		stamps, err, _ := c.GetLastUpdateTimestamps(context.Background(), client.InternalAdminToken, "")
		if err != nil {
			t.Fatal(err)
		}
		written := map[string]bool{}
		for _, s := range stamps {
			written[s.Collection] = true
		}
		for _, name := range []string{
			config.MongoAspectCollection, config.MongoAspectCollection + "_node", config.MongoAspectClassCollection,
			config.MongoConceptCollection, config.MongoCharacteristicCollection, config.MongoFunctionCollection,
			config.MongoDeviceClassCollection, config.MongoDeviceTypeCollection, config.MongoDeviceTypeCollection + "_criteria",
			config.MongoProtocolCollection,
		} {
			if !written[name] {
				t.Errorf("no last update timestamp for collection %q", name)
			}
		}
	})

	t.Run("group A: 304 until a collection of the fixed set is written", func(t *testing.T) {
		etag, _ := getOk(t, "/v2/aspects", user1jwt)
		expect304(t, "/v2/aspects", user1jwt, etag)
		expect304(t, "/v2/aspects", user2jwt, etag)
		expect304(t, "/v2/aspects", "", etag) //no token needed for a 304 of group A

		pause()
		_, err, _ := c.SetAspect(context.Background(), client.InternalAdminToken, models.Aspect{Id: aspectId + "2", Name: "water"})
		if err != nil {
			t.Fatal(err)
		}
		newEtag := expectChanged(t, "/v2/aspects", user1jwt, etag)
		expect304(t, "/v2/aspects", user1jwt, newEtag)

		// a write on a collection of the fixed set invalidates every group A url, here a protocol
		list := "/v3/device-types"
		listEtag, _ := getOk(t, list, user1jwt)
		pause()
		_, err, _ = c.SetProtocol(context.Background(), client.InternalAdminToken, models.Protocol{
			Id: "p2", Name: "p2", Handler: "p2", ProtocolSegments: []models.ProtocolSegment{{Id: "ps1", Name: "ps1"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		expectChanged(t, list, user1jwt, listEtag)
	})

	t.Run("group A: a device write does not change the ETag", func(t *testing.T) {
		etag, _ := getOk(t, "/v3/device-types", user1jwt)
		pause()
		_, err, _ := c.CreateDevice(context.Background(), user1jwt, models.Device{LocalId: "dev-a-1", Name: "dev-a-1", DeviceTypeId: dt1.Id})
		if err != nil {
			t.Fatal(err)
		}
		expect304(t, "/v3/device-types", user1jwt, etag)
	})

	t.Run("group A: the ETag is bound to the url", func(t *testing.T) {
		etag, _ := getOk(t, "/v2/aspects", user1jwt)
		resp := get(t, "/aspects/"+aspectId+"-unknown", user1jwt, etag)
		if resp.status != http.StatusNotFound {
			t.Errorf("expected 404, got %v %v", resp.status, resp.body)
		}
		if resp.header.Get("ETag") != "" || resp.header.Get("Cache-Control") != "" {
			t.Errorf("an error response must not carry cache headers: %v", resp.header)
		}
		resp = get(t, "/v2/aspects?limit=abc", user1jwt, etag)
		if resp.status != http.StatusBadRequest {
			t.Errorf("expected 400, got %v %v", resp.status, resp.body)
		}
		// the same url with another query is another representation
		other, _ := getOk(t, "/v2/aspects?limit=1", user1jwt)
		if other == etag {
			t.Errorf("expected different ETags for different urls, got %v", other)
		}
		resp = get(t, "/v2/aspects?limit=1", user1jwt, etag)
		if resp.status != http.StatusOK {
			t.Errorf("expected 200, got %v", resp.status)
		}
		// "*" never turns a 404 into a 304
		resp = get(t, "/aspects/"+aspectId+"-unknown", user1jwt, "*")
		if resp.status != http.StatusNotFound {
			t.Errorf("expected 404 for If-None-Match *, got %v", resp.status)
		}
		resp = get(t, "/v2/aspects", user1jwt, "*")
		if resp.status != http.StatusNotModified {
			t.Errorf("expected 304 for If-None-Match * on an existing resource, got %v", resp.status)
		}
	})

	t.Run("group A: If-None-Match forms", func(t *testing.T) {
		etag, _ := getOk(t, "/v2/aspects", user1jwt)
		for name, value := range map[string]string{
			"weak":       "W/" + etag,
			"list":       `"nope", ` + etag + `, "other"`,
			"list weak":  `W/"nope",W/` + etag,
			"whitespace": "  " + etag + "  ",
		} {
			resp := get(t, "/v2/aspects", user1jwt, value)
			if resp.status != http.StatusNotModified {
				t.Errorf("%v: expected 304, got %v", name, resp.status)
			}
		}
		for name, value := range map[string]string{
			"other":     `"nope"`,
			"unquoted":  strings.Trim(etag, `"`),
			"truncated": strings.TrimSuffix(etag, `"`),
			"garbage":   `,,, W/ ,`,
		} {
			resp := get(t, "/v2/aspects", user1jwt, value)
			if resp.status != http.StatusOK {
				t.Errorf("%v: expected 200, got %v", name, resp.status)
			}
		}
	})

	t.Run("group A: same response and ETag for different users", func(t *testing.T) {
		for _, path := range groupA() {
			etag1, resp1 := getOk(t, path, user1jwt)
			etag2, resp2 := getOk(t, path, user2jwt)
			if etag1 != etag2 {
				t.Errorf("%v: ETag differs between users: %v %v", path, etag1, etag2)
			}
			if resp1.body != resp2.body {
				t.Errorf("%v: body differs between users", path)
			}
			expect304(t, path, user2jwt, etag1)
		}
	})

	var d1, d2 models.Device
	t.Run("create devices", func(t *testing.T) {
		var err error
		d1, err, _ = c.CreateDevice(context.Background(), user1jwt, models.Device{LocalId: "d1", Name: "d1", DeviceTypeId: dt1.Id})
		if err != nil {
			t.Fatal(err)
		}
		d2, err, _ = c.CreateDevice(context.Background(), user2jwt, models.Device{LocalId: "d2", Name: "d2", DeviceTypeId: dt1.Id})
		if err != nil {
			t.Fatal(err)
		}
	})
	if t.Failed() {
		return
	}

	t.Run("group B: 304 and change", func(t *testing.T) {
		for _, path := range []string{"/devices", "/devices/" + d1.Id, "/extended-devices", "/extended-devices/" + d1.Id, "/local-devices", "/device-groups", "/hubs", "/extended-hubs", "/locations", "/extended-locations", "/things", "/graphs", "/invalid/device-types", "/defaults/devices/attributes"} {
			resp := get(t, path, user1jwt, "")
			if resp.status != http.StatusOK {
				t.Errorf("%v: unexpected status %v %v", path, resp.status, resp.body)
				continue
			}
			etag, _ := getOk(t, path, user1jwt)
			expect304(t, path, user1jwt, etag)
		}

		etag, _ := getOk(t, "/devices/"+d1.Id, user1jwt)
		d1.Name = "d1-renamed"
		_, err, _ := c.SetDevice(context.Background(), user1jwt, d1, client.DeviceUpdateOptions{})
		if err != nil {
			t.Fatal(err)
		}
		expectChanged(t, "/devices/"+d1.Id, user1jwt, etag)
	})

	t.Run("group B: revoked right changes the ETag without a write", func(t *testing.T) {
		before, _ := getOk(t, "/devices", user1jwt)
		resp := get(t, "/devices", user1jwt, "")
		if strings.Contains(resp.body, d2.Id) {
			t.Fatal("user1 must not see the device of user2 yet")
		}
		ownerPermissions := func(owner string, readers ...string) permclient.ResourcePermissions {
			result := permclient.ResourcePermissions{
				UserPermissions: map[string]permmodel.PermissionsMap{
					owner: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupPermissions: map[string]permmodel.PermissionsMap{"admin": {Read: true, Write: true, Execute: true, Administrate: true}},
				RolePermissions:  map[string]permmodel.PermissionsMap{"admin": {Read: true, Write: true, Execute: true, Administrate: true}},
			}
			for _, reader := range readers {
				result.UserPermissions[reader] = permmodel.PermissionsMap{Read: true}
			}
			return result
		}
		_, err, _ := permClient.SetPermission(permclient.InternalAdminToken, config.DeviceTopic, d2.Id, ownerPermissions(user2id, user1id))
		if err != nil {
			t.Fatal(err)
		}
		shared := expectChanged(t, "/devices", user1jwt, before)
		resp = get(t, "/devices", user1jwt, shared)
		if resp.status != http.StatusNotModified {
			t.Fatalf("expected 304 with the shared state, got %v", resp.status)
		}

		_, err, _ = permClient.SetPermission(permclient.InternalAdminToken, config.DeviceTopic, d2.Id, ownerPermissions(user2id))
		if err != nil {
			t.Fatal(err)
		}
		revoked := expectChanged(t, "/devices", user1jwt, shared)
		if revoked != before {
			t.Errorf("the revoked state has the content of the state before sharing, expected the same ETag: %v %v", revoked, before)
		}
	})

	t.Run("group B: connection state", func(t *testing.T) {
		path := "/extended-devices/" + d1.Id
		err, _ := c.SetDeviceConnectionState(context.Background(), client.InternalAdminToken, d1.Id, true)
		if err != nil {
			t.Fatal(err)
		}
		online, _ := getOk(t, path, user1jwt)
		err, _ = c.SetDeviceConnectionState(context.Background(), client.InternalAdminToken, d1.Id, false)
		if err != nil {
			t.Fatal(err)
		}
		offline := expectChanged(t, path, user1jwt, online)
		err, _ = c.SetDeviceConnectionState(context.Background(), client.InternalAdminToken, d1.Id, true)
		if err != nil {
			t.Fatal(err)
		}
		resp := get(t, path, user1jwt, offline)
		if resp.status != http.StatusOK {
			t.Fatalf("expected 200, got %v", resp.status)
		}
		if resp.header.Get("ETag") != online {
			t.Errorf("expected the ETag of the first online state again, got %v want %v", resp.header.Get("ETag"), online)
		}
	})

	t.Run("group B: X-Total-Count is part of the ETag", func(t *testing.T) {
		path := "/extended-devices?limit=1&sort=name.asc"
		etag, resp := getOk(t, path, user1jwt)
		total := resp.header.Get("X-Total-Count")
		if total == "" {
			t.Fatal("missing X-Total-Count")
		}
		_, err, _ := c.CreateDevice(context.Background(), user1jwt, models.Device{LocalId: "zzz", Name: "zzz-last", DeviceTypeId: dt1.Id})
		if err != nil {
			t.Fatal(err)
		}
		resp = get(t, path, user1jwt, etag)
		if resp.status != http.StatusOK {
			t.Fatalf("expected 200, got %v", resp.status)
		}
		if resp.header.Get("X-Total-Count") == total {
			t.Errorf("expected another total, got %v", total)
		}
		if resp.header.Get("ETag") == etag {
			t.Error("the ETag must change with the total")
		}
	})

	t.Run("group B: no 304 across users", func(t *testing.T) {
		etag1, resp1 := getOk(t, "/devices", user1jwt)
		etag2, resp2 := getOk(t, "/devices", user2jwt)
		if resp1.body == resp2.body {
			t.Fatal("test needs users with different visible devices")
		}
		if etag1 == etag2 {
			t.Errorf("expected different ETags for users with different devices, got %v", etag1)
		}
		resp := get(t, "/devices", user2jwt, etag1)
		if resp.status != http.StatusOK {
			t.Errorf("a validator of user1 must not give a 304 to user2, got %v", resp.status)
		}
		resp = get(t, "/devices", user1jwt, etag2)
		if resp.status != http.StatusOK {
			t.Errorf("a validator of user2 must not give a 304 to user1, got %v", resp.status)
		}
	})

	t.Run("user-device-types follows the devices of the user", func(t *testing.T) {
		path := "/user-device-types"
		etag, _ := getOk(t, path, user1jwt)
		expect304(t, path, user1jwt, etag)
		typesEtag, _ := getOk(t, "/v3/device-types", user1jwt)
		pause()
		_, err, _ := c.CreateDevice(context.Background(), user1jwt, models.Device{LocalId: "dev-dt2", Name: "dev-dt2", DeviceTypeId: dt2.Id})
		if err != nil {
			t.Fatal(err)
		}
		expectChanged(t, path, user1jwt, etag)
		expect304(t, "/v3/device-types", user1jwt, typesEtag) //the device type itself did not change
	})

	t.Run("hubs: HEAD carries no ETag", func(t *testing.T) {
		hub, err, _ := c.SetHub(context.Background(), user1jwt, models.Hub{Name: "h1", Hash: "h", DeviceLocalIds: []string{d1.LocalId}, DeviceIds: []string{d1.Id}}, model.HubUpdateOptions{})
		if err != nil {
			t.Fatal(err)
		}
		getEtag, _ := getOk(t, "/hubs/"+hub.Id, user1jwt)
		expect304(t, "/hubs/"+hub.Id, user1jwt, getEtag)
		resp := do(t, http.MethodHead, "/hubs/"+hub.Id, user1jwt, nil)
		if resp.status != http.StatusOK {
			t.Fatalf("expected 200, got %v", resp.status)
		}
		if resp.header.Get("ETag") != "" || resp.header.Get("Cache-Control") != "" {
			t.Errorf("HEAD must carry no cache headers: %v", resp.header)
		}
		resp = do(t, http.MethodHead, "/hubs/"+hub.Id, user1jwt, map[string]string{"If-None-Match": getEtag})
		if resp.status != http.StatusOK {
			t.Errorf("HEAD must not answer 304, got %v", resp.status)
		}
	})

	t.Run("clear-cache", func(t *testing.T) {
		resp := do(t, http.MethodPost, "/clear-cache", "", map[string]string{"Origin": "http://example.test"})
		if resp.status != http.StatusNoContent {
			t.Fatalf("expected 204, got %v %v", resp.status, resp.body)
		}
		if resp.header.Get("Clear-Site-Data") != `"cache"` {
			t.Errorf("unexpected Clear-Site-Data %q", resp.header.Get("Clear-Site-Data"))
		}
		if resp.header.Get("Access-Control-Allow-Origin") != "http://example.test" {
			t.Errorf("unexpected Access-Control-Allow-Origin %q", resp.header.Get("Access-Control-Allow-Origin"))
		}
		if resp.header.Get("Cache-Control") != "" || resp.header.Get("ETag") != "" {
			t.Errorf("clear-cache must not be cached: %v", resp.header)
		}
		// also with a token that is garbage: the endpoint does not look at it
		resp = do(t, http.MethodPost, "/clear-cache", "Bearer garbage", nil)
		if resp.status != http.StatusNoContent {
			t.Errorf("expected 204, got %v", resp.status)
		}
	})

	t.Run("cors", func(t *testing.T) {
		resp := do(t, http.MethodOptions, "/devices", "", map[string]string{"Origin": "http://example.test", "Access-Control-Request-Headers": "if-none-match"})
		if !strings.Contains(resp.header.Get("Access-Control-Allow-Headers"), "If-None-Match") {
			t.Errorf("If-None-Match missing in Access-Control-Allow-Headers: %q", resp.header.Get("Access-Control-Allow-Headers"))
		}
		if resp.header.Get("Access-Control-Expose-Headers") != "ETag, X-Total-Count" {
			t.Errorf("unexpected Access-Control-Expose-Headers %q", resp.header.Get("Access-Control-Expose-Headers"))
		}
	})

	t.Run("304 carries the cors header", func(t *testing.T) {
		for _, path := range []string{"/v2/aspects", "/devices"} {
			etag, _ := getOk(t, path, user1jwt)
			resp := do(t, http.MethodGet, path, user1jwt, map[string]string{"If-None-Match": etag, "Origin": "http://example.test"})
			if resp.status != http.StatusNotModified {
				t.Fatalf("%v: expected 304, got %v", path, resp.status)
			}
			if resp.header.Get("Access-Control-Allow-Origin") != "http://example.test" {
				t.Errorf("%v: unexpected Access-Control-Allow-Origin %q", path, resp.header.Get("Access-Control-Allow-Origin"))
			}
			if resp.header.Get("Cache-Control") == "" || resp.header.Get("ETag") != etag {
				t.Errorf("%v: 304 without Cache-Control or ETag: %v", path, resp.header)
			}
			if resp.header.Get("Access-Control-Expose-Headers") != "ETag, X-Total-Count" {
				t.Errorf("%v: unexpected Access-Control-Expose-Headers %q", path, resp.header.Get("Access-Control-Expose-Headers"))
			}
		}
	})

	t.Run("no cache headers outside the positive list", func(t *testing.T) {
		for _, path := range []string{"/last-update-timestamps", "/export", "/health", "/doc", "/helper/id?short_id=AAAAAAAAAAAAAAAAAAAAAA&prefix=b", "/"} {
			resp := get(t, path, client.InternalAdminToken, "")
			if resp.status >= 300 {
				t.Errorf("%v: unexpected status %v %v", path, resp.status, resp.body)
			}
			for _, h := range []string{"ETag", "Cache-Control", "Last-Modified"} {
				if resp.header.Get(h) != "" {
					t.Errorf("%v: unexpected header %v: %q", path, h, resp.header.Get(h))
				}
			}
		}
		resp := do(t, http.MethodPost, "/query/aspect-nodes", user1jwt, nil)
		if resp.header.Get("ETag") != "" || resp.header.Get("Cache-Control") != "" {
			t.Errorf("POST query must not carry cache headers: %v", resp.header)
		}
		resp = do(t, http.MethodGet, "/users/"+user1id, user1jwt, nil)
		if resp.header.Get("ETag") != "" || resp.header.Get("Cache-Control") != "" {
			t.Errorf("/users must not carry cache headers: %v", resp.header)
		}
	})
}
