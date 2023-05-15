/*
 * Copyright 2023 InfAI (CC SES)
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

package consumer

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/source/util"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestKafkaTimings(t *testing.T) {
	t.Skip("used to compare consumer configurations")
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, zkIp, err := docker.Zookeeper(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	kafkaUrl, err := docker.Kafka(ctx, wg, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("create test topics", func(t *testing.T) {
		for _, topic := range []string{"t1", "t2", "t3", "t4", "t5", "t6", "t7", "t8", "t9", "t10"} {
			err = util.InitTopic(kafkaUrl, topic)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	testFunc := func(msgCount int, msgDelay time.Duration, config kafka.ReaderConfig) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, stop := context.WithCancel(ctx)
			deltas := []time.Duration{}
			r := kafka.NewReader(config)
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer r.Close()
				defer log.Println("close consumer for topic ", config.Topic)
				for {
					select {
					case <-ctx.Done():
						return
					default:
						m, err := r.FetchMessage(ctx)
						if err == io.EOF || err == context.Canceled {
							return
						}
						if err != nil {
							log.Println("ERROR: while consuming topic ", config.Topic, err)
							t.Error(err)
							return
						}

						deltas = append(deltas, time.Now().Sub(m.Time))
						err = r.CommitMessages(ctx, m)
						if err != nil {
							log.Println("ERROR: while committing message ", config.Topic, string(m.Value), err)
							t.Error(err)
							return
						}
						if len(deltas) == msgCount {
							stop()
						}
					}
				}
			}()

			writer := &kafka.Writer{
				Addr:        kafka.TCP(kafkaUrl),
				Topic:       config.Topic,
				MaxAttempts: 10,
				BatchSize:   1,
				Balancer:    &kafka.Hash{},
			}

			for i := 0; i < msgCount; i++ {
				time.Sleep(msgDelay)
				err = writer.WriteMessages(ctx, kafka.Message{
					Key:   []byte("test"),
					Value: []byte(strconv.Itoa(i)),
					Time:  time.Now(),
				})
				if err != nil {
					t.Error(err)
					return
				}
			}

			wg.Wait()

			t.Log(deltas)

		}
	}

	t.Run("t1", testFunc(10, 1*time.Second, kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{kafkaUrl},
		GroupID:                "t1",
		Topic:                  "t1",
		MaxWait:                1 * time.Second,
		ReadBatchTimeout:       10 * time.Second,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	}))

	t.Run("t2", testFunc(10, 10*time.Second, kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{kafkaUrl},
		GroupID:                "t2",
		Topic:                  "t2",
		MaxWait:                1 * time.Second,
		ReadBatchTimeout:       10 * time.Second,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	}))

	t.Run("t3", testFunc(10, 1*time.Second, kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{kafkaUrl},
		GroupID:                "t3",
		Topic:                  "t3",
		MaxWait:                10 * time.Second,
		ReadBatchTimeout:       10 * time.Second,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	}))

	t.Run("t4", testFunc(10, 10*time.Second, kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{kafkaUrl},
		GroupID:                "t4",
		Topic:                  "t4",
		MaxWait:                10 * time.Second,
		ReadBatchTimeout:       10 * time.Second,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	}))

	t.Run("t5", testFunc(10, 1*time.Second, kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{kafkaUrl},
		GroupID:                "t5",
		Topic:                  "t5",
		MaxWait:                1 * time.Second,
		ReadBatchTimeout:       30 * time.Second,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	}))

	t.Run("t6", testFunc(10, 10*time.Second, kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{kafkaUrl},
		GroupID:                "t6",
		Topic:                  "t6",
		MaxWait:                1 * time.Second,
		ReadBatchTimeout:       30 * time.Second,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	}))

	t.Run("t7", testFunc(10, 1*time.Second, kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{kafkaUrl},
		GroupID:                "t7",
		Topic:                  "t7",
		MaxWait:                10 * time.Second,
		ReadBatchTimeout:       30 * time.Second,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	}))

	t.Run("t8", testFunc(10, 10*time.Second, kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{kafkaUrl},
		GroupID:                "t8",
		Topic:                  "t8",
		MaxWait:                10 * time.Second,
		ReadBatchTimeout:       30 * time.Second,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	}))
}
