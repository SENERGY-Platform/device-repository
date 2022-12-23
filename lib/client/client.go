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
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/api"
	"net/http"
)

type Client struct {
	baseUrl string
}

func NewClient(baseUrl string) api.Controller {
	return &Client{baseUrl: baseUrl}
}

func do[T any](req *http.Request) (result T, err error, code int) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if resp.StatusCode > 299 {
		return result, errors.New("unexpected statuscode"), resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return
}

func (c *Client) validate(path string, e interface{}) (err error, code int) {
	b, err := json.Marshal(e)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+path+"?dry-run=true", bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, resp.StatusCode
}

func (c *Client) validateDelete(path string) (err error, code int) {
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+path+"?dry-run=true", nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, resp.StatusCode
}
