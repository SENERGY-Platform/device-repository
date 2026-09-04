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

package controller

import (
	"slices"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
)

//AspectId of a filter-criteria is deprecated in favor of AspectIds. It is an alias for a
//single element list, and it is folded into that list at the controller boundary, so that
//everything behind it evaluates AspectIds only. A criteria naming more than one aspect is
//an AND: the same content variable has to carry all of them.

// setFilterCriteriaAspectIds folds the deprecated AspectId of a query into AspectIds.
// The list is edited in place, because a query is never handed back to the caller.
func setFilterCriteriaAspectIds(criteria []model.FilterCriteria) {
	for i := range criteria {
		criteria[i].AspectIds = addAspectIdToAspectIds(criteria[i].AspectId, criteria[i].AspectIds)
	}
}

// SetDeviceGroupCriteriaAspectIdsOnWrite folds the deprecated AspectId of every stored
// criteria into AspectIds. Exported for the mgw mirror, which writes device-groups from a
// remote instance straight into the database and may receive them without AspectIds.
func SetDeviceGroupCriteriaAspectIdsOnWrite(deviceGroup *models.DeviceGroup) {
	setDeviceGroupFilterCriteriaAspectIds(deviceGroup.Criteria)
}

// setDeviceGroupFilterCriteriaAspectIds folds the deprecated AspectId of every criteria into
// AspectIds. The list is edited in place, like setFilterCriteriaAspectIds.
func setDeviceGroupFilterCriteriaAspectIds(criteria []models.DeviceGroupFilterCriteria) {
	for i := range criteria {
		criteria[i].AspectIds = addAspectIdToAspectIds(criteria[i].AspectId, criteria[i].AspectIds)
	}
}

// setDeviceGroupCriteriaAspectIdsOnRead makes both fields consistent for stored
// device-groups. AspectIds is filled from the deprecated AspectId first, because
// device-groups written before the migration carry only AspectId and everything behind a
// read evaluates AspectIds. AspectId is then set to the alphabetically first entry, to
// serve clients that do not know AspectIds yet.
func setDeviceGroupCriteriaAspectIdsOnRead(deviceGroup *models.DeviceGroup) {
	SetDeviceGroupCriteriaAspectIdsOnWrite(deviceGroup)
	for i := range deviceGroup.Criteria {
		if len(deviceGroup.Criteria[i].AspectIds) == 0 {
			continue
		}
		sorted := slices.Clone(deviceGroup.Criteria[i].AspectIds)
		slices.Sort(sorted)
		deviceGroup.Criteria[i].AspectId = sorted[0]
	}
}

func setDeviceGroupCriteriaAspectIdsOnReadList(deviceGroups []models.DeviceGroup) {
	for i := range deviceGroups {
		setDeviceGroupCriteriaAspectIdsOnRead(&deviceGroups[i])
	}
}

func addAspectIdToAspectIds(aspectId string, aspectIds []string) []string {
	if aspectId != "" && !slices.Contains(aspectIds, aspectId) {
		return append(aspectIds, aspectId)
	}
	return aspectIds
}

// pathOptionAspectNodes loads the aspect nodes a path option offers. An unknown aspect is
// tolerated here, the way the aspect-node of a path option always was.
func (this *Controller) pathOptionAspectNodes(aspectCache *map[string]models.AspectNode, aspectIds []string) (nodes []models.AspectNode, deprecated models.AspectNode, err error) {
	return aspectNodesWithDeprecated(aspectIds, func(aspectId string) (models.AspectNode, error) {
		return this.getAspectNodeForDeviceTypeSelectables(aspectCache, aspectId)
	})
}

// configurableAspectNodes loads the aspect nodes a configurable offers. An unknown aspect
// fails the request here, the way it always did for a configurable.
func (this *Controller) configurableAspectNodes(aspectIds []string) (nodes []models.AspectNode, deprecated models.AspectNode, err error) {
	return aspectNodesWithDeprecated(aspectIds, func(aspectId string) (models.AspectNode, error) {
		node, err, _ := this.GetAspectNode(aspectId)
		return node, err
	})
}

// aspectNodesWithDeprecated resolves aspect ids in a stable order and derives the deprecated
// single node from the result. AspectNode is the alias for a single element list, so it gets
// the node with the alphabetically first id; without any aspect it stays the unset node the
// answer carried before the list existed.
func aspectNodesWithDeprecated(aspectIds []string, load func(aspectId string) (models.AspectNode, error)) (nodes []models.AspectNode, deprecated models.AspectNode, err error) {
	for _, aspectId := range slices.Sorted(slices.Values(aspectIds)) {
		if aspectId == "" {
			continue
		}
		node, err := load(aspectId)
		if err != nil {
			return nil, deprecated, err
		}
		nodes = append(nodes, node)
	}
	if len(nodes) > 0 {
		deprecated = nodes[0]
	}
	return nodes, deprecated, nil
}

// deviceTypeCriteriaMatch pairs a device-type criteria row with the aspects of that row the
// query criteria asked for. The pairing has to be kept, because the rows of all query
// criteria are merged afterwards and the aspect of a path option comes from the query.
type deviceTypeCriteriaMatch struct {
	criteria         model.DeviceTypeCriteria
	matchedAspectIds []string
}

func (this *Controller) newDeviceTypeCriteriaMatch(aspectCache *map[string]models.AspectNode, row model.DeviceTypeCriteria, query model.FilterCriteria) (result deviceTypeCriteriaMatch, err error) {
	matched, err := this.matchedAspectIds(aspectCache, row.AspectIds, query.AspectIds)
	if err != nil {
		return result, err
	}
	return deviceTypeCriteriaMatch{criteria: row, matchedAspectIds: matched}, nil
}

// matchedAspectIds returns the aspects of a device-type criteria row that the query asked
// for. A row carries every aspect of its content variable, while a path option is offered
// per matched aspect, so the aspects the query did not name are dropped here. A queried
// aspect covers its descendants, like the database filter that selected the row.
func (this *Controller) matchedAspectIds(aspectCache *map[string]models.AspectNode, rowAspectIds []string, queryAspectIds []string) (result []string, err error) {
	if len(queryAspectIds) == 0 {
		return rowAspectIds, nil
	}
	for _, queried := range queryAspectIds {
		node, err := this.getAspectNodeForDeviceTypeSelectables(aspectCache, queried)
		if err != nil {
			return nil, err
		}
		for _, aspectId := range rowAspectIds {
			if aspectId == queried || slices.Contains(node.DescendentIds, aspectId) {
				if !slices.Contains(result, aspectId) {
					result = append(result, aspectId)
				}
			}
		}
	}
	return result, nil
}

// SetContentVariableAspectIdsOnWrite hands the write normalization of the content variable
// aspects to the startup migration, which rebuilds the derived device-type criteria from
// stored device-types and may not import this package. It is the method form of the
// package function of the same name.
func (this *Controller) SetContentVariableAspectIdsOnWrite(deviceType *models.DeviceType) {
	SetContentVariableAspectIdsOnWrite(deviceType)
}
