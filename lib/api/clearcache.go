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

package api

import (
	"net/http"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
)

func init() {
	endpoints = append(endpoints, &ClearCacheEndpoints{})
}

type ClearCacheEndpoints struct{}

// Post godoc
// @Summary      clear the browser cache
// @Description  answers 204 with the header 'Clear-Site-Data: "cache"', which makes a browser discard its HTTP cache for the origin of this response. That is the origin of the platform API, so responses of all services behind it are affected. Needs no token: at logout it may have expired, and the response only affects the browser that receives it. To be called by a frontend before it redirects to the identity provider.
// @Success      204
// @Router       /clear-cache [POST]
func (this *ClearCacheEndpoints) Post(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("POST /clear-cache", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Clear-Site-Data", `"cache"`)
		writer.WriteHeader(http.StatusNoContent)
	})
}
