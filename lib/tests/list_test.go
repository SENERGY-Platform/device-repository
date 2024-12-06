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
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils"
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
	producer, err := testutils.NewPublisher(conf)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("aspects", func(t *testing.T) {
		testAspectList(t, producer, conf)
	})

	t.Run("characteristics", func(t *testing.T) {
		testCharacteristicList(t, producer, conf)
	})

	t.Run("concepts", func(t *testing.T) {
		testConceptList(t, producer, conf)
	})

	t.Run("device-classes", func(t *testing.T) {
		testListDeviceClasses(t, producer, conf)
	})
}
