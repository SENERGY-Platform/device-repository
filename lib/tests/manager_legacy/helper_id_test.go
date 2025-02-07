/*
 * Copyright 2024 InfAI (CC SES)
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
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/tests/manager_legacy/helper"
	"net/url"
	"testing"
)

func testHelperId(t *testing.T, port string) {
	shortId := "RJvtgd9yR1Sput8ocwRnNA"
	expectedUUID := "foobar:449bed81-df72-4754-a9ba-df2873046734"
	resp, err := helper.Jwtget(adminjwt, "http://localhost:"+port+"/helper/id?prefix="+url.QueryEscape("foobar:")+"&short_id="+url.QueryEscape(shortId))
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()
	result := ""
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
		return
	}
	if expectedUUID != result {
		t.Error(result)
		return
	}
}
