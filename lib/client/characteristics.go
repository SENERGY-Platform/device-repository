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
	"strconv"
)

func (c *Client) GetCharacteristics(leafsOnly bool) (result []models.Characteristic, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/characteristics?leafsOnly="+strconv.FormatBool(leafsOnly), nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.Characteristic](req)
}

func (c *Client) GetCharacteristic(id string) (result models.Characteristic, err error, errCode int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/characteristics/"+id, nil)
	if err != nil {
		return models.Characteristic{}, err, http.StatusInternalServerError
	}
	return do[models.Characteristic](req)
}

func (c *Client) ValidateCharacteristics(characteristic models.Characteristic) (err error, code int) {
	return c.validate("/characteristics", characteristic)
}

func (c *Client) ValidateCharacteristicDelete(id string) (err error, code int) {
	return c.validateDelete("/characteristics/" + id)
}
