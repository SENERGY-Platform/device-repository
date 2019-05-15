package mongo

import "github.com/SENERGY-Platform/iot-device-repository/lib/model"

func (this *Mongo) ListEndpointsOfDevice(deviceId string) ([]model.Endpoint, error) {
	panic("implement me") //TODO
}

func (this *Mongo) RemoveEndpoint(id string) error {
	panic("implement me") //TODO
}

func (this *Mongo) SetEndpoint(endpoint model.Endpoint) error {
	panic("implement me") //TODO
}
