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
	for i := 0; i < 20; i++ {
		err = producer.PublishDevice(model.DeviceInstance{Id: uuid.NewV4().String(), Name: uuid.NewV4().String(), Url: uuid.NewV4().String(), DeviceType: devicetype1id}, userid)
		if err != nil {
			return err
		}
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
	if len(result) != 21 {
		t.Error("unexpected result", result)
		return
	}
}

func TestDeviceListLimit10(t *testing.T) {
	t.Parallel()
	endpoint := "http://localhost:" + configuration.ServerPort + "/devices?limit=10"
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
	if len(result) != 10 {
		t.Error("unexpected result", result)
		return
	}
}

func TestDeviceListLimit10Offset20(t *testing.T) {
	t.Parallel()
	endpoint := "http://localhost:" + configuration.ServerPort + "/devices?limit=10&offset=20"
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
	if len(result) != 1 {
		t.Error("unexpected result", result)
		return
	}
}

func TestDeviceListSort(t *testing.T) {
	t.Parallel()
	ascendpoint := "http://localhost:" + configuration.ServerPort + "/devices?sort=name.asc"
	resp, err := userjwt.Get(ascendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", ascendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	ascresult := []model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&ascresult)
	if err != nil {
		t.Error(err)
	}
	if len(ascresult) != 21 {
		t.Error("unexpected result", ascresult)
		return
	}

	descendpoint := "http://localhost:" + configuration.ServerPort + "/devices?sort=name.desc"
	resp, err = userjwt.Get(descendpoint)
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		t.Error("unexpectet response", descendpoint, resp.Status, resp.StatusCode, string(b))
		return
	}
	descresult := []model.DeviceInstance{}
	err = json.NewDecoder(resp.Body).Decode(&descresult)
	if err != nil {
		t.Error(err)
	}
	if len(ascresult) != 21 {
		t.Error("unexpected result", descresult)
		return
	}

	for i := 0; i < 21; i++ {
		if descresult[i].Id != ascresult[20-i].Id {
			t.Error("unexpected sorting result", i, descresult[i].Id, ascresult[20-i].Id)
			return
		}
	}
}
