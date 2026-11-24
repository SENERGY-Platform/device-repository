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

package model

import (
	"errors"
	"time"

	"github.com/SENERGY-Platform/models/go/models"
)

// ErrUnsupportedSort is returned (wrapped) by ListThings if ThingListOptions.SortBy names a field that can not be sorted by.
var ErrUnsupportedSort = errors.New("unsupported sort field")

// ThingDeviceLinkCreate readings: from_value (and until_value) are in the unit the column is delivered in after the
// platform's characteristic conversion, i.e. kWh for energy and m³ for volume, not the raw device unit (e.g. Wh).
type ThingDeviceLinkCreate struct {
	DeviceId string                          `json:"device_id"`
	From     *time.Time                      `json:"from"`               // nil = now
	Until    *time.Time                      `json:"until"`              // nil = open
	Readings []models.ThingDeviceLinkReading `json:"readings,omitempty"` // only for meters; until_value only together with until
	Factor   *float64                        `json:"factor,omitempty"`   // only for meters; transformer ratio, nil = 1
}

type ThingDeviceLinkEnd struct {
	Until       *time.Time                  `json:"until"`                  // nil = now
	UntilValues []ThingDeviceLinkEndReading `json:"until_values,omitempty"` // readings at removal, each must name an existing reading of the link
}

// ThingDeviceLinkEndReading sets the reading at removal of the reading identified by ServiceId and Column.
// UntilValue is in the same unit as from_value: kWh for energy and m³ for volume, not the raw device unit (e.g. Wh).
type ThingDeviceLinkEndReading struct {
	ServiceId  string  `json:"service_id"`
	Column     string  `json:"column"`
	UntilValue float64 `json:"until_value"`
}

// ThingDeviceLinkSwap replaces the open link of a thing at the instant At: the link ends there with the readings
// at removal (UntilValues), and a link to DeviceId starts there with Readings and Factor. DeviceId may be the device
// of the ended link (counter reset, change of the factor) or another one. At may lie in the past, but must be after
// the start of the ended link; nil means now. Both changes are one write of the thing.
type ThingDeviceLinkSwap struct {
	At          *time.Time                      `json:"at"`                     // nil = now
	UntilValues []ThingDeviceLinkEndReading     `json:"until_values,omitempty"` // readings at removal of the ended link
	DeviceId    string                          `json:"device_id"`
	Readings    []models.ThingDeviceLinkReading `json:"readings,omitempty"` // readings at installation, only for meters; until_value is not allowed
	Factor      *float64                        `json:"factor,omitempty"`   // only for meters; nil = 1
}

// GraphNodeMove moves a node below another node from At on: the outgoing edges of the node end at At and one edge to
// ToNodeId with weight 100 starts there. With Until the move holds for [At, Until) only and the old edges are resumed.
type GraphNodeMove struct {
	At       *time.Time `json:"at"` // nil = now
	ToNodeId string     `json:"to_node_id"`
	Until    *time.Time `json:"until,omitempty"` // nil = open; must be after At
}

// GraphNodeEnd takes a node out of a graph from At on: the node and its edges end there and its children are
// rerouted to its parent from there. The history before At is not touched.
type GraphNodeEnd struct {
	At *time.Time `json:"at"` // nil = now
}
