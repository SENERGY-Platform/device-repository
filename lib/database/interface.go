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
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
)

type Database interface {
	Disconnect()
	Transaction(ctx context.Context) (context.Context, func(success bool) error, error)
	CreateId() string
	GetDevice(ctx context.Context, id string) (device model.DeviceInstance, exists bool, err error)
	SetDevice(ctx context.Context, device model.DeviceInstance) error
	RemoveDevice(ctx context.Context, id string) error
	GetDeviceType(ctx context.Context, id string) (deviceType model.DeviceType, exists bool, err error)
	SetDeviceType(ctx context.Context, deviceType model.DeviceType) error
	ListDevicesOfDeviceType(ctx context.Context, deviceTypeId string, options ...listoptions.ListOptions) ([]model.DeviceInstance, error)
	RemoveDeviceType(ctx context.Context, id string) error
	ListEndpoints(ctx context.Context, listoptions ...listoptions.ListOptions) (result []model.Endpoint, err error)
	RemoveEndpoint(ctx context.Context, id string) error
	SetEndpoint(ctx context.Context, endpoint model.Endpoint) error
	GetHub(ctx context.Context, id string) (model.Hub, bool, error)
	SetHub(ctx context.Context, hub model.Hub) error
	RemoveHub(ctx context.Context, id string) error
	ListDevicesWithHub(ctx context.Context, id string, options ...listoptions.ListOptions) ([]model.DeviceInstance, error)
	GetValueType(ctx context.Context, id string) (model.ValueType, bool, error)
	SetValueType(ctx context.Context, valueType model.ValueType) error
	RemoveValueType(ctx context.Context, id string) error
	ListDeviceTypesUsingValueType(ctx context.Context, id string, options ...listoptions.ListOptions) ([]model.DeviceType, error)
	ListValueTypesUsingValueType(ctx context.Context, id string, options ...listoptions.ListOptions) ([]model.ValueType, error)
	GetDeviceByUri(ctx context.Context, uri string) (model.DeviceInstance, bool, error)
	ListDeviceTypes(ctx context.Context, options listoptions.ListOptions) (result []model.DeviceType, err error)
}
