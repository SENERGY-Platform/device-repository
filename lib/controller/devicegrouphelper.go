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
	"errors"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/SENERGY-Platform/device-repository/v2/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/v2/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	permissions "github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

// DeviceGroupHelper answers what a device-group holding the given devices would look like: the
// criteria it would have, and the devices that may still be added to it.
func (this *Controller) DeviceGroupHelper(token string, deviceIds []string, options model.DeviceGroupHelperOptions) (result model.DeviceGroupHelperResult, err error, code int) {
	if options.Limit <= 0 {
		options.Limit = 100
	}
	//the criteria are read from the database without a permission check, so the ids of the
	//request are checked here - otherwise the endpoint would answer what a device measures to
	//anyone able to guess its id
	err, code = this.checkDeviceGroupHelperDeviceIds(token, deviceIds)
	if err != nil {
		return result, err, code
	}
	result.Criteria, err, code = this.GetDeviceGroupCriteria(deviceIds)
	if err != nil {
		return result, err, code
	}
	result.Options, err, code = this.getDeviceGroupHelperOptions(token, deviceIds, result.Criteria, options)
	if err != nil {
		return result, err, code
	}
	return result, nil, http.StatusOK
}

func (this *Controller) checkDeviceGroupHelperDeviceIds(token string, deviceIds []string) (err error, code int) {
	if len(deviceIds) == 0 {
		return nil, http.StatusOK
	}
	access, err, _ := this.permissionsV2Client.CheckMultiplePermissions(token, this.config.DeviceTopic, deviceIds, permissions.Permission(model.READ))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, id := range deviceIds {
		if !access[id] {
			return errors.New("access denied"), http.StatusForbidden
		}
	}
	return nil, http.StatusOK
}

// deviceGroupHelperEffect is what taking a device would do to the group. It depends only on the
// device-type, so it is computed once per device-type and reused for every device of it.
type deviceGroupHelperEffect struct {
	removes   []models.DeviceGroupFilterCriteria
	maintains bool
}

func (this *Controller) getDeviceGroupHelperOptions(token string, currentDeviceIds []string, criteria []models.DeviceGroupFilterCriteria, options model.DeviceGroupHelperOptions) (result []model.DeviceGroupOption, err error, code int) {
	result = []model.DeviceGroupOption{}
	blocked := blockedFunctionSet(options.FunctionBlockList)
	devices, err, code := this.getDeviceGroupHelperOptionDevices(token, currentDeviceIds, criteria, blocked, options)
	if err != nil {
		return result, err, code
	}

	//devices of the same device-type produce the same criteria, and a listing holds many of them
	effects := map[string]deviceGroupHelperEffect{}
	for _, device := range devices {
		effect, cached := effects[device.DeviceTypeId]
		if !cached {
			removes, keeps, deviceCriteria, err, code := this.getDeviceGroupHelperCriteriaSplit(criteria, device.Device)
			if err != nil {
				return result, err, code
			}
			effect = deviceGroupHelperEffect{removes: removes}
			if len(currentDeviceIds) == 0 {
				//without a group yet it is enough that the device brings a criteria of its own
				effect.maintains = containsCountingFunction(deviceCriteria, blocked)
			} else {
				//a device keeps the group usable if at least one criteria survives that counts
				effect.maintains = containsCountingFunction(keeps, blocked)
			}
			effects[device.DeviceTypeId] = effect
		}
		if options.MaintainsGroupUsability && !effect.maintains {
			continue
		}
		result = append(result, model.DeviceGroupOption{
			Device:                  device,
			RemovesCriteria:         effect.removes,
			MaintainsGroupUsability: effect.maintains,
		})
	}
	return result, nil, http.StatusOK
}

// getDeviceGroupHelperCriteriaSplit divides the criteria of the group into the ones the device
// would take away - those it does not answer itself - and the ones that survive. A criteria
// falls away with its function, so the check is the same Short() comparison
// GetDeviceGroupCriteria intersects by. The function block list does not enter here:
// RemovesCriteria reports what actually falls away, blocked or not.
func (this *Controller) getDeviceGroupHelperCriteriaSplit(currentCriteria []models.DeviceGroupFilterCriteria, device models.Device) (removes []models.DeviceGroupFilterCriteria, keeps []models.DeviceGroupFilterCriteria, deviceCriteria []models.DeviceGroupFilterCriteria, err error, code int) {
	removes = []models.DeviceGroupFilterCriteria{}
	keeps = []models.DeviceGroupFilterCriteria{}
	deviceCriteria, err, code = this.getDeviceGroupCriteriaOfDevice(device)
	if err != nil {
		return removes, keeps, deviceCriteria, err, code
	}
	deviceCriteriaSet := map[string]bool{}
	for _, criteria := range deviceCriteria {
		deviceCriteriaSet[criteria.Short()] = true
	}
	for _, criteria := range currentCriteria {
		if deviceCriteriaSet[criteria.Short()] {
			keeps = append(keeps, criteria)
		} else {
			removes = append(removes, criteria)
		}
	}
	return removes, keeps, deviceCriteria, nil, http.StatusOK
}

