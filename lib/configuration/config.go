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

package configuration

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerPort                             string `json:"server_port"`
	EnableSwaggerUi                        bool   `json:"enable_swagger_ui"`
	KafkaUrl                               string `json:"kafka_url"`
	DeviceTopic                            string `json:"device_topic"`
	DeviceTypeTopic                        string `json:"device_type_topic"`
	DeviceGroupTopic                       string `json:"device_group_topic"`
	HubTopic                               string `json:"hub_topic"`
	ProtocolTopic                          string `json:"protocol_topic"`
	ConceptTopic                           string `json:"concept_topic"`
	CharacteristicTopic                    string `json:"characteristic_topic"`
	AspectTopic                            string `json:"aspect_topic"`
	FunctionTopic                          string `json:"function_topic"`
	DeviceClassTopic                       string `json:"device_class_topic"`
	LocationTopic                          string `json:"location_topic"`
	PermissionsV2Url                       string `json:"permissions_v2_url"`
	MongoUrl                               string `json:"mongo_url"`
	MongoTable                             string `json:"mongo_table"`
	MongoRightsCollection                  string `json:"mongo_rights_collection"`
	MongoDeviceCollection                  string `json:"mongo_device_collection"`
	MongoDeviceTypeCollection              string `json:"mongo_device_type_collection"`
	MongoDeviceGroupCollection             string `json:"mongo_device_group_collection"`
	MongoProtocolCollection                string `json:"mongo_protocol_collection"`
	MongoHubCollection                     string `json:"mongo_hub_collection"`
	MongoAspectCollection                  string `json:"mongo_aspect_collection"`
	MongoCharacteristicCollection          string `json:"mongo_characteristic_collection"`
	MongoConceptCollection                 string `json:"mongo_concept_collection"`
	MongoDeviceClassCollection             string `json:"mongo_device_class_collection"`
	MongoFunctionCollection                string `json:"mongo_function_collection"`
	MongoLocationCollection                string `json:"mongo_location_collection"`
	MongoDefaultDeviceAttributesCollection string `json:"mongo_default_device_attributes_collection"`
	Debug                                  bool   `json:"debug"`
	HttpClientTimeout                      string `json:"http_client_timeout"`

	DeviceServiceGroupSelectionAllowNotFound     bool `json:"device_service_group_selection_allow_not_found"`
	AllowNoneLeafAspectNodesInDeviceTypesDefault bool `json:"allow_none_leaf_aspect_nodes_in_device_types_default"`

	InitialGroupRights   map[string]map[string]string `json:"initial_group_rights"`
	RunStartupMigrations bool                         `json:"run_startup_migrations"`

	LocalIdUniqueForOwner               bool `json:"local_id_unique_for_owner"`
	SkipDeviceGroupGenerationFromDevice bool `json:"skip_device_group_generation_from_device"`

	PreventEmptyCriteriaListsAllBehavior bool `json:"prevent_empty_criteria_lists_all_behavior"`

	SyncInterval     string `json:"sync_interval"`
	SyncLockDuration string `json:"sync_lock_duration"`

	DisableStrictValidationForTesting bool `json:"disable_strict_validation_for_testing"` //only for tests; disables validations and id generations

	StructLoggerLogLevel   string `json:"struct_logger_log_level"`
	StructLoggerLogHandler string `json:"struct_logger_log_handler"`

	ApiDocsProviderBaseUrl string `json:"api_docs_provider_base_url"`
}

// loads config from json in location and used environment variables (e.g ZookeeperUrl --> ZOOKEEPER_URL)
func Load(location string) (config Config, err error) {
	file, err := os.Open(location)
	if err != nil {
		log.Println("error on config load: ", err)
		return config, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Println("invalid config json: ", err)
		return config, err
	}
	handleEnvironmentVars(&config)
	setDefaultHttpClient(config)

	return config, nil
}

var camel = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

func fieldNameToEnvName(s string) string {
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToUpper(strings.Join(a, "_"))
}

// preparations for docker
func handleEnvironmentVars(config *Config) {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	configType := configValue.Type()
	for index := 0; index < configType.NumField(); index++ {
		fieldName := configType.Field(index).Name
		envName := fieldNameToEnvName(fieldName)
		envValue := os.Getenv(envName)
		if envValue != "" {
			fmt.Println("use environment variable: ", envName, " = ", envValue)
			if configValue.FieldByName(fieldName).Kind() == reflect.Int64 {
				i, _ := strconv.ParseInt(envValue, 10, 64)
				configValue.FieldByName(fieldName).SetInt(i)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.String {
				configValue.FieldByName(fieldName).SetString(envValue)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Bool {
				b, _ := strconv.ParseBool(envValue)
				configValue.FieldByName(fieldName).SetBool(b)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Float64 {
				f, _ := strconv.ParseFloat(envValue, 64)
				configValue.FieldByName(fieldName).SetFloat(f)
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Slice {
				val := []string{}
				for _, element := range strings.Split(envValue, ",") {
					val = append(val, strings.TrimSpace(element))
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(val))
			}
			if configValue.FieldByName(fieldName).Kind() == reflect.Map {
				value := map[string]string{}
				for _, element := range strings.Split(envValue, ",") {
					keyVal := strings.Split(element, ":")
					key := strings.TrimSpace(keyVal[0])
					val := strings.TrimSpace(keyVal[1])
					value[key] = val
				}
				configValue.FieldByName(fieldName).Set(reflect.ValueOf(value))
			}
		}
	}
}

func setDefaultHttpClient(config Config) {
	var err error
	http.DefaultClient.Timeout, err = time.ParseDuration(config.HttpClientTimeout)
	if err != nil {
		log.Println("WARNING: invalid http timeout --> no timeouts\n", err)
	}
}
