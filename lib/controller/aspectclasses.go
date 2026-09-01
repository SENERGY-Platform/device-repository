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
	"strconv"
	"strings"

	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/models/go/models"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
)

func (this *Controller) ListAspectClasses(listOptions model.AspectClassListOptions) (result []models.AspectClass, total int64, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, total, err = this.db.ListAspectClasses(ctx, listOptions)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	return result, total, nil, http.StatusOK
}

func (this *Controller) GetAspectClass(id string) (result models.AspectClass, err error, errCode int) {
	ctx, _ := getTimeoutContext()
	result, exists, err := this.db.GetAspectClass(ctx, id)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !exists {
		return result, errors.New("not found"), http.StatusNotFound
	}
	return result, nil, http.StatusOK
}

func (this *Controller) SetAspectClass(token string, aspectClass models.AspectClass) (result models.AspectClass, err error, code int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return result, errors.New("token is not an admin"), http.StatusUnauthorized
	}

	//ensure ids
	aspectClass.GenerateId()
	if !this.config.DisableStrictValidationForTesting {
		err, code = this.ValidateAspectClass(aspectClass)
		if err != nil {
			return result, err, code
		}
	}
	err = this.setAspectClass(aspectClass)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return aspectClass, nil, http.StatusOK
}

func (this *Controller) setAspectClass(aspectClass models.AspectClass) (err error) {
	ctx, _ := getTimeoutContext()
	return this.db.SetAspectClass(ctx, aspectClass, this.setAspectClassSyncHandler)
}

func (this *Controller) setAspectClassSyncHandler(aspectClass models.AspectClass) error {
	return this.publisher.PublishAspectClass(aspectClass)
}

func (this *Controller) DeleteAspectClass(token string, id string) (error, int) {
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if !jwtToken.IsAdmin() {
		return errors.New("token is not an admin"), http.StatusUnauthorized
	}
	err, code := this.ValidateAspectClassDelete(id)
	if err != nil {
		return err, code
	}
	ctx, _ := getTimeoutContext()
	err = this.db.RemoveAspectClass(ctx, id, this.deleteAspectClassSyncHandler)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func (this *Controller) deleteAspectClassSyncHandler(aspectClass models.AspectClass) error {
	return this.publisher.PublishAspectClassDelete(aspectClass.Id)
}

func (this *Controller) ValidateAspectClass(aspectClass models.AspectClass) (err error, code int) {
	if aspectClass.Id == "" {
		return errors.New("missing aspect class id"), http.StatusBadRequest
	}
	if !strings.HasPrefix(aspectClass.Id, model.URN_PREFIX) {
		return errors.New("invalid aspect class id"), http.StatusBadRequest
	}
	if aspectClass.Name == "" {
		return errors.New("missing aspect class name"), http.StatusBadRequest
	}
	return nil, http.StatusOK
}

// aspectClassUsageErrorLimit caps how many aspects the delete error names. The count
// itself comes from the listing's total, so it stays exact.
const aspectClassUsageErrorLimit = 20

func (this *Controller) ValidateAspectClassDelete(id string) (err error, code int) {
	ctx, _ := getTimeoutContext()
	//the aspect nodes are the flat index: an aspect tree is one nested document, while
	//every aspect of it has exactly one node carrying the class
	nodes, total, err := this.db.ListAspectNodes(ctx, model.AspectListOptions{
		AspectClassIds: []string{id},
		Limit:          aspectClassUsageErrorLimit,
	})
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if total == 0 {
		return nil, http.StatusOK
	}
	where := []string{}
	for _, node := range nodes {
		where = append(where, node.Name+" ("+node.Id+")")
	}
	if total > int64(len(nodes)) {
		where = append(where, "and "+strconv.FormatInt(total-int64(len(nodes)), 10)+" more")
	}
	return errors.New("still in use by " + strconv.FormatInt(total, 10) + " aspect(s): " + strings.Join(where, ", ")), http.StatusBadRequest
}

// ResolveAspectClassIds gives every aspect of the hierarchy the aspect-class of its
// root. A sub-aspect may repeat that value but may not carry a different one — an
// aspect hierarchy has at most one aspect-class. Exported because the mgw mirror
// writes aspects to the database without passing a controller write.
func ResolveAspectClassIds(aspect *models.Aspect) (err error, code int) {
	for i := range aspect.SubAspects {
		err, code = inheritAspectClassId(&aspect.SubAspects[i], aspect.AspectClassId, aspect.Id)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}

func inheritAspectClassId(aspect *models.Aspect, rootClassId string, rootId string) (err error, code int) {
	if aspect.AspectClassId == "" {
		aspect.AspectClassId = rootClassId
	} else if aspect.AspectClassId != rootClassId {
		return errors.New("sub aspect " + aspect.Id + " sets aspect class " + aspect.AspectClassId +
			", but its hierarchy is " + describeAspectClassOfRoot(rootClassId, rootId) +
			" — an aspect hierarchy has at most one aspect class, and only its root assigns it"), http.StatusBadRequest
	}
	for i := range aspect.SubAspects {
		err, code = inheritAspectClassId(&aspect.SubAspects[i], rootClassId, rootId)
		if err != nil {
			return err, code
		}
	}
	return nil, http.StatusOK
}

func describeAspectClassOfRoot(rootClassId string, rootId string) string {
	if rootClassId == "" {
		return "rooted in " + rootId + ", which has none"
	}
	return "rooted in " + rootId + ", which assigns " + rootClassId
}
