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

package model

import "github.com/SENERGY-Platform/models/go/models"

type DeviceListOptions struct {
	Ids             []string                //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	ConnectionState *models.ConnectionState //filter
	Search          string
	Limit           int64                 //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset          int64                 //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy          string                //default name.asc
	Permission      models.PermissionFlag //defaults to read
}

type HubListOptions struct {
	Ids             []string                ///filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	ConnectionState *models.ConnectionState //filter
	Search          string
	Limit           int64                 //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset          int64                 //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy          string                //default name.asc
	Permission      models.PermissionFlag //defaults to read
}
