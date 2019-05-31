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

package main

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/database"
	"github.com/SENERGY-Platform/device-repository/lib/database/listoptions"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	"github.com/SENERGY-Platform/iot-device-repository/lib/persistence/ordf"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"github.com/ory/dockertest"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

const userjwt = jwt_http_router.JwtImpersonate("Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ")

func TestMigrateFlags(t *testing.T) {
	//create source for migration

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal("Could not connect to docker:", err)
	}

	iotUrl, ontoUrl, iotClose, err := IotDependencies(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer iotClose()

	err = fillSourceDb(iotUrl)
	if err != nil {
		t.Fatal(err)
	}

	//create sink for migration

	closer, port, _, err := MongoTestServer(pool)
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	//migrate by calling main() with migrate flag

	os.Setenv("MONGO_URL", "mongodb://localhost:"+port)
	cmd := os.Args[0]
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{cmd, "-migrate", ontoUrl, "iot", "dba", "myDbaPassword"}

	main()

	//connect to sink db to test migration success

	conf, err := config.Load("config.json")
	if err != nil {
		t.Fatal("ERROR: unable to load config", err)
	}

	db, err := database.New(conf)
	if err != nil {
		log.Fatal("ERROR: unable to connect to database", err)
	}

	//tests

	t.Run("valuetypes", func(t *testing.T) {
		checkValueTypes(t, db)
	})
	t.Run("devicetypes", func(t *testing.T) {
		checkDeviceTypes(t, db)
	})
	t.Run("device", func(t *testing.T) {
		checkDevices(t, db)
	})
	t.Run("endpoint", func(t *testing.T) {
		checkEndpoints(t, db)
	})
	t.Run("hub", func(t *testing.T) {
		checkHubs(t, db)
	})
}

func checkValueTypes(t *testing.T, db database.Database) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	dt, exists, err := db.GetDeviceType(ctx, "iot#f8b43fd0-6318-4cca-82d3-71eb8e6fce79")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("unexpected result dt", dt)
	}

	result, exists, err := db.GetValueType(ctx, dt.Services[0].Input[0].Type.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !exists || result.Name != "test" {
		t.Fatal("unexpected result", result)
	}

	result, exists, err = db.GetValueType(ctx, dt.Services[0].Input[0].Type.Fields[0].Type.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !exists || result.Name != "test_int" {
		t.Fatal("unexpected result", result)
	}
}

func checkDeviceTypes(t *testing.T, db database.Database) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, exists, err := db.GetDeviceType(ctx, "iot#f8b43fd0-6318-4cca-82d3-71eb8e6fce79")
	if err != nil {
		t.Fatal(err)
	}
	if !exists || result.Name != "test" || result.Services[0].Name != "test" {
		t.Fatal("unexpected result", result)
	}
}

func checkDevices(t *testing.T, db database.Database) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := db.ListDevicesOfDeviceType(ctx, "iot#f8b43fd0-6318-4cca-82d3-71eb8e6fce79", listoptions.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Name != "test" || result[0].Url != "uri_1" {
		t.Fatal("unexpected result", len(result), result)
	}
}

func checkEndpoints(t *testing.T, db database.Database) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := db.ListEndpoints(ctx, listoptions.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Service != "iot#bd936af1-ad93-4dc9-b310-bf93264de0eb" || result[0].ProtocolHandler != "connector" || result[0].Endpoint != "uri_1/test" {
		t.Fatal("unexpected result", len(result), result)
	}
}

func checkHubs(t *testing.T, db database.Database) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, err := db.ListHubs(ctx, listoptions.New())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Name != "testhub" || !reflect.DeepEqual(result[0].Devices, []string{"uri_1"}) {
		t.Fatal("unexpected result", len(result), result)
	}
}

