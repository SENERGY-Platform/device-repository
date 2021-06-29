/*
 * Copyright 2021 InfAI (CC SES)
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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"testing"
)

func TestValidateServiceInputOk(t *testing.T) {
	dbmock := DbMock{}
	ctrl := &Controller{db: dbmock}
	err, _ := ctrl.ValidateService(model.Service{
		Id:         "foo",
		LocalId:    "foo",
		Name:       "foo",
		ProtocolId: "foo",
		Inputs: []model.Content{
			{
				Id: "foo",
				ContentVariable: model.ContentVariable{
					Id:   "foo",
					Name: "foo",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo",
			},
			{
				Id: "bar",
				ContentVariable: model.ContentVariable{
					Id:   "bar",
					Name: "bar",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo2",
			},
		},
		Outputs: nil,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestValidateServiceInputErr(t *testing.T) {
	dbmock := DbMock{}
	ctrl := &Controller{db: dbmock}
	err, _ := ctrl.ValidateService(model.Service{
		Id:         "foo",
		LocalId:    "foo",
		Name:       "foo",
		ProtocolId: "foo",
		Inputs: []model.Content{
			{
				Id: "foo",
				ContentVariable: model.ContentVariable{
					Id:   "foo",
					Name: "foo",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo",
			},
			{
				Id: "foo",
				ContentVariable: model.ContentVariable{
					Id:   "foo",
					Name: "foo",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo2",
			},
		},
		Outputs: nil,
	})
	if err == nil {
		t.Error(errors.New("expected error"))
	}
}

func TestValidateServiceOutputOk(t *testing.T) {
	dbmock := DbMock{}
	ctrl := &Controller{db: dbmock}
	err, _ := ctrl.ValidateService(model.Service{
		Id:         "foo",
		LocalId:    "foo",
		Name:       "foo",
		ProtocolId: "foo",
		Outputs: []model.Content{
			{
				Id: "foo",
				ContentVariable: model.ContentVariable{
					Id:   "foo",
					Name: "foo",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo",
			},
			{
				Id: "bar",
				ContentVariable: model.ContentVariable{
					Id:   "bar",
					Name: "bar",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo2",
			},
		},
		Inputs: nil,
	})
	if err != nil {
		t.Error(err)
	}
}

func TestValidateServiceOutputErr(t *testing.T) {
	dbmock := DbMock{}
	ctrl := &Controller{db: dbmock}
	err, _ := ctrl.ValidateService(model.Service{
		Id:         "foo",
		LocalId:    "foo",
		Name:       "foo",
		ProtocolId: "foo",
		Outputs: []model.Content{
			{
				Id: "foo",
				ContentVariable: model.ContentVariable{
					Id:   "foo",
					Name: "foo",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo",
			},
			{
				Id: "foo",
				ContentVariable: model.ContentVariable{
					Id:   "foo",
					Name: "foo",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo2",
			},
		},
		Inputs: nil,
	})
	if err == nil {
		t.Error(errors.New("expected error"))
	}
}

func TestValidateServiceInputOutputOk(t *testing.T) {
	dbmock := DbMock{}
	ctrl := &Controller{db: dbmock}
	err, _ := ctrl.ValidateService(model.Service{
		Id:         "foo",
		LocalId:    "foo",
		Name:       "foo",
		ProtocolId: "foo",
		Outputs: []model.Content{
			{
				Id: "foo",
				ContentVariable: model.ContentVariable{
					Id:   "foo",
					Name: "foo",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo",
			},
			{
				Id: "bar",
				ContentVariable: model.ContentVariable{
					Id:   "bar",
					Name: "bar",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo2",
			},
		},
		Inputs: []model.Content{
			{
				Id: "foo",
				ContentVariable: model.ContentVariable{
					Id:   "foo",
					Name: "foo",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo",
			},
			{
				Id: "bar",
				ContentVariable: model.ContentVariable{
					Id:   "bar",
					Name: "bar",
					Type: model.String,
				},
				Serialization:     model.JSON,
				ProtocolSegmentId: "foo2",
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

type DbMock struct{}

func (this DbMock) GetProtocol(ctx context.Context, id string) (result model.Protocol, exists bool, err error) {
	return model.Protocol{
		Id: "id",
		ProtocolSegments: []model.ProtocolSegment{
			{
				Id:   "foo",
				Name: "foo",
			},
			{
				Id:   "foo2",
				Name: "foo2",
			},
		},
	}, true, nil
}

func (this DbMock) Disconnect() {
	panic("implement me")
}

func (this DbMock) GetDevice(ctx context.Context, id string) (device model.Device, exists bool, err error) {
	panic("implement me")
}

func (this DbMock) SetDevice(ctx context.Context, device model.Device) error {
	panic("implement me")
}

func (this DbMock) RemoveDevice(ctx context.Context, id string) error {
	panic("implement me")
}

func (this DbMock) GetDeviceByLocalId(ctx context.Context, localId string) (device model.Device, exists bool, err error) {
	panic("implement me")
}

func (this DbMock) GetHub(ctx context.Context, id string) (hub model.Hub, exists bool, err error) {
	panic("implement me")
}

func (this DbMock) SetHub(ctx context.Context, hub model.Hub) error {
	panic("implement me")
}

func (this DbMock) RemoveHub(ctx context.Context, id string) error {
	panic("implement me")
}

func (this DbMock) GetHubsByDeviceLocalId(ctx context.Context, localId string) (hubs []model.Hub, err error) {
	panic("implement me")
}

func (this DbMock) GetDeviceType(ctx context.Context, id string) (deviceType model.DeviceType, exists bool, err error) {
	panic("implement me")
}

func (this DbMock) SetDeviceType(ctx context.Context, deviceType model.DeviceType) error {
	panic("implement me")
}

func (this DbMock) RemoveDeviceType(ctx context.Context, id string) error {
	panic("implement me")
}

func (this DbMock) ListDeviceTypes(ctx context.Context, limit int64, offset int64, sort string) (result []model.DeviceType, err error) {
	panic("implement me")
}

func (this DbMock) GetDeviceTypesByServiceId(ctx context.Context, serviceId string) ([]model.DeviceType, error) {
	panic("implement me")
}

func (this DbMock) GetDeviceGroup(ctx context.Context, id string) (deviceGroup model.DeviceGroup, exists bool, err error) {
	panic("implement me")
}

func (this DbMock) SetDeviceGroup(ctx context.Context, deviceGroup model.DeviceGroup) error {
	panic("implement me")
}

func (this DbMock) RemoveDeviceGroup(ctx context.Context, id string) error {
	panic("implement me")
}

func (this DbMock) ListDeviceGroups(ctx context.Context, limit int64, offset int64, sort string) (result []model.DeviceGroup, err error) {
	panic("implement me")
}

func (this DbMock) ListProtocols(ctx context.Context, limit int64, offset int64, sort string) ([]model.Protocol, error) {
	panic("implement me")
}

func (this DbMock) SetProtocol(ctx context.Context, protocol model.Protocol) error {
	panic("implement me")
}

func (this DbMock) RemoveProtocol(ctx context.Context, id string) error {
	panic("implement me")
}
