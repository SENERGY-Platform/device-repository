package semantic_legacy

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/tests/repo_legacy/testenv"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func TestConceptCharacteristic(t *testing.T) {
	conf, err := configuration.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()
	conf, ctrl, err := NewPartialMockEnv(ctx, wg, conf, t)
	if err != nil {
		t.Error(err)
		return
	}

	concept := models.Concept{
		Id:                "urn:ses:infai:concept:c",
		Name:              "cn",
		CharacteristicIds: []string{"urn:ses:infai:characteristicch1", "urn:ses:infai:characteristicch2"},
	}

	characterisitc1 := models.Characteristic{
		Id:   "urn:ses:infai:characteristicch1",
		Name: "chn",
		Type: models.Boolean,
	}
	characterisitc2 := models.Characteristic{
		Id:   "urn:ses:infai:characteristicch2",
		Name: "chn",
		Type: models.Boolean,
	}

	_, err, _ = ctrl.SetConcept(testenv.AdminToken, concept)
	if err != nil {
		t.Error(err)
		return
	}
	_, err, _ = ctrl.SetCharacteristic(testenv.AdminToken, characterisitc1)
	if err != nil {
		t.Error(err)
		return
	}
	_, err, _ = ctrl.SetCharacteristic(testenv.AdminToken, characterisitc2)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("check after init", checkConcept(ctrl, "urn:ses:infai:concept:c", concept))

	_, err, _ = ctrl.SetConcept(testenv.AdminToken, concept)
	if err != nil {
		t.Error(err)
		return
	}
	t.Run("check after reset concept", checkConcept(ctrl, "urn:ses:infai:concept:c", concept))

	_, err, _ = ctrl.SetCharacteristic(testenv.AdminToken, characterisitc1)
	if err != nil {
		t.Error(err)
		return
	}
	t.Run("check after reset characteristic", checkConcept(ctrl, "urn:ses:infai:concept:c", concept))
}

func checkConcept(ctrl *controller.Controller, id string, expected models.Concept) func(t *testing.T) {
	return func(t *testing.T) {
		result, err, _ := ctrl.GetConceptWithoutCharacteristics(id)
		if err != nil {
			t.Error(err, result)
			return
		}
		sort.Strings(result.CharacteristicIds)
		sort.Strings(expected.CharacteristicIds)
		if !reflect.DeepEqual(result, expected) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expected)
			t.Error(string(resultJson), string(expectedJson))
		}
	}
}
