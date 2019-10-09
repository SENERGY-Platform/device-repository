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

/*
this file contains code needed to create the test environment
*/

package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/device-repository/lib/config"
	"github.com/SENERGY-Platform/device-repository/lib/model"
	jwt_http_router "github.com/SmartEnergyPlatform/jwt-http-router"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/segmentio/kafka-go"
	"github.com/wvanbergen/kazoo-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

const userjwt = jwt_http_router.JwtImpersonate("Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ")
const userid = "dd69ea0d-f553-4336-80f3-7f4567f85c7b"

func jwtdelete(jwt jwt_http_router.JwtImpersonate, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", string(jwt))
	resp, err = http.DefaultClient.Do(req)
	return
}

func jwtput(jwt jwt_http_router.JwtImpersonate, url string, contenttype string, body *bytes.Buffer) (resp *http.Response, err error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", string(jwt))
	req.Header.Set("Content-Type", contenttype)
	resp, err = http.DefaultClient.Do(req)
	return
}

func createTestEnv() (closer func(), conf config.Config, err error) {
	conf, err = config.Load("../config.json")
	if err != nil {
		log.Println("ERROR: unable to load config: ", err)
		return func() {}, conf, err
	}
	conf.MongoReplSet = false
	conf, closer, err = NewDockerEnv(conf)
	if err != nil {
		log.Println("ERROR: unable to create docker env", err)
		return func() {}, conf, err
	}
	time.Sleep(1 * time.Second)
	stop, err := Start(conf)
	if err != nil {
		log.Println("ERROR: unable to connect to database", err)
		closer()
		return closer, conf, err
	}
	dockerCloser := closer
	closer = func() {
		stop()
		dockerCloser()
	}
	time.Sleep(10 * time.Second)
	return closer, conf, err
}

func NewDockerEnv(startConfig config.Config) (config config.Config, shutdown func(), err error) {
	config = startConfig

	whPort, err := getFreePort()
	if err != nil {
		log.Println("unable to find free port", err)
		return config, func() {}, err
	}
	config.ServerPort = strconv.Itoa(whPort)

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Println("Could not connect to docker:", err)
		return config, func() {}, err
	}

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

	//mongo
	wait.Add(1)
	go func() {
		defer wait.Done()
		closer, _, ip, err := MongoTestServer(pool)
		listMux.Lock()
		closerList = append(closerList, closer)
		listMux.Unlock()
		if err != nil {
			globalError = err
			return
		}
		config.MongoUrl = "mongodb://" + ip + ":27017"
	}()

	wait.Add(1)
	go func() {
		defer wait.Done()

		var wait2 sync.WaitGroup

		var elasticIp string

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

		wait2.Wait()

		if globalError != nil {
			return
		}

		closeZk, _, zkIp, err := Zookeeper(pool)
		listMux.Lock()
		closerList = append(closerList, closeZk)
		listMux.Unlock()
		if err != nil {
			globalError = err
			return
		}
		config.ZookeeperUrl = zkIp + ":2181"

		//kafka
		closeKafka, err := Kafka(pool, config.ZookeeperUrl)
		listMux.Lock()
		closerList = append(closerList, closeKafka)
		listMux.Unlock()
		if err != nil {
			globalError = err
			return
		}

		//permsearch
		closePerm, _, permIp, err := PermSearch(pool, config.ZookeeperUrl, elasticIp)
		listMux.Lock()
		closerList = append(closerList, closePerm)
		listMux.Unlock()
		if err != nil {
			globalError = err
			return
		}

		config.PermissionsUrl = "http://" + permIp + ":8080"
	}()

	wait.Wait()
	if globalError != nil {
		close(closerList)
		return config, shutdown, globalError
	}

	return config, func() { close(closerList) }, nil
}

func PermSearch(pool *dockertest.Pool, zk string, elasticIp string) (closer func(), hostPort string, ipAddress string, err error) {
	log.Println("start permsearch")
	repo, err := pool.Run("fgseitsrancher.wifa.intern.uni-leipzig.de:5000/permissionsearch", "test", []string{
		"ZOOKEEPER_URL=" + zk,
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
		return err
	})
	if err != nil {
		log.Println(err)
	}
	return func() { repo.Close() }, hostPort, repo.Container.NetworkSettings.IPAddress, err
}

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

