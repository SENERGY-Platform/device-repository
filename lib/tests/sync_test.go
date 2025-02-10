/*
 * Copyright 2025 InfAI (CC SES)
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
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"sync"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.SyncLockDuration = time.Second.String()

	_, mip, err := docker.MongoDB(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	config.MongoUrl = "mongodb://" + mip + ":27017"

	db, err := mongo.New(config)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("create", func(t *testing.T) {
		t.Run("successful", func(t *testing.T) {
			err = db.SetProtocol(ctx, models.Protocol{Id: "p1"}, func(protocol models.Protocol) error {
				return nil
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("failed", func(t *testing.T) {
			err = db.SetProtocol(ctx, models.Protocol{Id: "p2"}, func(protocol models.Protocol) error {
				return errors.New("error")
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("early retry", func(t *testing.T) {
			err = db.RetryProtocolSync(time.Second, func(protocol models.Protocol) error {
				err := errors.New("unexpected delete retry")
				t.Error(err)
				return err
			}, func(protocol models.Protocol) error {
				err := errors.New("protocol sync early retry")
				t.Error(err)
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		time.Sleep(2 * time.Second)

		t.Run("retry", func(t *testing.T) {
			inFErr := errors.New("missing retry")
			err = db.RetryProtocolSync(time.Second, func(protocol models.Protocol) error {
				err := errors.New("unexpected delete retry")
				t.Error(err)
				return err
			}, func(protocol models.Protocol) error {
				inFErr = nil
				if protocol.Id != "p2" {
					inFErr = errors.New("unexpected protocol id in retry")
					t.Error(inFErr)
				}
				return nil
			})
			if err != nil {
				t.Error(err)
				return
			}
			if inFErr != nil {
				t.Error(inFErr)
				return
			}
		})

		time.Sleep(2 * time.Second)

		t.Run("no new retries needed", func(t *testing.T) {
			err = db.RetryProtocolSync(time.Second, func(protocol models.Protocol) error {
				err := errors.New("unexpected delete retry")
				t.Error(err)
				return err
			}, func(protocol models.Protocol) error {
				err := errors.New("protocol sync early retry")
				t.Error(err)
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
		})
	})

	t.Run("update", func(t *testing.T) {
		t.Run("successful", func(t *testing.T) {
			err = db.SetProtocol(ctx, models.Protocol{Id: "p2", Name: "p2"}, func(protocol models.Protocol) error {
				return nil
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("failed", func(t *testing.T) {
			err = db.SetProtocol(ctx, models.Protocol{Id: "p1", Name: "p1"}, func(protocol models.Protocol) error {
				return errors.New("error")
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("early retry", func(t *testing.T) {
			err = db.RetryProtocolSync(time.Second, func(protocol models.Protocol) error {
				err := errors.New("unexpected delete retry")
				t.Error(err)
				return err
			}, func(protocol models.Protocol) error {
				err := errors.New("protocol sync early retry")
				t.Error(err)
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		time.Sleep(2 * time.Second)

		t.Run("retry", func(t *testing.T) {
			inFErr := errors.New("missing retry")
			err = db.RetryProtocolSync(time.Second, func(protocol models.Protocol) error {
				err := errors.New("unexpected delete retry")
				t.Error(err)
				return err
			}, func(protocol models.Protocol) error {
				inFErr = nil
				if protocol.Id != "p1" {
					inFErr = errors.New("unexpected protocol id in retry")
					t.Error(inFErr)
				}
				return nil
			})
			if err != nil {
				t.Error(err)
				return
			}
			if inFErr != nil {
				t.Error(inFErr)
				return
			}
		})

		time.Sleep(2 * time.Second)

		t.Run("no new retries needed", func(t *testing.T) {
			err = db.RetryProtocolSync(time.Second, func(protocol models.Protocol) error {
				err := errors.New("unexpected delete retry")
				t.Error(err)
				return err
			}, func(protocol models.Protocol) error {
				err := errors.New("protocol sync early retry")
				t.Error(err)
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
		})
	})

	t.Run("delete", func(t *testing.T) {
		t.Run("list before delete", func(t *testing.T) {
			list, err := db.ListProtocols(ctx, 10, 0, "name.asc")
			if err != nil {
				t.Error(err)
				return
			}
			if len(list) != 2 {
				t.Error(list)
				return
			}
		})
		t.Run("successful", func(t *testing.T) {
			err = db.RemoveProtocol(ctx, "p1", func(protocol models.Protocol) error {
				return nil
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("list after successful delete", func(t *testing.T) {
			list, err := db.ListProtocols(ctx, 10, 0, "name.asc")
			if err != nil {
				t.Error(err)
				return
			}
			if len(list) != 1 {
				t.Error(list)
				return
			}
		})

		t.Run("failed", func(t *testing.T) {
			err = db.RemoveProtocol(ctx, "p2", func(protocol models.Protocol) error {
				return errors.New("error")
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("list after failed delete", func(t *testing.T) {
			list, err := db.ListProtocols(ctx, 10, 0, "name.asc")
			if err != nil {
				t.Error(err)
				return
			}
			if len(list) != 0 {
				t.Error(list)
				return
			}
		})

		t.Run("early retry", func(t *testing.T) {
			err = db.RetryProtocolSync(time.Second, func(protocol models.Protocol) error {
				err := errors.New("protocol sync early retry")
				t.Error(err)
				return err
			}, func(protocol models.Protocol) error {
				err := errors.New("unexpected update retry")
				t.Error(err)
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		time.Sleep(2 * time.Second)

		t.Run("retry", func(t *testing.T) {
			inFErr := errors.New("missing retry")
			err = db.RetryProtocolSync(time.Second, func(protocol models.Protocol) error {
				inFErr = nil
				if protocol.Id != "p2" {
					inFErr = errors.New("unexpected protocol id in retry")
					t.Error(inFErr)
				}
				return nil
			}, func(protocol models.Protocol) error {
				err := errors.New("unexpected update retry")
				t.Error(err)
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
			if inFErr != nil {
				t.Error(inFErr)
				return
			}
		})

		time.Sleep(2 * time.Second)

		t.Run("no new retries needed", func(t *testing.T) {
			err = db.RetryProtocolSync(time.Second, func(protocol models.Protocol) error {
				err := errors.New("protocol sync early retry")
				t.Error(err)
				return err
			}, func(protocol models.Protocol) error {
				err := errors.New("unexpected update retry")
				t.Error(err)
				return err
			})
			if err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("list after retries", func(t *testing.T) {
			list, err := db.ListProtocols(ctx, 10, 0, "name.asc")
			if err != nil {
				t.Error(err)
				return
			}
			if len(list) != 0 {
				t.Error(list)
				return
			}
		})
	})

}
