/*
 * Copyright 2026 InfAI (CC SES)
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
	"net/http"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func (this *Controller) GetLastUpdateTimestamps(token string, userId string) (result []model.LastUpdateTimestamp, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if userId != "" && jwtToken.GetUserId() != userId && !jwtToken.IsAdmin() {
		return result, errors.New("only admins may set userId explicitly"), http.StatusBadRequest
	}
	if userId == "" {
		userId = jwtToken.GetUserId()
	}
	ctx, _ := getTimeoutContext()
	result, err = this.db.GetLastUpdateTimestampsForUser(ctx, jwtToken.GetUserId())
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}
