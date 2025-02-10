package semantic_legacy

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/client"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller"
	"github.com/SENERGY-Platform/device-repository/lib/tests/repo_legacy/testenv"
	"github.com/SENERGY-Platform/models/go/models"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func TestLocation(t *testing.T) {
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

	bath := models.Location{Id: "urn:infai:ses:location:bath", Name: "Bath", Description: "bath description", Image: "https://i.imgur.com/YHc7cbe.png", DeviceIds: []string{"urn:infai:ses:device:d1", "urn:infai:ses:device:d2"}}
	floor := models.Location{Id: "urn:infai:ses:location:floor", Name: "Floor", Description: "floor description", Image: "https://i.imgur.com/YHc7cbe.png", DeviceGroupIds: []string{"urn:infai:ses:device-group:dg1", "urn:infai:ses:device-group:dg2"}}

	t.Run("testProduceLocation bath", testProduceLocation(conf, bath))
	t.Run("testProduceLocation floor", testProduceLocation(conf, floor))
	t.Run("testLocationRead bath", testLocationRead(ctrl, bath.Id, &bath))
	t.Run("testLocationRead floor", testLocationRead(ctrl, floor.Id, &floor))
	t.Run("testLocationDelete bath", testLocationDelete(conf, bath.Id))
	t.Run("testLocationRead bath after delete", testLocationRead(ctrl, bath.Id, nil))
	t.Run("testLocationRead floor after delete", testLocationRead(ctrl, floor.Id, &floor))
}

func testProduceLocation(conf configuration.Config, location models.Location) func(t *testing.T) {
	return func(t *testing.T) {
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		_, err, _ := c.SetLocation(testenv.AdminToken, location)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testLocationRead(con *controller.Controller, id string, expectedLocation *models.Location) func(t *testing.T) {
	return func(t *testing.T) {
		result, err, code := con.GetLocation(id, testenv.AdminToken)
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
			t.Errorf("\na=%#v\ne=%#v\n", result, expected)
		}
	}
}

func testLocationDelete(conf configuration.Config, id string) func(t *testing.T) {
	return func(t *testing.T) {
		c := client.NewClient("http://localhost:"+conf.ServerPort, nil)
		err, _ := c.DeleteLocation(testenv.AdminToken, id)
		if err != nil {
			t.Error(err)
			return
		}
	}
}
