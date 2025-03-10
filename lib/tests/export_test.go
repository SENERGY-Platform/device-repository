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
	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestExport(t *testing.T) {
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

	_, mip, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MongoUrl = "mongodb://" + mip + ":27017"

	_, zkIp, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	config.KafkaUrl, err = docker.Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}

	_, permIp, err := docker.PermissionsV2(ctx, wg, config.MongoUrl, config.KafkaUrl)
	if err != nil {
		t.Error(err)
		return
	}
	config.PermissionsV2Url = "http://" + permIp + ":8080"

	freePort, err := docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	config.ServerPort = strconv.Itoa(freePort)

	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(time.Second)

	c := client.NewClient("http://localhost:"+config.ServerPort, nil)

	protocolId := ""
	protocols := []models.Protocol{}
	concepts := []models.Concept{}
	functions := []models.Function{}
	aspects := []models.Aspect{}
	deviceTypes := []models.DeviceType{}
	devices := []models.Device{}
	devicegroups := []models.DeviceGroup{}
	t.Run("create", func(t *testing.T) {
		t.Run("protocol", func(t *testing.T) {
			result, err, _ := c.SetProtocol(client.InternalAdminToken, models.Protocol{
				Name:    "test-protocol",
				Handler: "test-handler",
				ProtocolSegments: []models.ProtocolSegment{
					{
						Id:   "pl",
						Name: "pl",
					},
				},
				Constraints: nil,
			})
			if err != nil {
				t.Error(err)
				return
			}
			protocolId = result.Id
			protocols = append(protocols, result)

			result, err, _ = c.SetProtocol(client.InternalAdminToken, models.Protocol{
				Name:    "test-protocol-2",
				Handler: "test-handler-2",
				ProtocolSegments: []models.ProtocolSegment{
					{
						Id:   "pl",
						Name: "pl",
					},
				},
				Constraints: nil,
			})
			if err != nil {
				t.Error(err)
				return
			}
			protocols = append(protocols, result)
		})

		t.Run("concepts", func(t *testing.T) {
			temp, err, _ := c.SetConcept(client.InternalAdminToken, models.Concept{
				Name: "test-concept",
			})
			if err != nil {
				t.Error(err)
				return
			}
			concepts = append(concepts, temp)
			temp, err, _ = c.SetConcept(client.InternalAdminToken, models.Concept{
				Name: "test-concept-2",
			})
			if err != nil {
				t.Error(err)
				return
			}
			concepts = append(concepts, temp)
		})

		t.Run("functions", func(t *testing.T) {
			temp, err, _ := c.SetFunction(client.InternalAdminToken, models.Function{
				Name:      "test-function",
				ConceptId: concepts[0].Id,
				RdfType:   models.SES_ONTOLOGY_CONTROLLING_FUNCTION,
			})
			if err != nil {
				t.Error(err)
				return
			}
			functions = append(functions, temp)
			temp, err, _ = c.SetFunction(client.InternalAdminToken, models.Function{
				Name:      "test-function-2",
				ConceptId: concepts[0].Id,
				RdfType:   models.SES_ONTOLOGY_CONTROLLING_FUNCTION,
			})
			if err != nil {
				t.Error(err)
				return
			}
			functions = append(functions, temp)
		})

		t.Run("aspects", func(t *testing.T) {
			temp, err, _ := c.SetAspect(client.InternalAdminToken, models.Aspect{
				Name: "test-aspect",
				SubAspects: []models.Aspect{
					{Name: "sub1"},
					{Name: "sub2"},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}
			aspects = append(aspects, temp)
			temp, err, _ = c.SetAspect(client.InternalAdminToken, models.Aspect{
				Name: "test-aspect-2",
			})
			if err != nil {
				t.Error(err)
				return
			}
			aspects = append(aspects, temp)
		})

		t.Run("device-types", func(t *testing.T) {
			temp, err, _ := c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
				Name: "test-dt",
				Services: []models.Service{
					{
						LocalId:     "s1",
						Name:        "s1",
						Interaction: models.REQUEST,
						ProtocolId:  protocolId,
					},
				},
			}, client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			deviceTypes = append(deviceTypes, temp)
			temp, err, _ = c.SetDeviceType(client.InternalAdminToken, models.DeviceType{
				Name: "test-dt-2",
				Services: []models.Service{
					{
						LocalId:     "s2",
						Name:        "s2",
						Interaction: models.REQUEST,
						ProtocolId:  protocolId,
					},
				},
			}, client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			deviceTypes = append(deviceTypes, temp)
		})

		t.Run("devices", func(t *testing.T) {
			temp, err, _ := c.CreateDevice(client.InternalAdminToken, models.Device{
				LocalId:      "d1",
				Name:         "d1",
				Attributes:   nil,
				DeviceTypeId: deviceTypes[0].Id,
			})
			if err != nil {
				t.Error(err)
				return
			}
			devices = append(devices, temp)
			temp, err, _ = c.CreateDevice(client.InternalAdminToken, models.Device{
				LocalId:      "d2",
				Name:         "d2",
				Attributes:   nil,
				DeviceTypeId: deviceTypes[0].Id,
			})
			if err != nil {
				t.Error(err)
				return
			}
			devices = append(devices, temp)
		})
	})

	t.Run("get generated device-groups", func(t *testing.T) {
		temp, _, err, _ := c.ListDeviceGroups(client.InternalAdminToken, client.DeviceGroupListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		devicegroups = temp
	})

	t.Run("export", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			export, err, _ := c.Export(client.InternalAdminToken, client.ImportExportOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			expected := client.ImportExport{
				Protocols:   protocols,
				Functions:   functions,
				Aspects:     aspects,
				Concepts:    concepts,
				DeviceTypes: deviceTypes,
			}
			expected.Sort()
			if !reflect.DeepEqual(export, expected) {
				t.Errorf("\na:%#v\ne:%#v\n", export, expected)
			}
		})
		t.Run("with devices", func(t *testing.T) {
			export, err, _ := c.Export(client.InternalAdminToken, client.ImportExportOptions{IncludeOwnedInformation: true})
			if err != nil {
				t.Error(err)
				return
			}
			expected := client.ImportExport{
				Protocols:    protocols,
				Functions:    functions,
				Aspects:      aspects,
				Concepts:     concepts,
				DeviceTypes:  deviceTypes,
				Devices:      devices,
				DeviceGroups: devicegroups,
				Permissions: []client.Resource{
					{
						Id:                  model.DeviceIdToGeneratedDeviceGroupId(devices[0].Id),
						TopicId:             "device-groups",
						ResourcePermissions: controller.GetDefaultEntryPermissions(config, "device-groups", "dd69ea0d-f553-4336-80f3-7f4567f85c7b").ToPermV2Permissions(),
					},
					{
						Id:                  model.DeviceIdToGeneratedDeviceGroupId(devices[1].Id),
						TopicId:             "device-groups",
						ResourcePermissions: controller.GetDefaultEntryPermissions(config, "device-groups", "dd69ea0d-f553-4336-80f3-7f4567f85c7b").ToPermV2Permissions(),
					},
					{
						Id:                  devices[0].Id,
						TopicId:             "devices",
						ResourcePermissions: controller.GetDefaultEntryPermissions(config, "devices", "dd69ea0d-f553-4336-80f3-7f4567f85c7b").ToPermV2Permissions(),
					},
					{
						Id:                  devices[1].Id,
						TopicId:             "devices",
						ResourcePermissions: controller.GetDefaultEntryPermissions(config, "devices", "dd69ea0d-f553-4336-80f3-7f4567f85c7b").ToPermV2Permissions(),
					},
				},
			}
			expected.Sort()
			if !reflect.DeepEqual(export, expected) {
				t.Errorf("\na:%#v\ne:%#v\n", export, expected)
			}
		})
		t.Run("only device-types", func(t *testing.T) {
			export, err, _ := c.Export(client.InternalAdminToken, client.ImportExportOptions{FilterResourceTypes: []string{"device-types"}})
			if err != nil {
				t.Error(err)
				return
			}
			expected := client.ImportExport{
				DeviceTypes: deviceTypes,
			}
			expected.Sort()
			if !reflect.DeepEqual(export, expected) {
				t.Errorf("\na:%#v\ne:%#v\n", export, expected)
			}
		})
		t.Run("only first device-type", func(t *testing.T) {
			export, err, _ := c.Export(client.InternalAdminToken, client.ImportExportOptions{
				FilterResourceTypes: []string{"device-types"},
				FilterIds:           []string{deviceTypes[0].Id},
			})
			if err != nil {
				t.Error(err)
				return
			}
			expected := client.ImportExport{
				DeviceTypes: []models.DeviceType{deviceTypes[0]},
			}
			expected.Sort()
			if !reflect.DeepEqual(export, expected) {
				t.Errorf("\na:%#v\ne:%#v\n", export, expected)
			}
		})
		t.Run("only device-types and functions", func(t *testing.T) {
			export, err, _ := c.Export(client.InternalAdminToken, client.ImportExportOptions{FilterResourceTypes: []string{"device-types", "functions"}})
			if err != nil {
				t.Error(err)
				return
			}
			expected := client.ImportExport{
				DeviceTypes: deviceTypes,
				Functions:   functions,
			}
			expected.Sort()
			if !reflect.DeepEqual(export, expected) {
				t.Errorf("\na:%#v\ne:%#v\n", export, expected)
			}
		})
		t.Run("only first device-type and function", func(t *testing.T) {
			export, err, _ := c.Export(client.InternalAdminToken, client.ImportExportOptions{
				FilterResourceTypes: []string{"device-types", "functions"},
				FilterIds:           []string{deviceTypes[0].Id, functions[0].Id},
			})
			if err != nil {
				t.Error(err)
				return
			}
			expected := client.ImportExport{
				DeviceTypes: []models.DeviceType{deviceTypes[0]},
				Functions:   []models.Function{functions[0]},
			}
			expected.Sort()
			if !reflect.DeepEqual(export, expected) {
				t.Errorf("\na:%#v\ne:%#v\n", export, expected)
			}
		})
	})

}
