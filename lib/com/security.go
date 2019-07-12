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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"net/http"

	"net/url"

	"errors"

	"encoding/json"

	"strconv"

	"log"

	"github.com/SmartEnergyPlatform/jwt-http-router"
)

func NewSecurity(config config.Config) (*Security, error) {
	return &Security{config: config}, nil
}

type Security struct {
	config config.Config
}

func authActionToString(action model.AuthAction) (right string) {
	switch action {
	case model.READ:
		right = "r"
	case model.WRITE:
		right = "w"
	case model.EXECUTE:
		right = "x"
	case model.ADMINISTRATE:
		right = "a"
	}
	return
}

type IdWrapper struct {
	Id string `json:"id"`
}

func IsAdmin(jwt jwt_http_router.Jwt) bool {
	return contains(jwt.RealmAccess.Roles, "admin")
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (this *Security) Check(jwt jwt_http_router.Jwt, kind string, id string, action model.AuthAction) (err error) {
	if IsAdmin(jwt) {
		return nil
	}
	right := authActionToString(action)
	result := false
	err = jwt.Impersonate.GetJSON(this.config.PermissionsUrl+"/jwt/check/"+url.QueryEscape(kind)+"/"+url.QueryEscape(id)+"/"+right+"/bool", &result)
	if err != nil {
		log.Println("DEBUG: permissions.Check:", err)
		return err
	}
	if !result {
		err = errors.New("access denied")
	}
	return
}

func (this *Security) CheckBool(jwt jwt_http_router.Jwt, kind string, id string, action model.AuthAction) (allowed bool, err error) {
	if IsAdmin(jwt) {
		return true, nil
	}
	right := authActionToString(action)
	err = jwt.Impersonate.GetJSON(this.config.PermissionsUrl+"/jwt/check/"+url.QueryEscape(kind)+"/"+url.QueryEscape(id)+"/"+right+"/bool", &allowed)
	if err != nil {
		log.Println("DEBUG: permissions.Check:", err)
	}
	return
}

func (this *Security) CheckList(jwt jwt_http_router.Jwt, kind string, ids []string, action model.AuthAction) (result map[string]bool, err error) {
	right := authActionToString(action)
	err = jwt.Impersonate.PostJSON(this.config.PermissionsUrl+"/ids/check/"+url.QueryEscape(kind)+"/"+right, ids, &result)
	return
}

func (this *Security) ListAll(jwt jwt_http_router.Jwt, kind string, action model.AuthAction) (result []IdWrapper, err error) {
	//"/jwt/list/:resource_kind/:right"
	right := authActionToString(action)
	resp, err := jwt.Impersonate.Get(this.config.PermissionsUrl + "/jwt/list/" + url.QueryEscape(kind) + "/" + right)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("access denied")
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func (this *Security) List(jwt jwt_http_router.Jwt, kind string, action model.AuthAction, limit string, offset string) (ids []string, err error) {
	//"/jwt/list/:resource_kind/:right"
	right := authActionToString(action)
	resp, err := jwt.Impersonate.Get(this.config.PermissionsUrl + "/jwt/list/" + url.QueryEscape(kind) + "/" + right + "/" + limit + "/" + offset)
	if err != nil {
		return ids, err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("access denied")
		return ids, err
	}
	var result []IdWrapper
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return ids, err
	}
	for _, id := range result {
		ids = append(ids, id.Id)
	}
	return
}

func (this *Security) SortedList(jwt jwt_http_router.Jwt, kind string, action model.AuthAction, limit string, offset string, sortby string, sortdirection string) (ids []string, err error) {
	//"/jwt/list/:resource_kind/:right"
	right := authActionToString(action)
	resp, err := jwt.Impersonate.Get(this.config.PermissionsUrl + "/jwt/list/" + url.QueryEscape(kind) + "/" + right + "/" + limit + "/" + offset + "/" + sortby + "/" + sortdirection)
	if err != nil {
		return ids, err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("access denied")
		return ids, err
	}
	var result []IdWrapper
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return ids, err
	}
	for _, id := range result {
		ids = append(ids, id.Id)
	}
	return
}

func (this *Security) Search(jwt jwt_http_router.Jwt, kind string, query string, action model.AuthAction, limit string, offset string) (result []IdWrapper, err error) {
	//"/jwt/search/:resource_kind/:query/:right/:limit/:offset"
	right := authActionToString(action)
	resp, err := jwt.Impersonate.Get(this.config.PermissionsUrl + "/jwt/search/" + url.QueryEscape(kind) + "/" + url.QueryEscape(query) + "/" + right + "/" + limit + "/" + offset)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("access denied")
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func (this *Security) SearchAll(jwt jwt_http_router.Jwt, kind string, query string, action model.AuthAction) (result []IdWrapper, err error) {
	//"/jwt/search/:resource_kind/:query/:right"
	right := authActionToString(action)
	resp, err := jwt.Impersonate.Get(this.config.PermissionsUrl + "/jwt/search/" + url.QueryEscape(kind) + "/" + url.QueryEscape(query) + "/" + right)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("access denied")
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func (this *Security) SelectFieldAll(jwt jwt_http_router.Jwt, kind string, field string, value string, action model.AuthAction) (result []IdWrapper, err error) {
	//"/jwt/select/:resource_kind/:field/:value/:right"
	right := authActionToString(action)
	resp, err := jwt.Impersonate.Get(this.config.PermissionsUrl + "/jwt/select/" + url.QueryEscape(kind) + "/" + url.QueryEscape(field) + "/" + url.QueryEscape(value) + "/" + right)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("access denied")
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func (this *Security) SelectField(jwt jwt_http_router.Jwt, kind string, field string, value string, action model.AuthAction, limit string, offset string, sortby string, sortdirection string) (result []IdWrapper, err error) {
	//"/jwt/select/:resource_kind/:field/:value/:right"
	right := authActionToString(action)
	resp, err := jwt.Impersonate.Get(this.config.PermissionsUrl + "/jwt/select/" + url.QueryEscape(kind) + "/" + url.QueryEscape(field) + "/" + url.QueryEscape(value) + "/" + right + "/" + limit + "/" + offset + "/" + sortby + "/" + sortdirection)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("access denied")
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func (this *Security) Exists(jwt jwt_http_router.Jwt, kind string, id string) (exists bool, err error) {
	// /administrate/exists/:resource_kind/:resource
	resp, err := jwt.Impersonate.Get(this.config.PermissionsUrl + "/administrate/exists/" + url.QueryEscape(kind) + "/" + url.QueryEscape(id))
	if err != nil {
		return exists, err
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("existence check request responds with " + strconv.Itoa(resp.StatusCode))
		return exists, err
	}
	err = json.NewDecoder(resp.Body).Decode(&exists)
	return
}
