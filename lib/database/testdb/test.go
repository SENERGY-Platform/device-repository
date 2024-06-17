/*
 * Copyright 2022 InfAI (CC SES)
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

package testdb

import (
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

type DB struct {
	devices         map[string]models.Device
	hubs            map[string]models.Hub
	deviceTypes     map[string]models.DeviceType
	deviceGroups    map[string]models.DeviceGroup
	protocols       map[string]models.Protocol
	aspects         map[string]models.Aspect
	aspectNodes     map[string]models.AspectNode
	characteristics map[string]models.Characteristic
	concepts        map[string]models.Concept
	deviceClasses   map[string]models.DeviceClass
	functions       map[string]models.Function
	locations       map[string]models.Location
}

func NewTestDB() database.Database {
	return &DB{
		devices:         make(map[string]models.Device),
		hubs:            make(map[string]models.Hub),
		deviceTypes:     make(map[string]models.DeviceType),
		deviceGroups:    make(map[string]models.DeviceGroup),
		protocols:       make(map[string]models.Protocol),
		aspects:         make(map[string]models.Aspect),
		aspectNodes:     make(map[string]models.AspectNode),
		characteristics: make(map[string]models.Characteristic),
		concepts:        make(map[string]models.Concept),
		deviceClasses:   make(map[string]models.DeviceClass),
		functions:       make(map[string]models.Function),
		locations:       make(map[string]models.Location),
	}
}

func (db *DB) Disconnect() {}

func (db *DB) RunStartupMigrations(producer model.MigrationPublisher) error {
	return nil
}

func get[T any](id string, m map[string]T) (T, bool, error) {
	resp, ok := m[id]
	return resp, ok, nil
}

func set[T any](id string, m map[string]T, t T) error {
	m[id] = t
	return nil
}

func del[T any](id string, m map[string]T) error {
	delete(m, id)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
