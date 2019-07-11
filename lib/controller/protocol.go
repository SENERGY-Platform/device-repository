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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"net/http"
	"time"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadProtocol(id string, jwt jwt_http_router.Jwt) (result model.Protocol, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	protocol, exists, err := this.db.GetProtocol(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return protocol, nil, http.StatusOK
}

func (this *Controller) ListProtocols(jwt jwt_http_router.Jwt, limit int64, offset int64, sort string) (result []model.Protocol, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err = this.db.ListProtocols(ctx, limit, offset, sort)
	return
}

func (this *Controller) ValidateProtocol(protocol model.Protocol) (err error, code int) {
	if protocol.Id == "" {
		return errors.New("missing protocol id"), http.StatusBadRequest
	}
	if protocol.Name == "" {
		return errors.New("missing protocol name"), http.StatusBadRequest
	}
	if protocol.Handler == "" {
		return errors.New("missing protocol handler"), http.StatusBadRequest
	}
	if len(protocol.ProtocolSegments) == 0 {
		return errors.New("expect at least one protocol-segment"), http.StatusBadRequest
	}
	for _, segment := range protocol.ProtocolSegments {
		if segment.Id == "" {
			return errors.New("missing protocol-segment id"), http.StatusBadRequest
		}
		if segment.Name == "" {
			return errors.New("missing protocol-segment name"), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetProtocol(protocol model.Protocol, owner string) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.SetProtocol(ctx, protocol)
}

func (this *Controller) DeleteProtocol(id string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.RemoveProtocol(ctx, id)
}
