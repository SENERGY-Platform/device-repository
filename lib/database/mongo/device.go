package mongo

import (
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
)

func (this *Mongo) ReadDevice(id string) (device model.DeviceInstance, exists bool, err error) {
	panic("implement me")
}

func (this *Mongo) SetDevice(device model.DeviceInstance) error {
	panic("implement me")
}

func (this *Mongo) RemoveDevice(id string) error {
	panic("implement me")
}
