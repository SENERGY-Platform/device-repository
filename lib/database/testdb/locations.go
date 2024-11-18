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

func (db *DB) SetLocation(_ context.Context, location models.Location) error {
	return set(location.Id, db.locations, location)
}
func (db *DB) RemoveLocation(_ context.Context, id string) error {
	return del(id, db.locations)
}
func (db *DB) GetLocation(_ context.Context, id string) (result models.Location, exists bool, err error) {
	return get(id, db.locations)
}

func (db *DB) ListLocations(ctx context.Context, options model.LocationListOptions) (locations []models.Location, total int64, err error) {
	locations = []models.Location{}
	var r *regexp.Regexp
	if options.Search != "" {
		r, err = regexp.Compile("(?i)" + regexp.QuoteMeta(options.Search))
		if err != nil {
			return nil, total, err
		}

	}
	for _, location := range db.locations {
		if options.Ids != nil && !slices.Contains(options.Ids, location.Id) {
			continue
		}
		if options.Search != "" && r != nil {
			if !r.MatchString(location.Name) {
				continue
			}
		}
		locations = append(locations, location)
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
	slices.SortFunc(locations, func(a, b models.Location) int {
		afield := a.Name
		bfield := b.Name
		if sortby == "id" {
			afield = a.Id
			bfield = b.Id
		}
		return strings.Compare(afield, bfield) * direction
	})

	total = int64(len(locations))
	if options.Limit > 0 || options.Offset > 0 {
		if options.Offset >= int64(len(locations)) {
			return []models.Location{}, total, nil
		}
		if (options.Limit + options.Offset) >= int64(len(locations)) {
			return locations[options.Offset:], total, nil
		}
		return locations[options.Offset : options.Limit+options.Offset], total, nil
	}

	return locations, total, nil
}
