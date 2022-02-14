/*
 * Copyright 2019 InfAI (CC SES)
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

package com

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
)

func NewSecurity(config config.Config) (*Security, error) {
	return &Security{config: config}, nil
}

type Security struct {
	config config.Config
}

type IdWrapper struct {
	Id string `json:"id"`
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (this *Security) CheckBool(token string, kind string, id string, action model.AuthAction) (allowed bool, err error) {
	req, err := http.NewRequest("GET", this.config.PermissionsUrl+"/v3/resources/"+url.QueryEscape(kind)+"/"+url.QueryEscape(id)+"/access", nil)
	if err != nil {
		debug.PrintStack()
		return false, err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("ERROR: ", err)
		debug.PrintStack()
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return false, errors.New(buf.String())
	}
	err = json.NewDecoder(resp.Body).Decode(&allowed)
	if err != nil {
		debug.PrintStack()
		return false, err
	}
	return allowed, nil
}

func (this *Security) CheckMultiple(token string, kind string, ids []string, action model.AuthAction) (result map[string]bool, err error) {
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(map[string]interface{}{
		"resource": kind,
		"check_ids": map[string]interface{}{
			"rights": action,
			"ids":    ids,
		},
	})
	if err != nil {
		return result, err
	}

	req, err := http.NewRequest("POST", this.config.PermissionsUrl+"/v2/query", body)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	req.Header.Set("Authorization", token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return result, errors.New(buf.String())
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	return result, nil
}
