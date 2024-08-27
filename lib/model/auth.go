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

package model

import (
	"errors"
	"github.com/SENERGY-Platform/models/go/models"
	"net/url"
	"slices"
)

type AuthAction = models.PermissionFlag

const (
	READ         = models.Read
	WRITE        = models.Write
	EXECUTE      = models.Execute
	ADMINISTRATE = models.Administrate
)

func GetPermissionFlagFromQuery(query url.Values) (models.PermissionFlag, error) {
	if !query.Has("p") {
		return models.UnsetPermissionFlag, nil
	}
	runes := []rune(query.Get("p"))
	if len(runes) != 1 {
		return models.UnsetPermissionFlag, errors.New("invalid permission flag")
	}
	flag := models.PermissionFlag(runes[0])
	if !slices.Contains([]models.PermissionFlag{READ, WRITE, EXECUTE, ADMINISTRATE}, flag) {
		return models.UnsetPermissionFlag, errors.New("invalid permission flag")
	}
	return flag, nil
}
