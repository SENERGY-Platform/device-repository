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

package client

import (
	"github.com/SENERGY-Platform/models/go/models"
	"net/http"
)

func (c *Client) GetDeviceClasses() ([]models.DeviceClass, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes", nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.DeviceClass](req)
}

func (c *Client) GetDeviceClassesWithControllingFunctions() ([]models.DeviceClass, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes?function=controlling-function", nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.DeviceClass](req)
}

func (c *Client) GetDeviceClassesFunctions(id string) (result []models.Function, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes/"+id+"/functions", nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.Function](req)
}

func (c *Client) GetDeviceClassesControllingFunctions(id string) (result []models.Function, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes/"+id+"/controlling-functions", nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[[]models.Function](req)
}

func (c *Client) GetDeviceClass(id string) (result models.DeviceClass, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/device-classes/"+id, nil)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return do[models.DeviceClass](req)
}

func (c *Client) ValidateDeviceClass(deviceclass models.DeviceClass) (err error, code int) {
	return c.validate("/device-classes", deviceclass)
}

func (c *Client) ValidateDeviceClassDelete(id string) (err error, code int) {
	return c.validateDelete("/device-classes/" + id)
}