func fillSourceDb(url string) (err error) {
	dt := model.DeviceType{}
	json.Unmarshal([]byte(dtString), &dt)

	err = userjwt.PostJSON(url+"/import/deviceType", dt, nil)
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	instance_1 := model.DeviceInstance{}
	err = userjwt.PostJSON(url+"/deviceInstance", model.DeviceInstance{Name: "test", Url: "uri_1", DeviceType: "iot#f8b43fd0-6318-4cca-82d3-71eb8e6fce79"}, &instance_1)
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	hub := model.Hub{}
	err = userjwt.PostJSON(url+"/hubs", model.Hub{Name: "testhub", Devices: []string{"uri_1"}, Hash: "hash_1"}, &hub)
	if err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	return nil
}

const dtString = `{  
   "id":"iot#f8b43fd0-6318-4cca-82d3-71eb8e6fce79",
   "name":"test",
   "description":"test",
   "device_class":{  
      "id":"iot#3e522022-38ee-4a8b-b5c7-dbcb54b887d1",
      "name":"test"
   },
   "services":[  
      {  
         "id":"iot#bd936af1-ad93-4dc9-b310-bf93264de0eb",
         "service_type":"http://www.sepl.wifa.uni-leipzig.de/ontlogies/device-repo#Actuator",
         "name":"test",
         "description":"test",
		 "endpoint_format": "{{device_uri}}/{{service_uri}}",
         "protocol":{  
            "id":"iot#d6a462c5-d4e0-4396-b3f3-28cd37b647a8",
            "protocol_handler_url":"connector",
            "name":"standard-connector",
            "description":"Generic protocol for transporting data and metadata.",
            "msg_structure":[  
               {  
                  "id":"iot#37ff5298-a7dd-4744-9080-7cfdbda5dc72",
                  "name":"metadata",
                  "constraints":null
               },
               {  
                  "id":"iot#88cd5b0e-a451-4070-a20d-464ee23742dd",
                  "name":"data",
                  "constraints":null
               }
            ]
         },
         "input":[  
            {  
               "id":"iot#7398cf74-2194-4399-841b-cf401dc8a67e",
               "name":"test",
               "msg_segment":{  
                  "id":"iot#88cd5b0e-a451-4070-a20d-464ee23742dd",
                  "name":"data",
                  "constraints":null
               },
               "type":{  
                  "id":"iot#e69373a9-2ab9-4dc4-b5d5-ff57aa742c3e",
                  "name":"test",
                  "description":"test",
                  "base_type":"http://www.sepl.wifa.uni-leipzig.de/ontlogies/device-repo#structure",
                  "fields":[  
                     {  
                        "id":"iot#70908900-4acb-4b94-91ff-1c05b4f23c77",
                        "name":"a",
                        "type":{  
                           "id":"iot#e9104f3f-ffe1-410a-befa-dd68f0677ec6",
                           "name":"test_int",
                           "description":"test_int",
                           "base_type":"http://www.w3.org/2001/XMLSchema#integer",
                           "fields":null,
                           "literal":""
                        }
                     }
                  ],
                  "literal":""
               },
               "format":"http://www.sepl.wifa.uni-leipzig.de/ontlogies/device-repo#json",
               "additional_formatinfo":[  
                  {  
                     "id":"iot#32f58890-8b8f-4e06-9082-e7848c116154",
                     "field":{  
                        "id":"iot#70908900-4acb-4b94-91ff-1c05b4f23c77",
                        "name":"a",
                        "type":{  
                           "id":"iot#e9104f3f-ffe1-410a-befa-dd68f0677ec6",
                           "name":"test_int",
                           "description":"test_int",
                           "base_type":"http://www.w3.org/2001/XMLSchema#integer",
                           "fields":null,
                           "literal":""
                        }
                     },
                     "format_flag":""
                  }
               ]
            }
         ],
         "url":"test"
      }
   ],
   "vendor":{  
      "id":"iot#91bff598-bd63-44ce-aa5e-f66e092b7279",
      "name":"test"
   }
}`