func Kafka(pool *dockertest.Pool, zookeeperUrl string) (closer func(), err error) {
	kafkaport, err := getFreePort()
	if err != nil {
		log.Fatalf("Could not find new port: %s", err)
	}
	networks, _ := pool.Client.ListNetworks()
	hostIp := ""
	for _, network := range networks {
		if network.Name == "bridge" {
			hostIp = network.IPAM.Config[0].Gateway
		}
	}
	log.Println("host ip: ", hostIp)
	env := []string{
		"ALLOW_PLAINTEXT_LISTENER=yes",
		"KAFKA_LISTENERS=OUTSIDE://:9092",
		"KAFKA_ADVERTISED_LISTENERS=OUTSIDE://" + hostIp + ":" + strconv.Itoa(kafkaport),
		"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=OUTSIDE:PLAINTEXT",
		"KAFKA_INTER_BROKER_LISTENER_NAME=OUTSIDE",
		"KAFKA_ZOOKEEPER_CONNECT=" + zookeeperUrl,
	}
	log.Println("start kafka with env ", env)
	kafkaContainer, err := pool.RunWithOptions(&dockertest.RunOptions{Repository: "bitnami/kafka", Tag: "latest", Env: env, PortBindings: map[docker.Port][]docker.PortBinding{
		"9092/tcp": {{HostIP: "", HostPort: strconv.Itoa(kafkaport)}},
	}})
	if err != nil {
		return func() {}, err
	}
	err = pool.Retry(func() error {
		log.Println("try kafka connection...")
		conn, err := kafka.Dial("tcp", hostIp+":"+strconv.Itoa(kafkaport))
		if err != nil {
			log.Println(err)
			return err
		}
		defer conn.Close()
		return nil
	})
	return func() { kafkaContainer.Close() }, err
}

func Zookeeper(pool *dockertest.Pool) (closer func(), hostPort string, ipAddress string, err error) {
	zkport, err := getFreePort()
	if err != nil {
		log.Fatalf("Could not find new port: %s", err)
	}
	env := []string{}
	log.Println("start zookeeper on ", zkport)
	zkContainer, err := pool.RunWithOptions(&dockertest.RunOptions{Repository: "wurstmeister/zookeeper", Tag: "latest", Env: env, PortBindings: map[docker.Port][]docker.PortBinding{
		"2181/tcp": {{HostIP: "", HostPort: strconv.Itoa(zkport)}},
	}})
	if err != nil {
		return func() {}, "", "", err
	}
	hostPort = strconv.Itoa(zkport)
	err = pool.Retry(func() error {
		log.Println("try zk connection...")
		zookeeper := kazoo.NewConfig()
		zk, chroot := kazoo.ParseConnectionString(zkContainer.Container.NetworkSettings.IPAddress)
		zookeeper.Chroot = chroot
		kz, err := kazoo.NewKazoo(zk, zookeeper)
		if err != nil {
			log.Println("kazoo", err)
			return err
		}
		_, err = kz.Brokers()
		if err != nil && strings.TrimSpace(err.Error()) != strings.TrimSpace("zk: node does not exist") {
			log.Println("brokers", err)
			return err
		}
		return nil
	})
	return func() { zkContainer.Close() }, hostPort, zkContainer.Container.NetworkSettings.IPAddress, err
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func head(url string, token string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		log.Println("ERROR: request", err)
		return
	}
	req.Header.Set("Authorization", token)
	resp, err = http.DefaultClient.Do(req)
	return
}

type Publisher struct {
	config      config.Config
	devicetypes *kafka.Writer
	protocols   *kafka.Writer
	devices     *kafka.Writer
}

func NewPublisher(conf config.Config) (*Publisher, error) {
	broker, err := GetBroker(conf.ZookeeperUrl)
	if err != nil {
		return nil, err
	}
	if len(broker) == 0 {
		return nil, errors.New("missing kafka broker")
	}
	devicetypes, err := getProducer(broker, conf.DeviceTypeTopic, conf.LogLevel == "DEBUG")
	if err != nil {
		return nil, err
	}
	devices, err := getProducer(broker, conf.DeviceTopic, conf.LogLevel == "DEBUG")
	if err != nil {
		return nil, err
	}
	protocols, err := getProducer(broker, conf.ProtocolTopic, conf.LogLevel == "DEBUG")
	if err != nil {
		return nil, err
	}
	return &Publisher{config: conf, devicetypes: devicetypes, protocols: protocols, devices: devices}, nil
}

func getProducer(broker []string, topic string, debug bool) (writer *kafka.Writer, err error) {
	var logger *log.Logger
	if debug {
		logger = log.New(os.Stdout, "[KAFKA-PRODUCER] ", 0)
	} else {
		logger = log.New(ioutil.Discard, "", 0)
	}
	writer = kafka.NewWriter(kafka.WriterConfig{
		Brokers:     broker,
		Topic:       topic,
		MaxAttempts: 10,
		Logger:      logger,
	})
	return writer, err
}

func GetBroker(zk string) (brokers []string, err error) {
	return getBroker(zk)
}

func getBroker(zkUrl string) (brokers []string, err error) {
	zookeeper := kazoo.NewConfig()
	zookeeper.Logger = log.New(ioutil.Discard, "", 0)
	zk, chroot := kazoo.ParseConnectionString(zkUrl)
	zookeeper.Chroot = chroot
	if kz, err := kazoo.NewKazoo(zk, zookeeper); err != nil {
		return brokers, err
	} else {
		return kz.BrokerList()
	}
}

