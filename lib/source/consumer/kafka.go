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

package consumer

import (
	"context"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/source/util"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

func NewConsumerWithMultipleTopics(ctx context.Context, broker string, groupId string, topics []string, listener func(topic string, delivery []byte) error, errhandler func(topice string, err error)) error {
	if len(topics) == 0 {
		return nil
	}
	log.Println("consume:", topics)

	for _, topic := range topics {
		err := util.InitTopic(broker, topic)
		if err != nil {
			log.Println("ERROR: unable to create topic", err)
			return err
		}
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{broker},
		GroupID:                groupId,
		GroupTopics:            topics,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	})

	go func() {
		defer r.Close()
		defer log.Println("close consumer for topics ", topics)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m, err := r.FetchMessage(ctx)
				if err == io.EOF || err == context.Canceled {
					return
				}
				topic := m.Topic
				if err != nil {
					log.Println("ERROR: while consuming topic ", topic, err)
					errhandler(topic, err)
					return
				}

				err = retry(func() error {
					return listener(topic, m.Value)
				}, func(n int64) time.Duration {
					return time.Duration(n) * time.Second
				}, 10*time.Minute)

				if err != nil {
					log.Println("ERROR: unable to handle message (no commit)", err)
					errhandler(topic, err)
				} else {
					err = r.CommitMessages(ctx, m)
				}
			}
		}
	}()
	return nil
}

func NewConsumer(ctx context.Context, broker string, groupid string, topic string, listener func(topic string, msg []byte) error, errorhandler func(err error, consumer *Consumer)) (consumer *Consumer, err error) {
	consumer = &Consumer{ctx: ctx, groupId: groupid, broker: broker, topic: topic, listener: listener, errorhandler: errorhandler}
	err = consumer.start()
	return
}

type Consumer struct {
	count        int
	broker       string
	groupId      string
	topic        string
	ctx          context.Context
	listener     func(topic string, msg []byte) error
	errorhandler func(err error, consumer *Consumer)
	mux          sync.Mutex
}

func (this *Consumer) start() error {
	log.Println("DEBUG: consume topic: \"" + this.topic + "\"")

	err := util.InitTopic(this.broker, this.topic)
	if err != nil {
		log.Println("ERROR: unable to create topic", err)
		return err
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		CommitInterval:         0, //synchronous commits
		Brokers:                []string{this.broker},
		GroupID:                this.groupId,
		Topic:                  this.topic,
		Logger:                 log.New(io.Discard, "", 0),
		ErrorLogger:            log.New(os.Stdout, "[KAFKA-ERROR] ", log.Default().Flags()),
		WatchPartitionChanges:  true,
		PartitionWatchInterval: time.Minute,
	})
	go func() {
		defer r.Close()
		defer log.Println("close consumer for topic ", this.topic)
		for {
			select {
			case <-this.ctx.Done():
				return
			default:
				m, err := r.FetchMessage(this.ctx)
				if err == io.EOF || err == context.Canceled {
					return
				}
				if err != nil {
					log.Println("ERROR: while consuming topic ", this.topic, err)
					this.errorhandler(err, this)
					return
				}

				err = retry(func() error {
					return this.listener(m.Topic, m.Value)
				}, func(n int64) time.Duration {
					return time.Duration(n) * time.Second
				}, 10*time.Minute)

				if err != nil {
					log.Println("ERROR: unable to handle message (no commit)", err)
					this.errorhandler(err, this)
				} else {
					err = r.CommitMessages(this.ctx, m)
				}
			}
		}
	}()
	return err
}

func retry(f func() error, waitProvider func(n int64) time.Duration, timeout time.Duration) (err error) {
	err = errors.New("initial")
	start := time.Now()
	for i := int64(1); err != nil && time.Since(start) < timeout; i++ {
		err = f()
		if err != nil {
			log.Println("ERROR: kafka listener error:", err)
			wait := waitProvider(i)
			if time.Since(start)+wait < timeout {
				log.Println("ERROR: retry after:", wait.String())
				time.Sleep(wait)
			} else {
				return err
			}
		}
	}
	return err
}