func MongoTestServer(pool *dockertest.Pool) (closer func(), hostPort string, ipAddress string, err error) {
	log.Println("start mongodb")
	repo, err := pool.Run("mongo", "4.1.11", []string{})
	if err != nil {
		return func() {}, "", "", err
	}
	hostPort = repo.GetPort("27017/tcp")
	err = pool.Retry(func() error {
		log.Println("try mongodb connection...")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:"+hostPort))
		err = client.Ping(ctx, readpref.Primary())
		return err
	})
	return func() { repo.Close() }, hostPort, repo.Container.NetworkSettings.IPAddress, err
}

func IotOntology(pool *dockertest.Pool) (closer func(), hostPort string, ipAddress string, err error) {
	log.Println("start iot ontology")
	onto, err := pool.Run("fgseitsrancher.wifa.intern.uni-leipzig.de:5000/iot-ontology", "unstable", []string{
		"DBA_PASSWORD=myDbaPassword",
		"DEFAULT_GRAPH=iot",
	})
	if err != nil {
		return func() {}, "", "", err
	}
	hostPort = onto.GetPort("8890/tcp")
	err = pool.Retry(func() error {
		log.Println("try onto connection...")
		db := ordf.Persistence{
			Endpoint:  "http://" + onto.Container.NetworkSettings.IPAddress + ":8890/sparql",
			Graph:     "iot",
			User:      "dba",
			Pw:        "myDbaPassword",
			SparqlLog: "false",
		}
		_, err := db.IdExists("something")
		if err != nil {
			log.Println(err)
		}
		return err
	})
	return func() { onto.Close() }, hostPort, onto.Container.NetworkSettings.IPAddress, err
}

func IotRepo(pool *dockertest.Pool, ontoIp string, amqpIp string, permsearchIp string) (closer func(), hostPort string, ipAddress string, err error) {
	log.Println("start iot repo")
	repo, err := pool.Run("fgseitsrancher.wifa.intern.uni-leipzig.de:5000/iot-device-repository", "unstable", []string{
		"SPARQL_ENDPOINT=" + "http://" + ontoIp + ":8890/sparql",
		"AMQP_URL=" + "amqp://guest:guest@" + amqpIp + ":5672/",
		"PERMISSIONS_URL=" + "http://" + permsearchIp + ":8080",
	})
	if err != nil {
		return func() {}, "", "", err
	}
	hostPort = repo.GetPort("8080/tcp")
	err = pool.Retry(func() error {
		log.Println("try repo connection...")
		_, err := http.Get("http://" + repo.Container.NetworkSettings.IPAddress + ":8080/deviceType/foo")
		if err != nil {
			log.Println(err)
		}
		return err
	})
	return func() { repo.Close() }, hostPort, repo.Container.NetworkSettings.IPAddress, err
}

func PermSearch(pool *dockertest.Pool, amqpIp string, elasticIp string) (closer func(), hostPort string, ipAddress string, err error) {
	log.Println("start permsearch")
	repo, err := pool.Run("fgseitsrancher.wifa.intern.uni-leipzig.de:5000/permissionsearch", "unstable", []string{
		"AMQP_URL=" + "amqp://guest:guest@" + amqpIp + ":5672/",
		"ELASTIC_URL=" + "http://" + elasticIp + ":9200",
	})
	if err != nil {
		return func() {}, "", "", err
	}
	hostPort = repo.GetPort("8080/tcp")
	err = pool.Retry(func() error {
		log.Println("try permsearch connection...")
		_, err := http.Get("http://" + repo.Container.NetworkSettings.IPAddress + ":8080/jwt/check/deviceinstance/foo/r/bool")
		if err != nil {
			log.Println(err)
		}
		return err
	})
	return func() { repo.Close() }, hostPort, repo.Container.NetworkSettings.IPAddress, err
}

