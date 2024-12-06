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
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
)

/////////////////////////
//		api
/////////////////////////

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
			ids, err = this.db.ListAccessibleResourceIds(token, this.config.DeviceTopic, 0, 0, permissionFlag)
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err := this.db.CheckMultiple(token, this.config.DeviceTopic, options.Ids, permissionFlag)
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
			ids, err = this.db.ListAccessibleResourceIds(token, this.config.DeviceTopic, 0, 0, permissionFlag)
			if err != nil {
				return result, err, http.StatusInternalServerError
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err := this.db.CheckMultiple(token, this.config.DeviceTopic, options.Ids, permissionFlag)
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
	ok, err := this.db.CheckBool(token, this.config.DeviceTopic, id, action)
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
	ok, err := this.db.CheckBool(token, this.config.DeviceTopic, id, action)
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

	requestingUser, permissions, err := this.db.GetPermissionsInfo(token, this.config.DeviceTopic, device.Id)
	if err != nil {
		return models.ExtendedDevice{}, err
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
	ok, err := this.db.CheckBool(token, this.config.DeviceTopic, device.Id, action)
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
	ok, err := this.db.CheckBool(token, this.config.DeviceTopic, device.Id, action)
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
			admins, err := this.db.GetAdminUsers(token, this.config.DeviceTopic, device.Id)
			if errors.Is(err, model.PermissionCheckFailed) {
				return errors.New("requesting user must have admin rights to change owner"), http.StatusBadRequest
			}
			if err != nil {
				return err, http.StatusInternalServerError
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

/////////////////////////
//		source
/////////////////////////

func (this *Controller) SetDevice(device models.Device, owner string) (err error) {
	err = this.EnsureInitialRights(this.config.DeviceTopic, device.Id, owner)
	if err != nil {
		return err
	}

	//prevent collision of local ids
	//this if branch should be rarely needed if 2 devices are created at the same time with the same local_id (when the second device is validated before the creation of the first is finished)
	ctx, _ := getTimeoutContext()
	d, collision, err := this.db.GetDeviceByLocalId(ctx, device.OwnerId, device.LocalId)
	if err != nil {
		return err
	}
	if collision && d.Id != device.Id {

		//handle invalid device
		device.LocalId = ""
		device.OwnerId = owner
		ctx, _ = getTimeoutContext()
		err = this.db.SetDevice(ctx, model.DeviceWithConnectionState{
			Device:          device,
			ConnectionState: "",
		})
		if err != nil {
			return err
		}
		return this.PublishDeviceDelete(device.Id, owner)

	} else {

		//update hub about changed device.local_id
		ctx, _ = getTimeoutContext()
		old, exists, err := this.db.GetDevice(ctx, device.Id)
		if err != nil {
			return err
		}
		if exists && old.LocalId != device.LocalId {
			err = this.resetHubsForDeviceUpdate(old.Device)
			if err != nil {
				return err
			}
		}

		if device.OwnerId == "" {
			if exists {
				device.OwnerId = old.OwnerId
			} else {
				device.OwnerId = owner
			}
		}

		if !exists {
			err = this.CreateGeneratedDeviceGroup(device)
			if err != nil {
				return err
			}
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
		})
		if err != nil {
			return err
		}
		return nil
	}

}

func (this *Controller) SetDeviceConnectionState(id string, connected bool) error {
	state := models.ConnectionStateOffline
	if connected {
		state = models.ConnectionStateOnline
	}
	ctx, _ := getTimeoutContext()
	return this.db.SetDeviceConnectionState(ctx, id, state)
}

func (this *Controller) DeleteDevice(id string) error {
	ctx, _ := getTimeoutContext()
	old, exists, err := this.db.GetDevice(ctx, id)
	if err != nil {
		return err
	}
	err = this.RemoveGeneratedDeviceGroup(id, old.OwnerId)
	if err != nil {
		return err
	}
	if exists {
		ctx, _ := getTimeoutContext()
		err := this.db.RemoveDevice(ctx, id)
		if err != nil {
			return err
		}
		return this.resetHubsForDeviceUpdate(old.Device)
	}
	err = this.RemoveRights(this.config.DeviceTopic, id)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) PublishDeviceDelete(id string, owner string) error {
	return this.producer.PublishDeviceDelete(id, owner)
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
		err = this.producer.PublishHub(hub.Hub, hub.OwnerId)
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
