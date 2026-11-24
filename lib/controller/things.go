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
	if result == nil {
		//an empty list is answered as [] instead of null
		result = []models.Thing{}
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

	result, err, code = this.editThing(ctx, thing.Id, func(original models.Thing, exists bool) (models.Thing, bool, error, int) {
		updated := thing

		//set thing owner if none is given
		//prefer existing owner, fallback to requesting user
		if updated.OwnerId == "" {
			updated.OwnerId = original.OwnerId //may be empty for new things
		}
		if updated.OwnerId == "" {
			updated.OwnerId = jwtToken.GetUserId()
		}
		//a new thing belongs to the user who creates it; only admins may create things for others
		if isNew && !jwtToken.IsAdmin() {
			updated.OwnerId = jwtToken.GetUserId()
		}

		if exists && original.OwnerId != updated.OwnerId && original.OwnerId != "" {
			err, code := this.checkOwnerChange(ctx, token, jwtToken, this.config.ThingTopic, updated.Id, original.OwnerId, updated.OwnerId)
			if err != nil {
				return updated, false, err, code
			}
		}

		//device links are managed by AddThingDeviceLink, EndThingDeviceLink and SwapThingDeviceLink only
		updated.DeviceLinks = original.DeviceLinks
		//a body without kind keeps the stored one, like the links
		if updated.Kind == "" {
			updated.Kind = original.Kind
		}

		//empty lists are stored and answered as [] instead of null
		if updated.Attributes == nil {
			updated.Attributes = []models.Attribute{}
		}
		if updated.ContentVariables == nil {
			updated.ContentVariables = []models.ContentVariable{}
		}
		if updated.DeviceLinks == nil {
			updated.DeviceLinks = []models.ThingDeviceLink{}
		}

		err := updated.Valid()
		if err != nil {
			return updated, false, fmt.Errorf("invalid thing: %w", err), http.StatusBadRequest
		}
		return updated, true, nil, http.StatusOK
	})
	if err != nil && result.Id == "" {
		result = thing
	}
	return result, err, code
}

// thingWriteAttempts is how often a write that lost against a concurrent write of the same thing is applied again.
const thingWriteAttempts = 5

// editThing is the compare-and-set loop of every thing write, as for graphs: it reads the stored thing with its
// revision, lets edit compute the new document from it and writes it only if the revision is still the one that was
// read. Otherwise the edit is applied again to what is stored now, so a concurrent write of the same thing (a link
// swapped, the thing renamed) is never overwritten with a stale copy. After thingWriteAttempts lost races the answer is 409.
// edit answers write == false to leave the thing as it is; the stored thing is the result then.
// This protects one thing document. A check across things, i.e. the overlap of the links of one device on two things,
// stays a check before the write and is not atomic.
func (this *Controller) editThing(ctx context.Context, id string, edit func(stored models.Thing, exists bool) (updated models.Thing, write bool, err error, code int)) (result models.Thing, err error, code int) {
	for attempt := 0; attempt < thingWriteAttempts; attempt++ {
		readCtx, _ := getTimeoutContext(ctx)
		stored, revision, exists, err := this.db.GetThingWithRevision(readCtx, id)
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		updated, write, err, code := edit(stored, exists)
		if err != nil {
			return updated, err, code
		}
		if !write {
			return stored, nil, http.StatusOK
		}
		this.config.GetLogger().Debug("create/update thing", "id", updated.Id, "revision", revision)
		writeCtx, _ := getTimeoutContext(ctx)
		swapped, err := this.db.SetThingIfRevision(writeCtx, updated, revision, withContext(writeCtx, this.setThingSyncHandler))
		if err != nil {
			return updated, err, http.StatusInternalServerError
		}
		if swapped {
			return updated, nil, http.StatusOK
		}
	}
	return result, fmt.Errorf("thing %v was changed concurrently, please try again", id), http.StatusConflict
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

// checkLinkableDevice checks that the device exists and that the caller may read it, the precondition to link it.
func (this *Controller) checkLinkableDevice(ctx context.Context, token string, jwtToken jwt.Token, deviceId string) (err error, code int) {
	if deviceId == "" {
		return errors.New("missing device_id"), http.StatusBadRequest
	}
	_, deviceExists, err := this.db.GetDevice(ctx, deviceId)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !deviceExists {
		return fmt.Errorf("unknown device %v", deviceId), http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		ok, err, code := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.DeviceTopic, deviceId, client.Read)
		if err != nil {
			return err, code
		}
		if !ok {
			return errors.New("access denied to device"), http.StatusForbidden
		}
	}
	return nil, http.StatusOK
}

