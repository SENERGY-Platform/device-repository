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

package testdb

import (
	"context"
	"maps"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

func (db *DB) SetGraph(ctx context.Context, graph models.Graph, syncHandler func(models.Graph) error) error {
	return set(graph.Id, db.graphs, graph, func(graph models.Graph) error {
		return syncHandler(graph)
	})
}

func (db *DB) RemoveGraph(ctx context.Context, id string, syncDeleteHandler func(models.Graph) error) error {
	return del(id, db.graphs, syncDeleteHandler)
}

func (db *DB) GetGraph(ctx context.Context, id string) (graph models.Graph, exists bool, err error) {
	return get(id, db.graphs)
}

func (db *DB) ListGraphs(ctx context.Context, listOptions model.GraphListOptions) (result []models.Graph, total int64, err error) {
	return iterToSlice(maps.Values(db.graphs)), int64(len(db.graphs)), nil
}

func (db *DB) RetryGraphSync(lockduration time.Duration, syncDeleteHandler func(models.Graph) error, syncHandler func(models.Graph) error) error {
	return nil
}

func (db *DB) DesyncUnknownGraphs(ctx context.Context, knownGraphs []string) (err error) {
	return nil
}
