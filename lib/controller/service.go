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
	"net/http"
	"time"
)

func (this *Controller) ValidateService(service model.Service) (error, int) {
	if service.Id == "" {
		return errors.New("missing service id"), http.StatusBadRequest
	}
	if service.Name == "" {
		return errors.New("missing service name"), http.StatusBadRequest
	}
	if service.LocalId == "" {
		return errors.New("missing service local id"), http.StatusBadRequest
	}
	if service.ProtocolId == "" {
		return errors.New("missing service protocol id"), http.StatusBadRequest
	}
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	protocol, exists, err := this.db.GetProtocol(ctx, service.ProtocolId)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return errors.New("unknown protocol id: " + service.ProtocolId), http.StatusBadRequest
	}
	for _, content := range append(service.Inputs, service.Outputs...) {
		err, code := ValidateContent(content, protocol)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}
