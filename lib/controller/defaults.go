/*
 * Copyright 2025 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"net/http"
	"slices"
)

func (this *Controller) GetDefaultDeviceAttributes(token string) (attributes []models.Attribute, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return attributes, err, http.StatusInternalServerError
	}
	ctx, _ := getTimeoutContext()
	attributes, err = this.db.GetDefaultDeviceAttributes(ctx, jwtToken.GetUserId())
	if err != nil {
		return attributes, err, http.StatusInternalServerError
	}
	return attributes, nil, http.StatusOK
}

func (this *Controller) SetDefaultDeviceAttributes(token string, attributes []models.Attribute) (err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for i, attribute := range attributes {
		if attribute.Origin != "default" && attribute.Origin != "" {
			return errors.New("default attributes may not have an origin other than 'default'"), http.StatusBadRequest
		}
		attributes[i].Origin = "default"
	}
	ctx, _ := getTimeoutContext()
	err = this.db.SetDefaultDeviceAttributes(ctx, jwtToken.GetUserId(), attributes)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) applyDefaultDeviceAttributes(device model.DeviceWithConnectionState) (result model.DeviceWithConnectionState, err error) {
	ctx, _ := getTimeoutContext()
	defaultAttributes, err := this.db.GetDefaultDeviceAttributes(ctx, device.OwnerId)
	if err != nil {
		return result, err
	}
	for _, attribute := range defaultAttributes {
		if !slices.ContainsFunc(device.Attributes, func(a models.Attribute) bool { return a.Key == attribute.Key }) {
			attribute.Origin = "default"
			device.Attributes = append(device.Attributes, attribute)
		}
	}
	return device, nil
}

func (this *Controller) applyDefaultDeviceAttributesToList(devices []model.DeviceWithConnectionState) (result []model.DeviceWithConnectionState, err error) {
	cache := map[string][]models.Attribute{}
	getDefaultAttr := func(userId string) ([]models.Attribute, error) {
		if attributes, ok := cache[userId]; ok {
			return attributes, nil
		}
		ctx, _ := getTimeoutContext()
		attributes, err := this.db.GetDefaultDeviceAttributes(ctx, userId)
		if err != nil {
			return attributes, err
		}
		cache[userId] = attributes
		return attributes, nil
	}
	for i, device := range devices {
		defaultAttributes, err := getDefaultAttr(device.OwnerId)
		if err != nil {
			return result, err
		}
		for _, attribute := range defaultAttributes {
			if !slices.ContainsFunc(device.Attributes, func(a models.Attribute) bool { return a.Key == attribute.Key }) {
				attribute.Origin = "default"
				device.Attributes = append(device.Attributes, attribute)
			}
		}
		devices[i] = device
	}
	return devices, nil
}
