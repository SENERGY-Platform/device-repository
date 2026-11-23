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

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// healthcheckUrl targets the local server; the container HEALTHCHECK runs inside the same network namespace.
func healthcheckUrl() string {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	return "http://127.0.0.1:" + port + "/health"
}

// runHealthcheck prints the body of the health response so that `docker inspect` records the reason of an unhealthy state,
// which busybox wget drops on an error status. It returns the process exit code.
func runHealthcheck(url string, out io.Writer) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(out, err)
		return 1
	}
	if text := strings.TrimSpace(string(body)); text != "" {
		fmt.Fprintln(out, text)
	}
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
