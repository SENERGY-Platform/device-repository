/*
 * Copyright 2026 InfAI (CC SES)
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

package mgwmirror

import (
	"testing"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
)

func TestSourceIsNewer(t *testing.T) {
	const user = "ca4d1149-e3ed-4e0b-9e49-3bda908de436"
	ts := func(collection string, userId string, unix int64) model.LastUpdateTimestamp {
		return model.LastUpdateTimestamp{Collection: collection, UserId: userId, UnixTimestamp: unix}
	}
	cases := []struct {
		name     string
		local    []model.LastUpdateTimestamp
		source   []model.LastUpdateTimestamp
		expected bool
	}{
		{
			//SNRGY-4995: the first source entry of the collection is the older one without user
			name:     "newer user entry after older entry without user in source",
			local:    []model.LastUpdateTimestamp{ts("devicegroup", user, 200)},
			source:   []model.LastUpdateTimestamp{ts("devicegroup", "", 100), ts("devicegroup", user, 300)},
			expected: true,
		},
		{
			name:     "newer local user entry after older local entry without user",
			local:    []model.LastUpdateTimestamp{ts("devicegroup", "", 100), ts("devicegroup", user, 300)},
			source:   []model.LastUpdateTimestamp{ts("devicegroup", user, 200)},
			expected: false,
		},
		{
			name:     "source entry without user newer than local",
			local:    []model.LastUpdateTimestamp{ts("devicegroup", user, 200)},
			source:   []model.LastUpdateTimestamp{ts("devicegroup", user, 100), ts("devicegroup", "", 300)},
			expected: true,
		},
		{
			name:     "equal",
			local:    []model.LastUpdateTimestamp{ts("devicegroup", user, 300)},
			source:   []model.LastUpdateTimestamp{ts("devicegroup", "", 100), ts("devicegroup", user, 300)},
			expected: false,
		},
		{
			name:     "other collections are ignored",
			local:    []model.LastUpdateTimestamp{ts("device", user, 100), ts("devicegroup", user, 200)},
			source:   []model.LastUpdateTimestamp{ts("device", user, 300), ts("devicegroup", user, 100)},
			expected: false,
		},
		{
			name:     "no source entry",
			local:    []model.LastUpdateTimestamp{},
			source:   []model.LastUpdateTimestamp{ts("device", user, 300)},
			expected: false,
		},
		{
			name:     "no local entry",
			local:    []model.LastUpdateTimestamp{ts("device", user, 300)},
			source:   []model.LastUpdateTimestamp{ts("devicegroup", user, 100)},
			expected: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if actual := sourceIsNewer(configuration.Config{}, c.local, c.source, "devicegroup"); actual != c.expected {
				t.Errorf("expected %v, got %v", c.expected, actual)
			}
		})
	}
}
