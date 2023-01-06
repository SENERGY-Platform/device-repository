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
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/database/testdb"
	"github.com/SENERGY-Platform/device-repository/lib/tests/semantic_legacy"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/mocks"
)

func NewTestClient() (ctrl Interface, db database.Database, sec *mocks.Security, err error) {
	db = testdb.NewTestDB()
	sec = mocks.NewSecurity()
	ctrl, err = controller.New(config.Config{
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
	}, db, sec, semantic_legacy.VoidProducerMock{})
	if err != nil {
		return nil, nil, nil, err
	}
	return ctrl, db, sec, nil
}
