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

package database

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/model"
)

type Database interface {
	Disconnect()
	Transaction(ctx context.Context) (context.Context, func(success bool) error, error)
	GetDeviceType(ctx context.Context, id string) (deviceType model.DeviceType, exists bool, err error)
	SetDeviceType(ctx context.Context, deviceType model.DeviceType) error
	RemoveDeviceType(ctx context.Context, id string) error
	ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string) (result []model.DeviceType, err error)
	GetProtocol(ctx context.Context, id string) (result model.Protocol, exists bool, err error)
	ListProtocols(ctx context.Context, limit int64, offset int64, sort string) ([]model.Protocol, error)
	SetProtocol(ctx context.Context, protocol model.Protocol) error
	RemoveProtocol(ctx context.Context, id string) error
}
