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
	"github.com/SENERGY-Platform/models/go/models"
	"golang.org/x/exp/maps"
)

func (db *DB) AddAspectNode(_ context.Context, node models.AspectNode) error {
	return set(node.Id, db.aspectNodes, node)
}
func (db *DB) RemoveAspectNodesByRootId(_ context.Context, id string) error {
	panic("not implemented")
}

func (db *DB) GetAspectNode(_ context.Context, id string) (result models.AspectNode, exists bool, err error) {
	return get(id, db.aspectNodes)
}
func (db *DB) ListAllAspectNodes(_ context.Context) ([]models.AspectNode, error) {
	return maps.Values(db.aspectNodes), nil
}
func (db *DB) ListAspectNodesWithMeasuringFunction(_ context.Context, ancestors bool, descendants bool) ([]models.AspectNode, error) {
	panic("not implemented")
}
func (db *DB) ListAspectNodesByIdList(_ context.Context, ids []string) (result []models.AspectNode, err error) {
	for _, node := range db.aspectNodes {
		for _, id := range ids {
			if node.Id == id {
				result = append(result, node)
				break
			}
		}
	}
	return
}

func (db *DB) SetAspectNode(_ context.Context, node models.AspectNode) error {
	return set(node.Id, db.aspectNodes, node)
}
