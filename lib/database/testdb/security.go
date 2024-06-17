/*
 * Copyright 2024 InfAI (CC SES)
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

package testdb

import (
	"github.com/SENERGY-Platform/device-repository/lib/model"
)

func (this *DB) EnsureInitialRights(resourceKind string, resourceId string, owner string) error {
	panic("implement me")
}

func (this *DB) SetRights(resourceKind string, resourceId string, rights model.ResourceRights) error {
	panic("implement me")
}

func (this *DB) RemoveRights(topic string, id string) error {
	panic("implement me")
}

func (this *DB) CheckBool(token string, kind string, id string, action model.AuthAction) (allowed bool, err error) {
	panic("implement me")
}

func (this *DB) CheckMultiple(token string, kind string, ids []string, action model.AuthAction) (map[string]bool, error) {
	panic("implement me")
}

func (this *DB) GetAdminUsers(token string, topic string, resourceId string) (admins []string, err error) {
	panic("implement me")
}
