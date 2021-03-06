/*
 * Copyright 2019 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/ory/dockertest"
	"testing"
	"time"
)

func TestMongoProtocol(t *testing.T) {

	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Error("Could not connect to docker: ", err)
		return
	}
	closer, port, _, err := MongoTestServer(pool)
	if err != nil {
		t.Error(err)
		return
	}
	if true {
		defer closer()
	}
	conf.MongoUrl = "mongodb://localhost:" + port
	m, err := New(conf)
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	_, exists, err := m.GetProtocol(ctx, "does_not_exist")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("protocol type should not exist")
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetProtocol(ctx, model.Protocol{
		Id:   "foobar1",
		Name: "foo1",
		ProtocolSegments: []model.ProtocolSegment{
			{
				Id:   "segment1",
				Name: "s1name",
			},
			{
				Id:   "segment2",
				Name: "s2name",
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetProtocol(ctx, model.Protocol{
		Id:   "foobar2",
		Name: "foo2",
		ProtocolSegments: []model.ProtocolSegment{
			{
				Id:   "segment3",
				Name: "s3name",
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	protocol, exists, err := m.GetProtocol(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("protocol should exist")
		return
	}
	if protocol.Id != "foobar1" || protocol.Name != "foo1" {
		t.Error("unexpected result", protocol)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.SetProtocol(ctx, model.Protocol{
		Id:   "foobar1",
		Name: "foo1changed",
		ProtocolSegments: []model.ProtocolSegment{
			{
				Id:   "segment1",
				Name: "s1name",
			},
			{
				Id:   "segment2",
				Name: "s2name",
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	protocol, exists, err = m.GetProtocol(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}
	if !exists {
		t.Error("protocol should exist")
		return
	}
	if protocol.Id != "foobar1" || protocol.Name != "foo1changed" {
		t.Error("unexpected result", protocol)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	result, err := m.ListProtocols(ctx, 100, 0, "name.asc")
	if err != nil {
		t.Error(err)
		return
	}
	if len(result) != 2 {
		t.Error("unexpected result", result)
		return
	}
	if (result[0].Id != "foobar1" && result[1].Id != "foobar1") || (result[0].Id != "foobar2" && result[1].Id != "foobar2") {
		t.Error("unexpected result", result)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = m.RemoveProtocol(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	dt, exists, err := m.GetProtocol(ctx, "foobar1")
	if err != nil {
		t.Error(err)
		return
	}
	if exists {
		t.Error("dt should not exist", dt)
		return
	}
}
