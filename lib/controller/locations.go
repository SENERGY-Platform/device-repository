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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func (this *Controller) SetLocation(ctx context.Context, token string, location models.Location) (result models.Location, err error, errCode int) {
	if location.Id != "" {
		ctx, _ := getTimeoutContext(ctx)
		_, exists, err := this.db.GetLocation(ctx, location.Id)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		if exists {
			ok, err, _ := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.LocationTopic, location.Id, client.Write)
			if err != nil {
				return result, err, http.StatusInternalServerError
			}
			if !ok {
				return result, errors.New("access denied"), http.StatusForbidden
			}
		}
	}

	location.GenerateId()
	if !this.config.DisableStrictValidationForTesting {
		location.DeviceIds, err = this.filterInvalidDeviceIds(ctx, token, location.DeviceIds, "r")
		if err != nil {
			return location, err, http.StatusInternalServerError
		}
		err, code := this.ValidateLocation(ctx, location)
		if err != nil {
			return location, err, code
		}
	}

	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return location, err, http.StatusBadRequest
	}

	err = this.setLocation(ctx, location, jwtToken.GetUserId())
	if err != nil {
		return location, err, http.StatusInternalServerError
	}

	return location, nil, http.StatusOK
}

func (this *Controller) DeleteLocation(ctx context.Context, token string, id string) (err error, code int) {
	if err := preventIdModifier(id); err != nil {
		return err, http.StatusBadRequest
	}
	ctx, _ = getTimeoutContext(ctx)
	_, exists, err := this.db.GetLocation(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return nil, http.StatusOK
	}
	ok, err, _ := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.LocationTopic, id, client.Administrate)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("access denied"), http.StatusForbidden
	}
	err = this.deleteLocation(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) setLocationSyncHandler(ctx context.Context, location models.Location, user string) (err error) {
	err = this.EnsureInitialRights(ctx, this.config.LocationTopic, location.Id, user)
	if err != nil {
		return err
	}
	return this.publisher.PublishLocation(location)
}

func (this *Controller) setLocation(ctx context.Context, location models.Location, owner string) error {
	ctx, _ = getTimeoutContext(ctx)
	return this.db.SetLocation(ctx, location, withContext2(ctx, this.setLocationSyncHandler), owner)
}

func (this *Controller) deleteLocationSyncHandler(ctx context.Context, location models.Location) error {
	err := this.RemoveRights(ctx, this.config.LocationTopic, location.Id)
	if err != nil {
		return err
	}
	return this.publisher.PublishLocationDelete(location.Id)
}

func (this *Controller) deleteLocation(ctx context.Context, id string) error {
	ctx, _ = getTimeoutContext(ctx)
	return this.db.RemoveLocation(ctx, id, withContext(ctx, this.deleteLocationSyncHandler))
}

func (this *Controller) GetLocation(ctx context.Context, id string, token string) (result models.Location, err error, code int) {
	ok, err, _ := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.LocationTopic, id, client.Read)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	ctx, _ = getTimeoutContext(ctx)
	result, exists, err := this.db.GetLocation(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ValidateLocation(ctx context.Context, location models.Location) (err error, code int) {
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

func (this *Controller) ListLocations(ctx context.Context, token string, options model.LocationListOptions) (result []models.Location, total int64, err error, errCode int) {
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
			ids, err, _ = this.permissionsV2Client.ListAccessibleResourceIdsContext(ctx, token, this.config.LocationTopic, client.ListOptions{}, client.Permission(permissionFlag))
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
		idMap, err, _ := this.permissionsV2Client.CheckMultiplePermissionsContext(ctx, token, this.config.LocationTopic, options.Ids, client.Permission(permissionFlag))
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

	ctx, _ = getTimeoutContext(ctx)
	result, total, err = this.db.ListLocations(ctx, options)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}

	return result, total, nil, http.StatusOK
}

func (this *Controller) ListExtendedLocations(ctx context.Context, token string, options model.LocationListOptions) (result []models.ExtendedLocation, total int64, err error, errCode int) {
	locations, total, err, code := this.ListLocations(ctx, token, options)
	if err != nil {
		return result, total, err, code
	}

	result = make([]models.ExtendedLocation, len(locations))
	wg := sync.WaitGroup{}
	mux := sync.Mutex{}
	for i, location := range locations {
		wg.Add(1)
		go func(h models.Location, resultIndex int) {
			defer wg.Done()
			extended, temperr := this.extendLocation(ctx, token, h)
			mux.Lock()
			defer mux.Unlock()
			err = errors.Join(err, temperr)
			result[resultIndex] = extended
		}(location, i)
	}
	wg.Wait()
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) extendLocation(ctx context.Context, token string, location models.Location) (models.ExtendedLocation, error) {
	computedPermList, err, _ := this.permissionsV2Client.ListComputedPermissionsContext(ctx, token, this.config.LocationTopic, []string{location.Id})
	if err != nil {
		return models.ExtendedLocation{}, err
	}
	if len(computedPermList) == 0 {
		return models.ExtendedLocation{}, errors.New("no computation permissions")
	}
	permissions := models.Permissions{
		Read:         computedPermList[0].Read,
		Write:        computedPermList[0].Write,
		Execute:      computedPermList[0].Execute,
		Administrate: computedPermList[0].Administrate,
	}
	return models.ExtendedLocation{
		Location:    location,
		Permissions: permissions,
	}, nil
}
