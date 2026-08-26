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

package tests

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller/publisher"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	permclient "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/segmentio/kafka-go"
)

func TestGraphKafkaPublish(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const userid = "testOwner"
	const userjwt = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIwOGM0N2E4OC0yYzc5LTQyMGYtODEwNC02NWJkOWViYmU0MWUiLCJleHAiOjE1NDY1MDcyMzMsIm5iZiI6MCwiaWF0IjoxNTQ2NTA3MTczLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjgwMDEvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJ0ZXN0T3duZXIiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiOTJjNDNjOTUtNzViMC00NmNmLTgwYWUtNDVkZDk3M2I0YjdmIiwiYXV0aF90aW1lIjoxNTQ2NTA3MDA5LCJzZXNzaW9uX3N0YXRlIjoiNWRmOTI4ZjQtMDhmMC00ZWI5LTliNjAtM2EwYWUyMmVmYzczIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJ1c2VyIl19LCJyZXNvdXJjZV9hY2Nlc3MiOnsibWFzdGVyLXJlYWxtIjp7InJvbGVzIjpbInZpZXctcmVhbG0iLCJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsIm1hbmFnZS1pZGVudGl0eS1wcm92aWRlcnMiLCJpbXBlcnNvbmF0aW9uIiwiY3JlYXRlLWNsaWVudCIsIm1hbmFnZS11c2VycyIsInF1ZXJ5LXJlYWxtcyIsInZpZXctYXV0aG9yaXphdGlvbiIsInF1ZXJ5LWNsaWVudHMiLCJxdWVyeS11c2VycyIsIm1hbmFnZS1ldmVudHMiLCJtYW5hZ2UtcmVhbG0iLCJ2aWV3LWV2ZW50cyIsInZpZXctdXNlcnMiLCJ2aWV3LWNsaWVudHMiLCJtYW5hZ2UtYXV0aG9yaXphdGlvbiIsIm1hbmFnZS1jbGllbnRzIiwicXVlcnktZ3JvdXBzIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJyb2xlcyI6WyJ1c2VyIl19.ykpuOmlpzj75ecSI6cHbCATIeY4qpyut2hMc1a67Ycg`

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.SyncLockDuration = time.Second.String()
	config.Debug = true

	config, err = docker.NewEnv(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	//consume from the start, before the repository produces anything
	graphCommands, rightsCommands := listenToGraphTopic(t, ctx, wg, config)

	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	c := client.NewClient("http://localhost:"+config.ServerPort, nil)

	t.Run("topic is registered with kafka publishing", func(t *testing.T) {
		pc := permclient.New(config.PermissionsV2Url)
		topic, err, _ := pc.GetTopic(permclient.InternalAdminToken, config.GraphTopic)
		if err != nil {
			t.Error(err)
			return
		}
		if topic.PublishToKafkaTopic != config.GraphTopic {
			t.Error("unexpected publish_to_kafka_topic: ", topic.PublishToKafkaTopic)
		}
	})

	graph := models.Graph{
		Attributes: []models.Attribute{{Key: "name", Value: "g1"}},
		Nodes:      []models.Node{{Id: "1"}, {Id: "2"}},
		Edges:      []models.Edge{{Id: "1->2", FromNodeId: "1", ToNodeId: "2", Weight: 100}},
	}

	t.Run("create graph", func(t *testing.T) {
		graph, err, _ = c.SetGraph(userjwt, graph)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("create is published", func(t *testing.T) {
		cmd, err := graphCommands.wait(graph.Id, "PUT")
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(cmd.Graph, graph) {
			t.Errorf("unexpected graph in command:\n%#v\n%#v", cmd.Graph, graph)
		}
	})

	t.Run("initial rights are published", func(t *testing.T) {
		cmd, err := rightsCommands.wait(graph.Id, "RIGHTS")
		if err != nil {
			t.Error(err)
			return
		}
		if !cmd.Rights.UserRights[userid].Administrate {
			t.Error("owner is missing admin rights: ", cmd.Rights)
		}
	})

	t.Run("update is published", func(t *testing.T) {
		graph.Attributes = []models.Attribute{{Key: "name", Value: "g1-renamed"}}
		graph, err, _ = c.SetGraph(userjwt, graph)
		if err != nil {
			t.Error(err)
			return
		}
		cmd, err := graphCommands.wait(graph.Id, "PUT")
		if err != nil {
			t.Error(err)
			return
		}
		if !reflect.DeepEqual(cmd.Graph, graph) {
			t.Errorf("unexpected graph in command:\n%#v\n%#v", cmd.Graph, graph)
		}
	})

	t.Run("delete is published", func(t *testing.T) {
		err, _ := c.DeleteGraph(userjwt, graph.Id)
		if err != nil {
			t.Error(err)
			return
		}
		cmd, err := graphCommands.wait(graph.Id, "DELETE")
		if err != nil {
			t.Error(err)
			return
		}
		if cmd.Id != graph.Id {
			t.Error("unexpected id: ", cmd.Id)
		}
	})
}

// rightsCommand mirrors the permissions-v2 kafka command; it is not exported by the permissions-v2 client
type rightsCommand struct {
	Command string `json:"command"`
	Id      string `json:"id"`
	Rights  struct {
		UserRights map[string]struct {
			Read         bool `json:"read"`
			Write        bool `json:"write"`
			Execute      bool `json:"execute"`
			Administrate bool `json:"administrate"`
		} `json:"user_rights"`
	} `json:"rights"`
}

type commandBuffer[T any] struct {
	mux      sync.Mutex
	commands []T
	match    func(cmd T, id string, command string) bool
}

func (this *commandBuffer[T]) add(cmd T) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.commands = append(this.commands, cmd)
}

// wait returns the first buffered command for the given id and command name, or fails after a timeout
func (this *commandBuffer[T]) wait(id string, command string) (result T, err error) {
	timeout := time.After(30 * time.Second)
	for {
		this.mux.Lock()
		for i, cmd := range this.commands {
			if this.match(cmd, id, command) {
				this.commands = this.commands[i+1:]
				this.mux.Unlock()
				return cmd, nil
			}
		}
		this.mux.Unlock()
		select {
		case <-timeout:
			return result, errors.New("timeout while waiting for " + command + " on " + id)
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func listenToGraphTopic(t *testing.T, ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (graphs *commandBuffer[publisher.GraphCommand], rights *commandBuffer[rightsCommand]) {
	graphs = &commandBuffer[publisher.GraphCommand]{match: func(cmd publisher.GraphCommand, id string, command string) bool {
		return cmd.Id == id && cmd.Command == command
	}}
	rights = &commandBuffer[rightsCommand]{match: func(cmd rightsCommand, id string, command string) bool {
		return cmd.Id == id && cmd.Command == command
	}}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{config.KafkaUrl},
		Topic:       config.GraphTopic,
		Partition:   0,
		StartOffset: kafka.FirstOffset,
		MaxWait:     100 * time.Millisecond,
		Logger:      slog.NewLogLogger(config.GetLogger().Handler(), slog.LevelDebug),
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer reader.Close()
		for {
			msg, err := reader.ReadMessage(ctx)
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				t.Log("error while reading graph topic:", err)
				return
			}
			var command struct {
				Command string `json:"command"`
			}
			if err = json.Unmarshal(msg.Value, &command); err != nil {
				t.Log("unable to unmarshal graph topic message:", err)
				continue
			}
			if command.Command == "RIGHTS" {
				cmd := rightsCommand{}
				if err = json.Unmarshal(msg.Value, &cmd); err == nil {
					rights.add(cmd)
				}
				continue
			}
			cmd := publisher.GraphCommand{}
			if err = json.Unmarshal(msg.Value, &cmd); err == nil {
				graphs.add(cmd)
			}
		}
	}()

	return graphs, rights
}
