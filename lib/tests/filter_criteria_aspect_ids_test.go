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

package tests

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/controller/publisher"
	"github.com/SENERGY-Platform/device-repository/lib/database/mongo"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/docker"
	"github.com/SENERGY-Platform/models/go/models"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
)

// TestFilterCriteriaAspectIds covers the move of the filter-criteria from a single AspectId
// to an AspectIds list: the deprecated field stays usable as an alias for a single element
// list, several aspects in one criteria are ANDed, and the startup migration converts the
// stored data that still holds one aspect per row.
func TestFilterCriteriaAspectIds(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.SyncLockDuration = time.Second.String()
	config.Debug = true

	config, err = docker.NewEnv(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	err = lib.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(1 * time.Second)

	c := client.NewClient("http://localhost:"+config.ServerPort, nil)

	const measuringFunction = models.URN_PREFIX + "measuring-function:getTemperature"
	const (
		airAspect     = models.URN_PREFIX + "aspect:air"
		insideAspect  = models.URN_PREFIX + "aspect:inside"
		outsideAspect = models.URN_PREFIX + "aspect:outside"
		deviceClass   = models.URN_PREFIX + "device-class:dc1"
	)

	t.Run("create metadata", func(t *testing.T) {
		for _, aspectId := range []string{airAspect, insideAspect, outsideAspect} {
			_, err, _ := c.SetAspect(client.InternalAdminToken, models.Aspect{Id: aspectId, Name: aspectId})
			if err != nil {
				t.Error(err)
				return
			}
		}
		_, err, _ := c.SetFunction(client.InternalAdminToken, models.Function{
			Id:          measuringFunction,
			Name:        "getTemperature",
			DisplayName: "getTemperature",
			RdfType:     models.SES_ONTOLOGY_MEASURING_FUNCTION,
		})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetDeviceClass(client.InternalAdminToken, models.DeviceClass{Id: deviceClass, Name: "dc1"})
		if err != nil {
			t.Error(err)
			return
		}
		_, err, _ = c.SetProtocol(client.InternalAdminToken, models.Protocol{
			Id:               "p1",
			Name:             "p1",
			Handler:          "p1",
			ProtocolSegments: []models.ProtocolSegment{{Id: "ps1", Name: "ps1"}},
		})
		if err != nil {
			t.Error(err)
			return
		}
	})

	deviceType := func(id string, aspectIds []string) models.DeviceType {
		return models.DeviceType{
			Id:            id,
			Name:          id,
			DeviceClassId: deviceClass,
			Services: []models.Service{{
				Id:          id + "_s1",
				LocalId:     id + "_s1",
				Name:        "s1",
				Interaction: models.EVENT_AND_REQUEST,
				ProtocolId:  "p1",
				Outputs: []models.Content{{
					Id:                id + "_s1_c1",
					Serialization:     models.JSON,
					ProtocolSegmentId: "ps1",
					ContentVariable: models.ContentVariable{
						Id:         id + "_cv1",
						Name:       "temperature",
						Type:       models.String,
						FunctionId: measuringFunction,
						AspectIds:  aspectIds,
					},
				}},
			}},
		}
	}

	selectableDeviceTypeIds := func(t *testing.T, criteria []model.FilterCriteria) []string {
		t.Helper()
		selectables, err, _ := c.GetDeviceTypeSelectablesV2(criteria, "", false, false)
		if err != nil {
			t.Error(err)
			return nil
		}
		result := []string{}
		for _, selectable := range selectables {
			result = append(result, selectable.DeviceTypeId)
		}
		slices.Sort(result)
		return result
	}

	//dt_both carries both aspects on one content variable, dt_air only one of them
	t.Run("create device-types", func(t *testing.T) {
		for id, aspectIds := range map[string][]string{
			"dt_both": {airAspect, insideAspect},
			"dt_air":  {airAspect},
		} {
			_, err, _ := c.SetDeviceType(client.InternalAdminToken, deviceType(id, aspectIds), client.DeviceTypeUpdateOptions{})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})

	t.Run("selectables: aspect_id is an alias for a single element aspect_ids", func(t *testing.T) {
		byDeprecated := selectableDeviceTypeIds(t, []model.FilterCriteria{{FunctionId: measuringFunction, AspectId: airAspect}})
		byList := selectableDeviceTypeIds(t, []model.FilterCriteria{{FunctionId: measuringFunction, AspectIds: []string{airAspect}}})
		if !slices.Equal(byDeprecated, []string{"dt_air", "dt_both"}) {
			t.Error("unexpected result for aspect_id", byDeprecated)
		}
		if !slices.Equal(byDeprecated, byList) {
			t.Error("aspect_id and aspect_ids disagree", byDeprecated, byList)
		}
	})

	t.Run("selectables: several aspects in one criteria are anded", func(t *testing.T) {
		result := selectableDeviceTypeIds(t, []model.FilterCriteria{{
			FunctionId: measuringFunction,
			AspectIds:  []string{airAspect, insideAspect},
		}})
		if !slices.Equal(result, []string{"dt_both"}) {
			t.Error("expected only the device-type carrying both aspects", result)
		}
	})

	t.Run("selectables: an aspect no device-type carries matches nothing", func(t *testing.T) {
		result := selectableDeviceTypeIds(t, []model.FilterCriteria{{
			FunctionId: measuringFunction,
			AspectIds:  []string{airAspect, outsideAspect},
		}})
		if len(result) != 0 {
			t.Error("expected no result", result)
		}
	})

	t.Run("selectables: a path option names only the matched aspects", func(t *testing.T) {
		selectables, err, _ := c.GetDeviceTypeSelectables([]model.FilterCriteria{{
			FunctionId: measuringFunction,
			AspectIds:  []string{airAspect},
		}}, "", nil, false)
		if err != nil {
			t.Error(err)
			return
		}
		for _, selectable := range selectables {
			for _, options := range selectable.ServicePathOptions {
				for _, option := range options {
					//the unmatched aspect of dt_both must not show up on the path option
					if !slices.Equal(aspectNodeIds(option.AspectNodes), []string{airAspect}) {
						t.Error("unexpected aspect nodes", selectable.DeviceTypeId, aspectNodeIds(option.AspectNodes))
					}
					if option.AspectNode.Id != airAspect {
						t.Error("unexpected aspect node", selectable.DeviceTypeId, option.AspectNode.Id)
					}
				}
			}
		}
	})

	t.Run("selectables: a path option lists every matched aspect once", func(t *testing.T) {
		selectables, err, _ := c.GetDeviceTypeSelectablesV2([]model.FilterCriteria{{
			FunctionId: measuringFunction,
			AspectIds:  []string{airAspect, insideAspect},
		}}, "", false, false)
		if err != nil {
			t.Error(err)
			return
		}
		if len(selectables) != 1 || len(selectables[0].ServicePathOptions) != 1 {
			t.Error("expected the one device-type carrying both aspects", selectables)
			return
		}
		for _, options := range selectables[0].ServicePathOptions {
			if len(options) != 1 {
				t.Error("expected one path option per path, not one per aspect", options)
				return
			}
			if !slices.Equal(aspectNodeIds(options[0].AspectNodes), []string{airAspect, insideAspect}) {
				t.Error("expected both aspect nodes", aspectNodeIds(options[0].AspectNodes))
			}
			//the deprecated field can only carry one of them, and gets the alphabetically first
			if options[0].AspectNode.Id != airAspect {
				t.Error("unexpected aspect node", options[0].AspectNode.Id)
			}
		}
	})

	devices := map[string]models.Device{}
	multiAspectGroup := models.DeviceGroup{}

	t.Run("device-groups", func(t *testing.T) {
		for _, deviceTypeId := range []string{"dt_both", "dt_air"} {
			device, err, _ := c.CreateDevice(client.InternalAdminToken, models.Device{
				LocalId:      "d_" + deviceTypeId,
				Name:         "d_" + deviceTypeId,
				DeviceTypeId: deviceTypeId,
			})
			if err != nil {
				t.Error(err)
				return
			}
			devices[deviceTypeId] = device
		}

		listGroupIds := func(t *testing.T, criteria []model.FilterCriteria) []string {
			t.Helper()
			groups, _, err, _ := c.ListDeviceGroups(client.InternalAdminToken, client.DeviceGroupListOptions{Criteria: criteria})
			if err != nil {
				t.Error(err)
				return nil
			}
			result := []string{}
			for _, group := range groups {
				result = append(result, group.Id)
			}
			slices.Sort(result)
			return result
		}

		t.Run("generated criteria carry both aspect fields", func(t *testing.T) {
			group, err, _ := c.ReadDeviceGroup(model.DeviceIdToGeneratedDeviceGroupId(devices["dt_both"].Id), client.InternalAdminToken, false)
			if err != nil {
				t.Error(err)
				return
			}
			assertAspectFieldsConsistent(t, group.Criteria)
			//the content variable of dt_both carries both aspects, so one criteria holds both
			found := slices.ContainsFunc(group.Criteria, func(criteria models.DeviceGroupFilterCriteria) bool {
				return slices.Equal(criteria.AspectIds, []string{airAspect, insideAspect})
			})
			if !found {
				t.Error("expected a criteria over the whole aspect list of the content variable", group.Criteria)
			}
		})

		t.Run("aspect_id is an alias for a single element aspect_ids", func(t *testing.T) {
			byDeprecated := listGroupIds(t, []model.FilterCriteria{{FunctionId: measuringFunction, AspectId: airAspect}})
			byList := listGroupIds(t, []model.FilterCriteria{{FunctionId: measuringFunction, AspectIds: []string{airAspect}}})
			if len(byDeprecated) != 2 {
				t.Error("expected the generated group of both devices", byDeprecated)
			}
			if !slices.Equal(byDeprecated, byList) {
				t.Error("aspect_id and aspect_ids disagree", byDeprecated, byList)
			}
		})

		t.Run("several aspects in one criteria are anded", func(t *testing.T) {
			result := listGroupIds(t, []model.FilterCriteria{{
				FunctionId: measuringFunction,
				AspectIds:  []string{airAspect, insideAspect},
			}})
			expected := []string{model.DeviceIdToGeneratedDeviceGroupId(devices["dt_both"].Id)}
			if !slices.Equal(result, expected) {
				t.Error("expected only the group of the device carrying both aspects", result, expected)
			}
		})

		//a stored criteria over several aspects claims more than the generated ones do: the
		//devices answer the function on one content variable carrying all of the aspects
		multiAspectCriteria := []models.DeviceGroupFilterCriteria{{
			Interaction: models.EVENT,
			FunctionId:  measuringFunction,
			AspectIds:   []string{airAspect, insideAspect},
		}}

		t.Run("a stored criteria may name several aspects", func(t *testing.T) {
			var err error
			multiAspectGroup, err, _ = c.SetDeviceGroup(client.InternalAdminToken, models.DeviceGroup{
				Name:      "group with a multi aspect criteria",
				DeviceIds: []string{devices["dt_both"].Id},
				Criteria:  multiAspectCriteria,
			})
			if err != nil {
				t.Error(err)
				return
			}
			stored, err, _ := c.ReadDeviceGroup(multiAspectGroup.Id, client.InternalAdminToken, false)
			if err != nil {
				t.Error(err)
				return
			}
			if len(stored.Criteria) != 1 || !slices.Equal(stored.Criteria[0].AspectIds, []string{airAspect, insideAspect}) {
				t.Error("expected both aspects to survive the write", stored.Criteria)
				return
			}
			//the deprecated field can only carry one of them, and gets the alphabetically first
			if stored.Criteria[0].AspectId != airAspect {
				t.Error("unexpected aspect_id", stored.Criteria[0].AspectId)
			}
		})

		t.Run("a device that carries only one of the aspects is refused", func(t *testing.T) {
			_, err, code := c.SetDeviceGroup(client.InternalAdminToken, models.DeviceGroup{
				Name:      "group with a device that misses an aspect",
				DeviceIds: []string{devices["dt_air"].Id},
				Criteria:  multiAspectCriteria,
			})
			if err == nil {
				t.Error("expected the write to be refused")
				return
			}
			if code != 400 {
				t.Error("unexpected code", code, err)
			}
		})

		t.Run("a criteria without an aspect does not filter by aspect", func(t *testing.T) {
			result := listGroupIds(t, []model.FilterCriteria{{
				Interaction: models.EVENT,
				FunctionId:  measuringFunction,
			}})
			expected := []string{
				multiAspectGroup.Id,
				model.DeviceIdToGeneratedDeviceGroupId(devices["dt_air"].Id),
				model.DeviceIdToGeneratedDeviceGroupId(devices["dt_both"].Id),
			}
			for _, id := range expected {
				if !slices.Contains(result, id) {
					t.Error("every group whose criteria carry an aspect has to be found", id, result)
				}
			}
		})

		//a query names its aspects one at a time or together, and a stored criteria has to
		//carry all of them
		for _, query := range [][]string{{airAspect}, {insideAspect}, {airAspect, insideAspect}} {
			t.Run("multi aspect criteria is found by "+strings.Join(query, ","), func(t *testing.T) {
				result := listGroupIds(t, []model.FilterCriteria{{
					Interaction: models.EVENT,
					FunctionId:  measuringFunction,
					AspectIds:   query,
				}})
				if !slices.Contains(result, multiAspectGroup.Id) {
					t.Error("group with the multi aspect criteria not found", result)
				}
			})
		}
	})

	t.Run("startup migration", func(t *testing.T) {
		testStartupMigration(t, config, measuringFunction, airAspect, insideAspect, deviceClass,
			model.DeviceIdToGeneratedDeviceGroupId(devices["dt_both"].Id), multiAspectGroup.Id)
	})
}

// testStartupMigration builds the pre-migration shape of the two stored collections by
// hand. Neither shape is producible through the repository itself once the write path has
// been changed, so the documents are written with a raw mongo client.
func testStartupMigration(t *testing.T, config configuration.Config, measuringFunction string, airAspect string, insideAspect string, deviceClass string, generatedGroupId string, manualGroupId string) {
	ctx := context.Background()
	raw, err := mongodriver.Connect(ctx, mongooptions.Client().ApplyURI(config.MongoUrl))
	if err != nil {
		t.Error(err)
		return
	}
	defer raw.Disconnect(ctx)
	criteriaCollection := raw.Database(config.MongoTable).Collection(config.MongoDeviceTypeCollection + "_criteria")
	deviceGroupCollection := raw.Database(config.MongoTable).Collection(config.MongoDeviceGroupCollection)

	//a device-type that predates models.ContentVariable.AspectIds: its stored content
	//variable carries only the deprecated AspectId, so the migration can only rebuild its
	//criteria with the normalization handed in by the controller
	legacyDeviceType := models.DeviceType{
		Id:            "dt_migration",
		Name:          "dt_migration",
		DeviceClassId: deviceClass,
		Services: []models.Service{{
			Id:          "dt_migration_s1",
			LocalId:     "dt_migration_s1",
			Name:        "s1",
			Interaction: models.EVENT_AND_REQUEST,
			ProtocolId:  "p1",
			Outputs: []models.Content{{
				Id:                "dt_migration_s1_c1",
				Serialization:     models.JSON,
				ProtocolSegmentId: "ps1",
				ContentVariable: models.ContentVariable{
					Id:         "dt_migration_cv1",
					Name:       "temperature",
					Type:       models.String,
					FunctionId: measuringFunction,
					AspectId:   airAspect,
				},
			}},
		}},
	}
	db, err := mongo.New(config)
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Disconnect()
	err = db.SetDeviceType(ctx, legacyDeviceType, func(models.DeviceType) error { return nil })
	if err != nil {
		t.Error(err)
		return
	}

	//replace the criteria of every device-type by the pre-migration shape: one row per
	//aspect, holding the aspect in the single field the rows carried back then
	cursor, err := criteriaCollection.Find(ctx, bson.M{})
	if err != nil {
		t.Error(err)
		return
	}
	legacyRows := []interface{}{}
	for cursor.Next(ctx) {
		row := bson.M{}
		err = cursor.Decode(&row)
		if err != nil {
			t.Error(err)
			return
		}
		delete(row, "_id")
		aspectIds, _ := row["aspectids"].(bson.A)
		delete(row, "aspectids")
		if len(aspectIds) == 0 {
			row["aspectid"] = ""
			legacyRows = append(legacyRows, row)
			continue
		}
		for _, aspectId := range aspectIds {
			legacyRow := bson.M{}
			for key, value := range row {
				legacyRow[key] = value
			}
			legacyRow["aspectid"] = aspectId
			legacyRows = append(legacyRows, legacyRow)
		}
	}
	err = cursor.Err()
	if err != nil {
		t.Error(err)
		return
	}
	_, err = criteriaCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		t.Error(err)
		return
	}
	_, err = criteriaCollection.InsertMany(ctx, legacyRows)
	if err != nil {
		t.Error(err)
		return
	}

	//the generator used to write one criteria per aspect and none over the whole list, so
	//drop the criteria that name several aspects
	_, err = deviceGroupCollection.UpdateMany(ctx, bson.M{}, bson.M{"$pull": bson.M{
		"criteria": bson.M{"aspectids.1": bson.M{"$exists": true}},
	}})
	if err != nil {
		t.Error(err)
		return
	}

	//and strip aspect_ids from the criteria that name at most one aspect, which is every
	//criteria that can predate the field
	_, err = deviceGroupCollection.UpdateMany(ctx, bson.M{},
		bson.M{"$unset": bson.M{"criteria.$[element].aspectids": ""}},
		mongooptions.Update().SetArrayFilters(mongooptions.ArrayFilters{Filters: []interface{}{
			bson.M{"element.aspectids.1": bson.M{"$exists": false}},
		}}))
	if err != nil {
		t.Error(err)
		return
	}

	manualShortsBeforeMigration := deviceGroupCriteriaShorts(t, ctx, deviceGroupCollection, manualGroupId)

	//the migration needs the controller: it normalizes stored device-types and rebuilds the
	//criteria of the generated device-groups
	migrationController, err := controller.New(config, db, publisher.Void{}, nil)
	if err != nil {
		t.Error(err)
		return
	}
	err = db.RunStartupMigrations(migrationController)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("no criteria row holds the deprecated aspect field", func(t *testing.T) {
		count, err := criteriaCollection.CountDocuments(ctx, bson.M{"aspectid": bson.M{"$exists": true}})
		if err != nil {
			t.Error(err)
			return
		}
		if count != 0 {
			t.Error("expected every criteria row to be rebuilt", count)
		}
	})

	t.Run("a rebuilt row holds every aspect of its content variable", func(t *testing.T) {
		row := bson.M{}
		err := criteriaCollection.FindOne(ctx, bson.M{"contentvariableid": "dt_both_cv1"}).Decode(&row)
		if err != nil {
			t.Error(err)
			return
		}
		aspectIds := bsonStrings(row["aspectids"])
		if len(aspectIds) != 2 || !slices.Contains(aspectIds, airAspect) || !slices.Contains(aspectIds, insideAspect) {
			t.Error("unexpected aspect_ids", row["aspectids"])
		}
	})

	t.Run("a device-type stored with the deprecated aspect_id keeps its aspect", func(t *testing.T) {
		row := bson.M{}
		err := criteriaCollection.FindOne(ctx, bson.M{"contentvariableid": "dt_migration_cv1"}).Decode(&row)
		if err != nil {
			t.Error(err)
			return
		}
		aspectIds := bsonStrings(row["aspectids"])
		if !slices.Equal(aspectIds, []string{airAspect}) {
			t.Error("expected the deprecated content variable aspect to survive", row["aspectids"])
		}
	})

	t.Run("device-group criteria get aspect_ids without changing criteria_short", func(t *testing.T) {
		cursor, err := deviceGroupCollection.Find(ctx, bson.M{})
		if err != nil {
			t.Error(err)
			return
		}
		groups := []models.DeviceGroup{}
		err = cursor.All(ctx, &groups)
		if err != nil {
			t.Error(err)
			return
		}
		if len(groups) == 0 {
			t.Error("expected stored device-groups")
			return
		}
		for _, group := range groups {
			assertAspectFieldsConsistent(t, group.Criteria)
		}
		if !slices.Equal(manualShortsBeforeMigration, deviceGroupCriteriaShorts(t, ctx, deviceGroupCollection, manualGroupId)) {
			t.Error("a manually created group must keep its criteria, they are not derived data")
		}
	})

	t.Run("a generated device-group gets a criteria over the whole aspect list back", func(t *testing.T) {
		group := models.DeviceGroup{}
		err := deviceGroupCollection.FindOne(ctx, bson.M{"id": generatedGroupId}).Decode(&group)
		if err != nil {
			t.Error(err)
			return
		}
		found := slices.ContainsFunc(group.Criteria, func(criteria models.DeviceGroupFilterCriteria) bool {
			return criteria.FunctionId == measuringFunction &&
				slices.Equal(criteria.AspectIds, []string{airAspect, insideAspect})
		})
		if !found {
			t.Error("expected the criteria over both aspects to be rebuilt", group.Criteria)
		}
		//the single aspect criteria carry the intersection over the devices of a group and
		//have to survive next to it
		for _, aspectId := range []string{airAspect, insideAspect} {
			found = slices.ContainsFunc(group.Criteria, func(criteria models.DeviceGroupFilterCriteria) bool {
				return criteria.FunctionId == measuringFunction && slices.Equal(criteria.AspectIds, []string{aspectId})
			})
			if !found {
				t.Error("expected the single aspect criteria to survive", aspectId, group.Criteria)
			}
		}
	})

	t.Run("the migration is idempotent", func(t *testing.T) {
		before, err := criteriaCollection.CountDocuments(ctx, bson.M{})
		if err != nil {
			t.Error(err)
			return
		}
		err = db.RunStartupMigrations(migrationController)
		if err != nil {
			t.Error(err)
			return
		}
		after, err := criteriaCollection.CountDocuments(ctx, bson.M{})
		if err != nil {
			t.Error(err)
			return
		}
		if before != after {
			t.Error("unexpected criteria count after a second run", before, after)
		}
	})
}

func deviceGroupCriteriaShorts(t *testing.T, ctx context.Context, collection *mongodriver.Collection, deviceGroupId string) []string {
	t.Helper()
	cursor, err := collection.Find(ctx, bson.M{"id": deviceGroupId})
	if err != nil {
		t.Error(err)
		return nil
	}
	groups := []models.DeviceGroup{}
	err = cursor.All(ctx, &groups)
	if err != nil {
		t.Error(err)
		return nil
	}
	result := []string{}
	for _, group := range groups {
		for _, short := range group.CriteriaShort {
			result = append(result, group.Id+"::"+short)
		}
	}
	slices.Sort(result)
	return result
}

func bsonStrings(value interface{}) []string {
	array, ok := value.(bson.A)
	if !ok {
		return nil
	}
	result := []string{}
	for _, element := range array {
		text, _ := element.(string)
		result = append(result, text)
	}
	return result
}

// assertAspectFieldsConsistent checks the compatibility rule of a stored criteria: AspectIds
// is the truth, and the deprecated AspectId carries its alphabetically first entry.
func assertAspectFieldsConsistent(t *testing.T, criteria []models.DeviceGroupFilterCriteria) {
	t.Helper()
	for _, c := range criteria {
		if len(c.AspectIds) == 0 {
			if c.AspectId != "" {
				t.Error("expected aspect_id to be folded into aspect_ids", c)
			}
			continue
		}
		sorted := slices.Sorted(slices.Values(c.AspectIds))
		if c.AspectId != sorted[0] {
			t.Error("expected aspect_id to hold the alphabetically first aspect", c)
		}
	}
}

func aspectNodeIds(nodes []models.AspectNode) []string {
	result := []string{}
	for _, node := range nodes {
		result = append(result, node.Id)
	}
	return result
}
