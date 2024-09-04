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

package client

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/database/testdb"
	"github.com/SENERGY-Platform/device-repository/lib/tests/semantic_legacy"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
)

func NewTestClient() (ctrl Interface, db database.Database, err error) {
	conf := config.Config{
		ServerPort:                               "8080",
		DeviceTopic:                              "devices",
		DeviceTypeTopic:                          "device-types",
		DeviceGroupTopic:                         "device-groups",
		HubTopic:                                 "hubs",
		ProtocolTopic:                            "protocols",
		ConceptTopic:                             "concepts",
		CharacteristicTopic:                      "characteristics",
		AspectTopic:                              "aspects",
		FunctionTopic:                            "functions",
		DeviceClassTopic:                         "device-classes",
		LocationTopic:                            "locations",
		Debug:                                    true,
		DisableKafkaConsumer:                     false,
		DisableHttpApi:                           false,
		HttpClientTimeout:                        "30s",
		FatalErrHandler:                          nil,
		DeviceServiceGroupSelectionAllowNotFound: true,
		LocalIdUniqueForOwner:                    true,
	}
	db = testdb.NewTestDB(conf)

	permclient, err := client.NewTestClient(context.Background())
	if err != nil {
		return nil, nil, err
	}
	ctrl, err = controller.New(conf, db, semantic_legacy.VoidProducerMock{}, permclient)
	if err != nil {
		return nil, nil, err
	}
	return ctrl, db, nil
}
