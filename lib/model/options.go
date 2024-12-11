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
	LocalIds        []string                //filter; in combination with owner; fills ids filter; comma-seperated list; ignored if LocalIds == nil; LocalIds == []string{} will return an empty list;
	Owner           string                  //used in combination with local_ids to fill ids filter; defaults to requesting user
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

type LocationListOptions struct {
	Ids        []string //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	Search     string
	Limit      int64                 //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset     int64                 //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy     string                //default name.asc
	Permission models.PermissionFlag //defaults to read
}

type FunctionListOptions struct {
	Ids     []string //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	RdfType string   // model.SES_ONTOLOGY_CONTROLLING_FUNCTION || model.SES_ONTOLOGY_MEASURING_FUNCTION
	Search  string
	Limit   int64  //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset  int64  //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy  string //default name.asc
}

type AspectListOptions struct {
	Ids    []string //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	Search string
	Limit  int64  //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset int64  //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy string //default name.asc
}

type CharacteristicListOptions struct {
	Ids    []string //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	Search string
	Limit  int64  //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset int64  //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy string //default name.asc
}

type ConceptListOptions struct {
	Ids    []string //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	Search string
	Limit  int64  //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset int64  //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy string //default name.asc
}

type DeviceClassListOptions struct {
	Ids    []string //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	Search string
	Limit  int64  //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset int64  //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy string //default name.asc
}

type ExtendedDeviceListOptions struct {
	Ids             []string                //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	LocalIds        []string                //filter; in combination with owner; fills ids filter; comma-seperated list; ignored if LocalIds == nil; LocalIds == []string{} will return an empty list;
	Owner           string                  //used in combination with local_ids to fill ids filter; defaults to requesting user
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
		Owner:           this.Owner,
		LocalIds:        this.LocalIds,
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
	LocalDeviceId   string                //filter; list hubs if they contain the device-id
	OwnerId         string                //only used in combination with LocalDeviceId; defaults to requesting user
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
	ProtocolIds      []string
	IncludeModified  bool
	IgnoreUnmodified bool
}

type DeviceGroupListOptions struct {
	Ids                            []string //filter; ignores limit/offset if Ids != nil; ignored if Ids == nil; Ids == []string{} will return an empty list;
	Search                         string
	Limit                          int64                 //default 100, will be ignored if 'ids' is set (Ids != nil)
	Offset                         int64                 //default 0, will be ignored if 'ids' is set (Ids != nil)
	SortBy                         string                //default name.asc
	AttributeKeys                  []string              //filter; ignored if nil; AttributeKeys and AttributeValues are independently evaluated, needs local filtering if a search like "attr1"="value1" is needed
	AttributeValues                []string              //filter; ignored if nil; AttributeKeys and AttributeValues are independently evaluated, needs local filtering if a search like "attr1"="value1" is needed
	Criteria                       []FilterCriteria      //filter; ignored if nil
	Permission                     models.PermissionFlag //defaults to read
	IgnoreGenerated                bool                  //remove generated groups from result
	FilterGenericDuplicateCriteria bool
}
