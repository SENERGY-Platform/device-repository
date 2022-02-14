package semantic_legacy

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	"github.com/SENERGY-Platform/device-repository/lib/tests/semantic_legacy/producer"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

func TestLocation(t *testing.T) {
	conf, err := config.Load("../../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	defer cancel()
	conf, ctrl, prod, err := NewPartialMockEnv(ctx, wg, conf)
	if err != nil {
		t.Error(err)
		return
	}

	bath := model.Location{Id: "urn:infai:ses:location:bath", Name: "Bath", Description: "bath description", Image: "https://i.imgur.com/YHc7cbe.png", DeviceIds: []string{"urn:infai:ses:device:d1", "urn:infai:ses:device:d2"}}
	floor := model.Location{Id: "urn:infai:ses:location:floor", Name: "Floor", Description: "floor description", Image: "https://i.imgur.com/YHc7cbe.png", DeviceGroupIds: []string{"urn:infai:ses:device-group:dg1", "urn:infai:ses:device-group:dg2"}}

	t.Run("testProduceLocation bath", testProduceLocation(prod, bath))
	t.Run("testProduceLocation floor", testProduceLocation(prod, floor))
	time.Sleep(2 * time.Second)
	t.Run("testLocationRead bath", testLocationRead(ctrl, bath.Id, &bath))
	t.Run("testLocationRead floor", testLocationRead(ctrl, floor.Id, &floor))
	t.Run("testLocationDelete bath", testLocationDelete(prod, bath.Id))
	t.Run("testLocationRead bath after delete", testLocationRead(ctrl, bath.Id, nil))
	t.Run("testLocationRead floor after delete", testLocationRead(ctrl, floor.Id, &floor))
}

func testProduceLocation(producer *producer.Producer, location model.Location) func(t *testing.T) {
	return func(t *testing.T) {
		err := producer.PublishLocation(location, "sdfdsfsf")
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testLocationRead(con *controller.Controller, id string, expectedLocation *model.Location) func(t *testing.T) {
	return func(t *testing.T) {
		result, err, code := con.GetLocation(id, "")
		if err != nil {
			if expectedLocation != nil {
				t.Error(code, err)
				return
			}
			t.Log("expected error received:", err)
			err = nil
		} else {
			if expectedLocation == nil {
				t.Error("expected error, not result", result)
				return
			}
		}

		if expectedLocation == nil {
			return
		}
		expected := *expectedLocation //copy -> no side effects
		sort.Strings(expected.DeviceGroupIds)
		sort.Strings(expected.DeviceIds)
		sort.Strings(result.DeviceGroupIds)
		sort.Strings(result.DeviceIds)
		if !reflect.DeepEqual(result, expected) {
			resultJson, _ := json.Marshal(result)
			expectedJson, _ := json.Marshal(expected)
			t.Error(string(resultJson), "\n\n", string(expectedJson))
		}
	}
}

func testLocationDelete(producer *producer.Producer, id string) func(t *testing.T) {
	return func(t *testing.T) {
		err := producer.PublishLocationDelete(id, "sdfdsfsf")
		if err != nil {
			t.Error(err)
			return
		}
	}
}
