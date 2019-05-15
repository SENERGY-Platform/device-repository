package mongo

import "github.com/SENERGY-Platform/iot-device-repository/lib/model"

func (this *Mongo) GetValueType(id string) (model.ValueType, bool, error) {
	panic("implement me") //TODO
}

func (this *Mongo) SetValueType(valueType model.ValueType) error {
	panic("implement me") //TODO
}

func (this *Mongo) RemoveValueType(id string) error {
	panic("implement me") //TODO
}

func (this *Mongo) ListValueTypesUsingValueType(id string) ([]model.ValueType, error) {
	panic("implement me") //TODO
}
