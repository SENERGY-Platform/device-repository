/*
 * Copyright 2025 InfAI (CC SES)
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
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"net/http"
)

func init() {
	endpoints = append(endpoints, &UsersEndpoints{})
}

type UsersEndpoints struct{}

// UserDelete godoc
// @Summary      delete user
// @Description  delete user; only admins may use this method
// @Tags         users
// @Security Bearer
// @Param        id path string true "User Id"
// @Success      200
// @Failure      400
// @Failure      500
// @Router       /users/{id} [DELETE]
func (this *AspectEndpoints) UserDelete(config configuration.Config, router *http.ServeMux, control Controller) {
	router.HandleFunc("DELETE /users/{id}", func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token := util.GetAuthToken(request)
		err, errCode := control.DeleteUser(token, id)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(true)
		if err != nil {
			config.GetLogger().Info("unable to encode response", "error", err.Error())
		}
		return
	})
}
