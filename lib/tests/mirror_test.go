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
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/device-repository/lib/tests/repo_legacy/testenv"
	"github.com/SENERGY-Platform/mgw-cloud-proxy/cert-manager/lib/models/service"
	"github.com/SENERGY-Platform/models/go/models"
	client2 "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

func TestMirror(t *testing.T) {
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

	proxyReqCount := atomic.Int64{}

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReqCount.Add(1)
		target, err := url.Parse(r.URL.String())
		if err != nil {
			t.Error(err)
			http.Error(w, "unable to parse url: "+r.URL.String(), http.StatusBadRequest)
			return
		}
		target.Host = repoIp + ":8080"
		target.Port()
		target.Scheme = "http"
		endpoint := target.String()
		log.Println("proxy request forwarded to: ", r.Method, endpoint)
		req, err := http.NewRequest(r.Method, endpoint, r.Body)
		if err != nil {
			t.Error(err)
			http.Error(w, "unable to create request: "+target.String(), http.StatusBadRequest)
			return
		}
		req.Header = r.Header
		req.Header.Set("Authorization", testenv.SecondOwnerToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
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
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			t.Error(err)
		}
	}))
	defer proxyServer.Close()

	certManagerMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(service.NetworkInfo{
			UserID: testenv.SecendOwnerTokenUser,
		})
		if err != nil {
			t.Error(err)
			return
		}
	}))
	defer certManagerMockServer.Close()

	config.MgwMirrorSourceUrl = proxyServer.URL
	config.MgwCertManagerUrl = certManagerMockServer.URL
	config.AsMgwMirror = true
	config.MgwMirrorUpdateInterval = "10s"

	sourceClient := client.NewClient("http://"+repoIp+":8080", nil)
	mirrorClient := client.NewClient("http://localhost:"+config.ServerPort, nil)

	t.Run("create initial source state", func(t *testing.T) {
		p1, err, _ := sourceClient.SetProtocol(client.InternalAdminToken, models.Protocol{
			Id:      "p1",
			Name:    "p1",
			Handler: "p1",
			ProtocolSegments: []models.ProtocolSegment{{
				Name: "ps1",
			}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		a1, err, _ := sourceClient.SetAspect(client.InternalAdminToken, models.Aspect{
			Name:       "a1",
			SubAspects: []models.Aspect{{Name: "a1.1"}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		c1, err, _ := sourceClient.SetCharacteristic(client.InternalAdminToken, models.Characteristic{
			Name: "c1",
			Type: models.String,
		})
		if err != nil {
			t.Error(err)
			return
		}
		co1, err, _ := sourceClient.SetConcept(client.InternalAdminToken, models.Concept{
			Name:                 "co1",
			BaseCharacteristicId: c1.Id,
			CharacteristicIds:    []string{c1.Id},
		})
		if err != nil {
			t.Error(err)
			return
		}
		f1, err, _ := sourceClient.SetFunction(client.InternalAdminToken, models.Function{
			Name:        "f1",
			DisplayName: "f1",
			ConceptId:   co1.Id,
			RdfType:     models.SES_ONTOLOGY_MEASURING_FUNCTION,
		})
		if err != nil {
			t.Error(err)
			return
		}

		dc1, err, _ := sourceClient.SetDeviceClass(client.InternalAdminToken, models.DeviceClass{
			Name: "dc1",
		})

		dt1, err, _ := sourceClient.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Name: "dt1",
			Services: []models.Service{
				{
					LocalId:     "s1",
					Name:        "s1",
					Interaction: models.EVENT_AND_REQUEST,
					ProtocolId:  "p1",
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Name:             "cv1",
								Type:             models.String,
								CharacteristicId: c1.Id,
								FunctionId:       f1.Id,
								AspectId:         a1.Id,
							},
							Serialization:     models.JSON,
							ProtocolSegmentId: p1.ProtocolSegments[0].Id,
						},
					},
				},
			},
			DeviceClassId: dc1.Id,
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}

		_, err, _ = sourceClient.CreateDevice(testenv.SecondOwnerToken, models.Device{
			LocalId:      "d1",
			Name:         "d1",
			DeviceTypeId: dt1.Id,
			OwnerId:      testenv.SecendOwnerTokenUser,
		})
		if err != nil {
			t.Error(err)
			return
		}

		_, err, _ = sourceClient.CreateDevice(testenv.TestToken, models.Device{
			LocalId:      "shouldNotBeFound",
			Name:         "shouldNotBeFound",
			DeviceTypeId: dt1.Id,
			OwnerId:      testenv.TestTokenUser,
		})
		if err != nil {
			t.Error(err)
			return
		}

		_, err, _ = sourceClient.SetHub(testenv.SecondOwnerToken, models.Hub{
			Id:      "h1",
			Name:    "h1",
			OwnerId: testenv.SecendOwnerTokenUser,
		}, client.HubUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = sourceClient.SetHub(testenv.TestToken, models.Hub{
			Id:      "shouldNotBeFound2",
			Name:    "shouldNotBeFound2",
			OwnerId: testenv.TestTokenUser,
		}, client.HubUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}

		_, err, _ = sourceClient.SetDeviceGroup(testenv.SecondOwnerToken, models.DeviceGroup{
			Name: "dg1",
		})
		if err != nil {
			t.Error(err)
			return
		}

		_, err, _ = sourceClient.SetDeviceGroup(testenv.TestToken, models.DeviceGroup{
			Name: "shouldNotBeFound3",
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("start mirror", func(t *testing.T) {
		err = lib.Start(ctx, wg, config)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check initial mirror state", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		t.Run("protocols", func(t *testing.T) {
			result, err, _ := mirrorClient.ListProtocols("", 100, 0, "")
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 1 {
				t.Error("unexpected result: ", len(result))
				return
			}
		})
		t.Run("aspects", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListAspects(client.AspectListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 1 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 1 {
				t.Error("unexpected total: ", total)
			}
		})
		t.Run("aspect-nodes", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListAspectNodes(client.AspectListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 2 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
				return
			}
		})
		t.Run("devices", func(t *testing.T) {
			result, err, _ := mirrorClient.ListDevices("", client.DeviceListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 1 {
				t.Errorf("unexpected result: len=%v,\nresult=%#v", len(result), result)
				resources, err, _ := client2.New(sourceConfig.PermissionsV2Url).ListResourcesWithAdminPermission(client.InternalAdminToken, config.DeviceTopic, client2.ListOptions{})
				t.Errorf("permissions result: %v\n%#v", err, resources)
				return
			}
		})
		t.Run("device-groups", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListDeviceGroups("", client.DeviceGroupListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 2 {
				t.Errorf("unexpected result: len=%v,\nresult=%#v", len(result), result)
				resources, err, _ := client2.New(sourceConfig.PermissionsV2Url).ListResourcesWithAdminPermission(client.InternalAdminToken, config.DeviceGroupTopic, client2.ListOptions{})
				t.Errorf("permissions result: %v\n%#v", err, resources)
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
			}
		})
		t.Run("hubs", func(t *testing.T) {
			result, err, _ := mirrorClient.ListHubs("", client.HubListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 1 {
				t.Errorf("unexpected result: len=%v,\nresult=%#v", len(result), result)
				return
			}
		})
		t.Run("concepts", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListConcepts(client.ConceptListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 1 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 1 {
				t.Error("unexpected total: ", total)
			}
		})
		t.Run("functions", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListFunctions(client.FunctionListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			//the explicit one, plus the Get and Set of the one concept the source has
			if len(result) != 3 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 3 {
				t.Error("unexpected total: ", total)
				return
			}
		})
		t.Run("characteristics", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListCharacteristics(client.CharacteristicListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 1 {
				t.Error("unexpected result: ", len(result))
			}
			if total != 1 {
				t.Error("unexpected total: ", total)
			}
		})
		t.Run("device-types", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListDeviceTypesV3("", client.DeviceTypeListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 1 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 1 {
				t.Error("unexpected total: ", total)
				return
			}
		})

	})

	t.Run("update source state", func(t *testing.T) {
		p2, err, _ := sourceClient.SetProtocol(client.InternalAdminToken, models.Protocol{
			Id:      "p2",
			Name:    "p2",
			Handler: "p2",
			ProtocolSegments: []models.ProtocolSegment{{
				Name: "ps2",
			}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		a2, err, _ := sourceClient.SetAspect(client.InternalAdminToken, models.Aspect{
			Name:       "a2",
			SubAspects: []models.Aspect{{Name: "a2.2"}},
		})
		if err != nil {
			t.Error(err)
			return
		}
		c2, err, _ := sourceClient.SetCharacteristic(client.InternalAdminToken, models.Characteristic{
			Name: "c2",
			Type: models.String,
		})
		if err != nil {
			t.Error(err)
			return
		}
		co2, err, _ := sourceClient.SetConcept(client.InternalAdminToken, models.Concept{
			Name:                 "co2",
			BaseCharacteristicId: c2.Id,
			CharacteristicIds:    []string{c2.Id},
		})
		if err != nil {
			t.Error(err)
			return
		}
		f2, err, _ := sourceClient.SetFunction(client.InternalAdminToken, models.Function{
			Name:        "f2",
			DisplayName: "f2",
			ConceptId:   co2.Id,
			RdfType:     models.SES_ONTOLOGY_MEASURING_FUNCTION,
		})
		if err != nil {
			t.Error(err)
			return
		}

		dc2, err, _ := sourceClient.SetDeviceClass(client.InternalAdminToken, models.DeviceClass{
			Name: "dc2",
		})

		dt2, err, _ := sourceClient.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Name: "dt2",
			Services: []models.Service{
				{
					LocalId:     "s2",
					Name:        "s2",
					Interaction: models.EVENT_AND_REQUEST,
					ProtocolId:  "p2",
					Outputs: []models.Content{
						{
							ContentVariable: models.ContentVariable{
								Name:             "cv2",
								Type:             models.String,
								CharacteristicId: c2.Id,
								FunctionId:       f2.Id,
								AspectId:         a2.Id,
							},
							Serialization:     models.JSON,
							ProtocolSegmentId: p2.ProtocolSegments[0].Id,
						},
					},
				},
			},
			DeviceClassId: dc2.Id,
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}

		device, err, _ := sourceClient.CreateDevice(testenv.SecondOwnerToken, models.Device{
			LocalId:      "d2",
			Name:         "d2",
			DeviceTypeId: dt2.Id,
			OwnerId:      testenv.SecendOwnerTokenUser,
		})
		if err != nil {
			t.Error(err)
			return
		}
		t.Logf("device: %#v", device)

		_, err, _ = sourceClient.CreateDevice(testenv.TestToken, models.Device{
			LocalId:      "shouldNotBeFound_2",
			Name:         "shouldNotBeFound_2",
			DeviceTypeId: dt2.Id,
			OwnerId:      testenv.TestTokenUser,
		})
		if err != nil {
			t.Error(err)
			return
		}

		_, err, _ = sourceClient.SetHub(testenv.SecondOwnerToken, models.Hub{
			Id:      "h2",
			Name:    "h2",
			OwnerId: testenv.SecendOwnerTokenUser,
		}, client.HubUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = sourceClient.SetHub(testenv.TestToken, models.Hub{
			Id:      "shouldNotBeFound2_2",
			Name:    "shouldNotBeFound2_2",
			OwnerId: testenv.TestTokenUser,
		}, client.HubUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}

		_, err, _ = sourceClient.SetDeviceGroup(testenv.SecondOwnerToken, models.DeviceGroup{
			Name: "dg2",
		})
		if err != nil {
			t.Error(err)
			return
		}

		_, err, _ = sourceClient.SetDeviceGroup(testenv.TestToken, models.DeviceGroup{
			Name: "shouldNotBeFound3_2",
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("check updated mirror", func(t *testing.T) {
		time.Sleep(15 * time.Second)
		t.Run("protocols", func(t *testing.T) {
			result, err, _ := mirrorClient.ListProtocols("", 100, 0, "")
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 2 {
				t.Error("unexpected result: ", len(result))
				return
			}
		})
		t.Run("aspects", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListAspects(client.AspectListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 2 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
			}
		})
		t.Run("aspect-nodes", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListAspectNodes(client.AspectListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 4 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 4 {
				t.Error("unexpected total: ", total)
				return
			}
		})
		t.Run("devices", func(t *testing.T) {
			result, err, _ := mirrorClient.ListDevices("", client.DeviceListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 2 {
				t.Error("unexpected result: ", len(result))
				return
			}
		})
		t.Run("device-groups", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListDeviceGroups("", client.DeviceGroupListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 4 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 4 {
				t.Error("unexpected total: ", total)
			}
		})
		t.Run("hubs", func(t *testing.T) {
			result, err, _ := mirrorClient.ListHubs("", client.HubListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 2 {
				t.Error("unexpected result: ", len(result))
				return
			}
		})
		t.Run("concepts", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListConcepts(client.ConceptListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 2 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
			}
		})
		t.Run("functions", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListFunctions(client.FunctionListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			//the two explicit ones, plus the Get and Set of the two concepts
			if len(result) != 6 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 6 {
				t.Error("unexpected total: ", total)
				return
			}
		})
		t.Run("characteristics", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListCharacteristics(client.CharacteristicListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 2 {
				t.Error("unexpected result: ", len(result))
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
			}
		})
		t.Run("device-types", func(t *testing.T) {
			result, total, err, _ := mirrorClient.ListDeviceTypesV3("", client.DeviceTypeListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(result) != 2 {
				t.Error("unexpected result: ", len(result))
				return
			}
			if total != 2 {
				t.Error("unexpected total: ", total)
				return
			}
		})

	})

	t.Run("check limited proxy requests without changes", func(t *testing.T) {
		proxyReqCount.Store(0)
		t.Run("update other user", func(t *testing.T) {
			dts, _, err, _ := sourceClient.ListDeviceTypesV3("", client.DeviceTypeListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			if len(dts) == 0 {
				t.Error("missing device types")
				return
			}
			_, err, _ = sourceClient.CreateDevice(testenv.TestToken, models.Device{
				LocalId:      "shouldNotBeFound_3",
				Name:         "shouldNotBeFound_3",
				DeviceTypeId: dts[0].Id,
				OwnerId:      testenv.TestTokenUser,
			})
			if err != nil {
				t.Error(err)
				return
			}

			_, err, _ = sourceClient.SetHub(testenv.TestToken, models.Hub{
				Id:      "shouldNotBeFound_32",
				Name:    "shouldNotBeFound_32",
				OwnerId: testenv.TestTokenUser,
			}, client.HubUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}

			_, err, _ = sourceClient.SetDeviceGroup(testenv.TestToken, models.DeviceGroup{
				Name: "shouldNotBeFound_33",
			})
			if err != nil {
				t.Error(err)
				return
			}
		})
		t.Run("wait", func(t *testing.T) {
			time.Sleep(15 * time.Second)
		})
		t.Run("check proxy requests", func(t *testing.T) {
			if proxyReqCount.Load() > 2 || proxyReqCount.Load() < 1 {
				t.Error("unexpected request count: ", proxyReqCount.Load(), " (expected: 2 or 1 for timestamp requests)")
			}
		})
	})

	t.Run("forward updates", func(t *testing.T) {
		t.Run("create resources via mirror", func(t *testing.T) {
			protocols, err, _ := mirrorClient.ListProtocols("", 100, 0, "")
			if err != nil {
				t.Error(err)
				return
			}
			if len(protocols) == 0 {
				t.Error("missing protocols")
				return
			}
			p3 := protocols[0]

			dt3, err, _ := mirrorClient.SetDeviceType("", models.DeviceType{
				Name: "dt3",
				Services: []models.Service{
					{
						LocalId:     "s3",
						Name:        "s3",
						Interaction: models.EVENT_AND_REQUEST,
						ProtocolId:  p3.Id,
						Outputs: []models.Content{
							{
								ContentVariable: models.ContentVariable{
									Name: "cv3",
									Type: models.String,
								},
								Serialization:     models.JSON,
								ProtocolSegmentId: p3.ProtocolSegments[0].Id,
							},
						},
					},
				},
			}, client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}

			t.Run("create device", func(t *testing.T) {
				_, err, _ = mirrorClient.CreateDevice("", models.Device{
					LocalId:      "d3",
					Name:         "d3",
					DeviceTypeId: dt3.Id,
				})
				if err != nil {
					t.Error(err)
					return
				}
			})

			t.Run("create hub", func(t *testing.T) {
				_, err, _ = mirrorClient.SetHub("", models.Hub{
					Id:   "h3",
					Name: "h3",
				}, client.HubUpdateOptions{})
				if err != nil {
					t.Error(err)
					return
				}
			})

			t.Run("create device-group", func(t *testing.T) {
				_, err, _ = mirrorClient.SetDeviceGroup("", models.DeviceGroup{
					Name: "dg3",
				})
				if err != nil {
					t.Error(err)
					return
				}
			})

		})

		t.Run("check mirror", func(t *testing.T) {
			t.Run("devices", func(t *testing.T) {
				result, err, _ := mirrorClient.ListDevices("", client.DeviceListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if len(result) != 3 {
					t.Error("unexpected result: ", len(result))
					return
				}
			})
			t.Run("device-groups", func(t *testing.T) {
				result, total, err, _ := mirrorClient.ListDeviceGroups("", client.DeviceGroupListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if len(result) != 6 {
					t.Errorf("unexpected result: len=%v\n%#v", len(result), result)
					return
				}
				if total != 6 {
					t.Error("unexpected total: ", total)
				}
			})
			t.Run("hubs", func(t *testing.T) {
				result, err, _ := mirrorClient.ListHubs("", client.HubListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if len(result) != 3 {
					t.Error("unexpected result: ", len(result))
					return
				}
			})
			t.Run("device-types", func(t *testing.T) {
				result, total, err, _ := mirrorClient.ListDeviceTypesV3("", client.DeviceTypeListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if len(result) != 3 {
					t.Error("unexpected result: ", len(result))
					return
				}
				if total != 3 {
					t.Error("unexpected total: ", total)
					return
				}
			})
		})
		t.Run("check source", func(t *testing.T) {
			t.Run("devices", func(t *testing.T) {
				result, err, _ := sourceClient.ListDevices(testenv.SecondOwnerToken, client.DeviceListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if len(result) != 3 {
					t.Error("unexpected result: ", len(result))
					return
				}
			})
			t.Run("device-groups", func(t *testing.T) {
				result, total, err, _ := sourceClient.ListDeviceGroups(testenv.SecondOwnerToken, client.DeviceGroupListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if len(result) != 6 {
					t.Error("unexpected result: ", len(result))
					return
				}
				if total != 6 {
					t.Error("unexpected total: ", total)
				}
			})
			t.Run("hubs", func(t *testing.T) {
				result, err, _ := sourceClient.ListHubs(testenv.SecondOwnerToken, client.HubListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if len(result) != 3 {
					t.Error("unexpected result: ", len(result))
					return
				}
			})
			t.Run("device-types", func(t *testing.T) {
				result, total, err, _ := sourceClient.ListDeviceTypesV3(testenv.SecondOwnerToken, client.DeviceTypeListOptions{})
				if err != nil {
					t.Error(err)
					return
				}
				if len(result) != 3 {
					t.Error("unexpected result: ", len(result))
					return
				}
				if total != 3 {
					t.Error("unexpected total: ", total)
					return
				}
			})
		})
	})

}
