/*
 * Copyright 2021 InfAI (CC SES)
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

package docker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/controller/publisher"
	"github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// kafkaStarterScript is the file the container waits for before it starts the broker.
// The advertised address is the container ip, which is only known after the container exists.
// Tests reach the container by ip (like mongo and permissions-v2), so no host port is published.
const kafkaStarterScript = "/tmp/start_kafka.sh"

// kafkaStarterScriptContent sets the advertised listener and runs the normal start of the apache/kafka image.
// The controller listener is not advertised, in kraft mode it is only used inside the container.
const kafkaStarterScriptContent = `#!/bin/bash
export KAFKA_ADVERTISED_LISTENERS=BROKER://%s:9092
exec /etc/kafka/docker/run
`

// resetKafkaTransport drops the connections and the cached metadata of the shared kafka-go transport.
//
// The repository and the test readers produce through kafka.DefaultTransport, which pools connections
// and caches the cluster metadata per broker address. A container ip is reused by the next kafka container.
// The pool then still holds the connection to the removed broker, and the produce of the repository
// (RequiredAcks defaults to RequireNone, so there is no response) is written into the dead connection
// without any error. Observed as a graph command that is never delivered, only from the second test of a process on.
func resetKafkaTransport() {
	if transport, ok := kafka.DefaultTransport.(*kafka.Transport); ok {
		transport.CloseIdleConnections()
	}
}

func Kafka(ctx context.Context, wg *sync.WaitGroup) (kafkaUrl string, err error) {
	resetKafkaTransport()
	c, err := startContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "apache/kafka:3.9.1",
			Tmpfs: map[string]string{},
			// waits for the starter script, which is copied by the PostStarts hook below
			Entrypoint: []string{"sh"},
			Cmd:        []string{"-c", "while [ ! -f " + kafkaStarterScript + " ]; do sleep 0.1; done; exec bash " + kafkaStarterScript},
			// WaitingFor is not used: the default readiness hook runs before the hooks below,
			// the broker only starts after the script is copied and the wait would never end.
			LifecycleHooks: []testcontainers.ContainerLifecycleHooks{{
				PostStarts: []testcontainers.ContainerHook{
					func(ctx context.Context, c testcontainers.Container) error {
						ip, err := c.ContainerIP(ctx)
						if err != nil {
							return fmt.Errorf("kafka container ip: %w", err)
						}
						err = c.CopyToContainer(ctx, []byte(fmt.Sprintf(kafkaStarterScriptContent, ip)), kafkaStarterScript, 0o755)
						if err != nil {
							return fmt.Errorf("copy kafka starter script: %w", err)
						}
						return wait.ForLog("INFO Awaiting socket connections on").WaitUntilReady(ctx, c)
					},
				},
			}},
			AlwaysPullImage: false,
			Env: map[string]string{
				"KAFKA_HEAP_OPTS":                                "-Xmx256m -Xms256m",
				"KAFKA_NODE_ID":                                  "1",
				"CLUSTER_ID":                                     "IrzuggcFT-mWom7mj7PgtA",
				"KAFKA_PROCESS_ROLES":                            "controller,broker",
				"KAFKA_LISTENERS":                                "BROKER://:9092,CONTROLLER://:9093",
				"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
				"KAFKA_INTER_BROKER_LISTENER_NAME":               "BROKER",
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "CONTROLLER:PLAINTEXT,BROKER:PLAINTEXT",
				"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@localhost:9093",
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
				"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
				"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
				"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
				"KAFKA_NUM_PARTITIONS":                           "1",
			},
		},
		Started: true,
	})
	if c != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
			log.Println("DEBUG: remove container kafka", c.Terminate(context.Background()))
			resetKafkaTransport()
		}()
	}
	if err != nil {
		return kafkaUrl, err
	}

	ip, err := c.ContainerIP(ctx)
	if err != nil {
		return kafkaUrl, err
	}
	kafkaUrl = ip + ":9092"
	log.Println("kafkaUrl url: ", kafkaUrl)

	deadline := time.Now().Add(1 * time.Minute)
	for {
		err = tryKafkaConn(kafkaUrl)
		if err == nil {
			return kafkaUrl, nil
		}
		if time.Now().After(deadline) {
			return kafkaUrl, err
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func tryKafkaConn(kafkaUrl string) error {
	log.Println("try kafka connection to " + kafkaUrl + "...")
	conn, err := kafka.Dial("tcp", kafkaUrl)
	if err != nil {
		log.Println(err)
		return err
	}
	defer conn.Close()
	brokers, err := conn.Brokers()
	if err != nil {
		log.Println(err)
		return err
	}
	if len(brokers) == 0 {
		err = errors.New("missing brokers")
		log.Println(err)
		return err
	}
	p, err := conn.ReadPartitions("__consumer_offsets")
	if err != nil {
		log.Println(err)
		return err
	}
	if len(p) == 0 {
		err = errors.New("__consumer_offsets not ready")
		return err
	}
	log.Println("kafka connection ok")
	return nil
}

// InitTopics creates the topics and returns when every topic has a partition leader in the metadata of the broker.
// CreateTopics returns as soon as the controller has accepted the topics, the broker serves them in its metadata a moment later.
// A produce in between fails with "Unknown Topic Or Partition", and the sync handlers of the repository
// only retry after sync_interval.
func InitTopics(kafkaUrl string, topics ...string) error {
	err := publisher.InitTopic(kafkaUrl, topics...)
	if err != nil {
		return err
	}
	return waitForTopicLeaders(kafkaUrl, time.Minute, topics...)
}

func waitForTopicLeaders(kafkaUrl string, timeout time.Duration, topics ...string) error {
	deadline := time.Now().Add(timeout)
	for _, topic := range topics {
		for {
			err := topicHasLeader(kafkaUrl, topic)
			if err == nil {
				break
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("topic %v has no leader after %s: %w", topic, timeout, err)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	return nil
}

// topicHasLeader asks the leader of every partition for its last offset.
// The metadata alone is not enough: the broker lists the leader a moment before its replica manager
// has the partition, and a produce in that window fails with "Unknown Topic Or Partition" (observed with apache/kafka 3.9.1).
// The offset request is answered by the replica manager and does not write to the topic.
func topicHasLeader(kafkaUrl string, topic string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := kafka.Dial("tcp", kafkaUrl)
	if err != nil {
		return err
	}
	defer conn.Close()
	partitions, err := conn.ReadPartitions(topic)
	if err != nil {
		return err
	}
	if len(partitions) == 0 {
		return errors.New("no partitions")
	}
	for _, partition := range partitions {
		leader, err := kafka.DialLeader(ctx, "tcp", kafkaUrl, topic, partition.ID)
		if err != nil {
			return err
		}
		_, err = leader.ReadLastOffset()
		leader.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
