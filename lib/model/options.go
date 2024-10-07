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
	DeviceTypeIds   []string                //filter; ignored if DeviceTypeIds == nil; DeviceTypeIds == []string{} will return an empty list;
	ConnectionState *models.ConnectionState //filter
	Search          string
	Limit           int64                 //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset          int64                 //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy          string                //default name.asc
	Permission      models.PermissionFlag //defaults to read
	AttributeKeys   []string              //filter; ignored if nil; AttributeKeys and AttributeValues are independently evaluated, needs local filtering if a search like "attr1"="value1" is needed
	AttributeValues []string              //filter; ignored if nil; AttributeKeys and AttributeValues are independently evaluated, needs local filtering if a search like "attr1"="value1" is needed
}

type ExtendedDeviceListOptions struct {
	Ids             []string                //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	DeviceTypeIds   []string                //filter; ignored if DeviceTypeIds == nil; DeviceTypeIds == []string{} will return an empty list;
	ConnectionState *models.ConnectionState //filter
	Search          string
	Limit           int64                 //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset          int64                 //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy          string                //default name.asc
	Permission      models.PermissionFlag //defaults to read
	AttributeKeys   []string              //filter; ignored if nil; AttributeKeys and AttributeValues are independently evaluated, needs local filtering if a search like "attr1"="value1" is needed
	AttributeValues []string              //filter; ignored if nil; AttributeKeys and AttributeValues are independently evaluated, needs local filtering if a search like "attr1"="value1" is needed
	FullDt          bool                  //if true, result contains full device-type
}

func (this ExtendedDeviceListOptions) ToDeviceListOptions() DeviceListOptions {
	return DeviceListOptions{
		Ids:             this.Ids,
		DeviceTypeIds:   this.DeviceTypeIds,
		ConnectionState: this.ConnectionState,
		Search:          this.Search,
		Limit:           this.Limit,
		Offset:          this.Offset,
		SortBy:          this.SortBy,
		Permission:      this.Permission,
		AttributeKeys:   this.AttributeKeys,
		AttributeValues: this.AttributeValues,
	}
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

type DeviceTypeListOptions struct {
	Ids              []string
	Search           string
	Limit            int64            //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset           int64            //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy           string           //default name.asc
	AttributeKeys    []string         //filter; ignored if nil; AttributeKeys and AttributeValues are independently evaluated, needs local filtering if a search like "attr1"="value1" is needed
	AttributeValues  []string         //filter; ignored if nil; AttributeKeys and AttributeValues are independently evaluated, needs local filtering if a search like "attr1"="value1" is needed
	Criteria         []FilterCriteria //filter; ignored if nil
	IncludeModified  bool
	IgnoreUnmodified bool
}
