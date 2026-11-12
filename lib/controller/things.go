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
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/google/uuid"
)

// normalizeThingTime converts to UTC and cuts to milliseconds, the precision mongo stores for dates.
// Without it testdb and mongo would answer differently for the same input.
func normalizeThingTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Millisecond)
}

func (this *Controller) ListThings(ctx context.Context, token string, options model.ThingListOptions) (result []models.Thing, total int64, err error, errCode int) {
	ids := []string{}
	permissionFlag := options.Permission
	if permissionFlag == models.UnsetPermissionFlag {
		permissionFlag = models.Read
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, total, err, http.StatusBadRequest
	}

	if options.Ids == nil {
		if jwtToken.IsAdmin() {
			ids = nil //no auth check for admins -> no id filter
		} else {
			ids, err, _ = this.permissionsV2Client.ListAccessibleResourceIdsContext(ctx, token, this.config.ThingTopic, client.ListOptions{}, client.Permission(permissionFlag))
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
			if ids == nil {
				//nil would mean "no id filter" for the database
				ids = []string{}
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err, _ := this.permissionsV2Client.CheckMultiplePermissionsContext(ctx, token, this.config.ThingTopic, options.Ids, client.Permission(permissionFlag))
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

	ctx, _ = getTimeoutContext(ctx)
	result, total, err = this.db.ListThings(ctx, options)
	if err != nil {
		if errors.Is(err, model.ErrUnsupportedSort) {
			return result, total, err, http.StatusBadRequest
		}
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) ReadThing(ctx context.Context, token string, id string) (result models.Thing, err error, errCode int) {
	ok, err, _ := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.ThingTopic, id, client.Read)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	ctx, _ = getTimeoutContext(ctx)
	result, exists, err := this.db.GetThing(ctx, id)
	if err != nil {
		return models.Thing{}, err, http.StatusInternalServerError
	}
	if !exists {
		return models.Thing{}, fmt.Errorf("thing %w", model.ErrNotFound), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) SetThing(ctx context.Context, token string, thing models.Thing) (result models.Thing, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if thing.Id != "" && !jwtToken.IsAdmin() {
		ok, err, code := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.ThingTopic, thing.Id, client.Write)
		if err != nil {
			return thing, err, code
		}
		if !ok {
			return thing, errors.New("access denied"), http.StatusForbidden
		}
	}
	isNew := thing.Id == ""
	//sets the id of new things and the ids of content variables without id
	thing.GenerateId()

	ctx, _ = getTimeoutContext(ctx)
	original, exists, err := this.db.GetThing(ctx, thing.Id)
	if err != nil {
		return thing, err, http.StatusInternalServerError
	}

	//set thing owner if none is given
	//prefer existing owner, fallback to requesting user
	if thing.OwnerId == "" {
		thing.OwnerId = original.OwnerId //may be empty for new things
	}
	if thing.OwnerId == "" {
		thing.OwnerId = jwtToken.GetUserId()
	}
	//a new thing belongs to the user who creates it; only admins may create things for others
	if isNew && !jwtToken.IsAdmin() {
		thing.OwnerId = jwtToken.GetUserId()
	}

	if exists && original.OwnerId != thing.OwnerId && original.OwnerId != "" {
		err, code = this.checkOwnerChange(ctx, token, jwtToken, this.config.ThingTopic, thing.Id, original.OwnerId, thing.OwnerId)
		if err != nil {
			return thing, err, code
		}
	}

	//device links are managed by AddThingDeviceLink and EndThingDeviceLink only
	thing.DeviceLinks = original.DeviceLinks

	//empty lists are stored and answered as [] instead of null
	if thing.Attributes == nil {
		thing.Attributes = []models.Attribute{}
	}
	if thing.ContentVariables == nil {
		thing.ContentVariables = []models.ContentVariable{}
	}
	if thing.DeviceLinks == nil {
		thing.DeviceLinks = []models.ThingDeviceLink{}
	}

	err = thing.Valid()
	if err != nil {
		return thing, fmt.Errorf("invalid thing: %w", err), http.StatusBadRequest
	}

	err = this.setThing(ctx, thing)
	if err != nil {
		return thing, err, http.StatusInternalServerError
	}
	return thing, nil, http.StatusOK
}

func (this *Controller) setThing(ctx context.Context, thing models.Thing) (err error) {
	this.config.GetLogger().Debug("create/update thing", "id", thing.Id)
	ctx, _ = getTimeoutContext(ctx)
	return this.db.SetThing(ctx, thing, withContext(ctx, this.setThingSyncHandler))
}

func (this *Controller) setThingSyncHandler(ctx context.Context, thing models.Thing) (err error) {
	err = this.EnsureInitialRights(ctx, this.config.ThingTopic, thing.Id, thing.OwnerId)
	if err != nil {
		return fmt.Errorf("unable to ensure initial thing permissions: %w", err)
	}
	return this.publisher.PublishThing(thing)
}

func (this *Controller) DeleteThing(ctx context.Context, token string, id string) (error, int) {
	ctx, _ = getTimeoutContext(ctx)
	_, exists, err := this.db.GetThing(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return nil, http.StatusOK
	}
	ok, err, _ := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.ThingTopic, id, client.Administrate)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("access denied"), http.StatusForbidden
	}
	err = this.deleteThing(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteThing(ctx context.Context, id string) error {
	ctx, _ = getTimeoutContext(ctx)
	return this.db.RemoveThing(ctx, id, withContext(ctx, this.deleteThingSyncHandler))
}

func (this *Controller) deleteThingSyncHandler(ctx context.Context, old models.Thing) (err error) {
	err = this.removeResourceFromGraphs(ctx, models.GraphResourceTypeThing, old.Id)
	if err != nil {
		return err
	}
	err = this.RemoveRights(ctx, this.config.ThingTopic, old.Id)
	if err != nil {
		return err
	}
	return this.publisher.PublishThingDelete(old.Id)
}

// checkThingWriteAccess loads the thing for a change of its device links. Admins may write every thing.
func (this *Controller) checkThingWriteAccess(ctx context.Context, token string, thingId string) (thing models.Thing, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return thing, err, http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		ok, err, code := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.ThingTopic, thingId, client.Write)
		if err != nil {
			return thing, err, code
		}
		if !ok {
			return thing, errors.New("access denied"), http.StatusForbidden
		}
	}
	thing, exists, err := this.db.GetThing(ctx, thingId)
	if err != nil {
		return thing, err, http.StatusInternalServerError
	}
	if !exists {
		return thing, fmt.Errorf("thing %w", model.ErrNotFound), http.StatusNotFound
	}
	return thing, nil, http.StatusOK
}

func (this *Controller) AddThingDeviceLink(ctx context.Context, token string, thingId string, create model.ThingDeviceLinkCreate) (result models.Thing, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	thing, err, code := this.checkThingWriteAccess(ctx, token, thingId)
	if err != nil {
		return result, err, code
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusBadRequest
	}

	if create.DeviceId == "" {
		return result, errors.New("missing device_id"), http.StatusBadRequest
	}
	_, deviceExists, err := this.db.GetDevice(ctx, create.DeviceId)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !deviceExists {
		return result, fmt.Errorf("unknown device %v", create.DeviceId), http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		ok, err, code := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.DeviceTopic, create.DeviceId, client.Read)
		if err != nil {
			return result, err, code
		}
		if !ok {
			return result, errors.New("access denied to device"), http.StatusForbidden
		}
	}

	link := models.ThingDeviceLink{
		Id:       uuid.NewString(),
		DeviceId: create.DeviceId,
		From:     normalizeThingTime(time.Now()),
	}
	if create.From != nil {
		link.From = normalizeThingTime(*create.From)
	}
	if create.Until != nil {
		until := normalizeThingTime(*create.Until)
		link.Until = &until
		if !until.After(link.From) {
			return result, errors.New("until must be after from"), http.StatusBadRequest
		}
	}

	err, code = this.checkThingDeviceLinkOverlap(ctx, thing, link)
	if err != nil {
		return result, err, code
	}
	err, code = this.checkThingDeviceLinkGraphRule(ctx, thing.Id, link.DeviceId)
	if err != nil {
		return result, err, code
	}

	thing.DeviceLinks = append(slices.Clone(thing.DeviceLinks), link)
	err = thing.Valid()
	if err != nil {
		return result, fmt.Errorf("invalid thing: %w", err), http.StatusBadRequest
	}
	err = this.setThing(ctx, thing)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return thing, nil, http.StatusOK
}

