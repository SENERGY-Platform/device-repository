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
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

const startAttempts = 5

// startContainer starts the container and repeats the start if the docker daemon could not publish a port.
//
// testcontainers publishes every port of the image (EXPOSE) and of ExposedPorts on a host port the daemon picks
// from the ephemeral range. The rootless daemon does not check whether another socket of this network namespace
// already uses the port. The tests open many connections to the containers from the same namespace, so
// the pick collides now and then: "listen tcp4 0.0.0.0:<port>: bind: address already in use".
// The container is not running in this case and the next attempt gets another port.
// Every other error is returned unchanged.
func startContainer(ctx context.Context, req testcontainers.GenericContainerRequest) (c testcontainers.Container, err error) {
	for attempt := 1; ; attempt++ {
		c, err = testcontainers.GenericContainer(ctx, req)
		if err == nil || attempt >= startAttempts || !isPortPublishError(err) {
			return c, err
		}
		log.Printf("WARNING: docker could not publish a container port (attempt %d of %d), start again: %v", attempt, startAttempts, err)
		if c != nil {
			if terminateErr := c.Terminate(context.Background()); terminateErr != nil {
				log.Println("WARNING: unable to remove container of failed start:", terminateErr)
			}
		}
		select {
		case <-ctx.Done():
			return c, err
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func isPortPublishError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "driver failed programming external connectivity") &&
		strings.Contains(msg, "address already in use")
}
