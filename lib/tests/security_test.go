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

package tests

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/api"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/lib/idmodifier"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/source/consumer"
	"github.com/SENERGY-Platform/device-repository/lib/source/producer"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	permclient "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/jwt"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestSecurity(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.Load("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	conf.Debug = true

	_, mongoIp, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	conf.MongoUrl = "mongodb://" + mongoIp + ":27017"

	_, zkIp, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	conf.KafkaUrl, err = docker.Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}

	_, permV2Ip, err := docker.PermissionsV2(ctx, wg, conf.MongoUrl, conf.KafkaUrl)
	if err != nil {
		t.Error(err)
		return
	}
	conf.PermissionsV2Url = "http://" + permV2Ip + ":8080"

	m, err := mongo.New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	p, err := producer.New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	ctrl, err := controller.New(conf, m, p, nil)
	if err != nil {
		t.Error(err)
		return
	}

	err = consumer.Start(ctx, conf, ctrl)
	if err != nil {
		t.Error(err)
		return
	}

	whPort, err := docker.GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	conf.ServerPort = strconv.Itoa(whPort)
	err = api.Start(ctx, conf, ctrl)
	if err != nil {
		t.Error(err)
		return
	}

	deviceRepoClient := client.NewClient("http://localhost:" + conf.ServerPort)

	setRights := func(resourceKind string, resourceId string, rights model.ResourceRights) error {
		_, err, _ := deviceRepoClient.GetPermissionsClient().SetPermission(permclient.InternalAdminToken, resourceKind, resourceId, rights.ToPermV2Permissions())
		if err != nil {
			return err
		}
		return nil
	}

	checkBool := func(t *testing.T, token string, topic string, id string, action model.AuthAction) (allowed bool, err error) {
		//what is in the db?
		allowed, err = m.CheckBool(token, topic, id, action)
		if err != nil {
			t.Error(err)
			return false, err
		}
		var perm permclient.Permission
		switch action {
		case model.READ:
			perm = permclient.Read
		case model.WRITE:
			perm = permclient.Write
		case model.EXECUTE:
			perm = permclient.Execute
		case model.ADMINISTRATE:
			perm = permclient.Administrate
		}

		//what is in permissions-v2?
		expected := allowed

		jwtToken, err := jwt.Parse(token)
		if err != nil {
			t.Error(err)
			return false, err
		}
		if jwtToken.IsAdmin() {
			expected = true //admins may do everything in perm-v2
		}

		allowed2, err, _ := ctrl.GetPermissionsClient().CheckPermission(token, topic, id, perm)
		if err != nil {
			t.Error(err)
			return false, err
		}

		if expected != allowed2 {
			t.Error(topic, id, "expected != allowed2", expected, allowed2, jwtToken.IsAdmin())
		}

		//does api permissions embedding work?
		allowed3, err, _ := deviceRepoClient.GetPermissionsClient().CheckPermission(token, topic, id, perm)
		if err != nil {
			t.Error(err)
			return false, err
		}
		if expected != allowed3 {
			t.Error(topic, id, "expected != allowed3", expected, allowed3, jwtToken.IsAdmin())
		}
		return allowed, nil
	}

	var checkMultiple = func(t *testing.T, token string, topic string, ids []string, action model.AuthAction) (result map[string]bool, err error) {
		result, err = m.CheckMultiple(token, topic, ids, action)
		if err != nil {
			t.Error(err)
			return result, err
		}

		expected := map[string]bool{}
		jwtToken, err := jwt.Parse(token)
		if err != nil {
			t.Error(err)
			return result, err
		}
		for key, value := range result {
			if jwtToken.IsAdmin() {
				expected[key] = true //admins may do everything in perm-v2
			} else {
				expected[key] = value
			}
		}

		var perm permclient.Permission
		switch action {
		case model.READ:
			perm = permclient.Read
		case model.WRITE:
			perm = permclient.Write
		case model.EXECUTE:
			perm = permclient.Execute
		case model.ADMINISTRATE:
			perm = permclient.Administrate
		}
		result2, err, _ := ctrl.GetPermissionsClient().CheckMultiplePermissions(token, topic, ids, perm)
		if err != nil {
			t.Error(err)
			return result, err
		}
		if !reflect.DeepEqual(expected, result2) {
			t.Error(topic, ids, "expected != result2", expected, result2, jwtToken.IsAdmin())
		}
		result3, err, _ := deviceRepoClient.GetPermissionsClient().CheckMultiplePermissions(token, topic, ids, perm)
		if err != nil {
			t.Error(err)
			return result, err
		}
		if !reflect.DeepEqual(expected, result3) {
			t.Error(topic, ids, "expected != result3", expected, result3, jwtToken.IsAdmin())
		}
		return result, nil
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
			err = ctrl.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
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
			allowed, err := checkBool(t, ownerToken, topic, modId, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowedMap, err := checkMultiple(t, ownerToken, topic, []string{modId}, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4}, model.READ)
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
			err = setRights(topic, id1, model.ResourceRights{
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

			err = setRights(topic, id2, model.ResourceRights{
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

			err = setRights(topic, id3, model.ResourceRights{
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

			err = setRights(topic, id4, model.ResourceRights{
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

			time.Sleep(10 * time.Second)
		})

		t.Run("check rights with owner after update", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, ownerToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple after update", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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
			err = ctrl.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after ensure", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, ownerToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple after ensure", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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
			err = ctrl.DeleteDevice(id2)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after delete", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, ownerToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple after delete", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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
			err = ctrl.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4}, model.READ)
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
			err = setRights(topic, id1, model.ResourceRights{
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

			err = setRights(topic, id2, model.ResourceRights{
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

			err = setRights(topic, id3, model.ResourceRights{
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

			err = setRights(topic, id4, model.ResourceRights{
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

			time.Sleep(10 * time.Second)
		})

		t.Run("check rights with owner after update", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, ownerToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple after update", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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
			err = ctrl.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after ensure", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, ownerToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple after ensure", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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
			err = ctrl.DeleteHub(id2)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after delete", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, ownerToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple after delete", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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
			err = ctrl.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4}, model.READ)
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
			err = setRights(topic, id1, model.ResourceRights{
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

			err = setRights(topic, id2, model.ResourceRights{
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

			err = setRights(topic, id3, model.ResourceRights{
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

			err = setRights(topic, id4, model.ResourceRights{
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

			time.Sleep(10 * time.Second)
		})

		t.Run("check rights with owner after update", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, ownerToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple after update", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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
			err = ctrl.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id3, secondUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after ensure", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, ownerToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple after ensure", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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
			err = ctrl.DeleteDeviceGroup(id2)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("check rights with owner after delete", func(t *testing.T) {
			allowed, err := checkBool(t, ownerToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, ownerToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, ownerToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id5, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id5, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
		})

		t.Run("check multiple after delete", func(t *testing.T) {
			allowed, err := checkMultiple(t, ownerToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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

			allowed, err = checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4, id5}, model.READ)
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
			err = ctrl.EnsureInitialRights(topic, id1, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id2, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id3, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id4, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
			err = ctrl.EnsureInitialRights(topic, id5, ownerUser)
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("update rights", func(t *testing.T) {
			err = setRights(topic, id1, model.ResourceRights{
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

			err = setRights(topic, id2, model.ResourceRights{
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

			err = setRights(topic, id3, model.ResourceRights{
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

			err = setRights(topic, id4, model.ResourceRights{
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

			time.Sleep(10 * time.Second)
		})

		t.Run("check user read", func(t *testing.T) {
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowedMap, err := checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4}, model.READ)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id3, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.READ)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowedMap, err := checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4}, model.READ)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id3, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id4, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowedMap, err := checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4}, model.WRITE)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id3, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id4, model.WRITE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowedMap, err := checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4}, model.WRITE)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id3, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id4, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowedMap, err := checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4}, model.EXECUTE)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id3, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id4, model.EXECUTE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowedMap, err := checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4}, model.EXECUTE)
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
			allowed, err := checkBool(t, secondUserToken, topic, id1, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}
			allowed, err = checkBool(t, secondUserToken, topic, id2, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id3, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if allowed {
				t.Error("expected not allowed")
				return
			}

			allowed, err = checkBool(t, secondUserToken, topic, id4, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowedMap, err := checkMultiple(t, secondUserToken, topic, []string{id1, id2, id3, id4}, model.ADMINISTRATE)
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
			allowed, err := checkBool(t, adminToken, topic, id1, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}
			allowed, err = checkBool(t, adminToken, topic, id2, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id3, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowed, err = checkBool(t, adminToken, topic, id4, model.ADMINISTRATE)
			if err != nil {
				t.Error(err)
				return
			}
			if !allowed {
				t.Error("expected allowed")
				return
			}

			allowedMap, err := checkMultiple(t, adminToken, topic, []string{id1, id2, id3, id4}, model.ADMINISTRATE)
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
