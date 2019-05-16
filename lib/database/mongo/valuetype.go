package mongo

import (
	"context"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
)

func (this *Mongo) GetValueType(ctx context.Context, id string) (model.ValueType, bool, error) {
	panic("implement me")
}

func (this *Mongo) SetValueType(ctx context.Context, valueType model.ValueType) error {
	panic("implement me")
}

func (this *Mongo) RemoveValueType(ctx context.Context, id string) error {
	panic("implement me")
}

func (this *Mongo) ListValueTypesUsingValueType(ctx context.Context, id string) ([]model.ValueType, error) {
	panic("implement me")
}
