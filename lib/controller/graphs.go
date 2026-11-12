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

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
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
	return result, total, nil, http.StatusOK
}

func (this *Controller) ReadGraph(ctx context.Context, token string, id string) (result models.Graph, err error, errCode int) {
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
		return models.Graph{}, err, errCode
	}
	if !exists {
		return models.Graph{}, fmt.Errorf("graph %w", model.ErrNotFound), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

// removeResourceFromGraphs deletes every node of the given resource from all graphs containing it.
// It is internal and neither checks permissions nor the thing/device rule of SetGraph.
// ListGraphs is called with Limit 0: mongo returns all matches, testdb ignores the options and returns all graphs.
func (this *Controller) removeResourceFromGraphs(ctx context.Context, resourceType models.GraphResourceType, resourceId string) error {
	options := model.GraphListOptions{}
	switch resourceType {
	case models.GraphResourceTypeDevice:
		options.DeviceIds = []string{resourceId}
	case models.GraphResourceTypeThing:
		options.ThingIds = []string{resourceId}
	default:
		return fmt.Errorf("unsupported graph resource type %q", resourceType)
	}
	ctx, _ = getTimeoutContext(ctx)
	graphs, _, err := this.db.ListGraphs(ctx, options)
	if err != nil {
		return err
	}
	for _, graph := range graphs {
		removed := false
		for _, node := range slices.Clone(graph.Nodes) {
			if node.ResourceType == resourceType && node.ResourceId == resourceId {
				graph.DeleteNode(node.Id)
				removed = true
			}
		}
		if !removed {
			continue
		}
		err = graph.Valid()
		if err != nil {
			this.config.GetLogger().Error("graph.DeleteNode created invalid graph", "graphId", graph.Id, "resourceType", resourceType, "resourceId", resourceId, "error", err)
			return err
		}
		err = this.setGraph(ctx, graph)
		if err != nil {
			this.config.GetLogger().Error("unable to update graph to remove resource", "graphId", graph.Id, "resourceType", resourceType, "resourceId", resourceId, "error", err)
			return err
		}
	}
	return nil
}

func (this *Controller) SetGraph(ctx context.Context, token string, graph models.Graph) (result models.Graph, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
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

	ctx, _ = getTimeoutContext(ctx)
	original, exists, err := this.db.GetGraph(ctx, graph.Id)
	if err != nil {
		return graph, err, code
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
			return graph, err, code
		}
	}

	err = graph.Valid()
	if err != nil {
		return graph, fmt.Errorf("invalid graph: %w", err), http.StatusBadRequest
	}

	err, code = this.checkGraphThingRule(ctx, graph)
	if err != nil {
		return graph, err, code
	}

	err = this.setGraph(ctx, graph)
	if err != nil {
		return graph, err, http.StatusInternalServerError
	}

	return graph, nil, http.StatusOK
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

// checkGraphThingRule rejects graphs that contain a thing and a device that is (or was) linked to this thing.
// The things of the nodes must exist. Links are evaluated here and not by a database filter,
// so the result is the same for mongo and testdb.
func (this *Controller) checkGraphThingRule(ctx context.Context, graph models.Graph) (err error, code int) {
	deviceNodes := map[string]bool{}
	for _, node := range graph.Nodes {
		if node.ResourceType == models.GraphResourceTypeDevice {
			deviceNodes[node.ResourceId] = true
		}
	}
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
			if deviceNodes[link.DeviceId] {
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
