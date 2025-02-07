/*
 * Copyright 2025 InfAI (CC SES)
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

package repo_legacy

import (
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func testAspectList(t *testing.T, conf config.Config) {
	aspects := []models.Aspect{
		{
			Id:   "a1",
			Name: "a1",
		},
		{
			Id:   "a2",
			Name: "a2",
			SubAspects: []models.Aspect{
				{
					Id:   "a2.1",
					Name: "a2.1",
				},
				{
					Id:   "a2.2",
					Name: "a2.2",
					SubAspects: []models.Aspect{
						{
							Id:   "a2.2.1",
							Name: "a2.2.1",
						},
						{
							Id:   "a2.2.2",
							Name: "a2.2.2",
						},
					},
				},
			},
		},
		{
			Id:   "b3",
			Name: "b3",
		},
		{
			Id:   "c4",
			Name: "c4",
		},
	}
	aspectNodes := []models.AspectNode{}
	var createAspectNodes func(aspect models.Aspect, rootId string, parentId string, ancestors []string) (descendents []string)
	createAspectNodes = func(aspect models.Aspect, rootId string, parentId string, ancestors []string) (descendents []string) {
		descendents = []string{}
		children := []string{}
		for _, sub := range aspect.SubAspects {
			children = append(children, sub.Id)
			temp := createAspectNodes(sub, rootId, aspect.Id, append(ancestors, aspect.Id))
			descendents = append(descendents, temp...)
		}
		aspectNodes = append(aspectNodes, models.AspectNode{
			Id:            aspect.Id,
			Name:          aspect.Name,
			RootId:        rootId,
			ParentId:      parentId,
			ChildIds:      children,
			AncestorIds:   ancestors,
			DescendentIds: descendents,
		})
		return append(descendents, aspect.Id)
	}

	for _, aspect := range aspects {
		createAspectNodes(aspect, aspect.Id, "", []string{})
	}
	slices.SortFunc(aspectNodes, func(a models.AspectNode, b models.AspectNode) int {
		return strings.Compare(a.Id, b.Id)
	})

	t.Run("create aspects", func(t *testing.T) {
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		for _, aspect := range aspects {
			_, err, _ := c.SetAspect(AdminToken, aspect)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
	t.Run("list all aspects", func(t *testing.T) {
		list, total, err, _ := c.ListAspects(client.AspectListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 4 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, aspects) {
			t.Error(list)
			return
		}
	})
	t.Run("list b aspects", func(t *testing.T) {
		list, total, err, _ := c.ListAspects(client.AspectListOptions{Search: "b"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 1 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.Aspect{aspects[2]}) {
			t.Error(list)
			return
		}
	})
	t.Run("list a1,b3 aspects", func(t *testing.T) {
		list, total, err, _ := c.ListAspects(client.AspectListOptions{Ids: []string{"a1", "b3"}})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 2 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.Aspect{aspects[0], aspects[2]}) {
			t.Error(list)
			return
		}
	})

	t.Run("list all aspect-nodes", func(t *testing.T) {
		list, total, err, _ := c.ListAspectNodes(client.AspectListOptions{SortBy: "name.asc"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 8 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, aspectNodes) {
			t.Error(list)
			return
		}
	})

	t.Run("list b aspects-nodes", func(t *testing.T) {
		list, total, err, _ := c.ListAspectNodes(client.AspectListOptions{Search: "b"})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 1 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.AspectNode{aspectNodes[6]}) {
			t.Error(list)
			return
		}
	})
	t.Run("list a1,a2.2.1,b3, aspects-nodes", func(t *testing.T) {
		list, total, err, _ := c.ListAspectNodes(client.AspectListOptions{Ids: []string{"a1", "a2.2.1", "b3"}})
		if err != nil {
			t.Error(err)
			return
		}
		if total != 3 {
			t.Error(total)
			return
		}
		if !reflect.DeepEqual(list, []models.AspectNode{aspectNodes[0], aspectNodes[4], aspectNodes[6]}) {
			t.Error(list)
			return
		}
	})
}
