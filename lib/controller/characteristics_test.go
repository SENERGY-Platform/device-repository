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

package controller

import (
	"encoding/json"
	"github.com/SENERGY-Platform/models/go/models"
	"testing"
)

func TestValidateCharacteristicIdReuse(t *testing.T) {
	existing := []models.Characteristic{
		{
			Id: "a",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "a.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "a.1.1"},
						{Id: "a.1.2"},
					},
				},
				{
					Id: "a.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "a.2.1"},
						{Id: "a.2.2"},
					},
				},
			},
		},
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.1.1"},
						{Id: "b.1.2"},
					},
				},
				{
					Id: "b.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.2.1"},
						{Id: "b.2.2"},
					},
				},
			},
		},
	}

	allowedUpdates := []models.Characteristic{
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.1.1"},
						{Id: "b.1.2"},
					},
				},
				{
					Id: "b.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.2.1"},
						{Id: "b.2.2"},
					},
				},
			},
		},
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.1.1"},
						{Id: "b.1.2"},
					},
				},
				{
					Id: "b.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.2.1"},
						{Id: "b.2.2"},
					},
				},
				{
					Id: "b.3",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.3.1"},
						{Id: "b.3.2"},
					},
				},
			},
		},
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.1.1"},
						{Id: "b.1.2"},
					},
				},
				{
					Id: "b.3",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.3.1"},
						{Id: "b.3.2"},
					},
				},
			},
		},
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.3",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.3.1"},
						{Id: "b.3.2"},
					},
				},
			},
		},
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.1.1"},
						{Id: "b.1.2"},
					},
				},
			},
		},
		{
			Id: "b",
		},
	}

	allowedCreates := []models.Characteristic{
		{
			Id: "c",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "c.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "c.1.1"},
						{Id: "c.1.2"},
					},
				},
				{
					Id: "c.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "c.2.1"},
						{Id: "c.2.2"},
					},
				},
			},
		},
		{Id: "c"},
	}

	forbiddenCreates := []models.Characteristic{
		{
			Id: "c",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "c.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "c.1.1"},
						{Id: "b.1.2"}, //ref to b
					},
				},
				{
					Id: "c.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "c.2.1"},
						{Id: "c.2.2"},
					},
				},
			},
		},
		{
			Id: "c",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1", //ref to b
					SubCharacteristics: []models.Characteristic{
						{Id: "c.1.1"},
						{Id: "c.1.2"},
					},
				},
				{
					Id: "c.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "c.2.1"},
						{Id: "c.2.2"},
					},
				},
			},
		},
		{
			Id: "c",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1", //ref to b
					SubCharacteristics: []models.Characteristic{
						{Id: "b.1.1"}, //ref to b
						{Id: "c.1.2"},
					},
				},
			},
		},
		{
			Id: "c",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1", //ref to b
				},
			},
		},
	}

	forbiddenUpdates := []models.Characteristic{
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "a.1.1"},
						{Id: "b.1.2"},
					},
				},
				{
					Id: "b.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.2.1"},
						{Id: "b.2.2"},
					},
				},
			},
		},
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "b.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.1.1"},
						{Id: "b.1.2"},
					},
				},
				{
					Id: "b.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.2.1"},
						{Id: "b.2.2"},
					},
				},
				{
					Id: "a.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.3.1"},
						{Id: "b.3.2"},
					},
				},
			},
		},
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "a.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "a.1.1"},
						{Id: "b.1.2"},
					},
				},
				{
					Id: "b.3",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.3.1"},
						{Id: "b.3.2"},
					},
				},
			},
		},
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "a.2",
					SubCharacteristics: []models.Characteristic{
						{Id: "b.3.1"},
						{Id: "b.3.2"},
					},
				},
			},
		},
		{
			Id: "b",
			SubCharacteristics: []models.Characteristic{
				{
					Id: "a.1",
					SubCharacteristics: []models.Characteristic{
						{Id: "a.1.1"},
						{Id: "b.1.2"},
					},
				},
			},
		},
	}

	t.Run("check allowedUpdates", func(t *testing.T) {
		for _, allowedUpdate := range allowedUpdates {
			err, _ := validateCharacteristicIdReuse(allowedUpdate, existing)
			if err != nil {
				t.Errorf("err=%v\n%#v", err, allowedUpdate)
			}
		}
	})

	t.Run("check allowedCreates", func(t *testing.T) {
		for _, allowedCreate := range allowedCreates {
			err, _ := validateCharacteristicIdReuse(allowedCreate, existing)
			if err != nil {
				t.Errorf("err=%v\n%#v", err, allowedCreate)
			}
		}
	})

	t.Run("check forbiddenCreates", func(t *testing.T) {
		for _, forbiddenCreate := range forbiddenCreates {
			err, _ := validateCharacteristicIdReuse(forbiddenCreate, existing)
			if err == nil {
				t.Errorf("%#v", forbiddenCreate)
			}
		}
	})

	t.Run("check forbiddenUpdates", func(t *testing.T) {
		for _, forbiddenUpdate := range forbiddenUpdates {
			err, _ := validateCharacteristicIdReuse(forbiddenUpdate, existing)
			if err == nil {
				temp, _ := json.Marshal(forbiddenUpdate)
				t.Errorf("%v\n%#v", string(temp), forbiddenUpdate)
			}
		}
	})

}
