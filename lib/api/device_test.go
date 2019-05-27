package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

var devicetype1id = uuid.NewV4().String()
var devicetype1name = uuid.NewV4().String()
var device1id = uuid.NewV4().String()
var device1name = uuid.NewV4().String()
var device1uri = uuid.NewV4().String()

func init() {
	before = append(before, InitDevices)
}

func InitDevices() error {
	err := producer.PublishDeviceType(model.DeviceType{Id: devicetype1id, Name: devicetype1name}, userid)
	if err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	err = producer.PublishDevice(model.DeviceInstance{Id: device1id, Name: device1name, Url: device1uri, DeviceType: devicetype1id}, userid)
	if err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	return nil
}

func TestHeartbeat(t *testing.T) {
	t.Parallel()
	resp, err := userjwt.Get("http://localhost:" + configuration.ServerPort)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("no heart beat")
		return
	}
}

func TestDeviceRead(t *testing.T) {
	t.Parallel()

	endpoint := "http://localhost:" + configuration.ServerPort + "/devices/" + url.PathEscape(device1id)
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Name != device1name || result.Url != device1uri {
		t.Error("unexpected result", result)
		return
	}
}

func TestDeviceList(t *testing.T) {
	t.Parallel()
	endpoint := "http://localhost:" + configuration.ServerPort + "/devices"
	resp, err := userjwt.Get(endpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", endpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	result := []model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Error(err)
	}
	if len(result) != 1 || result[0].Id != device1id {
		t.Error("unexpected result", result)
		return
	}
}
