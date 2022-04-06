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
	"log"
)

type Security interface {
	CheckBool(token string, kind string, id string, action model.AuthAction) (allowed bool, err error)
	CheckMultiple(token string, kind string, ids []string, action model.AuthAction) (map[string]bool, error)
}

type Producer interface {
	PublishDeviceDelete(id string, owner string) error
	PublishHub(hub model.Hub) (err error)
	PublishAspectDelete(id string, owner string) error
	PublishAspectUpdate(aspect model.Aspect, owner string) error
}

type ErrorProducer struct{}

func (this ErrorProducer) PublishAspectDelete(id string, owner string) (err error) {
	err = errors.New("no producer usage expected")
	log.Println("ERROR:", err)
	return err
}

func (this ErrorProducer) PublishAspectUpdate(aspect model.Aspect, owner string) (err error) {
	err = errors.New("no producer usage expected")
	log.Println("ERROR:", err)
	return err
}

func (this ErrorProducer) PublishDeviceDelete(id string, owner string) (err error) {
	err = errors.New("no producer usage expected")
	log.Println("ERROR:", err)
	return err
}

func (this ErrorProducer) PublishHub(hub model.Hub) (err error) {
	err = errors.New("no producer usage expected")
	log.Println("ERROR:", err)
	return err
}
