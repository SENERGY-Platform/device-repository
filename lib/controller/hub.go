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
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
	"runtime/debug"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadHub(id string, jwt jwt_http_router.Jwt) (result model.Hub, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	hub, exists, err := this.db.GetHub(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err := this.security.CheckBool(jwt, this.config.HubTopic, id, model.READ)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	return hub, nil, http.StatusOK
}

func (this *Controller) ValidateHub(hub model.Hub) (err error, code int) {
	if hub.Id == "" {
		return errors.New("missing hub id"), http.StatusBadRequest
	}
	if hub.Name == "" {
		return errors.New("missing hub name"), http.StatusBadRequest
	}
	for _, localId := range hub.DeviceLocalIds {
		//device exists?
		ctx, _ := getTimeoutContext()
		_, exists, err := this.db.GetDeviceByLocalId(ctx, localId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown device local id: " + localId), http.StatusBadRequest
		}

	}

	return nil, http.StatusOK
}

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetHub(hub model.Hub, owner string) (err error) {
	if hub.Id == "" {
		log.Println("ERROR: received hub without id")
		return nil
	}
	if err, _ := this.ValidateHub(hub); err != nil {
		log.Println("ERROR: ", err)
		debug.PrintStack()
		if hub.Name == "" {
			hub.Name = "generated-name"
		}
		hub.DeviceLocalIds = []string{}
		hub.Hash = ""
		return this.producer.PublishHub(hub)
	}
	hubIndex := map[string]model.Hub{}
	for _, lid := range hub.DeviceLocalIds {
		ctx, _ := getTimeoutContext()
		hubs, err := this.db.GetHubsByDeviceLocalId(ctx, lid)
		if err != nil {
			return err
		}
		for _, hub2 := range hubs {
			if hub2.Id != hub.Id {
				hubIndex[hub2.Id] = hub2
			}
		}
	}
	for _, lid := range hub.DeviceLocalIds {
		for _, hub2 := range hubIndex {
			hub2.DeviceLocalIds = filter(hub2.DeviceLocalIds, lid)
			hubIndex[hub2.Id] = hub2
		}
	}
	for _, hub2 := range hubIndex {
		err := this.producer.PublishHub(hub2)
		if err != nil {
			return err
		}
	}

	ctx, _ := getTimeoutContext()
	return this.db.SetHub(ctx, hub)
}

func (this *Controller) DeleteHub(id string) error {
	ctx, _ := getTimeoutContext()
	return this.db.RemoveHub(ctx, id)
}
