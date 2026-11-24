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
	"strconv"
	"strings"
	"time"

	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

// batchDeleteTimeout replaces the 10 s of getTimeoutContext for the Delete<Resources> methods:
// a batch of several thousand ids does not fit into the timeout of a single delete.
const batchDeleteTimeout = 60 * time.Second

// batchDeleteDeniedErrorLimit caps how many ids the access denied error names.
const batchDeleteDeniedErrorLimit = 20

// getBatchTimeoutContext is getTimeoutContext with batchDeleteTimeout.
func getBatchTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), batchDeleteTimeout)
}

// requireAdminForBatchDelete is the admin check of the single delete methods of the admin-only resources.
func (this *Controller) requireAdminForBatchDelete(token string) (error, int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return errors.New("token is not an admin"), http.StatusUnauthorized
	}
	return nil, http.StatusOK
}

// validateEachForBatchDelete runs the delete precondition of the single delete method for every id.
// The first violated precondition is returned with its id; nothing has been deleted at that point.
func validateEachForBatchDelete(ctx context.Context, ids []string, validate func(ctx context.Context, id string) (error, int)) (error, int) {
	for _, id := range ids {
		err, code := validate(ctx, id)
		if err != nil {
			return fmt.Errorf("unable to delete %v: %w", id, err), code
		}
	}
	return nil, http.StatusOK
}

// checkAdministrateForBatchDelete checks the Administrate permission of the token for all ids with one request.
// ListComputedPermissionsContext sends the ids in the request body; CheckMultiplePermissionsContext would put them into the url.
// Ids missing in the answer are treated as denied.
// Admins pass without a request: the single delete lets an admin delete a resource that has no rights entry yet,
// while ListComputedPermissionsContext answers such ids with all permissions false, even for admins.
func (this *Controller) checkAdministrateForBatchDelete(ctx context.Context, token string, topic string, ids []string) (error, int) {
	if len(ids) == 0 {
		return nil, http.StatusOK
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusBadRequest
	}
	if jwtToken.IsAdmin() {
		return nil, http.StatusOK
	}
	permissions, err, code := this.permissionsV2Client.ListComputedPermissionsContext(ctx, token, topic, ids)
	if err != nil {
		if code < 400 {
			code = http.StatusInternalServerError
		}
		return err, code
	}
	allowed := map[string]bool{}
	for _, permission := range permissions {
		if permission.Administrate {
			allowed[permission.Id] = true
		}
	}
	denied := []string{}
	for _, id := range ids {
		if !allowed[id] {
			denied = append(denied, id)
		}
	}
	if len(denied) == 0 {
		return nil, http.StatusOK
	}
	shown := denied
	if len(shown) > batchDeleteDeniedErrorLimit {
		shown = append(append([]string{}, denied[:batchDeleteDeniedErrorLimit]...), "and "+strconv.Itoa(len(denied)-batchDeleteDeniedErrorLimit)+" more")
	}
	return errors.New("access denied: " + strings.Join(shown, ", ")), http.StatusForbidden
}

// existingIds returns the ids of the listed elements that were requested, without duplicates.
// The filter is applied here as well because not every database implementation applies the id filter of its list method.
func existingIds[T any](elements []T, requested []string, id func(T) string) []string {
	want := make(map[string]bool, len(requested))
	for _, r := range requested {
		want[r] = true
	}
	result := []string{}
	for _, element := range elements {
		elementId := id(element)
		if want[elementId] {
			result = append(result, elementId)
			delete(want, elementId)
		}
	}
	return result
}

// RemoveRightsOfMany removes the rights of all ids of one topic with a single request.
func (this *Controller) RemoveRightsOfMany(ctx context.Context, topic string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	err, _ := this.permissionsV2Client.RemoveResourcesContext(ctx, client.InternalAdminToken, topic, ids)
	return err
}
