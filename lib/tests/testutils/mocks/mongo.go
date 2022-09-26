/*
 * Copyright (c) 2022 InfAI (CC SES)
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

package mocks

import (
	"context"
	"errors"
	"github.com/tryvium-travels/memongo"
	"log"
	"sync"
)

func Mongo(ctx context.Context, wg *sync.WaitGroup) (mongoUrl string, err error) {
	mongoServer, err := memongo.StartWithOptions(&memongo.Options{MongoVersion: "4.2.1", ShouldUseReplica: true})
	if err != nil {
		return "", err
	}
	if mongoServer == nil {
		return "", errors.New("memongo.StartWithOptions() == nil")
	}
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("RECOVER:", r)
			}
			wg.Done()
		}()
		<-ctx.Done()
		mongoServer.Stop()
	}()

	return mongoServer.URI(), nil
}
