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
	"context"
	"errors"
	"net/http"

	"github.com/SENERGY-Platform/device-repository/v3/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func (this *Controller) GetLastUpdateTimestamps(ctx context.Context, token string, userId string) (result []model.LastUpdateTimestamp, err error, code int) {
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
	ctx, _ = getTimeoutContext(ctx)
	result, err = this.db.GetLastUpdateTimestampsForUser(ctx, userId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}

// GetCacheTimestamps returns the last-update timestamps of the fixed set of collections that the endpoints
// without owner read (aspects, aspect nodes, aspect classes, concepts, characteristics, functions, device classes,
// device types, device type criteria, protocols), for the user id "". A collection without timestamp is listed with 0.
// The set is the same for every caller; locations also write under the user id "" and are not part of it.
// There is no token, no cache and no per-user data: the result is the cache validator of the endpoints of these resources.
func (this *Controller) GetCacheTimestamps(ctx context.Context) (result []model.LastUpdateTimestamp, err error) {
	collections := []string{
		this.config.MongoAspectCollection,
		mongo.GetAspectNodeCollectionName(this.config),
		this.config.MongoAspectClassCollection,
		this.config.MongoConceptCollection,
		this.config.MongoCharacteristicCollection,
		this.config.MongoFunctionCollection,
		this.config.MongoDeviceClassCollection,
		this.config.MongoDeviceTypeCollection,
		mongo.GetDeviceTypeCriteriaCollectionName(this.config),
		this.config.MongoProtocolCollection,
	}
	ctx, cancel := getTimeoutContext(ctx)
	defer cancel()
	all, err := this.db.GetLastUpdateTimestampsForUser(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, collection := range collections {
		element := model.LastUpdateTimestamp{Collection: collection}
		for _, candidate := range all {
			if candidate.Collection == collection && candidate.UserId == "" && candidate.UnixTimestamp > element.UnixTimestamp {
				element = candidate
			}
		}
		result = append(result, element)
	}
	return result, nil
}
