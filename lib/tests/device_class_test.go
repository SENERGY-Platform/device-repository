/*
 * Copyright 2025 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestListControllingDeviceClasses(t *testing.T) {
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
	config.DisableStrictValidationForTesting = true

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

	t.Run("create device-classes", func(t *testing.T) {
		_, err, _ := c.SetDeviceClass(client.InternalAdminToken, models.DeviceClass{
			Id:   "a",
			Name: "a",
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetDeviceClass(client.InternalAdminToken, models.DeviceClass{
			Id:   "b",
			Name: "b",
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetDeviceClass(client.InternalAdminToken, models.DeviceClass{
			Id:   "c",
			Name: "c",
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("create functions", func(t *testing.T) {
		_, err, _ := c.SetFunction(client.InternalAdminToken, models.Function{
			Id:          models.URN_PREFIX + "controlling-function:f1",
			Name:        "f1",
			DisplayName: "f1",
			RdfType:     models.SES_ONTOLOGY_CONTROLLING_FUNCTION,
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetFunction(client.InternalAdminToken, models.Function{
			Id:          models.URN_PREFIX + "measuring-function:f2",
			Name:        "f2",
			DisplayName: "f2",
			RdfType:     models.SES_ONTOLOGY_MEASURING_FUNCTION,
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("create protocols", func(t *testing.T) {
		_, err, _ := c.SetProtocol(client.InternalAdminToken, models.Protocol{
			Id:      "p1",
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

	t.Run("create aspect", func(t *testing.T) {
		_, err, _ := c.SetAspect(client.InternalAdminToken, models.Aspect{
			Id:   "a1",
			Name: "a1",
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("create device-types", func(t *testing.T) {
		_, err, _ := c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Id:   "dt1",
			Name: "dt1",
			Services: []models.Service{
				{
					Id:          "s1",
					LocalId:     "s1",
					Name:        "s1",
					Interaction: models.REQUEST,
					ProtocolId:  "p1",
					Inputs: []models.Content{{
						Id: "dt1s1c1",
						ContentVariable: models.ContentVariable{
							Id:         "dt1s1c1cv1",
							Name:       "val",
							Type:       models.String,
							FunctionId: models.URN_PREFIX + "controlling-function:f1",
							AspectId:   "a1",
						},
						Serialization:     models.JSON,
						ProtocolSegmentId: "ps1",
					}},
				},
			},
			DeviceClassId: "a",
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}

		_, err, _ = c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
			Id:   "dt2",
			Name: "dt2",
			Services: []models.Service{
				{
					Id:          "s2",
					LocalId:     "s2",
					Name:        "s2",
					Interaction: models.REQUEST,
					ProtocolId:  "p1",
					Inputs: []models.Content{{
						Id: "dt2s2c1",
						ContentVariable: models.ContentVariable{
							Id:         "dt2s2c1cv1",
							Name:       "val",
							Type:       models.String,
							FunctionId: models.URN_PREFIX + "measuring-function:f2",
							AspectId:   "a1",
						},
						Serialization:     models.JSON,
						ProtocolSegmentId: "ps1",
					}},
				},
			},
			DeviceClassId: "b",
		}, client.DeviceTypeUpdateOptions{})
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("list controlling functions deprecated", func(t *testing.T) {
		//req, err := http.NewRequest(http.MethodGet, "http://localhost:"+config.ServerPort+"/device-classes?function=controlling-function", nil)
		resp, err := http.Get("http://localhost:" + config.ServerPort + "/device-classes?function=controlling-function")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			tmp, _ := io.ReadAll(resp.Body)
			t.Error("unexpected status code", resp.StatusCode, string(tmp))
			return
		}
		var dcList []models.DeviceClass
		err = json.NewDecoder(resp.Body).Decode(&dcList)
		if err != nil {
			t.Error(err)
			return
		}
		if len(dcList) != 1 {
			t.Error("unexpected device-class list", dcList)
			return
		}
		if dcList[0].Id != "a" {
			t.Error("unexpected device-class list", dcList)
			return
		}
	})

}
