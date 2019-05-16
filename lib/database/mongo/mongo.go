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
	"github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type Mongo struct {
	config config.Config
	client *mongo.Client
}

func New(conf config.Config) (*Mongo, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MongoUrl))
	if err != nil {
		return nil, err
	}
	return &Mongo{config: conf, client: client}, nil
}

func (this *Mongo) CreateId() string {
	return uuid.NewV4().String()
}

func (this *Mongo) Transaction(ctx context.Context) (resultCtx context.Context, close func(success bool) error, err error) {
	session, err := this.client.StartSession()
	if err != nil {
		return nil, nil, err
	}
	err = session.StartTransaction()
	if err != nil {
		return nil, nil, err
	}

	//create session context; callback is executed synchronously and the error is passed on as error of WithSession
	_ = mongo.WithSession(ctx, session, func(sessionContext mongo.SessionContext) error {
		resultCtx = sessionContext
		return nil
	})

	return resultCtx, func(success bool) error {
		defer session.EndSession(context.Background())
		var err error
		if success {
			err = session.CommitTransaction(resultCtx)
		} else {
			err = session.AbortTransaction(resultCtx)
		}
		if err != nil {
			log.Println("ERROR: unable to finish mongo transaction", err)
		}
		return err
	}, nil
}
