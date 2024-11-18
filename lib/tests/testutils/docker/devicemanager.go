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

package docker

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

func DeviceManager(ctx context.Context, wg *sync.WaitGroup, kafkaUrl string, permV2Url string, deviceRepoUrl string) (hostPort string, ipAddress string, err error) {
	log.Println("start device-manager")
	var hostAccessPorts []int
	if strings.Contains(deviceRepoUrl, testcontainers.HostInternal) {
		u, err := url.Parse(deviceRepoUrl)
		if err != nil {
			return "", "", err
		}
		addrPort, err := strconv.Atoi(u.Port())
		if err != nil {
			return "", "", err
		}
		hostAccessPorts = append(hostAccessPorts, addrPort)
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "ghcr.io/senergy-platform/device-manager:dev",
			Env: map[string]string{
				"DEVICE_REPO_URL":    deviceRepoUrl,
				"KAFKA_URL":          kafkaUrl,
				"PERMISSIONS_V2_URL": permV2Url,
			},
			HostAccessPorts: hostAccessPorts,
			ExposedPorts:    []string{"8080/tcp"},
			WaitingFor:      wait.ForListeningPort("8080/tcp"),
			AlwaysPullImage: true,
		},
		Started: true,
	})
	if err != nil {
		return "", "", err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			log.Println("DEBUG: remove container device-manager", c.Terminate(context.Background()))
		}()
		<-ctx.Done()
		reader, err := c.Logs(context.Background())
		if err != nil {
			log.Println("ERROR: unable to get container log")
			return
		}
		buf := new(strings.Builder)
		io.Copy(buf, reader)
		fmt.Println("DEVICE-MANAGER LOGS: ------------------------------------------")
		fmt.Println(buf.String())
		fmt.Println("\n---------------------------------------------------------------")
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
