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
	"runtime/debug"
	"slices"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"github.com/google/uuid"
)

func (this *Controller) ListGraphs(ctx context.Context, token string, options model.GraphListOptions) (result []models.Graph, total int64, err error, errCode int) {
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
			ids, err, _ = this.permissionsV2Client.ListAccessibleResourceIdsContext(ctx, token, this.config.GraphTopic, client.ListOptions{}, client.Permission(permissionFlag))
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
			if ids == nil {
				//nil would mean "no id filter" for the database: a caller who may read nothing gets nothing
				ids = []string{}
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err, _ := this.permissionsV2Client.CheckMultiplePermissionsContext(ctx, token, this.config.GraphTopic, options.Ids, client.Permission(permissionFlag))
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
	result, total, err = this.db.ListGraphs(ctx, options)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	now := time.Now()
	for i := range result {
		result[i] = result[i].At(now)
	}
	return result, total, nil, http.StatusOK
}

// ReadGraph answers the graph as it is now, the projection of the stored document at the current time.
func (this *Controller) ReadGraph(ctx context.Context, token string, id string) (result models.Graph, err error, errCode int) {
	return this.ReadGraphAt(ctx, token, id, time.Now())
}

// ReadGraphAt answers the projection of the graph at the given instant, an ordinary graph.
// If no node is valid then, the graph is answered without nodes and edges.
func (this *Controller) ReadGraphAt(ctx context.Context, token string, id string, at time.Time) (result models.Graph, err error, errCode int) {
	stored, err, errCode := this.readStoredGraph(ctx, token, id)
	if err != nil {
		return result, err, errCode
	}
	return stored.At(at), nil, http.StatusOK
}

// ReadGraphVersions answers the stored document with all versions of its elements.
func (this *Controller) ReadGraphVersions(ctx context.Context, token string, id string) (result models.Graph, err error, errCode int) {
	return this.readStoredGraph(ctx, token, id)
}

// ReadGraphSegments cuts the range at every validity bound inside it and answers each piece with its projection.
// A nil start or end leaves the range open on that side.
func (this *Controller) ReadGraphSegments(ctx context.Context, token string, id string, start *time.Time, end *time.Time) (result []models.GraphSegment, err error, errCode int) {
	stored, err, errCode := this.readStoredGraph(ctx, token, id)
	if err != nil {
		return result, err, errCode
	}
	result, err = stored.Segments(start, end)
	if err != nil {
		return nil, err, http.StatusBadRequest
	}
	return result, nil, http.StatusOK
}

func (this *Controller) readStoredGraph(ctx context.Context, token string, id string) (result models.Graph, err error, errCode int) {
	ok, err, _ := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.GraphTopic, id, client.Read)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	ctx, _ = getTimeoutContext(ctx)
	result, exists, err := this.db.GetGraph(ctx, id)
	if err != nil {
		return models.Graph{}, err, http.StatusInternalServerError
	}
	if !exists {
		return models.Graph{}, fmt.Errorf("graph %w", model.ErrNotFound), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

// removeResourceFromGraphs deletes every node of the given resource from all graphs containing it.
func (this *Controller) removeResourceFromGraphs(ctx context.Context, resourceType models.GraphResourceType, resourceId string) error {
	return this.removeResourcesFromGraphs(ctx, resourceType, []string{resourceId})
}

// graphWriteAttempts is how often a write that lost against a concurrent write of the same graph is applied again.
const graphWriteAttempts = 5

// removeResourcesFromGraphs deletes every node of the given resources from all graphs containing any of them.
// Each affected graph is read once and written once, however many of the resources it contains:
// a write is a mongo update, a rights call and a kafka message, so a bulk delete must not repeat it per resource.
// It is internal and neither checks permissions nor the thing/device rule of SetGraph.
// ListGraphs is called with Limit 0: mongo returns all matches, testdb ignores the options and returns all graphs.
//
// The nodes are removed from the graph as it is now. A graph without validities is written as that.
// A graph with validities is written as a change at the current time: the nodes and their edges end now,
// the rerouted edges start now, and the history stays.
func (this *Controller) removeResourcesFromGraphs(ctx context.Context, resourceType models.GraphResourceType, resourceIds []string) error {
	if len(resourceIds) == 0 {
		return nil
	}
	options := model.GraphListOptions{}
	switch resourceType {
	case models.GraphResourceTypeDevice:
		options.DeviceIds = resourceIds
	case models.GraphResourceTypeThing:
		options.ThingIds = resourceIds
	default:
		return fmt.Errorf("unsupported graph resource type %q", resourceType)
	}
	removedIds := make(map[string]bool, len(resourceIds))
	for _, id := range resourceIds {
		removedIds[id] = true
	}
	listCtx, _ := getTimeoutContext(ctx)
	graphs, _, err := this.db.ListGraphs(listCtx, options)
	if err != nil {
		return err
	}
	for _, listed := range graphs {
		err = this.removeResourcesFromGraph(ctx, listed.Id, resourceType, removedIds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Controller) removeResourcesFromGraph(ctx context.Context, id string, resourceType models.GraphResourceType, removedIds map[string]bool) error {
	for attempt := 0; attempt < graphWriteAttempts; attempt++ {
		readCtx, _ := getTimeoutContext(ctx)
		stored, revision, exists, err := this.db.GetGraphWithRevision(readCtx, id)
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
		now := normalizeThingTime(time.Now())
		graph := stored.At(now)
		removed := false
		for _, node := range slices.Clone(graph.Nodes) {
			if node.ResourceType == resourceType && removedIds[node.ResourceId] {
				graph.DeleteNode(node.Id)
				removed = true
			}
		}
		if !removed {
			return nil
		}
		if len(stored.ValidityBounds()) > 0 {
			graph, err = stored.Edit(graph, now, nil, models.GraphEditModeChange)
			if err != nil {
				this.config.GetLogger().Error("unable to remove resources from graph", "graphId", id, "resourceType", resourceType, "resources", len(removedIds), "error", err)
				return err
			}
		}
		err = graph.Valid()
		if err != nil {
			this.config.GetLogger().Error("graph.DeleteNode created invalid graph", "graphId", id, "resourceType", resourceType, "resources", len(removedIds), "error", err)
			return err
		}
		swapped, err := this.setGraphIfRevision(ctx, graph, revision)
		if err != nil {
			this.config.GetLogger().Error("unable to update graph to remove resources", "graphId", id, "resourceType", resourceType, "resources", len(removedIds), "error", err)
			return err
		}
		if swapped {
			return nil
		}
	}
	return fmt.Errorf("graph %v was changed concurrently, unable to remove resources", id)
}

// SetGraph creates a graph or corrects an existing one at the current time, see EditGraph.
func (this *Controller) SetGraph(ctx context.Context, token string, graph models.Graph) (result models.Graph, err error, code int) {
	return this.EditGraph(ctx, token, graph, model.GraphEditOptions{})
}

// EditGraph creates a graph, or writes an existing one: the given graph is the edited projection at options.At,
// and models.Graph.Edit turns the difference to the stored projection into the new stored document.
// Versions that are not valid at options.At are never touched. The stored document is validated for every segment
// between its validity bounds before it is written.
// For a graph that does not exist yet the options are ignored: the graph is stored as it is given.
// The result is the projection of the stored document at the current time.
func (this *Controller) EditGraph(ctx context.Context, token string, graph models.Graph, options model.GraphEditOptions) (result models.Graph, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	at, until, mode, err := normalizeGraphEditOptions(options)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if graph.Id == "" {
		graph.GenerateId()
	} else {
		if !jwtToken.IsAdmin() {
			ok, err, code := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.GraphTopic, graph.Id, client.Write)
			if err != nil {
				return graph, err, code
			}
			if !ok {
				return graph, errors.New("access denied"), http.StatusForbidden
			}
		}
	}

	body := graph
	for attempt := 0; attempt < graphWriteAttempts; attempt++ {
		result, swapped, err, code := this.editGraphOnce(ctx, token, jwtToken, body, at, until, mode, nil)
		if err != nil || swapped {
			return result, err, code
		}
		//somebody wrote the graph between reading and writing it. The body is a diff against the projection at 'at',
		//so it is applied again to what is stored now.
	}
	return body, fmt.Errorf("graph %v was changed concurrently, please try again", body.Id), http.StatusConflict
}

// editGraphOnce reads the stored graph, applies the edit and writes the result if the stored graph is still the one that was read.
//
// With transform the body is not given but built from the stored graph: it is called with the projection at 'at' of what
// was just read and returns the edited projection, so every attempt starts from the current state.
func (this *Controller) editGraphOnce(ctx context.Context, token string, jwtToken jwt.Token, graph models.Graph, at time.Time, until *time.Time, mode string, transform graphTransform) (result models.Graph, swapped bool, err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	original, revision, exists, err := this.db.GetGraphWithRevision(ctx, graph.Id)
	if err != nil {
		return graph, false, err, http.StatusInternalServerError
	}
	if transform != nil {
		if !exists {
			return graph, false, fmt.Errorf("graph %w", model.ErrNotFound), http.StatusNotFound
		}
		graph, err, code = transform(original, original.At(at))
		if err != nil {
			return graph, false, err, code
		}
	}

	//set graph owner if none is given
	//prefer existing owner, fallback to requesting user
	if graph.Owner == "" {
		graph.Owner = original.Owner //may be empty for new graphs
	}
	if graph.Owner == "" {
		graph.Owner = jwtToken.GetUserId()
	}

	if exists && original.Owner != graph.Owner && original.Owner != "" {
		err, code = this.checkOwnerChange(ctx, token, jwtToken, this.config.GraphTopic, graph.Id, original.Owner, graph.Owner)
		if err != nil {
			return graph, false, err, code
		}
	}

	if exists {
		graph, err = original.Edit(graph, at, until, mode)
		if err != nil {
			return graph, false, fmt.Errorf("invalid graph edit: %w", err), http.StatusBadRequest
		}
	} else {
		graph = normalizeGraphTimes(graph)
	}

	err = graph.Valid()
	if err != nil {
		return graph, false, fmt.Errorf("invalid graph: %w", err), http.StatusBadRequest
	}

	err, code = this.checkGraphThingRule(ctx, graph)
	if err != nil {
		return graph, false, err, code
	}

	swapped, err = this.setGraphIfRevision(ctx, graph, revision)
	if err != nil {
		return graph, false, err, http.StatusInternalServerError
	}
	return graph.At(time.Now()), swapped, nil, http.StatusOK
}

// graphTransform gets the stored document and its projection at one instant, and edits the projection of a stored graph at one instant and answers the edited projection.
type graphTransform func(stored models.Graph, projection models.Graph) (edited models.Graph, err error, code int)

// editGraphByTransform is EditGraph for an edit that is a function of the stored graph: it needs write on the graph,
// reads the projection at 'at', lets transform edit it and writes the result as a change (bounded by 'until' if given),
// with the same compare-and-set loop and validation as EditGraph. The graph must exist.
func (this *Controller) editGraphByTransform(ctx context.Context, token string, id string, at time.Time, until *time.Time, transform graphTransform) (result models.Graph, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusBadRequest
	}
	if !jwtToken.IsAdmin() {
		ok, err, code := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.GraphTopic, id, client.Write)
		if err != nil {
			return result, err, code
		}
		if !ok {
			return result, errors.New("access denied"), http.StatusForbidden
		}
	}
	for attempt := 0; attempt < graphWriteAttempts; attempt++ {
		result, swapped, err, code := this.editGraphOnce(ctx, token, jwtToken, models.Graph{Id: id}, at, until, models.GraphEditModeChange, transform)
		if err != nil || swapped {
			return result, err, code
		}
	}
	return result, fmt.Errorf("graph %v was changed concurrently, please try again", id), http.StatusConflict
}

// MoveGraphNode moves a node below another node from the instant on: the outgoing edges of the node in the projection
// at that instant are replaced by one edge to the new parent with weight 100. With until the move holds for
// [at, until) and the old edges are resumed there (A, B, A). The node needs an outgoing edge, so the end node
// can not be moved. Needs write on the graph.
func (this *Controller) MoveGraphNode(ctx context.Context, token string, graphId string, nodeId string, move model.GraphNodeMove) (result models.Graph, err error, code int) {
	at := normalizeThingTime(time.Now())
	if move.At != nil {
		at = normalizeThingTime(*move.At)
	}
	var until *time.Time
	if move.Until != nil {
		u := normalizeThingTime(*move.Until)
		if !u.After(at) {
			return result, errors.New("until must be after at"), http.StatusBadRequest
		}
		until = &u
	}
	if move.ToNodeId == "" {
		return result, errors.New("missing to_node_id"), http.StatusBadRequest
	}
	if move.ToNodeId == nodeId {
		return result, errors.New("a node can not be moved below itself"), http.StatusBadRequest
	}
	return this.editGraphByTransform(ctx, token, graphId, at, until, func(stored models.Graph, graph models.Graph) (models.Graph, error, int) {
		atText := at.Format(time.RFC3339Nano)
		if !slices.ContainsFunc(graph.Nodes, func(n models.Node) bool { return n.Id == nodeId }) {
			return graph, fmt.Errorf("node %v is not valid at %v", nodeId, atText), http.StatusBadRequest
		}
		if !slices.ContainsFunc(graph.Nodes, func(n models.Node) bool { return n.Id == move.ToNodeId }) {
			return graph, fmt.Errorf("node %v is not valid at %v", move.ToNodeId, atText), http.StatusBadRequest
		}
		outgoing := 0
		for _, edge := range graph.Edges {
			if edge.FromNodeId != nodeId {
				continue
			}
			outgoing++
			if edge.ToNodeId == move.ToNodeId && edge.Weight == 100 {
				return graph, fmt.Errorf("node %v is already below %v at %v", nodeId, move.ToNodeId, atText), http.StatusBadRequest
			}
		}
		if outgoing == 0 {
			return graph, fmt.Errorf("node %v has no outgoing edge at %v and can not be moved", nodeId, atText), http.StatusBadRequest
		}
		if later := plannedChangeAfter(stored, nodeId, at, until, false); later != nil {
			return graph, fmt.Errorf("node %v has changes from %v; move it after that or bound the move with until, to end before it", nodeId, later.Format(time.RFC3339Nano)), http.StatusBadRequest
		}
		graph.Edges = slices.DeleteFunc(slices.Clone(graph.Edges), func(e models.Edge) bool { return e.FromNodeId == nodeId })
		graph.Edges = append(graph.Edges, models.Edge{Id: uuid.NewString(), FromNodeId: nodeId, ToNodeId: move.ToNodeId, Weight: 100})
		return graph, nil, http.StatusOK
	})
}

// EndGraphNode takes a node out of the graph from the instant on: the node and its edges end there, and the edges
// that led to it are rerouted to where it led, as Graph.DeleteNode does, from there. Everything before the instant stays.
// A thing or device the node stands for is not touched; ending the device link of a meter is the job of
// SwapThingDeviceLink / EndThingDeviceLink. Needs write on the graph.
func (this *Controller) EndGraphNode(ctx context.Context, token string, graphId string, nodeId string, end model.GraphNodeEnd) (result models.Graph, err error, code int) {
	at := normalizeThingTime(time.Now())
	if end.At != nil {
		at = normalizeThingTime(*end.At)
	}
	return this.editGraphByTransform(ctx, token, graphId, at, nil, func(stored models.Graph, graph models.Graph) (models.Graph, error, int) {
		if !slices.ContainsFunc(graph.Nodes, func(n models.Node) bool { return n.Id == nodeId }) {
			return graph, fmt.Errorf("node %v is not valid at %v", nodeId, at.Format(time.RFC3339Nano)), http.StatusBadRequest
		}
		if later := plannedChangeAfter(stored, nodeId, at, nil, true); later != nil {
			return graph, fmt.Errorf("node %v has changes from %v; end it after that", nodeId, later.Format(time.RFC3339Nano)), http.StatusBadRequest
		}
		graph.DeleteNode(nodeId)
		return graph, nil, http.StatusOK
	})
}

// plannedChangeAfter answers the first instant after 'at' at which an edge of the node starts in the stored document: its
// outgoing edges, and with incoming the edges that lead to it. Only starts inside [at, until) count (after 'at' if until is nil).
// An action built from the projection at 'at' does not know those versions and would produce a document that is
// invalid at that instant, so the actions refuse up front and name it. nil if there is none.
func plannedChangeAfter(stored models.Graph, nodeId string, at time.Time, until *time.Time, incoming bool) (first *time.Time) {
	for _, edge := range stored.Edges {
		if edge.FromNodeId != nodeId && !(incoming && edge.ToNodeId == nodeId) {
			continue
		}
		from := edge.ValidFrom
		if from == nil || !from.After(at) || (until != nil && !from.Before(*until)) {
			continue
		}
		if first == nil || from.Before(*first) {
			first = from
		}
	}
	return first
}

// normalizeGraphEditOptions fills the defaults and cuts the instants to milliseconds in UTC,
// the precision of mongo, so the test database and mongo answer the same.
func normalizeGraphEditOptions(options model.GraphEditOptions) (at time.Time, until *time.Time, mode string, err error) {
	at = normalizeThingTime(time.Now())
	if options.At != nil {
		at = normalizeThingTime(*options.At)
	}
	mode = options.Mode
	if mode == "" {
		mode = models.GraphEditModeCorrection
	}
	if mode != models.GraphEditModeChange && mode != models.GraphEditModeCorrection {
		return at, nil, mode, fmt.Errorf("unknown mode %q, use %v or %v", mode, models.GraphEditModeChange, models.GraphEditModeCorrection)
	}
	if options.Until != nil {
		if mode != models.GraphEditModeChange {
			return at, nil, mode, fmt.Errorf("until is only allowed with mode %v", models.GraphEditModeChange)
		}
		u := normalizeThingTime(*options.Until)
		until = &u
		if !until.After(at) {
			return at, nil, mode, errors.New("until must be after at")
		}
	}
	return at, until, mode, nil
}

// normalizeGraphTimes cuts every validity bound of a graph to milliseconds in UTC. The given graph is not modified.
func normalizeGraphTimes(graph models.Graph) models.Graph {
	normalize := func(t *time.Time) *time.Time {
		if t == nil {
			return nil
		}
		n := normalizeThingTime(*t)
		return &n
	}
	attributes := func(list []models.Attribute) []models.Attribute {
		if list == nil {
			return nil
		}
		list = slices.Clone(list)
		for i := range list {
			list[i].ValidFrom = normalize(list[i].ValidFrom)
			list[i].ValidUntil = normalize(list[i].ValidUntil)
		}
		return list
	}
	graph.Attributes = attributes(graph.Attributes)
	if graph.Nodes != nil {
		graph.Nodes = slices.Clone(graph.Nodes)
		for i := range graph.Nodes {
			graph.Nodes[i].ValidFrom = normalize(graph.Nodes[i].ValidFrom)
			graph.Nodes[i].ValidUntil = normalize(graph.Nodes[i].ValidUntil)
			graph.Nodes[i].Attributes = attributes(graph.Nodes[i].Attributes)
		}
	}
	if graph.Edges != nil {
		graph.Edges = slices.Clone(graph.Edges)
		for i := range graph.Edges {
			graph.Edges[i].ValidFrom = normalize(graph.Edges[i].ValidFrom)
			graph.Edges[i].ValidUntil = normalize(graph.Edges[i].ValidUntil)
			graph.Edges[i].Attributes = attributes(graph.Edges[i].Attributes)
		}
	}
	return graph
}

// checkOwnerChange checks a change of the owner of an existing resource: only users with administrate rights may do it,
// and the new owner must already have administrate rights on the resource.
// Callers only use it for existing resources with a previous owner.
func (this *Controller) checkOwnerChange(ctx context.Context, token string, jwtToken jwt.Token, topic string, id string, originalOwner string, newOwner string) (err error, code int) {
	if !jwtToken.IsAdmin() {
		ok, err, code := this.permissionsV2Client.CheckPermissionContext(ctx, token, topic, id, client.Administrate)
		if err != nil {
			return err, code
		}
		if !ok {
			return errors.New("only admins may set new owner"), http.StatusBadRequest
		}
	}

	rights, err, code := this.permissionsV2Client.GetResourceContext(ctx, client.InternalAdminToken, topic, id)
	if err != nil && code != http.StatusNotFound {
		this.config.GetLogger().Error("unable to get permission resource", "error", err)
		debug.PrintStack()
		return err, code
	}

	//new owner-id must be existing admin user (ignore for resources without permission entry or with unchanged owner)
	if code != http.StatusNotFound && newOwner != originalOwner && !rights.UserPermissions[newOwner].Administrate {
		return errors.New("new owner must have existing user admin rights"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

// checkGraphThingRule rejects graphs that contain a thing and a device that is (or was) linked to this thing
// at the same time: the validity of the thing node, the validity of the device node and the period of a link between
// them must have no instant in common. Nodes and links without bounds are valid always, so there any link is rejected.
// That lets a device node be replaced by a thing node: the device node ends where the thing node starts.
// The things of the nodes must exist, whatever the validity of the node. Links are evaluated here and not by a
// database filter, so the result is the same for mongo and testdb.
func (this *Controller) checkGraphThingRule(ctx context.Context, graph models.Graph) (err error, code int) {
	ctx, _ = getTimeoutContext(ctx)
	for _, node := range graph.Nodes {
		if node.ResourceType != models.GraphResourceTypeThing {
			continue
		}
		thing, exists, err := this.db.GetThing(ctx, node.ResourceId)
		if err != nil {
			return err, http.StatusInternalServerError
		}
		if !exists {
			return fmt.Errorf("graph node %v references unknown thing %v", node.Id, node.ResourceId), http.StatusBadRequest
		}
		for _, link := range thing.DeviceLinks {
			if graphNodesCoexist(graph, thing.Id, link) {
				return fmt.Errorf("graph contains thing %v and device %v, which is linked to this thing", thing.Id, link.DeviceId), http.StatusBadRequest
			}
		}
	}
	return nil, http.StatusOK
}

func (this *Controller) setGraph(ctx context.Context, graph models.Graph) (err error) {
	this.config.GetLogger().Debug("create/update graph", "id", graph.Id)
	ctx, _ = getTimeoutContext(ctx)
	err = this.db.SetGraph(ctx, graph, withContext(ctx, this.setGraphSyncHandler))
	if err != nil {
		return err
	}
	return nil
}

// setGraphIfRevision writes the graph if the stored graph still has the given revision, see database.SetGraphIfRevision.
func (this *Controller) setGraphIfRevision(ctx context.Context, graph models.Graph, revision int64) (swapped bool, err error) {
	this.config.GetLogger().Debug("create/update graph", "id", graph.Id, "revision", revision)
	ctx, _ = getTimeoutContext(ctx)
	return this.db.SetGraphIfRevision(ctx, graph, revision, withContext(ctx, this.setGraphSyncHandler))
}

func (this *Controller) setGraphSyncHandler(ctx context.Context, graph models.Graph) (err error) {
	err = this.EnsureInitialRights(ctx, this.config.GraphTopic, graph.Id, graph.Owner)
	if err != nil {
		return fmt.Errorf("unable to ensure initial graph permissions: %w", err)
	}
	return this.publisher.PublishGraph(graph)
}

func (this *Controller) DeleteGraph(ctx context.Context, token string, id string) (error, int) {
	ctx, _ = getTimeoutContext(ctx)
	_, exists, err := this.db.GetGraph(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return nil, http.StatusOK
	}
	ok, err, _ := this.permissionsV2Client.CheckPermissionContext(ctx, token, this.config.GraphTopic, id, client.Administrate)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("access denied"), http.StatusForbidden
	}
	err = this.deleteGraph(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteGraphSyncHandler(ctx context.Context, old models.Graph) (err error) {
	err = this.RemoveRights(ctx, this.config.GraphTopic, old.Id)
	if err != nil {
		return err
	}
	return this.publisher.PublishGraphDelete(old.Id)
}

func (this *Controller) deleteGraph(ctx context.Context, id string) error {
	ctx, _ = getTimeoutContext(ctx)
	err := this.db.RemoveGraph(ctx, id, withContext(ctx, this.deleteGraphSyncHandler))
	if err != nil {
		return err
	}
	return nil
}

// DeleteGraphs is the batch variant of DeleteGraph: the rights of all existing ids are checked before the first one is deleted.
// Unknown ids are ignored.
func (this *Controller) DeleteGraphs(ctx context.Context, token string, ids []string) (err error, code int) {
	ids = removeDuplicates(ids)
	if len(ids) == 0 {
		return nil, http.StatusOK
	}
	ctx, cancel := getBatchTimeoutContext(ctx)
	defer cancel()
	list, _, err := this.db.ListGraphs(ctx, model.GraphListOptions{Ids: ids, Limit: int64(len(ids))})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	existing := existingIds(list, ids, func(e models.Graph) string { return e.Id })
	if len(existing) == 0 {
		return nil, http.StatusOK
	}
	err, code = this.checkAdministrateForBatchDelete(ctx, token, this.config.GraphTopic, existing)
	if err != nil {
		return err, code
	}
	err = this.db.RemoveGraphs(ctx, existing, withContext(ctx, this.deleteGraphsSyncHandler))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteGraphsSyncHandler(ctx context.Context, olds []models.Graph) (err error) {
	ids := make([]string, 0, len(olds))
	for _, old := range olds {
		ids = append(ids, old.Id)
	}
	err = this.RemoveRightsOfMany(ctx, this.config.GraphTopic, ids)
	if err != nil {
		return err
	}
	return this.publisher.PublishGraphDeletes(ids)
}
