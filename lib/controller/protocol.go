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
	"errors"
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadProtocol(id string, token string) (result models.Protocol, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	protocol, exists, err := this.db.GetProtocol(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return protocol, nil, http.StatusOK
}

func (this *Controller) ListProtocols(token string, limit int64, offset int64, sort string) (result []models.Protocol, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, err = this.db.ListProtocols(ctx, limit, offset, sort)
	return
}

func (this *Controller) ValidateProtocol(protocol models.Protocol) (err error, code int) {
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
	exists := map[string]bool{}
	for _, segment := range protocol.ProtocolSegments {
		if segment.Id == "" {
			return errors.New("missing protocol-segment id"), http.StatusBadRequest
		}
		if segment.Name == "" {
			return errors.New("missing protocol-segment name"), http.StatusBadRequest
		}
		if _, found := exists[segment.Name]; found {
			return errors.New("repeated protocol-segment name"), http.StatusBadRequest
		}
		exists[segment.Name] = true
	}
	return nil, http.StatusOK
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetProtocol(protocol models.Protocol, owner string) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.SetProtocol(ctx, protocol)
}

func (this *Controller) DeleteProtocol(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveProtocol(ctx, id)
}
