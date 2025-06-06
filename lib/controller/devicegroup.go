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
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"log"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"
)

const FilterDevicesOfGroupByAccess = true

func (this *Controller) ListDeviceGroups(token string, options model.DeviceGroupListOptions) (result []models.DeviceGroup, total int64, err error, errCode int) {
	ids := []string{}
	permissionFlag := options.Permission
	if permissionFlag == models.UnsetPermissionFlag {
		permissionFlag = models.Read
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, total, err, http.StatusBadRequest
	}

	//check permissions
	if options.Ids == nil {
		if jwtToken.IsAdmin() {
			ids = nil //no auth check for admins -> no id filter
		} else {
			ids, err, _ = this.permissionsV2Client.ListAccessibleResourceIds(token, this.config.DeviceGroupTopic, client.ListOptions{}, client.Permission(permissionFlag))
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
			if len(ids) == 0 {
				ids = []string{}
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err, _ := this.permissionsV2Client.CheckMultiplePermissions(token, this.config.DeviceGroupTopic, options.Ids, client.Permission(permissionFlag))
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
		for id, ok := range idMap {
			if ok {
				ids = append(ids, id)
			}
		}
	}
	options.Ids = ids

	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListDeviceGroups(ctx, options)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}

	if options.FilterGenericDuplicateCriteria {
		for i, group := range result {
			group, err = DeviceGroupFilterGenericDuplicateCriteria(group, this.db)
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
			result[i] = group
		}

	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) ReadDeviceGroup(id string, token string, filterGenericDuplicateCriteria bool) (result models.DeviceGroup, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetDeviceGroup(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.DeviceGroupTopic, id, client.Read)
	if err != nil {
		result = models.DeviceGroup{}
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		result = models.DeviceGroup{}
		return result, errors.New("access denied"), http.StatusForbidden
	}

	//ref https://bitnify.atlassian.net/browse/SNRGY-3027
	if filterGenericDuplicateCriteria {
		result, err = DeviceGroupFilterGenericDuplicateCriteria(result, this.db)
		if err != nil {
			result = models.DeviceGroup{}
			return result, err, http.StatusInternalServerError
		}
	}

	if FilterDevicesOfGroupByAccess {
		return this.FilterDevicesOfGroupByAccess(token, result)
	} else {
		return result, nil, http.StatusOK
	}
}

func (this *Controller) FilterDevicesOfGroupByAccess(token string, group models.DeviceGroup) (result models.DeviceGroup, err error, code int) {
	if len(group.DeviceIds) == 0 {
		return group, nil, http.StatusOK
	}
	access, err, _ := this.permissionsV2Client.CheckMultiplePermissions(token, this.config.DeviceTopic, group.DeviceIds, client.Execute)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	result = group
	result.DeviceIds = []string{}
	for _, id := range group.DeviceIds {
		if access[id] {
			result.DeviceIds = append(result.DeviceIds, id)
		} else if this.config.Debug {
			log.Println("DEBUG: filtered " + id + " from result, because user lost execution access to the device")
		}
	}
	return result, nil, http.StatusOK
}

func (this *Controller) checkAccessToDevicesOfGroup(token string, group models.DeviceGroup) (err error, code int) {
	if len(group.DeviceIds) == 0 {
		return nil, http.StatusOK
	}
	access, err, _ := this.permissionsV2Client.CheckMultiplePermissions(token, this.config.DeviceTopic, group.DeviceIds, client.Execute)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	//looping one element of group.Devices is enough because ValidateDeviceGroup() ensures that every used device is referenced in each group.Devices element
	for _, id := range group.DeviceIds {
		if !access[id] {
			return errors.New("no execution access to device " + id), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) ValidateDeviceGroup(token string, group models.DeviceGroup) (err error, code int) {
	if group.Id == "" {
		return errors.New("missing device-group id"), http.StatusBadRequest
	}
	if group.Name == "" {
		return errors.New("missing device-group name"), http.StatusBadRequest
	}
	ctx, _ := getTimeoutContext()
	old, exists, err := this.db.GetDeviceGroup(ctx, group.Id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists && group.AutoGeneratedByDevice != "" {
		return errors.New("manually created device groups may not fill auto_generated_by_device"), http.StatusBadRequest
	}
	if exists && old.AutoGeneratedByDevice != group.AutoGeneratedByDevice {
		return fmt.Errorf("user may not change auto_generated_by_device (from '%v' to '%v')", old.AutoGeneratedByDevice, group.AutoGeneratedByDevice), http.StatusBadRequest
	}
	err, code = this.checkAccessToDevicesOfGroup(token, group)
	if err != nil {
		return err, code
	}
	return this.ValidateDeviceGroupSelection(group.Criteria, group.DeviceIds)
}

func (this *Controller) ValidateDeviceGroupSelection(criteria []models.DeviceGroupFilterCriteria, devices []string) (error, int) {
	deviceCache := map[string]models.Device{}
	deviceTypeCache := map[string]models.DeviceType{}
	deviceUsageCount := map[string]int{}
	for _, c := range criteria {
		deviceUsedInMapping := map[string]bool{}
		for _, deviceId := range devices {
			if deviceUsedInMapping[deviceId] {
				return errors.New("multiple uses of device-id " + deviceId + " for the same filter-criteria"), http.StatusBadRequest
			}
			deviceUsedInMapping[deviceId] = true
			deviceUsageCount[deviceId] = deviceUsageCount[deviceId] + 1
			err, code := this.selectionMatchesCriteria(&deviceCache, &deviceTypeCache, c, deviceId)
			if err != nil {
				return err, code
			}
		}
	}
	return nil, http.StatusOK
}

type AspectNodeProvider interface {
	ListAspectNodesByIdList(ctx context.Context, ids []string) ([]models.AspectNode, error)
}

// DeviceGroupFilterGenericDuplicateCriteria removes criteria without aspect, that are already present with an aspect
// ref: https://bitnify.atlassian.net/browse/SNRGY-3027
func DeviceGroupFilterGenericDuplicateCriteria(dg models.DeviceGroup, aspectNodeProvider AspectNodeProvider) (result models.DeviceGroup, err error) {
	result = dg

	//get used aspect ids
	aspectIds := []string{}
	for _, criteria := range result.Criteria {
		if criteria.AspectId != "" && !slices.Contains(aspectIds, criteria.AspectId) {
			aspectIds = append(aspectIds, criteria.AspectId)
		}
	}

	//get used aspect nodes
	aspectNodes, err := aspectNodeProvider.ListAspectNodesByIdList(context.Background(), aspectIds)
	if err != nil {
		return result, err
	}

	//prepare index of descendents of aspects
	descendents := map[string][]string{}
	for _, aspect := range aspectNodes {
		descendents[aspect.Id] = append(descendents[aspect.Id], aspect.DescendentIds...)
	}

	//function to check if candidate aspect is descendent of criteria aspect
	candidateUsesDescendentAspect := func(criteria models.DeviceGroupFilterCriteria, candidate models.DeviceGroupFilterCriteria) bool {
		if criteria.AspectId == "" && candidate.AspectId != "" {
			return true
		}
		if criteria.AspectId == candidate.AspectId {
			return false
		}
		return slices.Contains(descendents[criteria.AspectId], candidate.AspectId)
	}

	//function to check if the candidate is a more specialized variant of criteria
	isDuplicateCriteriaWithDescendentAspect := func(criteria models.DeviceGroupFilterCriteria, candidate models.DeviceGroupFilterCriteria) bool {
		return candidateUsesDescendentAspect(criteria, candidate) &&
			candidate.FunctionId == criteria.FunctionId &&
			candidate.DeviceClassId == criteria.DeviceClassId &&
			candidate.Interaction == criteria.Interaction
	}

	//filter criteria where more specialized variants exist
	newCriteriaList := []models.DeviceGroupFilterCriteria{}
	for _, criteria := range result.Criteria {
		duplicateWithAspectExists := slices.ContainsFunc(result.Criteria, func(element models.DeviceGroupFilterCriteria) bool {
			return isDuplicateCriteriaWithDescendentAspect(criteria, element)
		})
		if !duplicateWithAspectExists {
			newCriteriaList = append(newCriteriaList, criteria)
			continue
		}
	}
	result.Criteria = newCriteriaList
	result.SetShortCriteria()
	return result, nil
}

func (this *Controller) selectionMatchesCriteria(
	dcache *map[string]models.Device,
	dtcache *map[string]models.DeviceType,
	criteria models.DeviceGroupFilterCriteria,
	deviceId string) (err error, code int) {

	ctx, _ := getTimeoutContext()
	var exists bool

	var aspectNode models.AspectNode
	if criteria.AspectId != "" {
		aspectNode, exists, err = this.db.GetAspectNode(ctx, criteria.AspectId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return errors.New("unknown aspect-node-id: " + criteria.AspectId), http.StatusBadRequest
		}
	}

	device, ok := (*dcache)[deviceId]
	if !ok {
		temp, err, code := this.readDevice(deviceId, true)
		if err != nil {
			return fmt.Errorf("unable to read device %v: %w", deviceId, err), code
		}
		device = temp.Device
		(*dcache)[deviceId] = device
	}

	deviceType, ok := (*dtcache)[device.DeviceTypeId]
	if !ok {
		deviceType, err, code = this.readDeviceType(device.DeviceTypeId)
		if err != nil {
			return fmt.Errorf("unable to read device-type %v: %w", device.DeviceTypeId, err), code
		}
		(*dtcache)[device.DeviceTypeId] = deviceType
	}

	deviceClassMatches := criteria.DeviceClassId == "" || criteria.DeviceClassId == deviceType.DeviceClassId
	if !deviceClassMatches {
		return errors.New("device " + deviceId + " does not match device-class of filter-criteria"), http.StatusBadRequest
	}

	serviceMatches := false
	for _, service := range deviceType.Services {
		interactionMatches := service.Interaction == criteria.Interaction
		if service.Interaction == models.EVENT_AND_REQUEST {
			interactionMatches = true
		}
		contentMatches := false
		for _, content := range service.Inputs {
			if contentVariableContainsCriteria(content.ContentVariable, criteria, aspectNode) {
				contentMatches = true
				break
			}
		}
		for _, content := range service.Outputs {
			if contentVariableContainsCriteria(content.ContentVariable, criteria, aspectNode) {
				contentMatches = true
				break
			}
		}
		if interactionMatches && contentMatches {
			serviceMatches = true
			break
		}
	}
	if !serviceMatches {
		return errors.New("no service of the device " + deviceId + " matches filter-criteria"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func contentVariableContainsCriteria(variable models.ContentVariable, criteria models.DeviceGroupFilterCriteria, aspectNode models.AspectNode) bool {
	if variable.FunctionId == criteria.FunctionId &&
		(criteria.AspectId == "" ||
			variable.AspectId == criteria.AspectId ||
			listContains(aspectNode.DescendentIds, variable.AspectId)) {
		return true
	}
	for _, sub := range variable.SubContentVariables {
		if contentVariableContainsCriteria(sub, criteria, aspectNode) {
			return true
		}
	}
	return false
}

func listContains(list []string, search string) bool {
	for _, element := range list {
		if element == search {
			return true
		}
	}
	return false
}

func (this *Controller) ValidateDeviceGroupDelete(token string, id string) (err error, code int) {
	ctx, _ := getTimeoutContext()
	dg, exists, err := this.db.GetDeviceGroup(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if exists {
		ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.DeviceGroupTopic, id, client.Administrate)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !ok {
			return errors.New("access denied"), http.StatusForbidden
		}
	}
	if exists && dg.AutoGeneratedByDevice != "" {
		return errors.New("device-group is auto generated by device " + dg.AutoGeneratedByDevice), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

func (this *Controller) SetDeviceGroup(token string, dg models.DeviceGroup) (result models.DeviceGroup, err error, errCode int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	dg.GenerateId()
	dg.SetShortCriteria()
	if !this.config.DisableStrictValidationForTesting {
		dg.DeviceIds, err = this.filterInvalidDeviceIds(token, dg.DeviceIds, "r")
		if err != nil {
			return dg, err, http.StatusInternalServerError
		}
	}

	ctx, _ := getTimeoutContext()
	old, exists, err := this.db.GetDeviceGroup(ctx, dg.Id)
	if err != nil {
		return dg, err, http.StatusInternalServerError
	}

	if exists && !jwtToken.IsAdmin() && !this.config.DisableStrictValidationForTesting {
		ok, err, code := this.permissionsV2Client.CheckPermission(token, this.config.DeviceGroupTopic, dg.Id, client.Write)
		if err != nil {
			debug.PrintStack()
			return dg, err, code
		}
		if !ok {
			return dg, errors.New("access denied"), http.StatusForbidden
		}
	}

	if dg.AutoGeneratedByDevice == "" && exists {
		dg.AutoGeneratedByDevice = old.AutoGeneratedByDevice
	}

	if !this.config.DisableStrictValidationForTesting {
		err, code := this.ValidateDeviceGroup(token, dg)
		if err != nil {
			debug.PrintStack()
			return dg, err, code
		}
	}

	err = this.setDeviceGroup(dg, jwtToken.GetUserId())
	if err != nil {
		debug.PrintStack()
		return dg, err, http.StatusInternalServerError
	}

	return dg, nil, http.StatusOK
}

func (this *Controller) filterInvalidDeviceIds(token string, ids []string, rights string) (result []string, err error) {
	deviceIsAccessible, err, _ := this.permissionCheckForDeviceList(token, ids, rights)
	if err != nil {
		return result, err
	}
	result = []string{}
	for _, id := range ids {
		if deviceIsAccessible[id] {
			result = append(result, id)
		} else {
			log.Println("WARNING: remove device from device-group because its inaccessible", id)
		}
	}
	return result, nil
}

func (this *Controller) permissionCheckForDeviceList(token string, ids []string, rights string) (result map[string]bool, err error, code int) {
	ids = append(ids, removeIdModifiers(ids)...)
	ids = removeDuplicates(ids)
	permissions, err := permissionListFromString(rights)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return this.permissionsV2Client.CheckMultiplePermissions(token, this.config.DeviceTopic, ids, permissions...)
}

func permissionListFromString(str string) (result client.PermissionList, err error) {
	for _, p := range str {
		switch client.Permission(p) {
		case client.Read, client.Write, client.Execute, client.Administrate:
			result = append(result, client.Permission(p))
		default:
			return result, fmt.Errorf("unknown permission '%v'", p)
		}
	}
	return result, nil
}

func (this *Controller) setDeviceGroupSyncHandler(dg models.DeviceGroup, user string) error {
	if user != "" {
		err := this.EnsureInitialRights(this.config.DeviceGroupTopic, dg.Id, user)
		if err != nil {
			return err
		}
	}
	return this.publisher.PublishDeviceGroup(dg)
}

func (this *Controller) setDeviceGroup(deviceGroup models.DeviceGroup, owner string) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.SetDeviceGroup(ctx, deviceGroup, this.setDeviceGroupSyncHandler, owner)
}

func (this *Controller) DeleteDeviceGroup(token string, id string) (err error, code int) {
	err = preventIdModifier(id)
	if err != nil {
		return err, http.StatusBadRequest
	}
	ctx, _ := getTimeoutContext()
	_, exists, err := this.db.GetDeviceGroup(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return nil, http.StatusOK
	}
	ok, err, code := this.permissionsV2Client.CheckPermission(token, this.config.DeviceGroupTopic, id, client.Administrate)
	if err != nil {
		return err, code
	}
	if !ok {
		return errors.New("access denied"), http.StatusForbidden
	}

	err, code = this.ValidateDeviceGroupDelete(token, id)
	if err != nil {
		return err, code
	}

	err = this.deleteDeviceGroup(id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteDeviceGroupSyncHandler(dg models.DeviceGroup) (err error) {
	err = this.RemoveRights(this.config.DeviceGroupTopic, dg.Id)
	if err != nil {
		return err
	}
	return this.publisher.PublishDeviceGroupDelete(dg.Id)
}

func (this *Controller) deleteDeviceGroup(id string) error {
	ctx, _ := getTimeoutContext()
	err := this.db.RemoveDeviceGroup(ctx, id, this.deleteDeviceGroupSyncHandler)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) UpdateDeviceGroupCriteria(dg models.DeviceGroup) (err error) {
	user, exists, err := this.db.GetDeviceGroupSyncUser(context.Background(), dg.Id)
	if err != nil {
		return err
	}
	if !exists {
		log.Println("WARNING: tried to update unknown device-group criteria")
		return nil
	}
	dg.Criteria, err, _ = this.GetDeviceGroupCriteria(dg.DeviceIds)
	if err != nil {
		return err
	}
	slices.SortFunc(dg.Criteria, func(a, b models.DeviceGroupFilterCriteria) int {
		return strings.Compare(a.Short(), b.Short())
	})
	dg.SetShortCriteria()
	return this.setDeviceGroup(dg, user)
}