// blockedFunctionSet reads the function block list. A blocked function does not count towards
// keeping a group usable: a caller building a group for a purpose says with it which functions
// are too common to make a device a meaningful member - a device that answers only those adds
// nothing to the group, even though it technically leaves a criteria standing.
func blockedFunctionSet(functionBlockList []string) map[string]bool {
	result := map[string]bool{}
	for _, functionId := range functionBlockList {
		functionId = strings.TrimSpace(functionId)
		if functionId != "" {
			result[functionId] = true
		}
	}
	return result
}

func containsCountingFunction(criteria []models.DeviceGroupFilterCriteria, blocked map[string]bool) bool {
	return slices.ContainsFunc(criteria, func(c models.DeviceGroupFilterCriteria) bool {
		return !blocked[c.FunctionId]
	})
}

// getDeviceGroupHelperOptionDevices lists the devices that may be offered for the group. Both
// unmodified devices and the modified variants a device-type with service-groups produces are
// candidates - a device-group may hold either.
func (this *Controller) getDeviceGroupHelperOptionDevices(token string, currentDeviceIds []string, criteria []models.DeviceGroupFilterCriteria, blocked map[string]bool, options model.DeviceGroupHelperOptions) (result []models.ExtendedDevice, err error, code int) {
	//when only usable options are wanted, the device-types that answer none of the counting
	//criteria are dropped before paging, so that the limit counts options the caller actually
	//gets. With every criteria blocked the set is empty and no device is offered, which is the
	//honest answer: nothing can make this group meaningful under the caller's own rule.
	var validDeviceTypeIds []string //nil means: no device-type filter
	if options.MaintainsGroupUsability && len(criteria) > 0 {
		validDeviceTypeIds, err = this.getDeviceTypeIdsForAnyDeviceGroupCriteria(criteria, blocked)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
	}

	unmodified, err, code := this.getDeviceGroupHelperUnmodifiedOptionDevices(token, currentDeviceIds, validDeviceTypeIds, options)
	if err != nil {
		return result, err, code
	}
	modified, err, code := this.getDeviceGroupHelperModifiedOptionDevices(token, currentDeviceIds, validDeviceTypeIds, options)
	if err != nil {
		return result, err, code
	}
	devices := append(unmodified, modified...)

	//a device already in the group is no option, and neither is the unmodified device behind a
	//modified group member
	blockedDeviceIds := map[string]bool{}
	for _, id := range currentDeviceIds {
		blockedDeviceIds[id] = true
		pureId, modifier := idmodifier.SplitModifier(id)
		if len(modifier) > 0 {
			blockedDeviceIds[pureId] = true
		}
	}

	result = []models.ExtendedDevice{}
	seen := map[string]bool{}
	for _, device := range devices {
		if blockedDeviceIds[device.Id] || seen[device.Id] {
			continue
		}
		seen[device.Id] = true
		result = append(result, device)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return extendedDeviceDisplayName(result[i]) < extendedDeviceDisplayName(result[j])
	})
	return result, nil, http.StatusOK
}

func (this *Controller) getDeviceGroupHelperUnmodifiedOptionDevices(token string, currentDeviceIds []string, validDeviceTypeIds []string, options model.DeviceGroupHelperOptions) (result []models.ExtendedDevice, err error, code int) {
	var deviceTypeIds []string //nil means: no device-type filter
	if validDeviceTypeIds != nil {
		deviceTypeIds = []string{}
		for _, id := range validDeviceTypeIds {
			if pureId, modifier := idmodifier.SplitModifier(id); len(modifier) == 0 {
				deviceTypeIds = append(deviceTypeIds, pureId)
			}
		}
	}
	result, _, err, code = this.ListExtendedDevices(token, model.ExtendedDeviceListOptions{
		DeviceTypeIds: deviceTypeIds,
		Search:        options.Search,
		//the devices already in the group are filtered out afterwards, so the page is read
		//large enough to still hold the requested number of options
		Limit:  options.Limit + int64(len(currentDeviceIds)),
		Offset: options.Offset,
		SortBy: "name.asc",
	})
	return result, err, code
}

