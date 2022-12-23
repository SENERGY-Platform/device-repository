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

func (c *Client) GetAspects() ([]models.Aspect, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspects", nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.Aspect](req)
}
func (c *Client) GetAspectsWithMeasuringFunction(ancestors bool, descendants bool) ([]models.Aspect, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspects?function=measuring-function&ancestors="+strconv.FormatBool(ancestors)+"&descendants="+strconv.FormatBool(descendants), nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	return do[[]models.Aspect](req)
}

func (c *Client) GetAspect(id string) (models.Aspect, error, int) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl+"/aspects/"+id, nil)
	if err != nil {
		return models.Aspect{}, err, http.StatusInternalServerError
	}
	return do[models.Aspect](req)
}

func (c *Client) ValidateAspect(aspect models.Aspect) (err error, code int) {
	return c.validate("/aspects", aspect)
}

func (c *Client) ValidateAspectDelete(id string) (err error, code int) {
	return c.validateDelete("/aspects/" + id)
}
