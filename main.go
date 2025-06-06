/*
 * Copyright 2019 InfAI (CC SES)
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
	"flag"
	"github.com/SENERGY-Platform/api-docs-provider/lib/client"
	"github.com/SENERGY-Platform/device-repository/docs"
	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	configLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	conf, err := configuration.Load(*configLocation)
	if err != nil {
		log.Fatal("ERROR: unable to load config", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	err = lib.Start(ctx, wg, conf)
	if err != nil {
		log.Fatal(err)
	}

	if conf.ApiDocsProviderBaseUrl != "" && conf.ApiDocsProviderBaseUrl != "-" {
		err = PublishAsyncApiDoc(conf)
		if err != nil {
			log.Fatal(err)
		}
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-shutdown
	log.Println("received shutdown signal", sig)
	cancel()
	wg.Wait() //wait for clean disconnects
}

func PublishAsyncApiDoc(conf configuration.Config) error {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return client.New(http.DefaultClient, conf.ApiDocsProviderBaseUrl).AsyncapiPutDoc(ctx, "github_com_SENERGY-Platform_device-repository", docs.AsyncApiDoc)
}
