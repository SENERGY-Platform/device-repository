/*
 *
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *
 */

package api

import (
	"encoding/json"
	"net/http"

	"github.com/SENERGY-Platform/device-repository/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
)

func init() {
	endpoints = append(endpoints, &LastUpdateTimestampsEndpoints{})
}

type LastUpdateTimestampsEndpoints struct{}

// Get godoc
// @Summary      last update timestamps
// @Description  list last update timestamps for a user
// @Produce      json
// @Security Bearer
// @Param        user_id query string false "admins may explicitly set the userId which is by default inferred from the jwt"
// @Param        sort query string false "default name.asc"
// @Param        ids query string false "filter; ignores limit/offset; comma-seperated list"
// @Success      200 {array}  model.LastUpdateTimestamp
// @Failure      400
// @Failure      500
// @Router       /last-update-timestamps [GET]
func (this *LastUpdateTimestampsEndpoints) Get(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("GET /last-update-timestamps", func(writer http.ResponseWriter, request *http.Request) {
		userId := request.URL.Query().Get("user_id")
		result, err, errCode := control.GetLastUpdateTimestamps(util.GetAuthToken(request), userId)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}
