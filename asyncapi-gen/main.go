/*
 * Copyright 2025 InfAI (CC SES)
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

package main

import (
	"flag"
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/lib/controller/publisher"
	"log"
	"os"

	"github.com/swaggest/go-asyncapi/reflector/asyncapi-2.4.0"
	"github.com/swaggest/go-asyncapi/spec-2.4.0"
)

//go:generate go run main.go

func main() {
	configLocation := flag.String("config", "../config.json", "configuration file")
	flag.Parse()

	conf, err := configuration.Load(*configLocation)
	if err != nil {
		log.Fatal("ERROR: unable to load config", err)
	}

	asyncAPI := spec.AsyncAPI{}
	asyncAPI.Info.Title = "Device-Repository"

	asyncAPI.AddServer("kafka", spec.Server{
		URL:      conf.KafkaUrl,
		Protocol: "kafka",
	})

	reflector := asyncapi.Reflector{}
	reflector.Schema = &asyncAPI

	mustNotFail := func(err error) {
		if err != nil {
			panic(err.Error())
		}
	}

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.AspectTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "AspectCommand",
				Title: "AspectCommand",
			},
			MessageSample: new(publisher.AspectCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.CharacteristicTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "CharacteristicCommand",
				Title: "CharacteristicCommand",
			},
			MessageSample: new(publisher.CharacteristicCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.ConceptTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "ConceptCommand",
				Title: "ConceptCommand",
			},
			MessageSample: new(publisher.ConceptCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.DeviceClassTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "DeviceClassCommand",
				Title: "DeviceClassCommand",
			},
			MessageSample: new(publisher.DeviceClassCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.DeviceGroupTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "DeviceGroupCommand",
				Title: "DeviceGroupCommand",
			},
			MessageSample: new(publisher.DeviceGroupCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.FunctionTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "FunctionCommand",
				Title: "FunctionCommand",
			},
			MessageSample: new(publisher.FunctionCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.HubTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "HubCommand",
				Title: "HubCommand",
			},
			MessageSample: new(publisher.HubCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.LocationTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "LocationCommand",
				Title: "LocationCommand",
			},
			MessageSample: new(publisher.LocationCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.DeviceTypeTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "DeviceTypeCommand",
				Title: "DeviceTypeCommand",
			},
			MessageSample: new(publisher.DeviceTypeCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.ProtocolTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "ProtocolCommand",
				Title: "ProtocolCommand",
			},
			MessageSample: new(publisher.ProtocolCommand),
		},
	}))

	mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
		Name: conf.DeviceTopic,
		Subscribe: &asyncapi.MessageSample{
			MessageEntity: spec.MessageEntity{
				Name:  "DeviceCommand",
				Title: "DeviceCommand",
			},
			MessageSample: new(publisher.DeviceCommand),
		},
	}))

	buff, err := reflector.Schema.MarshalJSON()
	mustNotFail(err)

	fmt.Println(string(buff))
	mustNotFail(os.WriteFile("asyncapi.json", buff, 0o600))
}
