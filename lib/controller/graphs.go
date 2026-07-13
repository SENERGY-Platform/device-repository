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
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func (this *Controller) ListGraphs(token string, options model.GraphListOptions) (result []models.Graph, total int64, err error, errCode int) {
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
			ids, err, _ = this.permissionsV2Client.ListAccessibleResourceIds(token, this.config.GraphTopic, client.ListOptions{}, client.Permission(permissionFlag))
			if err != nil {
				return result, total, err, http.StatusInternalServerError
			}
		}
	} else {
		options.Limit = 0
		options.Offset = 0
		idMap, err, _ := this.permissionsV2Client.CheckMultiplePermissions(token, this.config.GraphTopic, options.Ids, client.Permission(permissionFlag))
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
	result, total, err = this.db.ListGraphs(ctx, options)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) ReadGraph(token string, id string) (result models.Graph, err error, errCode int) {
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.GraphTopic, id, client.Read)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !ok {
		return result, errors.New("access denied"), http.StatusForbidden
	}
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetGraph(ctx, id)
	if err != nil {
		return models.Graph{}, err, errCode
	}
	if !exists {
		return models.Graph{}, fmt.Errorf("graph %w", model.ErrNotFound), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) removeDeviceFromGraphs(deviceId string) error {
	ctx, _ := getTimeoutContext()
	graphs, _, err := this.db.ListGraphs(ctx, model.GraphListOptions{DeviceIds: []string{deviceId}})
	if err != nil {
		return err
	}
	for _, graph := range graphs {
		for _, node := range graph.Nodes {
			if node.ResourceType == models.GraphResourceTypeDevice && node.ResourceId == deviceId {
				graph.DeleteNode(node.Id)
			}
		}
		err = graph.Valid()
		if err != nil {
			this.config.GetLogger().Error("graph.DeleteNode created invalid graph", "graphId", graph.Id, "deviceId", deviceId, "error", err)
			return err
		}
		err = this.setGraph(graph)
		if err != nil {
			this.config.GetLogger().Error("unable to update graph to remove device", "graphId", graph.Id, "deviceId", deviceId, "error", err)
			return err
		}
	}
	return nil
}

func (this *Controller) SetGraph(token string, graph models.Graph) (result models.Graph, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if graph.Id == "" {
		graph.GenerateId()
	} else {
		if !jwtToken.IsAdmin() {
			ok, err, code := this.permissionsV2Client.CheckPermission(token, this.config.GraphTopic, graph.Id, client.Write)
			if err != nil {
				return graph, err, code
			}
			if !ok {
				return graph, errors.New("access denied"), http.StatusForbidden
			}
		}
	}

	ctx, _ := getTimeoutContext()
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
		if !jwtToken.IsAdmin() {
			ok, err, code := this.permissionsV2Client.CheckPermission(token, this.config.GraphTopic, graph.Id, client.Administrate)
			if err != nil {
				return graph, err, code
			}
			if !ok {
				return graph, fmt.Errorf("only admins may set new owner: %w", err), http.StatusBadRequest
			}
		}

		rights, err, code := this.permissionsV2Client.GetResource(client.InternalAdminToken, this.config.GraphTopic, graph.Id)
		if err != nil && code != http.StatusNotFound {
			this.config.GetLogger().Error("unable to get permission resource", "error", err)
			debug.PrintStack()
			return graph, err, code
		}

		//new graph owner-id must be existing admin user (ignore for new graphs or graphs with unchanged owner)
		if code != http.StatusNotFound && graph.Owner != original.Owner && !rights.UserPermissions[graph.Owner].Administrate {
			return graph, errors.New("new owner must have existing user admin rights"), http.StatusBadRequest
		}

	}

	err = graph.Valid()
	if err != nil {
		return graph, fmt.Errorf("invalid graph: %w", err), http.StatusBadRequest
	}

	err = this.setGraph(graph)
	if err != nil {
		return graph, err, http.StatusInternalServerError
	}

	return graph, nil, http.StatusOK
}

func (this *Controller) setGraph(graph models.Graph) (err error) {
	this.config.GetLogger().Debug("create/update graph", "id", graph.Id)
	ctx, _ := getTimeoutContext()
	err = this.db.SetGraph(ctx, graph, this.setGraphSyncHandler)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) setGraphSyncHandler(graph models.Graph) (err error) {
	err = this.EnsureInitialRights(this.config.GraphTopic, graph.Id, graph.Owner)
	if err != nil {
		return fmt.Errorf("unable to ensure initial graph permissions: %w", err)
	}
	return nil
}

func (this *Controller) DeleteGraph(token string, id string) (error, int) {
	ctx, _ := getTimeoutContext()
	_, exists, err := this.db.GetGraph(ctx, id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !exists {
		return nil, http.StatusOK
	}
	ok, err, _ := this.permissionsV2Client.CheckPermission(token, this.config.GraphTopic, id, client.Administrate)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !ok {
		return errors.New("access denied"), http.StatusForbidden
	}
	err = this.deleteGraph(id)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteGraphSyncHandler(old models.Graph) (err error) {
	err = this.RemoveRights(this.config.GraphTopic, old.Id)
	if err != nil {
		return err
	}
	return nil
}

func (this *Controller) deleteGraph(id string) error {
	ctx, _ := getTimeoutContext()
	err := this.db.RemoveGraph(ctx, id, this.deleteGraphSyncHandler)
	if err != nil {
		return err
	}
	return nil
}
