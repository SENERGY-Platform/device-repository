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
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"regexp"
	"slices"
	"strings"
)

func (db *DB) GetHub(_ context.Context, id string) (hub model.HubWithConnectionState, exists bool, err error) {
	return get(id, db.hubs)
}
func (db *DB) SetHub(_ context.Context, hub model.HubWithConnectionState) error {
	return set(hub.Id, db.hubs, hub)
}
func (db *DB) RemoveHub(_ context.Context, id string) error {
	return del(id, db.hubs)
}
func (db *DB) GetHubsByDeviceId(_ context.Context, id string) (hubs []model.HubWithConnectionState, err error) {
	for i := range db.hubs {
		for j := range db.hubs[i].DeviceIds {
			if db.hubs[i].DeviceLocalIds[j] == id {
				hubs = append(hubs, db.hubs[i])
				break
			}
		}
	}
	return
}

func (db *DB) ListHubs(ctx context.Context, options model.HubListOptions) (hubs []model.HubWithConnectionState, err error) {
	hubs = []model.HubWithConnectionState{}
	var r *regexp.Regexp
	if options.Search != "" {
		r, err = regexp.Compile("(?i)" + regexp.QuoteMeta(options.Search))
		if err != nil {
			return nil, err
		}

	}
	for _, hub := range db.hubs {
		if options.Ids != nil && !slices.Contains(options.Ids, hub.Id) {
			continue
		}
		if options.ConnectionState != nil && *options.ConnectionState != hub.ConnectionState {
			continue
		}
		if options.Search != "" && r != nil {
			if !r.MatchString(hub.Name) {
				continue
			}
		}
		hubs = append(hubs, hub)
	}
	if options.SortBy == "" {
		options.SortBy = "name.asc"
	}
	sortby := options.SortBy
	sortby = strings.TrimSuffix(sortby, ".asc")
	sortby = strings.TrimSuffix(sortby, ".desc")

	direction := 1
	if strings.HasSuffix(options.SortBy, ".desc") {
		direction = -1
	}
	slices.SortFunc(hubs, func(a, b model.HubWithConnectionState) int {
		afield := a.Name
		bfield := b.Name
		if sortby == "id" {
			afield = a.Id
			bfield = b.Id
		}
		return strings.Compare(afield, bfield) * direction
	})

	if options.Limit > 0 || options.Offset > 0 {
		if options.Offset >= int64(len(hubs)) {
			return []model.HubWithConnectionState{}, nil
		}
		if (options.Limit + options.Offset) >= int64(len(hubs)) {
			return hubs[options.Offset:], nil
		}
		return hubs[options.Offset : options.Limit+options.Offset], nil
	}

	return hubs, nil
}

func (db *DB) SetHubConnectionState(ctx context.Context, id string, state models.ConnectionState) error {
	hub, ok := db.hubs[id]
	if !ok {
		return nil
	}
	hub.ConnectionState = state
	return db.SetHub(ctx, hub)
}
