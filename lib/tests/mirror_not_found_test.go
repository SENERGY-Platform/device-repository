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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib"
	"github.com/SENERGY-Platform/device-repository/v3/lib/client"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/docker"
	"github.com/SENERGY-Platform/device-repository/v3/lib/tests/repo_legacy/testenv"
	"github.com/SENERGY-Platform/mgw-cloud-proxy/cert-manager/lib/models/service"
	"github.com/SENERGY-Platform/models/go/models"
)

// TestMirrorPullsOnNotFound checks that a GET the mirror can not answer asks the source for changes before returning 404.
// The update interval is set to an hour, so only the not found reads can bring the new resources into the mirror.
func TestMirrorPullsOnNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.SyncLockDuration = time.Second.String()
	config.Debug = true
	config.RestLogger()

	_, mongoIp, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MongoUrl = "mongodb://" + mongoIp + ":27017"
	config.KafkaUrl = "-"
	config.PermissionsV2Url = "-"

	sourceConfig, err := docker.NewEnv(ctx, wg, configuration.Config{})
	if err != nil {
		t.Error(err)
		return
	}

	_, repoIp, err := docker.DeviceRepo(ctx, wg, "../..", sourceConfig.KafkaUrl, sourceConfig.MongoUrl, sourceConfig.PermissionsV2Url)
	if err != nil {
		t.Error(err)
		return
	}

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, "unable to parse url: "+r.URL.String(), http.StatusBadRequest)
			return
		}
		target.Host = repoIp + ":8080"
		target.Scheme = "http"
		req, err := http.NewRequest(r.Method, target.String(), r.Body)
		if err != nil {
			http.Error(w, "unable to create request: "+target.String(), http.StatusBadRequest)
			return
		}
		req.Header = r.Header
		req.Header.Set("Authorization", testenv.SecondOwnerToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "unable to send request: "+target.String(), http.StatusBadRequest)
			return
		}
		defer resp.Body.Close()
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}))
	defer proxyServer.Close()

	certManagerMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(service.NetworkInfo{
			UserID: testenv.SecendOwnerTokenUser,
		})
	}))
	defer certManagerMockServer.Close()

	config.MgwMirrorSourceUrl = proxyServer.URL
	config.MgwCertManagerUrl = certManagerMockServer.URL
	config.AsMgwMirror = true
	config.MgwMirrorUpdateInterval = "1h"

	sourceClient := client.NewClient("http://"+repoIp+":8080", nil)
	mirrorClient := client.NewClient("http://localhost:"+config.ServerPort, nil)

	t.Run("start mirror", func(t *testing.T) {
		err = lib.Start(ctx, wg, config)
		if err != nil {
			t.Error(err)
			return
		}
		//let the initial pull of the empty source finish
		time.Sleep(2 * time.Second)
	})

	t.Run("unknown id stays not found", func(t *testing.T) {
		_, err, code := mirrorClient.ReadProtocol("unknown", "")
		if err == nil || code != http.StatusNotFound {
			t.Error("expected 404, got", code, err)
		}
	})

	protocol := models.Protocol{}
	t.Run("protocol created in source after the initial pull", func(t *testing.T) {
		protocol, err, _ = sourceClient.SetProtocol(client.InternalAdminToken, models.Protocol{
			Id:               "p1",
			Name:             "p1",
			Handler:          "p1",
			ProtocolSegments: []models.ProtocolSegment{{Name: "ps1"}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		result, err, code := mirrorClient.ReadProtocol(protocol.Id, "")
		if err != nil {
			t.Error(code, err)
			return
		}
		if result.Name != "p1" {
			t.Errorf("unexpected result: %#v", result)
		}
	})

	t.Run("device created in source after the initial pull", func(t *testing.T) {
		dc, err, _ := sourceClient.SetDeviceClass(client.InternalAdminToken, models.DeviceClass{Name: "dc1"})
		if err != nil {
			t.Error(err)
			return
		}
		dt, err, _ := sourceClient.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Name:          "dt1",
			DeviceClassId: dc.Id,
			Services: []models.Service{{
				LocalId:     "s1",
				Name:        "s1",
				Interaction: models.EVENT_AND_REQUEST,
				ProtocolId:  protocol.Id,
			}},
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		device, err, _ := sourceClient.CreateDevice(testenv.SecondOwnerToken, models.Device{
			LocalId:      "d1",
			Name:         "d1",
			DeviceTypeId: dt.Id,
			OwnerId:      testenv.SecendOwnerTokenUser,
		})
		if err != nil {
			t.Error(err)
			return
		}
		result, err, code := mirrorClient.ReadDeviceByLocalId(testenv.SecendOwnerTokenUser, "d1", "", model.READ)
		if err != nil {
			t.Error(code, err)
			return
		}
		if result.Id != device.Id {
			t.Errorf("unexpected result: %#v", result)
		}
		//the device-type changed in the same pull
		_, err, code = mirrorClient.ReadDeviceType(dt.Id, "")
		if err != nil {
			t.Error(code, err)
		}
	})

}