// newThingDeviceLink builds the link of a create request: from defaults to now, until to open.
func newThingDeviceLink(create model.ThingDeviceLinkCreate) (link models.ThingDeviceLink, err error, code int) {
	link = models.ThingDeviceLink{
		Id:       uuid.NewString(),
		DeviceId: create.DeviceId,
		From:     normalizeThingTime(time.Now()),
		Readings: slices.Clone(create.Readings),
		Factor:   create.Factor,
	}
	if create.From != nil {
		link.From = normalizeThingTime(*create.From)
	}
	if create.Until != nil {
		until := normalizeThingTime(*create.Until)
		link.Until = &until
		if !until.After(link.From) {
			return link, errors.New("until must be after from"), http.StatusBadRequest
		}
	}
	return link, nil, http.StatusOK
}

// endThingDeviceLinkAt ends an open link at the given instant, which must be after its start, and sets the readings
// at removal. The given link is not modified.
func endThingDeviceLinkAt(link models.ThingDeviceLink, until time.Time, untilValues []model.ThingDeviceLinkEndReading) (models.ThingDeviceLink, error) {
	if link.Until != nil {
		return link, errors.New("device link is already ended")
	}
	if !until.After(link.From) {
		return link, errors.New("until must be after from")
	}
	link.Until = &until
	link.Readings = slices.Clone(link.Readings)
	seen := map[[2]string]bool{}
	for _, untilValue := range untilValues {
		key := [2]string{untilValue.ServiceId, untilValue.Column}
		if seen[key] {
			return link, fmt.Errorf("duplicate until_value for %v %v", untilValue.ServiceId, untilValue.Column)
		}
		seen[key] = true
		readingIndex := slices.IndexFunc(link.Readings, func(r models.ThingDeviceLinkReading) bool {
			return r.ServiceId == untilValue.ServiceId && r.Column == untilValue.Column
		})
		if readingIndex < 0 {
			return link, fmt.Errorf("until_value for unknown reading %v %v", untilValue.ServiceId, untilValue.Column)
		}
		value := untilValue.UntilValue
		link.Readings[readingIndex].UntilValue = &value
	}
	return link, nil
}

func (this *Controller) AddThingDeviceLink(ctx context.Context, token string, thingId string, create model.ThingDeviceLinkCreate) (result models.Thing, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	_, err, code = this.checkThingWriteAccess(ctx, token, thingId)
	if err != nil {
		return result, err, code
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusBadRequest
	}

	err, code = this.checkLinkableDevice(ctx, token, jwtToken, create.DeviceId)
	if err != nil {
		return result, err, code
	}
	link, err, code := newThingDeviceLink(create)
	if err != nil {
		return result, err, code
	}

	return this.editThing(ctx, thingId, func(thing models.Thing, exists bool) (models.Thing, bool, error, int) {
		if !exists {
			return thing, false, fmt.Errorf("thing %w", model.ErrNotFound), http.StatusNotFound
		}
		err, code := this.checkThingDeviceLinkOverlap(ctx, thing, link)
		if err != nil {
			return thing, false, err, code
		}
		err, code = this.checkThingDeviceLinkGraphRule(ctx, thing.Id, link)
		if err != nil {
			return thing, false, err, code
		}
		thing.DeviceLinks = append(slices.Clone(thing.DeviceLinks), link)
		err = thing.Valid()
		if err != nil {
			return thing, false, fmt.Errorf("invalid thing: %w", err), http.StatusBadRequest
		}
		return thing, true, nil, http.StatusOK
	})
}