func Elasticsearch(pool *dockertest.Pool) (closer func(), hostPort string, ipAddress string, err error) {
	log.Println("start elasticsearch")
	repo, err := pool.Run("docker.elastic.co/elasticsearch/elasticsearch", "6.4.3", []string{"discovery.type=single-node"})
	if err != nil {
		return func() {}, "", "", err
	}
	hostPort = repo.GetPort("9200/tcp")
	err = pool.Retry(func() error {
		log.Println("try elastic connection...")
		_, err := http.Get("http://" + repo.Container.NetworkSettings.IPAddress + ":9200/_cluster/health")
		if err != nil {
			log.Println(err)
		}
		return err
	})
	return func() { repo.Close() }, hostPort, repo.Container.NetworkSettings.IPAddress, err
}

func Amqp(pool *dockertest.Pool) (closer func(), hostPort string, ipAddress string, err error) {
	log.Println("start rabbitmq")
	rabbitmq, err := pool.Run("rabbitmq", "3-management", []string{})
	if err != nil {
		return func() {}, "", "", err
	}
	hostPort = rabbitmq.GetPort("5672/tcp")
	err = pool.Retry(func() error {
		log.Println("try amqp connection...")
		conn, err := amqp.Dial("amqp://guest:guest@" + rabbitmq.Container.NetworkSettings.IPAddress + ":5672/")
		if err != nil {
			return err
		}
		defer conn.Close()
		c, err := conn.Channel()
		defer c.Close()
		return err
	})
	return func() { rabbitmq.Close() }, hostPort, rabbitmq.Container.NetworkSettings.IPAddress, err
}

func IotDependencies(pool *dockertest.Pool) (iotUrl string, ontoUrl string, shutdown func(), err error) {
	var wait sync.WaitGroup

	listMux := sync.Mutex{}
	var globalError error
	closerList := []func(){}
	close := func(list []func()) {
		for i := len(list)/2 - 1; i >= 0; i-- {
			opp := len(list) - 1 - i
			list[i], list[opp] = list[opp], list[i]
		}
		for _, c := range list {
			if c != nil {
				c()
			}
		}
	}

	wait.Add(1)
	go func() {
		defer wait.Done()

		var wait2 sync.WaitGroup

		var elasticIp string
		var ontoIp string
		var amqpIp string

		wait2.Add(1)
		go func() {
			defer wait2.Done()
			//amqp
			closeAmqp, _, ip, err := Amqp(pool)
			amqpIp = ip
			listMux.Lock()
			closerList = append(closerList, closeAmqp)
			listMux.Unlock()
			if err != nil {
				globalError = err
				return
			}
		}()

		wait2.Add(1)
		go func() {
			defer wait2.Done()
			//elasticsearch
			closeElastic, _, ip, err := Elasticsearch(pool)
			elasticIp = ip
			listMux.Lock()
			closerList = append(closerList, closeElastic)
			listMux.Unlock()
			if err != nil {
				globalError = err
				return
			}
		}()

		wait2.Add(1)
		go func() {
			defer wait2.Done()
			//iot-onto
			closeOnto, _, ip, err := IotOntology(pool)
			ontoIp = ip
			listMux.Lock()
			closerList = append(closerList, closeOnto)
			listMux.Unlock()
			if err != nil {
				globalError = err
				return
			}
			ontoUrl = "http://" + ontoIp + ":8890/sparql"
		}()

		wait2.Wait()

		if globalError != nil {
			return
		}

		//permsearch
		closePerm, _, permIp, err := PermSearch(pool, amqpIp, elasticIp)
		listMux.Lock()
		closerList = append(closerList, closePerm)
		listMux.Unlock()
		if err != nil {
			globalError = err
			return
		}

		//iot-repo
		closeIot, _, iotIp, err := IotRepo(pool, ontoIp, amqpIp, permIp)
		listMux.Lock()
		closerList = append(closerList, closeIot)
		listMux.Unlock()
		if err != nil {
			globalError = err
			return
		}
		iotUrl = "http://" + iotIp + ":8080"
	}()

	wait.Wait()
	return iotUrl, ontoUrl, func() { close(closerList) }, nil
}
