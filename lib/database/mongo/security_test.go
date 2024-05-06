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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"reflect"
	"sync"
	"testing"
)

func TestSecurity(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	port, _, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	conf.MongoUrl = "mongodb://localhost:" + port
	m, err := New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("test device rights", func(t *testing.T) {
		topic := conf.DeviceTopic
		id1 := "device-test-id-1"
		id2 := "device-test-id-2"
		id3 := "device-test-id-3"
		id4 := "device-test-id-4"
		id5 := "device-test-id-5"
		ownerUser := Userid
		ownerToken := Userjwt
		secondUser := SecendOwnerTokenUser
		secondUserToken := SecondOwnerToken
		//adminUser := testenv.AdminTokenUser
		adminToken := AdminToken
		adminGroup := "admin"

		t.Run("initial rights", func(t *testing.T) {
			err = m.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with modified id", func(t *testing.T) {
			modId := id1 + idmodifier.Seperator + idmodifier.EncodeModifierParameter(map[string][]string{"service_group_selection": {"sg1"}})
			allowed, err := m.CheckBool(ownerToken, topic, modId, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowedMap, err := m.CheckMultiple(ownerToken, topic, []string{modId}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowedMap, map[string]bool{
				modId: true,
			}) {
				t.Errorf("%#v", allowedMap)
				return
			}

		})

		t.Run("check rights with second user", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: false,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: false,
				id2: false,
				id3: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

		t.Run("update rights", func(t *testing.T) {
			err = m.SetRights(topic, id1, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: true, Write: true, Execute: true, Administrate: true},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id2, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: map[string]model.Right{},
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id3, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: nil,
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id4, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: true, Write: true, Execute: true, Administrate: true},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

		})

		t.Run("check rights with owner after update", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(ownerToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user after update", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin after update", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple after update", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: false,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

		t.Run("ensure initial rights (second init should change nothing)", func(t *testing.T) {
			err = m.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after ensure", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(ownerToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user after ensure", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin after ensure", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple after ensure", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: false,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

		t.Run("delete id2", func(t *testing.T) {
			err = m.RemoveRights(topic, id2)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after delete", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(ownerToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user after delete", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin after delete", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple after delete", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

	})

	t.Run("test hubs rights", func(t *testing.T) {
		topic := conf.HubTopic
		id1 := "hub-test-id-1"
		id2 := "hub-test-id-2"
		id3 := "hub-test-id-3"
		id4 := "hub-test-id-4"
		id5 := "hub-test-id-5"
		ownerUser := Userid
		ownerToken := Userjwt
		secondUser := SecendOwnerTokenUser
		secondUserToken := SecondOwnerToken
		//adminUser := testenv.AdminTokenUser
		adminToken := AdminToken
		adminGroup := "admin"

		t.Run("initial rights", func(t *testing.T) {
			err = m.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: false,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: false,
				id2: false,
				id3: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

		t.Run("update rights", func(t *testing.T) {
			err = m.SetRights(topic, id1, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: true, Write: true, Execute: true, Administrate: true},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id2, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: map[string]model.Right{},
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id3, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: nil,
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id4, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: true, Write: true, Execute: true, Administrate: true},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

		})

		t.Run("check rights with owner after update", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(ownerToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user after update", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin after update", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple after update", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: false,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

		t.Run("ensure initial rights (second init should change nothing)", func(t *testing.T) {
			err = m.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after ensure", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(ownerToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user after ensure", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin after ensure", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple after ensure", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: false,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

		t.Run("delete id2", func(t *testing.T) {
			err = m.RemoveRights(topic, id2)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after delete", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(ownerToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user after delete", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin after delete", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple after delete", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

	})

	t.Run("test device-groups rights", func(t *testing.T) {
		topic := conf.DeviceGroupTopic
		id1 := "group-test-id-1"
		id2 := "group-test-id-2"
		id3 := "group-test-id-3"
		id4 := "group-test-id-4"
		id5 := "group-test-id-5"
		ownerUser := Userid
		ownerToken := Userjwt
		secondUser := SecendOwnerTokenUser
		secondUserToken := SecondOwnerToken
		//adminUser := testenv.AdminTokenUser
		adminToken := AdminToken
		adminGroup := "admin"

		t.Run("initial rights", func(t *testing.T) {
			err = m.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: false,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: false,
				id2: false,
				id3: false,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: false,
				id2: false,
				id3: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

		t.Run("update rights", func(t *testing.T) {
			err = m.SetRights(topic, id1, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: true, Write: true, Execute: true, Administrate: true},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id2, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: map[string]model.Right{},
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id3, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: nil,
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id4, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: true, Execute: true, Administrate: true},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: true, Write: true, Execute: true, Administrate: true},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

		})

		t.Run("check rights with owner after update", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(ownerToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user after update", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin after update", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple after update", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: false,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

		t.Run("ensure initial rights (second init should change nothing)", func(t *testing.T) {
			err = m.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after ensure", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(ownerToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user after ensure", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin after ensure", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple after ensure", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: false,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id2: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

		t.Run("delete id2", func(t *testing.T) {
			err = m.RemoveRights(topic, id2)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after delete", func(t *testing.T) {
			allowed, err := m.CheckBool(ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(ownerToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with second user after delete", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check rights with admin after delete", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
		})

		t.Run("check multiple after delete", func(t *testing.T) {
			allowed, err := m.CheckMultiple(ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}

			allowed, err = m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(allowed, map[string]bool{
				id1: true,
				id3: true,
				id4: true,
			}) {
				t.Errorf("%#v", allowed)
				return
			}
		})

	})

	t.Run("test right flags", func(t *testing.T) {
		topic := conf.DeviceTopic
		id1 := "flags-test-id-1"
		id2 := "flags-test-id-2"
		id3 := "flags-test-id-3"
		id4 := "flags-test-id-4"
		id5 := "flags-test-id-5"
		ownerUser := Userid
		secondUser := SecendOwnerTokenUser
		secondUserToken := SecondOwnerToken
		//adminUser := testenv.AdminTokenUser
		adminToken := AdminToken
		adminGroup := "admin"

		t.Run("initial rights", func(t *testing.T) {
			err = m.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id3, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id4, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = m.EnsureInitialRights(topic, id5, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("update rights", func(t *testing.T) {
			err = m.SetRights(topic, id1, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: true, Write: false, Execute: false, Administrate: false},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: true, Write: false, Execute: false, Administrate: false},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id2, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: false, Write: true, Execute: false, Administrate: false},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: false, Write: true, Execute: false, Administrate: false},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id3, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: false, Write: false, Execute: true, Administrate: false},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: false, Write: false, Execute: true, Administrate: false},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

			err = m.SetRights(topic, id4, model.ResourceRights{
				UserRights: map[string]model.Right{
					ownerUser:  {Read: true, Write: true, Execute: true, Administrate: true},
					secondUser: {Read: false, Write: false, Execute: false, Administrate: true},
				},
				GroupRights: map[string]model.Right{
					adminGroup: {Read: false, Write: false, Execute: false, Administrate: true},
				},
			})
			if err != nil {
				t.Error(err)
				return
			}

		})

		t.Run("check user read", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowedMap, err := m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(allowedMap, map[string]bool{
				id1: true,
				id2: false,
				id3: false,
				id4: false,
			}) {
				t.Errorf("%#v", allowedMap)
			}
		})
		t.Run("check admin read", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowedMap, err := m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4}, model.READ)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(allowedMap, map[string]bool{
				id1: true,
				id2: false,
				id3: false,
				id4: false,
			}) {
				t.Errorf("%#v", allowedMap)
			}
		})

		t.Run("check user write", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowedMap, err := m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4}, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(allowedMap, map[string]bool{
				id1: false,
				id2: true,
				id3: false,
				id4: false,
			}) {
				t.Errorf("%#v", allowedMap)
			}
		})
		t.Run("check admin write", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id3, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id4, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowedMap, err := m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4}, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(allowedMap, map[string]bool{
				id1: false,
				id2: true,
				id3: false,
				id4: false,
			}) {
				t.Errorf("%#v", allowedMap)
			}
		})

		t.Run("check user execute", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowedMap, err := m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4}, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(allowedMap, map[string]bool{
				id1: false,
				id2: false,
				id3: true,
				id4: false,
			}) {
				t.Errorf("%#v", allowedMap)
			}
		})
		t.Run("check admin execute", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id3, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id4, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowedMap, err := m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4}, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(allowedMap, map[string]bool{
				id1: false,
				id2: false,
				id3: true,
				id4: false,
			}) {
				t.Errorf("%#v", allowedMap)
			}
		})

		t.Run("check user administrate", func(t *testing.T) {
			allowed, err := m.CheckBool(secondUserToken, topic, id1, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(secondUserToken, topic, id2, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id3, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowed, err = m.CheckBool(secondUserToken, topic, id4, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowedMap, err := m.CheckMultiple(secondUserToken, topic, []string{id1, id2, id3, id4}, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(allowedMap, map[string]bool{
				id1: false,
				id2: false,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowedMap)
			}
		})
		t.Run("check admin administrate", func(t *testing.T) {
			allowed, err := m.CheckBool(adminToken, topic, id1, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = m.CheckBool(adminToken, topic, id2, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id3, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowed, err = m.CheckBool(adminToken, topic, id4, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowedMap, err := m.CheckMultiple(adminToken, topic, []string{id1, id2, id3, id4}, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(allowedMap, map[string]bool{
				id1: false,
				id2: false,
				id3: false,
				id4: true,
			}) {
				t.Errorf("%#v", allowedMap)
			}
		})
	})

}

const Userid = "testOwner"
const Userjwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`

const AdminTokenUser = "admin"
const AdminToken = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJhZG1pbiIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwibm9uY2UiOiI5MmM0M2M5NS03NWIwLTQ2Y2YtODBhZS00NWRkOTczYjRiN2YiLCJhdXRoX3RpbWUiOjE1NDY1MDcwMDksInNlc3Npb25fc3RhdGUiOiI1ZGY5MjhmNC0wOGYwLTRlYjktOWI2MC0zYTBhZTIyZWZjNzMiLCJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbIioiXSwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbInVzZXIiLCJhZG1pbiJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LXJlYWxtIiwidmlldy1pZGVudGl0eS1wcm92aWRlcnMiLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidXNlciIsImFkbWluIl19.ggcFFFEsjwdfSzEFzmZt_m6W4IiSQub2FRhZVfWttDI`

const SecendOwnerTokenUser = "secondOwner"
const SecondOwnerToken = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJzZWNvbmRPd25lciIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kIiwibm9uY2UiOiI5MmM0M2M5NS03NWIwLTQ2Y2YtODBhZS00NWRkOTczYjRiN2YiLCJhdXRoX3RpbWUiOjE1NDY1MDcwMDksInNlc3Npb25fc3RhdGUiOiI1ZGY5MjhmNC0wOGYwLTRlYjktOWI2MC0zYTBhZTIyZWZjNzMiLCJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbIioiXSwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbInVzZXIiXX0sInJlc291cmNlX2FjY2VzcyI6eyJtYXN0ZXItcmVhbG0iOnsicm9sZXMiOlsidmlldy1yZWFsbSIsInZpZXctaWRlbnRpdHktcHJvdmlkZXJzIiwibWFuYWdlLWlkZW50aXR5LXByb3ZpZGVycyIsImltcGVyc29uYXRpb24iLCJjcmVhdGUtY2xpZW50IiwibWFuYWdlLXVzZXJzIiwicXVlcnktcmVhbG1zIiwidmlldy1hdXRob3JpemF0aW9uIiwicXVlcnktY2xpZW50cyIsInF1ZXJ5LXVzZXJzIiwibWFuYWdlLWV2ZW50cyIsIm1hbmFnZS1yZWFsbSIsInZpZXctZXZlbnRzIiwidmlldy11c2VycyIsInZpZXctY2xpZW50cyIsIm1hbmFnZS1hdXRob3JpemF0aW9uIiwibWFuYWdlLWNsaWVudHMiLCJxdWVyeS1ncm91cHMiXX0sImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInJvbGVzIjpbInVzZXIiXX0.cq8YeUuR0jSsXCEzp634fTzNbGkq_B8KbVrwBPgceJ4`
