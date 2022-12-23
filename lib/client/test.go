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
	"flag"
	"github.com/SENERGY-Platform/device-repository/lib/api"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/database/testdb"
	"github.com/SENERGY-Platform/device-repository/lib/tests/semantic_legacy"
	"github.com/SENERGY-Platform/device-repository/lib/tests/testutils/mocks"
)

func NewTestClient() (ctrl api.Controller, db database.Database, err error) {
	configLocation := flag.String("config", "../../config.json", "configuration file")
	flag.Parse()

	conf, err := config.Load(*configLocation)
	if err != nil {
		return nil, nil, err
	}
	db = testdb.NewTestDB()
	ctrl, err = controller.New(conf, db, mocks.NewSecurity(), semantic_legacy.VoidProducerMock{})
	if err != nil {
		return nil, nil, err
	}
	return ctrl, db, nil
}
