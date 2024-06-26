/*
 * Copyright 2020 InfAI (CC SES)
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

package mocks

import (
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"strings"
)

type Security struct {
	access map[string]bool
	admins map[string][]string
}

func NewSecurity() *Security {
	return &Security{access: map[string]bool{}, admins: map[string][]string{}}
}

func (this *Security) CheckBool(token string, kind string, id string, action model.AuthAction) (allowed bool, err error) {
	return this.access[this.getKey(kind, id)], nil
}

func (this *Security) CheckMultiple(token string, kind string, ids []string, action model.AuthAction) (map[string]bool, error) {
	result := map[string]bool{}
	for _, id := range ids {
		result[id], _ = this.CheckBool(token, kind, id, action)
	}
	return result, nil
}

func (this *Security) getKey(kind string, id string) string {
	return kind + "/" + id
}

func (this *Security) Set(kind string, id string, access bool) {
	this.access[this.getKey(kind, id)] = access
}

func (this *Security) GetAdminUsers(token string, kind string, id string) (admins []string, err error) {
	return this.admins[this.getKey(kind, id)], nil
}

func (this *Security) SetAdmins(kind string, id string, admins []string) {
	this.admins[this.getKey(kind, id)] = admins
}

func (this *Security) ListAccessibleResourceIds(token string, topic string, limit int64, offset int64, action model.AuthAction) (result []string, err error) {
	for key, access := range this.access {
		if access && strings.HasPrefix(key, topic+"/") {
			result = append(result, strings.TrimPrefix(key, topic+"/"))
		}
	}
	return result, nil
}
