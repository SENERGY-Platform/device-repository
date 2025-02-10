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

package repo_legacy

import (
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"testing"
)

func testListDeviceClasses(t *testing.T, conf configuration.Config) {
	deviceClasses := []models.DeviceClass{
		{
			Id:   "c1",
			Name: "c1",
		},
		{
			Id:   "c2",
			Name: "c2",
		},
		{
			Id:   "c3",
			Name: "c3",
		},
		{
			Id:   "c4",
			Name: "c4",
		},
		{
			Id:   "c5",
			Name: "c5",
		},
	}

	c := client.NewClient("http://localhost:"+conf.ServerPort, nil)

	t.Run("create device-classes", func(t *testing.T) {
		for _, dc := range deviceClasses {
			_, err, _ := c.SetDeviceClass(AdminToken, dc)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("list all device-classes", func(t *testing.T) {
		list, total, err, _ := c.ListDeviceClasses(client.DeviceClassListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 5 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, deviceClasses) {
			t.Error(list)
			return
		}
	})

	t.Run("search c3 device-classes", func(t *testing.T) {
		list, total, err, _ := c.ListDeviceClasses(client.DeviceClassListOptions{Search: "c3"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 1 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.DeviceClass{deviceClasses[2]}) {
			t.Error(list)
			return
		}
	})

	t.Run("list c2,c4 device-classes", func(t *testing.T) {
		list, total, err, _ := c.ListDeviceClasses(client.DeviceClassListOptions{Ids: []string{"c2", "c4"}})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 2 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.DeviceClass{deviceClasses[1], deviceClasses[3]}) {
			t.Error(list)
			return
		}
	})
}