// checkThingDeviceLinkOverlap rejects a link whose time range overlaps with a link of the same device,
// in the given thing or in any other thing. The database filter only narrows the things down, the links are compared here.
// ListThings is called with Limit 0: mongo returns all matches, testdb ignores the options and returns all things.
func (this *Controller) checkThingDeviceLinkOverlap(ctx context.Context, thing models.Thing, link models.ThingDeviceLink) (err error, code int) {
	things, _, err := this.db.ListThings(ctx, model.ThingListOptions{DeviceIds: []string{link.DeviceId}})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	//the given version of the thing replaces the stored one: a swap checks against the links as they will be
	things = slices.DeleteFunc(slices.Clone(things), func(t models.Thing) bool { return t.Id == thing.Id })
	things = append(things, thing)
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

// checkThingDeviceLinkGraphRule rejects a link if a graph contains the thing and the device at the same time
// while the link holds: the validities of the thing node, of the device node and the period of the link must have no
// instant in common. Nodes without validity are valid always, so for them any link is rejected.
// Graph nodes are evaluated here, the database filter only narrows the graphs down.
// ListGraphs is called with Limit 0: mongo returns all matches, testdb ignores the options and returns all graphs.
func (this *Controller) checkThingDeviceLinkGraphRule(ctx context.Context, thingId string, link models.ThingDeviceLink) (err error, code int) {
	graphs, _, err := this.db.ListGraphs(ctx, model.GraphListOptions{ThingIds: []string{thingId}})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	for _, graph := range graphs {
		if graphNodesCoexist(graph, thingId, link) {
			//no graph id in the message: the graph may belong to someone else
			return fmt.Errorf("a graph contains thing %v and device %v", thingId, link.DeviceId), http.StatusBadRequest
		}
	}
	return nil, http.StatusOK
}

// graphNodesCoexist reports whether the graph has a node of the thing and a node of the linked device that are valid
// at the same time during the link.
func graphNodesCoexist(graph models.Graph, thingId string, link models.ThingDeviceLink) bool {
	linkRange := [2]*time.Time{&link.From, link.Until}
	for _, thingNode := range graph.Nodes {
		if thingNode.ResourceType != models.GraphResourceTypeThing || thingNode.ResourceId != thingId {
			continue
		}
		for _, deviceNode := range graph.Nodes {
			if deviceNode.ResourceType != models.GraphResourceTypeDevice || deviceNode.ResourceId != link.DeviceId {
				continue
			}
			if models.ValidityIntersects(
				[2]*time.Time{thingNode.ValidFrom, thingNode.ValidUntil},
				[2]*time.Time{deviceNode.ValidFrom, deviceNode.ValidUntil},
				linkRange,
			) {
				return true
			}
		}
	}
	return false
}

func (this *Controller) EndThingDeviceLink(ctx context.Context, token string, thingId string, linkId string, end model.ThingDeviceLinkEnd) (result models.Thing, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	_, err, code = this.checkThingWriteAccess(ctx, token, thingId)
	if err != nil {
		return result, err, code
	}
	return this.editThing(ctx, thingId, func(thing models.Thing, exists bool) (models.Thing, bool, error, int) {
		if !exists {
			return thing, false, fmt.Errorf("thing %w", model.ErrNotFound), http.StatusNotFound
		}
		index := slices.IndexFunc(thing.DeviceLinks, func(link models.ThingDeviceLink) bool { return link.Id == linkId })
		if index < 0 {
			return thing, false, fmt.Errorf("device link %w", model.ErrNotFound), http.StatusNotFound
		}
		thing.DeviceLinks = slices.Clone(thing.DeviceLinks)
		link := thing.DeviceLinks[index]
		if link.Until != nil {
			return thing, false, errors.New("device link is already ended"), http.StatusBadRequest
		}
		until := normalizeThingTime(time.Now())
		if !until.After(link.From) {
			//the link starts now or in the future: end it one millisecond after its start, as endOpenThingLinksOfDevice does
			until = link.From.Add(time.Millisecond)
		}
		if end.Until != nil {
			until = normalizeThingTime(*end.Until)
		}
		link, err := endThingDeviceLinkAt(link, until, end.UntilValues)
		if err != nil {
			return thing, false, err, http.StatusBadRequest
		}
		thing.DeviceLinks[index] = link
		err = thing.Valid()
		if err != nil {
			return thing, false, fmt.Errorf("invalid thing: %w", err), http.StatusBadRequest
		}
		return thing, true, nil, http.StatusOK
	})
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
	for _, listed := range things {
		_, err, _ = this.editThing(ctx, listed.Id, func(thing models.Thing, exists bool) (models.Thing, bool, error, int) {
			if !exists {
				return thing, false, nil, http.StatusOK
			}
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
			thing.DeviceLinks = links
			return thing, changed, nil, http.StatusOK
		})
		if err != nil {
			this.config.GetLogger().Error("unable to end thing device links of deleted device", "thingId", listed.Id, "deviceId", deviceId, "error", err)
			return err
		}
	}
	return nil
}

// DeleteThings is the batch variant of DeleteThing: the rights of all existing ids are checked before the first one is deleted.
// Unknown ids are ignored.
func (this *Controller) DeleteThings(ctx context.Context, token string, ids []string) (err error, code int) {
	ids = removeDuplicates(ids)
	if len(ids) == 0 {
		return nil, http.StatusOK
	}
	ctx, cancel := getBatchTimeoutContext(ctx)
	defer cancel()
	list, _, err := this.db.ListThings(ctx, model.ThingListOptions{Ids: ids, Limit: int64(len(ids))})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	existing := existingIds(list, ids, func(e models.Thing) string { return e.Id })
	if len(existing) == 0 {
		return nil, http.StatusOK
	}
	err, code = this.checkAdministrateForBatchDelete(ctx, token, this.config.ThingTopic, existing)
	if err != nil {
		return err, code
	}
	err = this.db.RemoveThings(ctx, existing, withContext(ctx, this.deleteThingsSyncHandler))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteThingsSyncHandler(ctx context.Context, olds []models.Thing) (err error) {
	ids := make([]string, 0, len(olds))
	for _, old := range olds {
		ids = append(ids, old.Id)
	}
	err = this.removeResourcesFromGraphs(ctx, models.GraphResourceTypeThing, ids)
	if err != nil {
		return err
	}
	err = this.RemoveRightsOfMany(ctx, this.config.ThingTopic, ids)
	if err != nil {
		return err
	}
	return this.publisher.PublishThingDeletes(ids)
}

// SwapThingDeviceLink ends the open link of a thing at the given instant and starts a link to a device from there,
// the device may be the same (counter reset, change of the factor) or another one (device or meter replaced).
// Ending and starting are one write of the thing, so either both hold or neither: every check runs on the thing as it
// will be before the write, and the write is compare-and-set like every thing write (see editThing).
// The new link follows the rules of AddThingDeviceLink, i.e. the device may not be linked
// elsewhere from the instant on, and the thing and the device may not share a graph then.
// Needs write on the thing and read on the device.
func (this *Controller) SwapThingDeviceLink(ctx context.Context, token string, thingId string, linkId string, swap model.ThingDeviceLinkSwap) (result models.Thing, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	_, err, code = this.checkThingWriteAccess(ctx, token, thingId)
	if err != nil {
		return result, err, code
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	err, code = this.checkLinkableDevice(ctx, token, jwtToken, swap.DeviceId)
	if err != nil {
		return result, err, code
	}
	for _, reading := range swap.Readings {
		if reading.UntilValue != nil {
			return result, errors.New("until_value is not allowed in the readings of a new link, end the link to set it"), http.StatusBadRequest
		}
	}
	at := normalizeThingTime(time.Now())
	if swap.At != nil {
		at = normalizeThingTime(*swap.At)
	}
	link, err, code := newThingDeviceLink(model.ThingDeviceLinkCreate{DeviceId: swap.DeviceId, From: &at, Readings: swap.Readings, Factor: swap.Factor})
	if err != nil {
		return result, err, code
	}

	return this.editThing(ctx, thingId, func(thing models.Thing, exists bool) (models.Thing, bool, error, int) {
		if !exists {
			return thing, false, fmt.Errorf("thing %w", model.ErrNotFound), http.StatusNotFound
		}
		index := slices.IndexFunc(thing.DeviceLinks, func(link models.ThingDeviceLink) bool { return link.Id == linkId })
		if index < 0 {
			return thing, false, fmt.Errorf("device link %w", model.ErrNotFound), http.StatusNotFound
		}
		thing.DeviceLinks = slices.Clone(thing.DeviceLinks)
		if from := thing.DeviceLinks[index].From; !at.After(from) {
			return thing, false, fmt.Errorf("at %v must be after the start of the device link, %v", at.Format(time.RFC3339Nano), from.Format(time.RFC3339Nano)), http.StatusBadRequest
		}
		ended, err := endThingDeviceLinkAt(thing.DeviceLinks[index], at, swap.UntilValues)
		if err != nil {
			return thing, false, err, http.StatusBadRequest
		}
		thing.DeviceLinks[index] = ended
		err, code := this.checkThingDeviceLinkOverlap(ctx, thing, link)
		if err != nil {
			return thing, false, err, code
		}
		err, code = this.checkThingDeviceLinkGraphRule(ctx, thing.Id, link)
		if err != nil {
			return thing, false, err, code
		}
		thing.DeviceLinks = append(thing.DeviceLinks, link)
		err = thing.Valid()
		if err != nil {
			return thing, false, fmt.Errorf("invalid thing: %w", err), http.StatusBadRequest
		}
		return thing, true, nil, http.StatusOK
	})
}
