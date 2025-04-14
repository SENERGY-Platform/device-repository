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
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/go-service-base/struct-logger"
	"github.com/SENERGY-Platform/go-service-base/struct-logger/attributes"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"log/slog"
	"os"
	"time"
)

func New(config configuration.Config, db database.Database, p Publisher, permClient client.Client) (ctrl *Controller, err error) {
	if permClient == nil {
		permClient = client.New(config.PermissionsV2Url)
	}
	ctrl = &Controller{
		publisher:           p,
		db:                  db,
		config:              config,
		permissionsV2Client: permClient,
		logger:              getLogger(config),
	}
	if permClient != nil {
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
		_, err, _ = ctrl.permissionsV2Client.SetTopic(client.InternalAdminToken, client.Topic{
			Id:                  config.LocationTopic,
			PublishToKafkaTopic: config.LocationTopic,
		})
		if err != nil {
			return nil, err
		}
	}
	return
}

func getLogger(config configuration.Config) *slog.Logger {
	options := &slog.HandlerOptions{
		AddSource: false,
		Level:     struct_logger.GetLevel(config.StructLoggerLogLevel, slog.LevelWarn),
	}
	handler := struct_logger.GetHandler(config.StructLoggerLogHandler, os.Stdout, options, slog.Default().Handler())
	handler = handler.WithAttrs([]slog.Attr{
		slog.String(attributes.Provider.ProjectKey(), "device-repository"),
	})
	return slog.New(handler)
}

type Controller struct {
	publisher           Publisher
	db                  database.Database
	config              configuration.Config
	permissionsV2Client client.Client
	logger              *slog.Logger
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
