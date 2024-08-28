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
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/api"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Interface = api.Controller

type Client struct {
	baseUrl string
}

func NewClient(baseUrl string) Interface {
	return &Client{baseUrl: baseUrl}
}

func do[T any](req *http.Request) (result T, err error, code int) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
		return result, fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp)), resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		_, _ = io.ReadAll(resp.Body) //ensure resp.Body is read to EOF
		return result, err, http.StatusInternalServerError
	}
	return
}

func doWithTotalInResult[T any](req *http.Request) (result T, total int64, err error, code int) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
		return result, total, fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp)), resp.StatusCode
	}
	total, err = strconv.ParseInt(resp.Header.Get("X-Total-Count"), 10, 64)
	if err != nil {
		return result, total, fmt.Errorf("unable to read X-Total-Count header %w", err), http.StatusInternalServerError
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		_, _ = io.ReadAll(resp.Body) //ensure resp.Body is read to EOF
		return result, total, err, http.StatusInternalServerError
	}
	return
}

func (c *Client) validateWithToken(token string, path string, e interface{}) (err error, code int) {
	return c.validateWithTokenAndOptions(token, path, e, nil)
}

func (c *Client) validateWithTokenAndOptions(token string, path string, e interface{}, options url.Values) (err error, code int) {
	b, err := json.Marshal(e)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	optQuery := options.Encode()
	if optQuery != "" {
		optQuery = "&" + optQuery
	}
	req, err := http.NewRequest(http.MethodPut, c.baseUrl+path+"?dry-run=true", bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("validation error: %v", string(temp)), resp.StatusCode
	}
	return nil, resp.StatusCode
}

func (c *Client) validate(path string, e interface{}) (err error, code int) {
	return c.validateWithOptions(path, e, nil)
}

func (c *Client) validateWithOptions(path string, e interface{}, options url.Values) (err error, code int) {
	b, err := json.Marshal(e)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	optQuery := options.Encode()
	if optQuery != "" {
		optQuery = "&" + optQuery
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
