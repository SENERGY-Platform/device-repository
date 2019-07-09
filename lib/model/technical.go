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

type Hub struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Hash           string   `json:"hash"`
	DeviceLocalIds []string `json:"device_local_ids"`
}

type Protocol struct {
	Id               string            `json:"id"`
	Name             string            `json:"name"`
	Handler          string            `json:"handler"`
	ProtocolSegments []ProtocolSegment `json:"protocol_segments"`
}

type ProtocolSegment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Content struct {
	Id                   string                `json:"id"`
	Variable             Variable              `json:"variable"`
	SerializationId      string                `json:"serialization_id"`
	SerializationOptions []SerializationOption `json:"serialization_options"`
	ProtocolSegmentId    string                `json:"protocol_segment_id"`
}

type SerializationOption struct {
	Id         string `json:"id"`
	Option     string `json:"option"`
	VariableId string `json:"variable_id"`
}

type Serialization struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
