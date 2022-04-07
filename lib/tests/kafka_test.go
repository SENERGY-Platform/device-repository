/*
 * Copyright 2022 InfAI (CC SES)
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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/source/consumer"
	"github.com/SENERGY-Platform/device-repository/lib/source/producer"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/docker"
	"github.com/ory/dockertest/v3"
	"github.com/segmentio/kafka-go"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestKafkaRetry(t *testing.T) {
	t.Skip("experiment to check kafka retry")
	conf, err := config.Load("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	conf.FatalErrHandler = t.Fatal
	conf.MongoReplSet = false
	conf.Debug = true

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Error(err)
		return
	}

	_, zkIp, err := docker.Zookeeper(pool, ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}
	zookeeperUrl := zkIp + ":2181"

	conf.KafkaUrl, err = docker.Kafka(pool, ctx, wg, zookeeperUrl)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(1 * time.Second)

	tryCount := 0
	doneCount := 0

	kwg := sync.WaitGroup{}

	_, err = consumer.NewConsumer(ctx, conf.KafkaUrl, "test", "test", func(topic string, msg []byte) error {
		log.Println("consume: ", string(msg), tryCount, tryCount%3)
		isSuccess := tryCount%3 == 2
		tryCount = tryCount + 1
		if !isSuccess {
			return errors.New("test error: " + string(msg))
		}
		doneCount = doneCount + 1
		kwg.Done()
		return nil
	}, func(err error, consumer *consumer.Consumer) {
		t.Error(err)
	})
	if err != nil {
		t.Error(err)
		return
	}

	w, err := producer.GetKafkaWriter([]string{conf.KafkaUrl}, "test", true)
	if err != nil {
		t.Error(err)
		return
	}
	for i := 0; i < 10; i++ {
		kwg.Add(1)
		err = w.WriteMessages(ctx, kafka.Message{
			Key:   []byte("test"),
			Value: []byte(strconv.Itoa(i)),
			Time:  time.Now(),
		})
		if err != nil {
			t.Error(err)
			return
		}
	}

	kwg.Wait()

	if doneCount != 10 {
		t.Error(doneCount)
	}

	if tryCount != 30 {
		t.Error(tryCount)
	}
}
