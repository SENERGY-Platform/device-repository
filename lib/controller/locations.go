/*
 * Copyright 2022 InfAI (CC SES)
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
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"net/http"
	"net/url"
	"strings"
)

func (this *Controller) SetLocation(location models.Location, owner string) error {
	err := this.EnsureInitialRights(this.config.LocationTopic, location.Id, owner)
	if err != nil {
		return err
	}
	ctx, _ := getTimeoutContext()
	return this.db.SetLocation(ctx, location)
}

func (this *Controller) DeleteLocation(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveLocation(ctx, id)
}

func (this *Controller) GetLocation(id string, token string) (result models.Location, err error, code int) {
	ok, err := this.db.CheckBool(token, this.config.LocationTopic, id, model.READ)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetLocation(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ValidateLocation(location models.Location) (err error, code int) {
	if location.Id == "" {
		return errors.New("missing device class id"), http.StatusBadRequest
	}
	if !strings.HasPrefix(location.Id, model.URN_PREFIX) {
		return errors.New("invalid location id"), http.StatusBadRequest
	}
	if location.Name == "" {
		return errors.New("missing device class name"), http.StatusBadRequest
	}
	if location.Image != "" {
		if _, err := url.ParseRequestURI(location.Image); err != nil {
			return fmt.Errorf("image is not valid URL: %w", err), http.StatusBadRequest
		}
	}
	for _, did := range location.DeviceIds {
		if _, err := url.Parse(did); err != nil {
			return fmt.Errorf("device is not valid URI: %w", err), http.StatusBadRequest
		}
	}
	for _, dgid := range location.DeviceGroupIds {
		if _, err := url.Parse(dgid); err != nil {
			return fmt.Errorf("device-group is not valid URI: %w", err), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) ListLocations(token string, options model.LocationListOptions) (result []models.Location, total int64, err error, errCode int) {
	ids := []string{}
	permissionFlag := options.Permission
	if permissionFlag == models.UnsetPermissionFlag {
		permissionFlag = models.Read
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, total, err, http.StatusBadRequest
	}

	if options.Ids == nil {
		if jwtToken.IsAdmin() {
			ids = nil //no auth check for admins -> no id filter
		} else {
			ids, err = this.db.ListAccessibleResourceIds(token, this.config.LocationTopic, 0, 0, permissionFlag)
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
			if len(ids) == 0 {
				ids = []string{}
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err := this.db.CheckMultiple(token, this.config.LocationTopic, options.Ids, permissionFlag)
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
		for id, ok := range idMap {
			if ok {
				ids = append(ids, id)
			}
		}
	}
	options.Ids = ids

	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListLocations(ctx, options)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}

	return result, total, nil, http.StatusOK
}
