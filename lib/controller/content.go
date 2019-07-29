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

package controller

import (
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"net/http"
)

func ValidateContent(content model.Content, protocol model.Protocol) (err error, code int) {
	if content.Id == "" {
		return errors.New("missing service id"), http.StatusBadRequest
	}
	if !content.Serialization.Valid() {
		return errors.New("unknown serialization " + string(content.Serialization)), http.StatusBadRequest
	}
	if content.ProtocolSegmentId == "" {
		return errors.New("missing protocol_segment_id"), http.StatusBadRequest
	}
	if !protocolContainsSegment(protocol, content.ProtocolSegmentId) {
		return errors.New("protocol_segment_id does not match to protocol"), http.StatusBadRequest
	}
	err, code = ValidateVariable(content.ContentVariable, content.Serialization)
	if err != nil {
		return err, code
	}
	return nil, http.StatusOK
}

func protocolContainsSegment(protocol model.Protocol, segmentId string) bool {
	for _, segment := range protocol.ProtocolSegments {
		if segment.Id == segmentId {
			return true
		}
	}
	return false
}
