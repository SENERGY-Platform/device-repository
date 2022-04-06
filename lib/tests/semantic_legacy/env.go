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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/semantic_legacy/producer"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/ory/dockertest/v3"
	"log"
	"strconv"
	"sync"
	"testing"
)

func NewPartialMockEnv(baseCtx context.Context, wg *sync.WaitGroup, startConfig config.Config, t *testing.T) (config config.Config, ctrl *controller.Controller, prod *producer.Producer, err error) {
	config = startConfig
	config.FatalErrHandler = t.Fatal
	ctx, cancel := context.WithCancel(baseCtx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	whPort, err := docker.GetFreePort()
	if err != nil {
		log.Println("unable to find free port", err)
		return config, ctrl, prod, err
	}
	config.ServerPort = strconv.Itoa(whPort)

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Println("Could not connect to docker:", err)
		return config, ctrl, prod, err
	}

	_, ip, err := docker.MongoDB(pool, ctx, wg)
	if err != nil {
		return config, ctrl, prod, err
	}
	config.MongoUrl = "mongodb://" + ip + ":27017"

	db, err := database.New(config)
	if err != nil {
		return config, ctrl, prod, err
	}

	ctrl, err = controller.New(config, db, SecurityMock{}, VoidProducerMock{})
	if err != nil {
		return config, ctrl, prod, err
	}

	prod, err = producer.StartSourceMock(config, ctrl)
	if err != nil {
		return config, ctrl, prod, err
	}

	return config, ctrl, prod, err
}

type SecurityMock struct{}

func (s SecurityMock) CheckBool(token string, kind string, id string, action model.AuthAction) (allowed bool, err error) {
	return true, nil
}

func (s SecurityMock) CheckMultiple(token string, kind string, ids []string, action model.AuthAction) (result map[string]bool, err error) {
	result = map[string]bool{}
	for _, id := range ids {
		result[id] = true
	}
	return
}

type VoidProducerMock struct{}

func (v VoidProducerMock) PublishAspectDelete(id string, owner string) error {
	//TODO implement me
	panic("implement me")
}

func (v VoidProducerMock) PublishAspectUpdate(aspect model.Aspect, owner string) error {
	//TODO implement me
	panic("implement me")
}

func (v VoidProducerMock) PublishDeviceDelete(id string, owner string) error {
	//TODO implement me
	panic("implement me")
}

func (v VoidProducerMock) PublishHub(hub model.Hub) (err error) {
	//TODO implement me
	panic("implement me")
}
