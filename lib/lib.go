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

package lib

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/api"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/source/consumer"
	"github.com/SENERGY-Platform/device-repository/lib/source/producer"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"log"
	"sync"
)

// set wg if you want to wait for clean disconnects after ctx is done
func Start(baseCtx context.Context, wg *sync.WaitGroup, conf config.Config) (err error) {
	ctx, cancel := context.WithCancel(baseCtx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	db, err := database.New(conf)
	if err != nil {
		log.Println("ERROR: unable to connect to database", err)
		return err
	}
	if wg != nil {
		wg.Add(1)
	}
	go func() {
		<-ctx.Done()
		db.Disconnect()
		if wg != nil {
			wg.Done()
		}
	}()

	var p controller.Producer = controller.ErrorProducer{}
	if !conf.DisableKafkaConsumer {
		p, err = producer.New(conf)
		if err != nil {
			log.Println("ERROR: unable to create producer", err)
			return err
		}
	}

	permClient := client.New(conf.PermissionsV2Url)

	ctrl, err := controller.New(conf, db, p, permClient)
	if err != nil {
		db.Disconnect()
		log.Println("ERROR: unable to start control", err)
		return err
	}

	if conf.RunStartupMigrations && !conf.DisableKafkaConsumer {
		err = db.RunStartupMigrations(ctrl)
		if err != nil {
			db.Disconnect()
			log.Println("ERROR: RunStartupMigrations()", err)
			return err
		}
	}

	if !conf.DisableKafkaConsumer {
		err = consumer.Start(ctx, conf, ctrl)
		if err != nil {
			log.Println("ERROR: unable to start source", err)
			return err
		}
	}

	if !conf.DisableHttpApi {
		err = api.Start(ctx, conf, ctrl)
		if err != nil {
			log.Println("ERROR: unable to start api", err)
			return err
		}
	}

	return err
}
