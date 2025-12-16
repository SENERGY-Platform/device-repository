/*
 * Copyright 2021 InfAI (CC SES)
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

package docker

import (
	"context"
	"errors"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller/publisher"
)

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func retry(timeout time.Duration, f func() error) (err error) {
	err = errors.New("initial")
	start := time.Now()
	for i := int64(1); err != nil && time.Since(start) < timeout; i++ {
		err = f()
		if err != nil {
			log.Println("ERROR: :", err)
			wait := time.Duration(i) * time.Second
			if time.Since(start)+wait < timeout {
				log.Println("ERROR: retry after:", wait.String())
				time.Sleep(wait)
			} else {
				time.Sleep(time.Since(start) + wait - timeout)
				return f()
			}
		}
	}
	return err
}

func NewEnv(baseCtx context.Context, wg *sync.WaitGroup, startConfig configuration.Config) (config configuration.Config, err error) {
	config = startConfig
	ctx, cancel := context.WithCancel(baseCtx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	whPort, err := GetFreePort()
	if err != nil {
		log.Println("unable to find free port", err)
		return config, err
	}
	config.ServerPort = strconv.Itoa(whPort)

	_, ip, err := MongoDB(ctx, wg)
	if err != nil {
		return config, err
	}
	config.MongoUrl = "mongodb://" + ip + ":27017"

	config.KafkaUrl, err = Kafka(ctx, wg)
	if err != nil {
		return config, err
	}

	err = publisher.InitTopic(config.KafkaUrl,
		"concepts",
		"device-groups",
		"aspects",
		"characteristics",
		"processmodel",
		"device-types",
		"hubs",
		"devices",
		"device-classes",
		"functions",
		"protocols",
		"import-types",
		"locations",
		"smart_service_releases",
		"gateway_log",
		"device_log")
	if err != nil {
		return config, err
	}

	_, permV2Ip, err := PermissionsV2(ctx, wg, config.MongoUrl, config.KafkaUrl)
	if err != nil {
		return config, err
	}
	config.PermissionsV2Url = "http://" + permV2Ip + ":8080"

	time.Sleep(5 * time.Second)

	return
}
