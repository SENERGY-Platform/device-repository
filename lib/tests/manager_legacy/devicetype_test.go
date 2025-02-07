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

package tests

import (
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/tests/manager_legacy/helper"
	"github.com/SENERGY-Platform/models/go/models"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func testDeviceType(t *testing.T, port string) {
	t.Run("create empty device-type", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true", models.DeviceType{})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		//expect validation error
		if resp.StatusCode == http.StatusOK {
			t.Fatal(resp.Status, resp.StatusCode)
		}
	})

	protocol := models.Protocol{}
	t.Run("create protocol", func(t *testing.T) {
		resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+port+"/protocols?wait=true", models.Protocol{
			Name:             "pname1",
			Handler:          "ph1",
			ProtocolSegments: []models.ProtocolSegment{{Name: "ps1"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		err = json.NewDecoder(resp.Body).Decode(&protocol)
		if err != nil {
			t.Fatal(err)
		}
	})

	dt := models.DeviceType{}
	t.Run("create device-type", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true", models.DeviceType{
			Name:          "foo",
			DeviceClassId: "dc1",
			Services: []models.Service{
				{
					Name:    "s1name",
					LocalId: "lid1",
					Inputs: []models.Content{
						{
							ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
							Serialization:     "json",
							ContentVariable: models.ContentVariable{
								Name:       "v1name",
								Type:       models.String,
								FunctionId: f1Id,
								AspectId:   a1Id,
							},
						},
					},
					ProtocolId: protocol.Id,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		err = json.NewDecoder(resp.Body).Decode(&dt)
		if err != nil {
			t.Fatal(err)
		}

		if dt.Id == "" {
			t.Fatal(dt)
		}
	})

	t.Run("read device-type", func(t *testing.T) {
		result := models.DeviceType{}
		resp, err := helper.Jwtget(userjwt, "http://localhost:"+port+"/device-types/"+url.PathEscape(dt.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Log("http://localhost:" + port + "/device-types/" + url.PathEscape(dt.Id))
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		result = models.DeviceType{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatal(err)
		}

		if result.Name != "foo" ||
			result.DeviceClassId != "dc1" ||
			len(result.Services) != 1 ||
			result.Services[0].Name != "s1name" ||
			result.Services[0].ProtocolId != protocol.Id ||
			result.Services[0].Inputs[0].ContentVariable.AspectId != a1Id ||
			result.Services[0].Inputs[0].ContentVariable.FunctionId != f1Id {
			t.Fatal(result)
		}
	})

	t.Run("delete device-type", func(t *testing.T) {
		resp, err := helper.Jwtdelete(adminjwt, "http://localhost:"+port+"/device-types/"+url.PathEscape(dt.Id)+"?wait=true")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/device-types/"+url.PathEscape(dt.Id))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Fatal(resp.Status, resp.StatusCode)
		}
	})
}

func testDeviceTypeWithDistinctAttributes(t *testing.T, port string) {
	protocol := models.Protocol{}
	t.Run("create protocol", func(t *testing.T) {
		resp, err := helper.Jwtpost(adminjwt, "http://localhost:"+port+"/protocols?wait=true", models.Protocol{
			Name:             "pname1",
			Handler:          "ph1",
			ProtocolSegments: []models.ProtocolSegment{{Name: "ps1"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		err = json.NewDecoder(resp.Body).Decode(&protocol)
		if err != nil {
			t.Fatal(err)
		}
	})

	dt1 := models.DeviceType{}
	t.Run("create first device-type", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true&distinct_attributes="+url.QueryEscape("senergy/vendor,senergy/model"), models.DeviceType{
			Name:          "foo",
			DeviceClassId: "dc1",
			Attributes: []models.Attribute{
				{Key: "senergy/vendor", Value: "Philips"},
				{Key: "senergy/model", Value: "model1"},
			},
			Services: []models.Service{
				{
					Name:    "s1name",
					LocalId: "lid1",
					Inputs: []models.Content{
						{
							ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
							Serialization:     "json",
							ContentVariable: models.ContentVariable{
								Name:       "v1name",
								Type:       models.String,
								FunctionId: f1Id,
								AspectId:   a1Id,
							},
						},
					},
					ProtocolId: protocol.Id,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		err = json.NewDecoder(resp.Body).Decode(&dt1)
		if err != nil {
			t.Fatal(err)
		}

		if dt1.Id == "" {
			t.Fatal(dt1)
		}
	})

	t.Run("try to create conflicting device-type", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true&distinct_attributes="+url.QueryEscape("senergy/vendor,senergy/model"), models.DeviceType{
			Name:          "foo2",
			DeviceClassId: "dc1",
			Attributes: []models.Attribute{
				{Key: "senergy/vendor", Value: "Philips"},
				{Key: "senergy/model", Value: "model1"},
			},
			Services: []models.Service{
				{
					Name:    "s1name",
					LocalId: "lid1",
					Inputs: []models.Content{
						{
							ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
							Serialization:     "json",
							ContentVariable: models.ContentVariable{
								Name:       "v1name",
								Type:       models.String,
								FunctionId: f1Id,
								AspectId:   a1Id,
							},
						},
					},
					ProtocolId: protocol.Id,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}
	})

	dt2 := models.DeviceType{}
	t.Run("create second device-type", func(t *testing.T) {
		resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true&distinct_attributes="+url.QueryEscape("senergy/vendor,senergy/model"), models.DeviceType{
			Name:          "foo",
			DeviceClassId: "dc1",
			Attributes: []models.Attribute{
				{Key: "senergy/vendor", Value: "Philips"},
				{Key: "senergy/model", Value: "model2"},
			},
			Services: []models.Service{
				{
					Name:    "s1name",
					LocalId: "lid1",
					Inputs: []models.Content{
						{
							ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
							Serialization:     "json",
							ContentVariable: models.ContentVariable{
								Name:       "v1name",
								Type:       models.String,
								FunctionId: f1Id,
								AspectId:   a1Id,
							},
						},
					},
					ProtocolId: protocol.Id,
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}

		err = json.NewDecoder(resp.Body).Decode(&dt2)
		if err != nil {
			t.Fatal(err)
		}

		if dt2.Id == "" {
			t.Fatal(dt1)
		}
	})

	t.Run("delete device-type 1", func(t *testing.T) {
		resp, err := helper.Jwtdelete(adminjwt, "http://localhost:"+port+"/device-types/"+url.PathEscape(dt1.Id)+"?wait=true")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}
	})

	t.Run("delete device-type 2", func(t *testing.T) {
		resp, err := helper.Jwtdelete(adminjwt, "http://localhost:"+port+"/device-types/"+url.PathEscape(dt2.Id)+"?wait=true")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatal(resp.Status, resp.StatusCode, string(b))
		}
	})
}

func testDeviceTypeWithServiceGroups(t *testing.T, port string) {
	resp, err := helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true", models.DeviceType{})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	//expect validation error
	if resp.StatusCode == http.StatusOK {
		t.Fatal(resp.Status, resp.StatusCode)
	}

	resp, err = helper.Jwtpost(adminjwt, "http://localhost:"+port+"/protocols?wait=true", models.Protocol{
		Name:             "pname2",
		Handler:          "ph2",
		ProtocolSegments: []models.ProtocolSegment{{Name: "ps2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	protocol := models.Protocol{}
	err = json.NewDecoder(resp.Body).Decode(&protocol)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = helper.Jwtpost(userjwt, "http://localhost:"+port+"/device-types?wait=true", models.DeviceType{
		Name:          "foo",
		DeviceClassId: "dc1",
		ServiceGroups: []models.ServiceGroup{
			{
				Key:         "sg1",
				Name:        "service group 1",
				Description: "foo  bar",
			},
		},
		Services: []models.Service{
			{
				Name:    "s1name",
				LocalId: "lid1",
				Inputs: []models.Content{
					{
						ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
						Serialization:     "json",
						ContentVariable: models.ContentVariable{
							Name:       "v1name",
							Type:       models.String,
							FunctionId: f1Id,
							AspectId:   a1Id,
						},
					},
				},
				ProtocolId: protocol.Id,
			},
			{
				Name:    "s2name",
				LocalId: "lid2",
				Inputs: []models.Content{
					{
						ProtocolSegmentId: protocol.ProtocolSegments[0].Id,
						Serialization:     "json",
						ContentVariable: models.ContentVariable{
							Name:       "v1name",
							Type:       models.String,
							FunctionId: f1Id,
							AspectId:   a1Id,
						},
					},
				},
				ProtocolId:      protocol.Id,
				ServiceGroupKey: "sg1",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	dt := models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&dt)
	if err != nil {
		t.Fatal(err)
	}

	if dt.Id == "" {
		t.Fatal(dt)
	}

	result := models.DeviceType{}
	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/device-types/"+url.PathEscape(dt.Id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Log("http://localhost:" + port + "/device-types/" + url.PathEscape(dt.Id))
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	result = models.DeviceType{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Name != "foo" ||
		result.DeviceClassId != "dc1" ||
		len(result.Services) != 2 {
		t.Fatal(result.Name, result.DeviceClassId, len(result.Services))
	}

	if !reflect.DeepEqual(result.ServiceGroups, []models.ServiceGroup{
		{
			Key:         "sg1",
			Name:        "service group 1",
			Description: "foo  bar",
		},
	}) {
		t.Fatal(result.ServiceGroups)
	}

	if result.Services[0].Name != "s1name" ||
		result.Services[0].LocalId != "lid1" ||
		result.Services[0].ServiceGroupKey != "" ||
		result.Services[0].ProtocolId != protocol.Id ||
		result.Services[0].Inputs[0].ContentVariable.AspectId != a1Id ||
		result.Services[0].Inputs[0].ContentVariable.FunctionId != f1Id {

		t.Fatal(result.Services[0])
	}

	if result.Services[1].Name != "s2name" ||
		result.Services[1].LocalId != "lid2" ||
		result.Services[1].ServiceGroupKey != "sg1" ||
		result.Services[1].ProtocolId != protocol.Id ||
		result.Services[1].Inputs[0].ContentVariable.AspectId != a1Id ||
		result.Services[1].Inputs[0].ContentVariable.FunctionId != f1Id {
		temp, _ := json.Marshal(result.Services[1])
		t.Fatal(string(temp))
	}

	resp, err = helper.Jwtdelete(adminjwt, "http://localhost:"+port+"/device-types/"+url.PathEscape(dt.Id)+"?wait=true")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(resp.Status, resp.StatusCode, string(b))
	}

	resp, err = helper.Jwtget(userjwt, "http://localhost:"+port+"/device-types/"+url.PathEscape(dt.Id))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Fatal(resp.Status, resp.StatusCode)
	}
}
