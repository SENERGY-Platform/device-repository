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
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"log"
	"net/http"
	"runtime/debug"
	"slices"
	"sort"
	"strings"
	"sync"
)

func (this *Controller) ListExtendedDevices(token string, options model.ExtendedDeviceListOptions) (result []models.ExtendedDevice, total int64, err error, errCode int) {
	ids := []string{}
	permissionFlag := options.Permission
	if permissionFlag == models.UnsetPermissionFlag {
		permissionFlag = models.Read
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, total, err, http.StatusBadRequest
	}

	if options.LocalIds != nil {
		if options.Owner == "" {
			options.Owner = jwtToken.GetUserId()
		}
		ctx, _ := getTimeoutContext()
		options.Ids, err = this.db.DeviceLocalIdsToIds(ctx, options.Owner, options.LocalIds)
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
	}

	//check permissions
	if options.Ids == nil {
		if jwtToken.IsAdmin() {
			ids = nil //no auth check for admins -> no id filter
		} else {
			ids, err, _ = this.permissionsV2Client.ListAccessibleResourceIds(token, this.config.DeviceTopic, client.ListOptions{}, client.Permission(permissionFlag))
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err, _ := this.permissionsV2Client.CheckMultiplePermissions(token, this.config.DeviceTopic, options.Ids, client.Permission(permissionFlag))
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
		for id, ok := range idMap {
			if ok {
				ids = append(ids, id)
			}
		}
	}

	//handle and preserve id modifiers
	pureIds := []string{}
	pureIdToRawIds := map[string][]string{}
	if ids == nil {
		pureIds = nil
	}
	for _, id := range ids {
		pureId, _ := idmodifier.SplitModifier(id)
		if !slices.Contains(pureIds, pureId) {
			pureIds = append(pureIds, pureId)
		}
		pureIdToRawIds[pureId] = append(pureIdToRawIds[pureId], id)
	}

	options.Ids = pureIds

	ctx, _ := getTimeoutContext()
	devices, total, err := this.db.ListDevices(ctx, options.ToDeviceListOptions(), true)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}

	for _, device := range devices {
		if len(pureIdToRawIds[device.Id]) == 0 {
			pureIdToRawIds[device.Id] = []string{device.Id}
		}
	}

	//get device-types for use in extendDevice()
	deviceTypes := []models.DeviceType{}
	for _, device := range devices {
		if !slices.ContainsFunc(deviceTypes, func(deviceType models.DeviceType) bool {
			return deviceType.Id == device.DeviceTypeId
		}) {
			pureDtId, _ := idmodifier.SplitModifier(device.DeviceTypeId)
			dt, exists, err := this.db.GetDeviceType(ctx, pureDtId)
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
			if exists {
				deviceTypes = append(deviceTypes, dt)
			} else {
				log.Println("WARNING: unable to find device type for ListExtendedDevices device.id=", device.Id)
			}
		}
	}

	//pre allocate result slice. no append() use to preserve sort
	resultSize := 0
	for _, device := range devices {
		for range pureIdToRawIds[device.Id] {
			resultSize = resultSize + 1
		}
	}
	result = make([]models.ExtendedDevice, resultSize)

	//transform db devices to extended devices; use go-routines; applies modified ids
	wg := sync.WaitGroup{}
	mux := sync.Mutex{}
	errCode = http.StatusOK
	index := 0
	for _, device := range devices {
		for _, id := range pureIdToRawIds[device.Id] {
			wg.Add(1)
			go func(d model.DeviceWithConnectionState, rawId string, resultIndex int) {
				defer wg.Done()
				d.Id = rawId
				_, modifier := idmodifier.SplitModifier(rawId)
				modDevice, modErr, modErrCode := this.modifyDevice(d.Device, modifier)
				if modErr != nil {
					mux.Lock()
					defer mux.Unlock()
					err = errors.Join(err, modErr)
					if errCode != http.StatusInternalServerError {
						errCode = modErrCode
					}
					return
				}
				d.Device = modDevice
				extendedDevice, extendErr := this.extendDevice(token, d, deviceTypes, options.FullDt)
				if extendErr != nil {
					mux.Lock()
					defer mux.Unlock()
					err = errors.Join(err, extendErr)
					errCode = http.StatusInternalServerError
					return
				}
				mux.Lock()
				defer mux.Unlock()
				result[resultIndex] = extendedDevice
			}(device, id, index)
			index = index + 1
		}
	}
	wg.Wait()
	if err != nil {
		return result, total, err, errCode
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) ListDevices(token string, options model.DeviceListOptions) (result []models.Device, err error, errCode int) {
	ids := []string{}
	permissionFlag := options.Permission
	if permissionFlag == models.UnsetPermissionFlag {
		permissionFlag = models.Read
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusBadRequest
	}

	if options.LocalIds != nil {
		if options.Owner == "" {
			options.Owner = jwtToken.GetUserId()
		}
		ctx, _ := getTimeoutContext()
		options.Ids, err = this.db.DeviceLocalIdsToIds(ctx, options.Owner, options.LocalIds)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
	}

	if options.Ids == nil {
		if jwtToken.IsAdmin() {
			ids = nil //no auth check for admins -> no id filter
		} else {
			ids, err, _ = this.permissionsV2Client.ListAccessibleResourceIds(token, this.config.DeviceTopic, client.ListOptions{}, client.Permission(permissionFlag))
			if err != nil {
				return result, err, http.StatusInternalServerError
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err, _ := this.permissionsV2Client.CheckMultiplePermissions(token, this.config.DeviceTopic, options.Ids, client.Permission(permissionFlag))
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		for id, ok := range idMap {
			if ok {
				ids = append(ids, id)
			}
		}
	}

	pureIds := []string{}
	pureIdToRawIds := map[string][]string{}
	if ids == nil {
		pureIds = nil
	}
	for _, id := range ids {
		pureId, _ := idmodifier.SplitModifier(id)
		if !slices.Contains(pureIds, pureId) {
			pureIds = append(pureIds, pureId)
		}
		pureIdToRawIds[pureId] = append(pureIdToRawIds[pureId], id)
	}

	options.Ids = pureIds

	ctx, _ := getTimeoutContext()
	devices, _, err := this.db.ListDevices(ctx, options, false)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	for _, device := range devices {
		if len(pureIdToRawIds[device.Id]) == 0 {
			pureIdToRawIds[device.Id] = []string{device.Id}
		}
	}
	result = []models.Device{}
	for _, device := range devices {
		for _, id := range pureIdToRawIds[device.Id] {
			device.Id = id
			_, modifier := idmodifier.SplitModifier(id)
			device.Device, err, errCode = this.modifyDevice(device.Device, modifier)
			if err != nil {
				return result, err, errCode
			}
			result = append(result, device.Device)
		}
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ReadDevice(id string, token string, action model.AuthAction) (result models.Device, err error, errCode int) {
	temp, err, errCode := this.readDevice(id)
	if err != nil {
		return result, err, errCode
	}
	result = temp.Device
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.DeviceTopic, id, client.Permission(action))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	return result, nil, http.StatusOK
}

func (this *Controller) ReadExtendedDevice(id string, token string, action model.AuthAction, fullDt bool) (result models.ExtendedDevice, err error, errCode int) {
	temp, err, errCode := this.readDevice(id)
	if err != nil {
		return result, err, errCode
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.DeviceTopic, id, client.Permission(action))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	ctx, _ := getTimeoutContext()
	pureDtId, _ := idmodifier.SplitModifier(temp.DeviceTypeId)
	dt, _, _ := this.db.GetDeviceType(ctx, pureDtId)
	result, err = this.extendDevice(token, temp, []models.DeviceType{dt}, fullDt)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}

func (this *Controller) readDevice(id string) (result model.DeviceWithConnectionState, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	pureId, modifier := idmodifier.SplitModifier(id)
	device, exists, err := this.db.GetDevice(ctx, pureId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	device.Id = id
	if modifier != nil && len(modifier) > 0 {
		device.Device, err, errCode = this.modifyDevice(device.Device, modifier)
		if err != nil {
			return result, err, errCode
		}
	}
	return device, nil, http.StatusOK
}

func (this *Controller) extendDevice(token string, device model.DeviceWithConnectionState, deviceTypes []models.DeviceType, fullDt bool) (result models.ExtendedDevice, err error) {
	var dtp *models.DeviceType
	deviceTypeName := ""
	for _, dt := range deviceTypes {
		pure, modifier := idmodifier.SplitModifier(device.DeviceTypeId)
		if dt.Id == pure {
			selectedDt := dt
			if len(modifier) > 0 {
				selectedDt.Id = device.DeviceTypeId //modifyDeviceType() does not modify id
				selectedDt, err, _ = this.modifyDeviceType(selectedDt, modifier)
				if err != nil {
					return models.ExtendedDevice{}, err
				}
			}
			if fullDt {
				dtp = &selectedDt
			}
			deviceTypeName = selectedDt.Name
			break
		}
	}

	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err
	}
	requestingUser := jwtToken.GetUserId()

	pureDeviceId, _ := idmodifier.SplitModifier(device.Id)
	computedPermissionList, err, _ := this.permissionsV2Client.ListComputedPermissions(token, this.config.DeviceTopic, []string{pureDeviceId})
	if err != nil {
		return models.ExtendedDevice{}, err
	}
	if len(computedPermissionList) != 1 {
		return models.ExtendedDevice{}, errors.New("unexpected response from permissions-v2 ListComputedPermissions()")
	}
	permissions := models.Permissions{
		Read:         computedPermissionList[0].Read,
		Write:        computedPermissionList[0].Write,
		Execute:      computedPermissionList[0].Execute,
		Administrate: computedPermissionList[0].Administrate,
	}

	return models.ExtendedDevice{
		Device:          device.Device,
		ConnectionState: device.ConnectionState,
		DisplayName:     getDeviceDisplayName(device.Device),
		DeviceTypeName:  deviceTypeName,
		Shared:          requestingUser != device.OwnerId,
		Permissions:     permissions,
		DeviceType:      dtp,
	}, err
}

func (this *Controller) ReadDeviceByLocalId(ownerId string, localId string, token string, action model.AuthAction) (result models.Device, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	device, exists, err := this.db.GetDeviceByLocalId(ctx, ownerId, localId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.DeviceTopic, device.Id, client.Permission(action))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	return device.Device, nil, http.StatusOK
}

func (this *Controller) ReadExtendedDeviceByLocalId(ownerId string, localId string, token string, action model.AuthAction, fullDt bool) (result models.ExtendedDevice, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	device, exists, err := this.db.GetDeviceByLocalId(ctx, ownerId, localId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.DeviceTopic, device.Id, client.Permission(action))
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	pureDtId, _ := idmodifier.SplitModifier(device.DeviceTypeId)
	dt, _, _ := this.db.GetDeviceType(ctx, pureDtId)
	result, err = this.extendDevice(token, device, []models.DeviceType{dt}, fullDt)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}

const DisplayNameAttributeName = "shared/nickname"

func ValidateDeviceName(device models.Device) (err error) {
	if device.Name == "" {
		hasDisplayNameAttribute := false
		for _, attr := range device.Attributes {
			if attr.Key == DisplayNameAttributeName {
				hasDisplayNameAttribute = true
				break
			}
		}
		if !hasDisplayNameAttribute {
			return errors.New("missing device name")
		}
	}
	return nil
}

func (this *Controller) ValidateDevice(token string, device models.Device) (err error, code int) {
	if device.Id == "" {
		return errors.New("missing device id"), http.StatusBadRequest
	}
	err = ValidateDeviceName(device)
	if err != nil {
		return err, http.StatusBadRequest
	}
	if device.LocalId == "" {
		return errors.New("missing device local id"), http.StatusBadRequest
	}
	if device.DeviceTypeId == "" {
		return errors.New("missing device type id"), http.StatusBadRequest
	}

	ctx, _ := getTimeoutContext()

	original, exists, err := this.db.GetDevice(ctx, device.Id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if exists {
		if original.DeviceTypeId != device.DeviceTypeId {
			return errors.New("device type id mismatch"), http.StatusBadRequest
		}
		if device.OwnerId != original.OwnerId {
			resource, err, code := this.permissionsV2Client.GetResource(token, this.config.DeviceTopic, device.Id)
			if err != nil {
				return err, code
			}
			admins := []string{}
			for user, perm := range resource.UserPermissions {
				if perm.Administrate {
					admins = append(admins, user)
				}
			}
			if len(admins) == 0 {
				//o admins indicates the requesting user has not the needed admin rights to see other admins
				return errors.New("requesting user must have admin rights to change owner"), http.StatusBadRequest
			}
			//new device owner-id must be existing admin user (ignore for new devices or devices with unchanged owner)
			if !slices.Contains(admins, device.OwnerId) {
				return errors.New("new owner must have existing user admin rights"), http.StatusBadRequest
			}
		}
	}

	//device-type exists
	dt, ok, err := this.db.GetDeviceType(ctx, device.DeviceTypeId)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("unknown device type id"), http.StatusBadRequest
	}

	protocolConstraints := map[string][]string{}
	for _, service := range dt.Services {
		if _, ok = protocolConstraints[service.ProtocolId]; !ok {
			p, exists, err := this.db.GetProtocol(ctx, service.ProtocolId)
			if err != nil {
				return err, http.StatusInternalServerError
			}
			if exists {
				protocolConstraints[p.Id] = p.Constraints
			}
		}
	}

	constraints := []string{}
	for _, pc := range protocolConstraints {
		constraints = append(constraints, pc...)
	}

	if contains(constraints, model.SenergyConnectorLocalIdConstraint) {
		if strings.ContainsAny(device.LocalId, "+#/") {
			return errors.New("device local id may not contain any +#/"), http.StatusBadRequest
		}
	}

	//local ids are globally unique
	ctx, _ = getTimeoutContext()
	d, ok, err := this.db.GetDeviceByLocalId(ctx, device.OwnerId, device.LocalId)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if ok && d.Id != device.Id {
		if !this.config.LocalIdUniqueForOwner {
			return errors.New("local id should be empty or globally unique"), http.StatusBadRequest
		}
		return errors.New("local id should be empty or for the owner unique"), http.StatusBadRequest
	}

	return nil, http.StatusOK
}

func (this *Controller) CreateDevice(token string, device models.Device) (result models.Device, err error, code int) {
	if device.Id != "" {
		return result, errors.New("device id already set"), http.StatusBadRequest
	}
	if !this.config.DisableStrictValidationForTesting {
		device.GenerateId()
	}

	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	if device.OwnerId != "" && device.OwnerId != jwtToken.GetUserId() {
		return device, errors.New("new devices must be initialised with the requesting user as owner-id"), http.StatusBadRequest
	}
	if device.OwnerId == "" {
		device.OwnerId = jwtToken.GetUserId()
	}
	if !this.config.DisableStrictValidationForTesting {
		err, code = this.ValidateDevice(token, device)
		if err != nil {
			this.logger.Warn("device validation failed with error: " + err.Error())
			return device, err, code
		}
	}

	return this.setDevice(device)
}

func (this *Controller) SetDevice(token string, device models.Device, options model.DeviceUpdateOptions) (result models.Device, err error, code int) {
	if device.Id == "" {
		return result, errors.New("missing device id"), http.StatusBadRequest
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() && !this.config.DisableStrictValidationForTesting {
		ok, err, code := this.permissionsV2Client.CheckPermission(token, this.config.DeviceTopic, device.Id, client.Write)
		if err != nil {
			return device, err, code
		}
		if !ok {
			return device, errors.New("access denied"), http.StatusForbidden
		}
	}

	var original model.DeviceWithConnectionState
	var exists bool
	original, err, code = this.readDevice(device.Id)
	if err != nil && code != http.StatusNotFound {
		return device, err, code
	}
	if err != nil {
		err, code = nil, 200
		exists = false
	} else {
		exists = true
	}

	if exists && len(options.UpdateOnlySameOriginAttributes) > 0 {
		device.Attributes = updateSameOriginAttributes(original.Attributes, device.Attributes, options.UpdateOnlySameOriginAttributes)
	}

	//set device owner-id if none is given
	//prefer existing owner, fallback to requesting user
	if device.OwnerId == "" {
		device.OwnerId = original.OwnerId //may be empty for new devices
	}
	if device.OwnerId == "" {
		device.OwnerId = jwtToken.GetUserId()
	}

	if exists && original.OwnerId != device.OwnerId && original.OwnerId != "" && !jwtToken.IsAdmin() {
		ok, err, code := this.permissionsV2Client.CheckPermission(token, this.config.DeviceTopic, device.Id, client.Administrate)
		if err != nil {
			return device, err, code
		}
		if !ok {
			return device, fmt.Errorf("only admins may set new owner: %w", err), http.StatusBadRequest
		}
	}

	if !this.config.DisableStrictValidationForTesting {
		err, code = this.ValidateDevice(token, device)
		if err != nil {
			this.logger.Warn("device validation failed with error: " + err.Error())
			return device, err, code
		}
	}

	rights, err, code := this.permissionsV2Client.GetResource(token, this.config.DeviceTopic, device.Id)
	if err != nil && code != http.StatusNotFound {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return device, err, code
	}

	//new device owner-id must be existing admin user (ignore for new devices or devices with unchanged owner)
	if code != http.StatusNotFound && device.OwnerId != original.OwnerId && !rights.UserPermissions[device.OwnerId].Administrate {
		return device, errors.New("new owner must have existing user admin rights"), http.StatusBadRequest
	}

	device, err, code = this.setDevice(device)
	if err != nil {
		return device, err, code
	}

	return device, nil, http.StatusOK
}

func (this *Controller) setDevice(device models.Device) (result models.Device, err error, code int) {
	//update hub about changed device.local_id
	ctx, _ := getTimeoutContext()
	old, exists, err := this.db.GetDevice(ctx, device.Id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	connectionState := models.ConnectionStateUnknown
	if exists {
		connectionState = old.ConnectionState
	}

	//save device
	ctx, _ = getTimeoutContext()
	err = this.db.SetDevice(ctx, model.DeviceWithConnectionState{
		Device:          device,
		ConnectionState: connectionState,
	}, this.setDeviceSyncHandler)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}

	return device, nil, http.StatusOK
}

func (this *Controller) setDeviceSyncHandler(device model.DeviceWithConnectionState) (err error) {
	err = this.EnsureInitialRights(this.config.DeviceTopic, device.Id, device.OwnerId)
	if err != nil {
		return err
	}

	err = this.EnsureGeneratedDeviceGroup(device.Device)
	if err != nil {
		return err
	}

	//ensure that changed device-local-ids are mirrored in hubs
	ctx, _ := getTimeoutContext()
	hubs, err := this.db.GetHubsByDeviceId(ctx, device.Id)
	if err != nil {
		return err
	}
	for _, hub := range hubs {
		if !slices.Contains(hub.DeviceLocalIds, device.LocalId) {
			devices, _, err := this.db.ListDevices(ctx, model.DeviceListOptions{Ids: hub.DeviceIds}, false)
			if err != nil {
				return err
			}
			hub.DeviceLocalIds = []string{}
			for _, d := range devices {
				hub.DeviceLocalIds = append(hub.DeviceLocalIds, d.LocalId)
			}
			hub.Hash = ""
			err = this.setHub(hub)
			if err != nil {
				return err
			}
		}
	}
	err = this.publisher.PublishDevice(device.Device)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) DeleteDevice(token string, id string) (error, int) {
	ctx, _ := getTimeoutContext()
	_, exists, err := this.db.GetDevice(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return nil, http.StatusOK
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.DeviceTopic, id, client.Administrate)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("access denied"), http.StatusForbidden
	}
	err = this.deleteDevice(id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteDeviceSyncHandler(old model.DeviceWithConnectionState) (err error) {
	err = this.resetHubsForDeviceUpdate(old.Device)
	if err != nil {
		return err
	}
	err = this.RemoveGeneratedDeviceGroup(old.Id, old.OwnerId)
	if err != nil {
		return err
	}
	err = this.RemoveRights(this.config.DeviceTopic, old.Id)
	if err != nil {
		return err
	}
	err = this.publisher.PublishDeviceDelete(old.Id)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) deleteDevice(id string) error {
	ctx, _ := getTimeoutContext()
	err := this.db.RemoveDevice(ctx, id, this.deleteDeviceSyncHandler)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) resetHubsForDeviceUpdate(old models.Device) error {
	ctx, _ := getTimeoutContext()
	hubs, err := this.db.GetHubsByDeviceId(ctx, old.Id)
	if err != nil {
		return err
	}
	for _, hub := range hubs {
		hub.DeviceLocalIds = filter(hub.DeviceLocalIds, old.LocalId)
		hub.DeviceIds = filter(hub.DeviceIds, old.Id)
		hub.Hash = ""
		err = this.setHub(hub)
		if err != nil {
			return err
		}
	}
	return nil
}

func filter(in []string, not string) (out []string) {
	for _, str := range in {
		if str != not {
			out = append(out, str)
		}
	}
	return
}

func updateSameOriginAttributes(attributes []models.Attribute, update []models.Attribute, origin []string) (result []models.Attribute) {
	for _, attr := range attributes {
		if !contains(origin, attr.Origin) {
			result = append(result, attr)
		}
	}
	for _, attr := range update {
		if contains(origin, attr.Origin) {
			result = append(result, attr)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

func (this *Controller) SetDeviceConnectionState(token string, id string, connected bool) (error, int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.DeviceTopic, id, client.Write)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !ok {
			return errors.New("access denied"), http.StatusForbidden
		}
	}
	state := models.ConnectionStateOffline
	if connected {
		state = models.ConnectionStateOnline
	}
	ctx, _ := getTimeoutContext()
	err = this.db.SetDeviceConnectionState(ctx, id, state)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}
