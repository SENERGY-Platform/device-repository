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

package controller

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"time"
)

func New(config config.Config, db database.Database, producer Producer) (ctrl *Controller, err error) {
	ctrl = &Controller{
		db:                  db,
		producer:            producer,
		config:              config,
		permissionsV2Client: client.New(config.PermissionsV2Url),
	}
	if config.PermissionsV2Url != "" && config.PermissionsV2Url != "-" {
		_, err, _ = ctrl.permissionsV2Client.SetTopic(client.InternalAdminToken, client.Topic{
			Id:                  config.DeviceTopic,
			PublishToKafkaTopic: config.DeviceTopic,
		})
		if err != nil {
			return nil, err
		}
		_, err, _ = ctrl.permissionsV2Client.SetTopic(client.InternalAdminToken, client.Topic{
			Id:                  config.HubTopic,
			PublishToKafkaTopic: config.HubTopic,
		})
		if err != nil {
			return nil, err
		}
		_, err, _ = ctrl.permissionsV2Client.SetTopic(client.InternalAdminToken, client.Topic{
			Id:                  config.DeviceGroupTopic,
			PublishToKafkaTopic: config.DeviceGroupTopic,
		})
		if err != nil {
			return nil, err
		}
	}
	return
}

type Controller struct {
	db                  database.Database
	producer            Producer
	config              config.Config
	permissionsV2Client client.Client
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
