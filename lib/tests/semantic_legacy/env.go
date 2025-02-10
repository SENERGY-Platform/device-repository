/*
 * Copyright 2022 InfAI (CC SES)
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

package semantic_legacy

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/api"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	docker2 "github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/device-repository/lib/tests/repo_legacy/testenv"
	permclient "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
)

func NewPartialMockEnv(baseCtx context.Context, wg *sync.WaitGroup, startConfig configuration.Config, t *testing.T) (config configuration.Config, ctrl *controller.Controller, err error) {
	config = startConfig
	config.DisableStrictValidationForTesting = true
	ctx, cancel := context.WithCancel(baseCtx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	whPort, err := docker2.GetFreePort()
	if err != nil {
		log.Println("unable to find free port", err)
		return config, ctrl, err
	}
	config.ServerPort = strconv.Itoa(whPort)

	_, ip, err := docker2.MongoDB(ctx, wg)
	if err != nil {
		return config, ctrl, err
	}
	config.MongoUrl = "mongodb://" + ip + ":27017"

	db, err := database.New(config)
	if err != nil {
		return config, ctrl, err
	}

	pc, err := permclient.NewTestClient(ctx)
	if err != nil {
		return config, ctrl, err
	}

	ctrl, err = controller.New(config, db, testenv.VoidProducerMock{}, pc)
	if err != nil {
		return config, ctrl, err
	}

	err = api.Start(ctx, config, ctrl)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(time.Second)

	return config, ctrl, err
}