// getDeviceGroupHelperModifiedOptionDevices lists the modified variants of the candidate
// devices. A modified device is not stored; its id is the id of the stored device joined with
// the modifier of a modified device-type id, and reading it back applies the modifier.
func (this *Controller) getDeviceGroupHelperModifiedOptionDevices(token string, currentDeviceIds []string, validDeviceTypeIds []string, options model.DeviceGroupHelperOptions) (result []models.ExtendedDevice, err error, code int) {
	result = []models.ExtendedDevice{}
	if validDeviceTypeIds == nil {
		//no criteria filter: every modified device-type is a candidate
		validDeviceTypeIds, err = this.getModifiedDeviceTypeIds()
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
	}

	//one device-type may carry several service-groups, so it produces one option per modifier
	pureDeviceTypeIds := []string{}
	modifiersByPureDeviceTypeId := map[string][]map[string][]string{}
	for _, deviceTypeId := range validDeviceTypeIds {
		pureId, modifier := idmodifier.SplitModifier(deviceTypeId)
		if len(modifier) == 0 {
			continue
		}
		if !slices.Contains(pureDeviceTypeIds, pureId) {
			pureDeviceTypeIds = append(pureDeviceTypeIds, pureId)
		}
		modifiersByPureDeviceTypeId[pureId] = append(modifiersByPureDeviceTypeId[pureId], modifier)
	}
	if len(pureDeviceTypeIds) == 0 {
		return result, nil, http.StatusOK
	}

	devices, err, code := this.ListDevices(token, model.DeviceListOptions{
		DeviceTypeIds: pureDeviceTypeIds,
		Search:        options.Search,
		Limit:         options.Limit + int64(len(currentDeviceIds)),
		Offset:        options.Offset,
		SortBy:        "name.asc",
	})
	if err != nil {
		return result, err, code
	}

	modifiedDeviceIds := []string{}
	for _, device := range devices {
		for _, modifier := range modifiersByPureDeviceTypeId[device.DeviceTypeId] {
			modifiedDeviceIds = append(modifiedDeviceIds, device.Id+idmodifier.Seperator+idmodifier.EncodeModifierParameter(modifier))
		}
	}
	if len(modifiedDeviceIds) == 0 {
		return result, nil, http.StatusOK
	}

	result, _, err, code = this.ListExtendedDevices(token, model.ExtendedDeviceListOptions{
		Ids:    modifiedDeviceIds,
		SortBy: "name.asc",
	})
	return result, err, code
}

// getDeviceTypeIdsForAnyDeviceGroupCriteria lists the device-types answering at least one of
// the criteria whose function counts, modified ids included. That is exactly the set of
// device-types whose devices leave the group with a criteria worth having, so it is the
// pre-filter for the usable options.
func (this *Controller) getDeviceTypeIdsForAnyDeviceGroupCriteria(criteria []models.DeviceGroupFilterCriteria, blocked map[string]bool) (result []string, err error) {
	result = []string{}
	seen := map[string]bool{}
	ctx, _ := getTimeoutContext()
	for _, c := range criteria {
		if blocked[c.FunctionId] {
			continue
		}
		ids, err := this.db.GetDeviceTypeIdsByFilterCriteriaV2(ctx, []model.FilterCriteria{{
			Interaction:   c.Interaction,
			FunctionId:    c.FunctionId,
			DeviceClassId: c.DeviceClassId,
			AspectIds:     addAspectIdToAspectIds(c.AspectId, c.AspectIds),
		}}, true)
		if err != nil {
			return result, err
		}
		for _, id := range ids {
			idStr, ok := id.(string)
			if ok && !seen[idStr] {
				seen[idStr] = true
				result = append(result, idStr)
			}
		}
	}
	return result, nil
}

// getModifiedDeviceTypeIds lists every modified device-type id. The empty criteria matches
// every row of the device-type criteria collection, which is where the modified ids live.
func (this *Controller) getModifiedDeviceTypeIds() (result []string, err error) {
	result = []string{}
	ctx, _ := getTimeoutContext()
	ids, err := this.db.GetDeviceTypeIdsByFilterCriteriaV2(ctx, []model.FilterCriteria{{}}, true)
	if err != nil {
		return result, err
	}
	seen := map[string]bool{}
	for _, id := range ids {
		idStr, ok := id.(string)
		if !ok || seen[idStr] {
			continue
		}
		if _, modifier := idmodifier.SplitModifier(idStr); len(modifier) == 0 {
			continue
		}
		seen[idStr] = true
		result = append(result, idStr)
	}
	return result, nil
}

func extendedDeviceDisplayName(device models.ExtendedDevice) string {
	if device.DisplayName != "" {
		return device.DisplayName
	}
	return device.Name
}
