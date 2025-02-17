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
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"net/http"
	"sync"
	"testing"
)

func TestResourceList(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := createTestEnv(ctx, wg, t)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("health", func(t *testing.T) {
		resp, err := http.Get("http://localhost:" + conf.ServerPort)
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			t.Error("unexpected status code", resp.StatusCode)
			return
		}
		resp, err = http.Get("http://localhost:" + conf.ServerPort + "/")
		if err != nil {
			t.Error(err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			t.Error("unexpected status code", resp.StatusCode)
			return
		}
	})

	t.Run("EOF /aspect-nodes/", func(t *testing.T) {
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, _, code := c.GetAspectNode("")
		if code != http.StatusNotFound {
			t.Error(err, code)
			return
		}
	})

	t.Run("aspects", func(t *testing.T) {
		testAspectList(t, conf)
	})

	t.Run("characteristics", func(t *testing.T) {
		testCharacteristicList(t, conf)
	})

	t.Run("concepts", func(t *testing.T) {
		testConceptList(t, conf)
	})

	t.Run("device-classes", func(t *testing.T) {
		testListDeviceClasses(t, conf)
	})
}
