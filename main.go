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
	"flag"
	"github.com/SENERGY-Platform/device-repository/lib/api"
	"github.com/SENERGY-Platform/device-repository/lib/com"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/source"
	"github.com/SENERGY-Platform/device-repository/lib/source/publisher"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	conf, err := config.Load(*configLocation)
	if err != nil {
		log.Fatal("ERROR: unable to load config", err)
	}

	db, err := database.New(conf)
	if err != nil {
		log.Fatal("ERROR: unable to connect to database", err)
	}

	perm, err := com.NewSecurity(conf)
	if err != nil {
		log.Fatal("ERROR: unable to create permission handler", err)
	}

	ctrl, err := controller.New(conf, db, perm, func(ctrl *controller.Controller) (controller.Publisher, error) {
		conn, err := source.Start(conf, ctrl)
		if err != nil {
			log.Println("ERROR: unable to start source", err)
			return nil, err
		}
		//return publisher.New(conn, conf)	//TODO: use when old iot-repo is updated
		return publisher.NewMute(conn, conf)
	})

	if err != nil {
		log.Fatal("ERROR: unable to start control", err)
	}

	err = api.Start(conf, ctrl)
	if err != nil {
		log.Fatal("ERROR: unable to start api", err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	sig := <-shutdown
	log.Println("received shutdown signal", sig)
}