func GetKafkaController(zkUrl string) (controller string, err error) {
	zookeeper := kazoo.NewConfig()
	zookeeper.Logger = log.New(ioutil.Discard, "", 0)
	zk, chroot := kazoo.ParseConnectionString(zkUrl)
	zookeeper.Chroot = chroot
	kz, err := kazoo.NewKazoo(zk, zookeeper)
	if err != nil {
		return controller, err
	}
	controllerId, err := kz.Controller()
	if err != nil {
		return controller, err
	}
	brokers, err := kz.Brokers()
	if err != nil {
		return controller, err
	}
	return brokers[controllerId], err
}

func InitTopic(zkUrl string, topics ...string) (err error) {
	return InitTopicWithConfig(zkUrl, 1, 1, topics...)
}

func InitTopicWithConfig(zkUrl string, numPartitions int, replicationFactor int, topics ...string) (err error) {
	controller, err := GetKafkaController(zkUrl)
	if err != nil {
		log.Println("ERROR: unable to find controller", err)
		return err
	}
	if controller == "" {
		log.Println("ERROR: unable to find controller")
		return errors.New("unable to find controller")
	}
	initConn, err := kafka.Dial("tcp", controller)
	if err != nil {
		log.Println("ERROR: while init topic connection ", err)
		return err
	}
	defer initConn.Close()
	for _, topic := range topics {
		err = initConn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
			ConfigEntries: []kafka.ConfigEntry{
				{ConfigName: "cleanup.policy", ConfigValue: "compact"},
				{ConfigName: "delete.retention.ms", ConfigValue: "100"},
				{ConfigName: "segment.ms", ConfigValue: "100"},
				{ConfigName: "min.cleanable.dirty.ratio", ConfigValue: "0.01"},
			},
		})
		if err != nil {
			return
		}
	}
	return nil
}

type DeviceTypeCommand struct {
	Command    string           `json:"command"`
	Id         string           `json:"id"`
	Owner      string           `json:"owner"`
	DeviceType model.DeviceType `json:"device_type"`
}

type DeviceCommand struct {
	Command string       `json:"command"`
	Id      string       `json:"id"`
	Owner   string       `json:"owner"`
	Device  model.Device `json:"device"`
}

func (this *Publisher) PublishDeviceType(device model.DeviceType, userId string) (err error) {
	cmd := DeviceTypeCommand{Command: "PUT", Id: device.Id, DeviceType: device, Owner: userId}
	return this.PublishDeviceTypeCommand(cmd)
}

func (this *Publisher) PublishDeviceTypeDelete(id string, userId string) error {
	cmd := DeviceTypeCommand{Command: "DELETE", Id: id, Owner: userId}
	return this.PublishDeviceTypeCommand(cmd)
}

func (this *Publisher) PublishDeviceTypeCommand(cmd DeviceTypeCommand) error {
	if this.config.LogLevel == "DEBUG" {
		log.Println("DEBUG: produce", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.devicetypes.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(cmd.Id),
			Value: message,
			Time:  time.Now(),
		},
	)
	if err != nil {
		debug.PrintStack()
	}
	return err
}

func (this *Publisher) PublishDevice(device model.Device, userId string) (err error) {
	cmd := DeviceCommand{Command: "PUT", Id: device.Id, Device: device, Owner: userId}
	return this.PublishDeviceCommand(cmd)
}

func (this *Publisher) PublishDeviceCommand(cmd DeviceCommand) error {
	if this.config.LogLevel == "DEBUG" {
		log.Println("DEBUG: produce", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.devices.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(cmd.Id),
			Value: message,
			Time:  time.Now(),
		},
	)
	if err != nil {
		debug.PrintStack()
	}
	return err
}

type ProtocolCommand struct {
	Command  string         `json:"command"`
	Id       string         `json:"id"`
	Owner    string         `json:"owner"`
	Protocol model.Protocol `json:"protocol"`
}

func (this *Publisher) PublishProtocol(device model.Protocol, userId string) (err error) {
	cmd := ProtocolCommand{Command: "PUT", Id: device.Id, Protocol: device, Owner: userId}
	return this.PublishProtocolCommand(cmd)
}

func (this *Publisher) PublishProtocolDelete(id string, userId string) error {
	cmd := ProtocolCommand{Command: "DELETE", Id: id, Owner: userId}
	return this.PublishProtocolCommand(cmd)
}

func (this *Publisher) PublishProtocolCommand(cmd ProtocolCommand) error {
	if this.config.LogLevel == "DEBUG" {
		log.Println("DEBUG: produce", cmd)
	}
	message, err := json.Marshal(cmd)
	if err != nil {
		debug.PrintStack()
		return err
	}
	err = this.protocols.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   []byte(cmd.Id),
			Value: message,
			Time:  time.Now(),
		},
	)
	if err != nil {
		debug.PrintStack()
	}
	return err
}
