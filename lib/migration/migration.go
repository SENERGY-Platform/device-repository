/*
 * Copyright 2019 InfAI (CC SES)
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

package migration

import (
	"context"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"github.com/SENERGY-Platform/iot-device-repository/lib/persistence/ordf"
	"log"
	"time"
)

func Migrate(source *ordf.Persistence, sink database.Database) (err error) {
	err = migrateValueTypes(source, sink)
	if err != nil {
		return err
	}
	err = migrateDeviceTypes(source, sink)
	if err != nil {
		return err
	}
	err = migrateDevices(source, sink)
	if err != nil {
		return err
	}
	err = migrateEndpoints(source, sink)
	if err != nil {
		return err
	}
	err = migrateHubs(source, sink)
	if err != nil {
		return err
	}
	return nil
}

func migrateValueTypes(source *ordf.Persistence, sink database.Database) (err error) {
	limit := 100
	offset := 0
	for {
		list := []model.ValueType{}
		err = source.List(&list, limit, offset)
		if err != nil || len(list) == 0 {
			return err
		}
		offset += limit
		for _, e := range list {
			err = source.SelectLevel(&e, -1)
			if err != nil {
				return err
			}
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = sink.SetValueType(ctx, e)
			if err != nil {
				return err
			}
		}
	}
}

func migrateDeviceTypes(source *ordf.Persistence, sink database.Database) (err error) {
	limit := 100
	offset := 0
	for {
		list := []model.DeviceType{}
		err = source.List(&list, limit, offset)
		if err != nil || len(list) == 0 {
			return err
		}
		offset += limit
		for _, e := range list {
			err = source.SelectLevel(&e, -1)
			if err != nil {
				return err
			}
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = sink.SetDeviceType(ctx, e)
			if err != nil {
				return err
			}
		}
	}
}

func migrateDevices(source *ordf.Persistence, sink database.Database) (err error) {
	limit := 100
	offset := 0
	for {
		list := []model.DeviceInstance{}
		err = source.List(&list, limit, offset)
		if err != nil || len(list) == 0 {
			return err
		}
		offset += limit
		for _, e := range list {
			err = source.SelectLevel(&e, -1)
			if err != nil {
				return err
			}
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = sink.SetDevice(ctx, e)
			if err != nil {
				return err
			}
		}
	}
}

func migrateEndpoints(source *ordf.Persistence, sink database.Database) (err error) {
	limit := 100
	offset := 0
	for {
		list := []model.Endpoint{}
		err = source.List(&list, limit, offset)
		if err != nil || len(list) == 0 {
			return err
		}
		offset += limit
		for _, e := range list {
			err = source.SelectLevel(&e, -1)
			if err != nil {
				return err
			}
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			err = sink.SetEndpoint(ctx, e)
			if err != nil {
				return err
			}
		}
	}
}

func migrateHubs(source *ordf.Persistence, sink database.Database) (err error) {
	limit := 100
	offset := 0
	for {
		list := []model.GatewayFlat{}
		err = source.List(&list, limit, offset)
		if err != nil || len(list) == 0 {
			return err
		}
		offset += limit
		for _, e := range list {
			err = source.SelectLevel(&e, -1)
			if err != nil {
				return err
			}
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			valid, hub, err := gatewayToHub(sink, ctx, e)
			if err != nil {
				return err
			}
			if !valid {
				log.Println("ERROR: inconsistent gateway will be skipped", e)
				continue
			}
			err = sink.SetHub(ctx, hub)
			if err != nil {
				return err
			}
		}
	}
}

func gatewayToHub(db database.Database, ctx context.Context, flat model.GatewayFlat) (valid bool, result model.Hub, err error) {
	result.Id = flat.Id
	result.Name = flat.Name
	result.Hash = flat.Hash
	used := map[string]bool{}
	for _, deviceId := range flat.Devices {
		if !used[deviceId] {
			used[deviceId] = true
			device, exists, err := db.GetDevice(ctx, deviceId)
			if err != nil || !exists {
				return exists, result, err
			}
			result.Devices = append(result.Devices, device.Url)
		}
	}
	return true, result, nil
}