// checkThingDeviceLinkOverlap rejects a link whose time range overlaps with a link of the same device,
// in the given thing or in any other thing. The database filter only narrows the things down, the links are compared here.
// ListThings is called with Limit 0: mongo returns all matches, testdb ignores the options and returns all things.
func (this *Controller) checkThingDeviceLinkOverlap(ctx context.Context, thing models.Thing, link models.ThingDeviceLink) (err error, code int) {
	things, _, err := this.db.ListThings(ctx, model.ThingListOptions{DeviceIds: []string{link.DeviceId}})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !slices.ContainsFunc(things, func(t models.Thing) bool { return t.Id == thing.Id }) {
		things = append(things, thing)
	}
	for _, other := range things {
		for _, otherLink := range other.DeviceLinks {
			if otherLink.DeviceId != link.DeviceId || !otherLink.Overlaps(link) {
				continue
			}
			if other.Id == thing.Id {
				return fmt.Errorf("time range overlaps with device link %v of this thing", otherLink.Id), http.StatusBadRequest
			}
			//the other thing may belong to someone else: no id in the message
			return fmt.Errorf("time range overlaps with a link of device %v to another thing", link.DeviceId), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

// checkThingDeviceLinkGraphRule rejects a link if a graph contains the thing and the device.
// Graph nodes are evaluated here, the database filter only narrows the graphs down.
// ListGraphs is called with Limit 0: mongo returns all matches, testdb ignores the options and returns all graphs.
func (this *Controller) checkThingDeviceLinkGraphRule(ctx context.Context, thingId string, deviceId string) (err error, code int) {
	graphs, _, err := this.db.ListGraphs(ctx, model.GraphListOptions{ThingIds: []string{thingId}})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, graph := range graphs {
		hasThing := slices.ContainsFunc(graph.Nodes, func(node models.Node) bool {
			return node.ResourceType == models.GraphResourceTypeThing && node.ResourceId == thingId
		})
		hasDevice := slices.ContainsFunc(graph.Nodes, func(node models.Node) bool {
			return node.ResourceType == models.GraphResourceTypeDevice && node.ResourceId == deviceId
		})
		if hasThing && hasDevice {
			//no graph id in the message: the graph may belong to someone else
			return fmt.Errorf("a graph contains thing %v and device %v", thingId, deviceId), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) EndThingDeviceLink(ctx context.Context, token string, thingId string, linkId string, end model.ThingDeviceLinkEnd) (result models.Thing, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	thing, err, code := this.checkThingWriteAccess(ctx, token, thingId)
	if err != nil {
		return result, err, code
	}
	index := slices.IndexFunc(thing.DeviceLinks, func(link models.ThingDeviceLink) bool { return link.Id == linkId })
	if index < 0 {
		return result, fmt.Errorf("device link %w", model.ErrNotFound), http.StatusNotFound
	}
	thing.DeviceLinks = slices.Clone(thing.DeviceLinks)
	link := thing.DeviceLinks[index]
	if link.Until != nil {
		return result, errors.New("device link is already ended"), http.StatusBadRequest
	}
	until := normalizeThingTime(time.Now())
	if !until.After(link.From) {
		//the link starts now or in the future: end it one millisecond after its start, as endOpenThingLinksOfDevice does
		until = link.From.Add(time.Millisecond)
	}
	if end.Until != nil {
		until = normalizeThingTime(*end.Until)
	}
	if !until.After(link.From) {
		return result, errors.New("until must be after from"), http.StatusBadRequest
	}
	link.Until = &until
	thing.DeviceLinks[index] = link

	err = thing.Valid()
	if err != nil {
		return result, fmt.Errorf("invalid thing: %w", err), http.StatusBadRequest
	}
	err = this.setThing(ctx, thing)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return thing, nil, http.StatusOK
}

// endOpenThingLinksOfDevice is part of the device delete sync handler: open links of the device end now.
// Ended links stay unchanged, so a retry after a partial failure only touches what is still open.
// ListThings is called with Limit 0: mongo returns all matches, testdb ignores the options and returns all things.
func (this *Controller) endOpenThingLinksOfDevice(ctx context.Context, deviceId string) error {
	ctx, _ = getTimeoutContext(ctx)
	things, _, err := this.db.ListThings(ctx, model.ThingListOptions{DeviceIds: []string{deviceId}})
	if err != nil {
		return err
	}
	now := normalizeThingTime(time.Now())
	for _, thing := range things {
		changed := false
		links := slices.Clone(thing.DeviceLinks)
		for i, link := range links {
			if link.DeviceId != deviceId || link.Until != nil {
				continue
			}
			until := now
			if !until.After(link.From) {
				//link starts in the future: end it one millisecond after its start to keep the range valid
				until = link.From.Add(time.Millisecond)
			}
			link.Until = &until
			links[i] = link
			changed = true
		}
		if !changed {
			continue
		}
		thing.DeviceLinks = links
		err = this.setThing(ctx, thing)
		if err != nil {
			this.config.GetLogger().Error("unable to end thing device links of deleted device", "thingId", thing.Id, "deviceId", deviceId, "error", err)
			return err
		}
	}
	return nil
}
