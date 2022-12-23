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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"strconv"
)

func init() {
	endpoints = append(endpoints, AspectNodesEndpoints)
}

type AspectNodeQuery struct {
	Ids *[]string `json:"ids,omitempty"`
}

func AspectNodesEndpoints(config config.Config, control Controller, router *httprouter.Router) {
	resource := "/aspect-nodes"

	router.POST("/query"+resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		query := AspectNodeQuery{}
		err := json.NewDecoder(request.Body).Decode(&query)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		if query.Ids != nil {
			result, err, errCode := control.GetAspectNodesByIdList(*query.Ids)
			if err != nil {
				http.Error(writer, err.Error(), errCode)
				return
			}
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			err = json.NewEncoder(writer).Encode(result)
			if err != nil {
				log.Println("ERROR: unable to encode response", err)
			}
			return
		}
		http.Error(writer, "no known query content found", http.StatusBadRequest)
		return
	})

	router.GET(resource, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		var result []models.AspectNode
		var err error
		var errCode int

		function := request.URL.Query().Get("function")

		if function == "" {
			result, err, errCode = control.GetAspectNodes()
			if err != nil {
				http.Error(writer, err.Error(), errCode)
				return
			}
		} else {
			ancestors := false
			descendants := true
			ancestorsQuery := request.URL.Query().Get("ancestors")
			if ancestorsQuery != "" {
				ancestors, err = strconv.ParseBool(ancestorsQuery)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusBadRequest)
					return
				}
			}
			descendantsQuery := request.URL.Query().Get("descendants")
			if descendantsQuery != "" {
				descendants, err = strconv.ParseBool(descendantsQuery)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusBadRequest)
					return
				}
			}
			if function == "measuring-function" {
				result, err, errCode = control.GetAspectNodesWithMeasuringFunction(ancestors, descendants)
				if err != nil {
					http.Error(writer, err.Error(), errCode)
					return
				}
			}
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

	router.GET(resource+"/:id", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		result, err, errCode := control.GetAspectNode(id)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

	router.GET(resource+"/:id/measuring-functions", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		id := params.ByName("id")
		ancestors := false
		descendants := true
		var err error
		ancestorsQuery := request.URL.Query().Get("ancestors")
		if ancestorsQuery != "" {
			ancestors, err = strconv.ParseBool(ancestorsQuery)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		descendantsQuery := request.URL.Query().Get("descendants")
		if descendantsQuery != "" {
			descendants, err = strconv.ParseBool(descendantsQuery)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}
		result, err, errCode := control.GetAspectNodesMeasuringFunctions(id, ancestors, descendants)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(writer).Encode(result)
		if err != nil {
			log.Println("ERROR: unable to encode response", err)
		}
		return
	})

}
