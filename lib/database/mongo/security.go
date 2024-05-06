/*
 * Copyright 2024 InfAI (CC SES)
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

package mongo

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"slices"
)

func init() {
	CreateCollections = append(CreateCollections, func(db *Mongo) error {
		var err error
		collection := db.client.Database(db.config.MongoTable).Collection(db.config.MongoRightsCollection)
		err = db.ensureCompoundIndex(collection, "rightsbykindandid", true, true, "kind", "id")
		if err != nil {
			return err
		}
		return nil
	})
}

type RightsEntry struct {
	Kind          Kind     `json:"kind" bson:"kind"`
	Id            string   `json:"id" bson:"id"`
	AdminUsers    []string `json:"admin_users" bson:"admin_users"`
	AdminGroups   []string `json:"admin_groups" bson:"admin_groups"`
	ReadUsers     []string `json:"read_users" bson:"read_users"`
	ReadGroups    []string `json:"read_groups" bson:"read_groups"`
	WriteUsers    []string `json:"write_users" bson:"write_users"`
	WriteGroups   []string `json:"write_groups" bson:"write_groups"`
	ExecuteUsers  []string `json:"execute_users" bson:"execute_users"`
	ExecuteGroups []string `json:"execute_groups" bson:"execute_groups"`
}

type Kind string

func (this *Mongo) rightsCollection() *mongo.Collection {
	return this.client.Database(this.config.MongoTable).Collection(this.config.MongoRightsCollection)
}

func (this *Mongo) EnsureInitialRights(topic string, resourceId string, owner string) error {
	kind, err := this.getInternalKind(topic)
	if err != nil {
		return err
	}
	ctx, _ := getTimeoutContext()
	exists, err := this.rightsElementExists(ctx, kind, resourceId)
	if err != nil {
		return err
	}
	if !exists {
		element := this.getDefaultEntryPermissions(kind, owner)
		element.Id = resourceId
		element.Kind = kind
		_, err = this.rightsCollection().InsertOne(ctx, element)
		return err
	}
	return nil
}

func (this *Mongo) rightsElementExists(ctx context.Context, kind Kind, resourceId string) (exists bool, err error) {
	err = this.rightsCollection().FindOne(ctx, bson.M{"kind": kind, "id": resourceId}).Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (this *Mongo) SetRights(topic string, resourceId string, rights model.ResourceRights) (err error) {
	element := RightsEntry{
		Id:            resourceId,
		AdminUsers:    []string{},
		AdminGroups:   []string{},
		ReadUsers:     []string{},
		ReadGroups:    []string{},
		WriteUsers:    []string{},
		WriteGroups:   []string{},
		ExecuteUsers:  []string{},
		ExecuteGroups: []string{},
	}
	element.Kind, err = this.getInternalKind(topic)
	if err != nil {
		return err
	}
	element.setResourceRights(rights)

	ctx, _ := getTimeoutContext()

	_, err = this.rightsCollection().ReplaceOne(ctx, bson.M{"kind": element.Kind, "id": element.Id}, element, options.Replace().SetUpsert(true))

	return err
}

func (this *Mongo) RemoveRights(topic string, id string) error {
	kind, err := this.getInternalKind(topic)
	if err != nil {
		return err
	}
	ctx, _ := getTimeoutContext()
	_, err = this.rightsCollection().DeleteOne(ctx, bson.M{"kind": kind, "id": id})
	return err
}

func (this *Mongo) CheckBool(token string, topic string, id string, action model.AuthAction) (allowed bool, err error) {
	pureid, _ := idmodifier.SplitModifier(id)
	kind, err := this.getInternalKind(topic)
	if err != nil {
		return false, err
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return false, err
	}
	ctx, _ := getTimeoutContext()
	result := this.rightsCollection().FindOne(ctx, bson.M{"kind": kind, "id": pureid})
	err = result.Err()
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var element RightsEntry
	err = result.Decode(&element)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return checkRights(jwtToken, element, action), err
}

func (this *Mongo) CheckMultiple(token string, topic string, ids []string, action model.AuthAction) (result map[string]bool, err error) {
	pureIds := []string{}
	rawIdsIndex := map[string]string{}
	for _, id := range ids {
		pureId, _ := idmodifier.SplitModifier(id)
		pureIds = append(pureIds, pureId)
		rawIdsIndex[pureId] = id
	}
	kind, err := this.getInternalKind(topic)
	if err != nil {
		return result, err
	}
	jwtToken, err := jwt.Parse(token)
	if err != nil {
		return result, err
	}
	ctx, _ := getTimeoutContext()
	cursor, err := this.rightsCollection().Find(ctx, bson.M{"kind": kind, "id": bson.M{"$in": pureIds}})
	if err != nil {
		return result, err
	}
	result = map[string]bool{}
	for cursor.Next(context.Background()) {
		element := RightsEntry{}
		err = cursor.Decode(&element)
		if err != nil {
			return nil, err
		}
		result[rawIdsIndex[element.Id]] = checkRights(jwtToken, element, action)
	}

	err = cursor.Err()
	return result, err
}

func checkRights(token jwt.Token, element RightsEntry, right model.AuthAction) bool {
	user := token.GetUserId()
	groups := token.GetRoles()
	switch right {
	case model.ADMINISTRATE:
		if !slices.Contains(element.AdminUsers, user) && !containsAny(element.AdminGroups, groups) {
			return false
		}
	case model.READ:
		if !slices.Contains(element.ReadUsers, user) && !containsAny(element.ReadGroups, groups) {
			return false
		}
	case model.WRITE:
		if !slices.Contains(element.WriteUsers, user) && !containsAny(element.WriteGroups, groups) {
			return false
		}
	case model.EXECUTE:
		if !slices.Contains(element.ExecuteUsers, user) && !containsAny(element.ExecuteGroups, groups) {
			return false
		}
	}
	return true
}

func containsAny(list []string, any []string) bool {
	for _, e := range any {
		if slices.Contains(list, e) {
			return true
		}
	}
	return false
}

func (this *RightsEntry) setResourceRights(rights model.ResourceRights) {
	for group, right := range rights.GroupRights {
		if right.Administrate {
			this.AdminGroups = append(this.AdminGroups, group)
		}
		if right.Execute {
			this.ExecuteGroups = append(this.ExecuteGroups, group)
		}
		if right.Write {
			this.WriteGroups = append(this.WriteGroups, group)
		}
		if right.Read {
			this.ReadGroups = append(this.ReadGroups, group)
		}
	}
	for user, right := range rights.UserRights {
		if right.Administrate {
			this.AdminUsers = append(this.AdminUsers, user)
		}
		if right.Execute {
			this.ExecuteUsers = append(this.ExecuteUsers, user)
		}
		if right.Write {
			this.WriteUsers = append(this.WriteUsers, user)
		}
		if right.Read {
			this.ReadUsers = append(this.ReadUsers, user)
		}
	}
}

func (this *Mongo) getDefaultEntryPermissions(kind Kind, owner string) (entry RightsEntry) {
	entry = RightsEntry{
		AdminUsers:    []string{},
		AdminGroups:   []string{},
		ReadUsers:     []string{},
		ReadGroups:    []string{},
		WriteUsers:    []string{},
		WriteGroups:   []string{},
		ExecuteUsers:  []string{},
		ExecuteGroups: []string{},
	}
	if owner != "" {
		entry.AdminUsers = []string{owner}
		entry.ReadUsers = []string{owner}
		entry.WriteUsers = []string{owner}
		entry.ExecuteUsers = []string{owner}
	}
	for group, rights := range this.config.InitialGroupRights[string(kind)] {
		for _, right := range rights {
			switch right {
			case 'a':
				entry.AdminGroups = append(entry.AdminGroups, group)
			case 'r':
				entry.ReadGroups = append(entry.ReadGroups, group)
			case 'w':
				entry.WriteGroups = append(entry.WriteGroups, group)
			case 'x':
				entry.ExecuteGroups = append(entry.AdminGroups, group)
			}
		}
	}
	return
}

func (this *Mongo) getInternalKind(topic string) (Kind, error) {
	switch topic {
	case this.config.DeviceTopic:
		return "devices", nil
	case this.config.DeviceGroupTopic:
		return "device-groups", nil
	case this.config.HubTopic:
		return "hubs", nil
	}
	return "", errors.New("unknown topic to rights entry kind mapping: " + topic)
}
