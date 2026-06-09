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

package docker

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func DeviceRepo(ctx context.Context, wg *sync.WaitGroup, relativePathToRoot string, kafkaUrl string, mongoUrl string, permv2Url string) (hostPort string, ipAddress string, err error) {
	log.Println("start device-repository")
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:        relativePathToRoot,
				Repo:           "ghcr.io/senergy-platform/device-repository",
				Tag:            "dev",
				BuildLogWriter: os.Stdout,
			},
			//Image:          "ghcr.io/senergy-platform/device-repository:dev",
			Env: map[string]string{
				"DEBUG":              "true",
				"KAFKA_URL":          kafkaUrl,
				"PERMISSIONS_V2_URL": permv2Url,
				"MONGO_URL":          mongoUrl,
			},
			ExposedPorts:    []string{"8080/tcp"},
			WaitingFor:      wait.ForListeningPort("8080/tcp"),
			AlwaysPullImage: true,
			LogConsumerCfg:  &testcontainers.LogConsumerConfig{Consumers: []testcontainers.LogConsumer{LogConsumer{Prefix: "DEVICE-REPOSITORY: "}}},
		},
		Started: true,
	})
	if err != nil {
		return "", "", err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		log.Println("DEBUG: remove container device-repository", c.Terminate(context.Background()))
	}()

	ipAddress, err = c.ContainerIP(ctx)
	if err != nil {
		return "", "", err
	}
	temp, err := c.MappedPort(ctx, "8080/tcp")
	if err != nil {
		return "", "", err
	}
	hostPort = temp.Port()

	return hostPort, ipAddress, err
}
