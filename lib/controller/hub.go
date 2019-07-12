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
	"context"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

/////////////////////////
//		api
/////////////////////////

func (this *Controller) ReadHub(id string, jwt jwt_http_router.Jwt) (result model.Hub, err error, errCode int) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
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
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
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
		hub.DeviceLocalIds = []string{}
		hub.Hash = ""
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.SetHub(ctx, hub)
}

func (this *Controller) DeleteHub(id string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return this.db.RemoveHub(ctx, id)
}
