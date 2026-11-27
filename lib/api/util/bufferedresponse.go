/*
 * Copyright 2026 InfAI (CC SES)
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

package util

import (
	"bytes"
	"net/http"
)

// bufferedResponse holds a complete response in memory, so that a middleware can look at it before it is sent.
type bufferedResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newBufferedResponse() *bufferedResponse {
	return &bufferedResponse{header: http.Header{}, status: http.StatusOK}
}

func (this *bufferedResponse) Header() http.Header {
	return this.header
}

func (this *bufferedResponse) WriteHeader(status int) {
	this.status = status
}

func (this *bufferedResponse) Write(b []byte) (int, error) {
	return this.body.Write(b)
}

func (this *bufferedResponse) writeTo(w http.ResponseWriter) {
	for k, vv := range this.header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(this.status)
	_, _ = w.Write(this.body.Bytes())
}
