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

package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/device-repository/v2/lib/configuration"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"slices"
	"strings"
	"time"
)

type Mongo struct {
	config configuration.Config
	client *mongo.Client
}

var CreateCollections = []func(db *Mongo) error{}

const startupCheckTimeout = 10 * time.Second

var (
	errEmptyDatabase   = errors.New("mongo database name must not be empty")
	errMissingPassword = errors.New("mongo password must not be empty when a mongo user is set")
)

func New(conf configuration.Config) (*Mongo, error) {
	if err := validateConfig(conf); err != nil {
		return nil, err
	}
	return start(conf, clientOptions(conf), startupCheckTimeout)
}

// start disconnects the client on every failure path, so a failed startup leaves nothing connected.
func start(conf configuration.Config, opts *options.ClientOptions, timeout time.Duration) (*Mongo, error) {
	c, err := connect(opts, conf.MongoDatabase, timeout)
	if err != nil {
		return nil, err
	}
	db := &Mongo{config: conf, client: c}
	for _, creators := range CreateCollections {
		err = creators(db)
		if err != nil {
			disconnect(c, timeout)
			return nil, err
		}
	}
	return db, nil
}

// connect runs listCollections on the service database because Connect is lazy and ping needs no
// authentication; unreachable servers and wrong or missing credentials then fail before the collection hooks.
func connect(opts *options.ClientOptions, database string, timeout time.Duration) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	c, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	listOpts := options.ListCollections().SetNameOnly(true).SetAuthorizedCollections(true)
	if _, err = c.Database(database).ListCollectionNames(ctx, bson.D{}, listOpts); err != nil {
		disconnect(c, timeout)
		return nil, fmt.Errorf("mongo startup check failed: %w", err)
	}
	return c, nil
}

func disconnect(c *mongo.Client, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_ = c.Disconnect(ctx)
}

func validateConfig(conf configuration.Config) error {
	if conf.MongoDatabase == "" {
		return errEmptyDatabase
	}
	if conf.MongoUser != "" && conf.MongoPassword == "" {
		return errMissingPassword
	}
	return nil
}

// clientOptions applies the credentials after the URI so they replace any given in MONGO_URL.
func clientOptions(conf configuration.Config) *options.ClientOptions {
	opts := options.Client().ApplyURI(conf.MongoUrl).SetReadConcern(readconcern.Majority())
	if conf.MongoUser != "" {
		opts.SetAuth(options.Credential{
			Username:   conf.MongoUser,
			Password:   conf.MongoPassword,
			AuthSource: conf.MongoAuthSource,
		})
	}
	return opts
}

func (this *Mongo) CreateId() string {
	return uuid.NewString()
}

func readCursorResult[T any](ctx context.Context, cursor *mongo.Cursor) (result []T, err error, code int) {
	err = cursor.All(ctx, &result)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	return result, nil, http.StatusOK
}

func (this *Mongo) removeIndex(collection *mongo.Collection, indexname string) error {
	_, err := collection.Indexes().DropOne(context.Background(), indexname)
	if err != nil {
		if strings.Contains(err.Error(), "IndexNotFound") {
			return nil
		} else {
			debug.PrintStack()
			return err
		}
	}
	return nil
}

func (this *Mongo) ensureIndex(collection *mongo.Collection, indexname string, indexKey string, asc bool, unique bool) error {
	ctx, _ := getTimeoutContext()
	var direction int32 = -1
	if asc {
		direction = 1
	}
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{indexKey, direction}},
		Options: options.Index().SetName(indexname).SetUnique(unique),
	})
	if err != nil {
		debug.PrintStack()
	}
	return err
}

func (this *Mongo) ensureCompoundIndex(collection *mongo.Collection, indexname string, asc bool, unique bool, indexKeys ...string) error {
	ctx, _ := getTimeoutContext()
	var direction int32 = -1
	if asc {
		direction = 1
	}
	keys := []bson.E{}
	for _, key := range indexKeys {
		keys = append(keys, bson.E{Key: key, Value: direction})
	}
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D(keys),
		Options: options.Index().SetName(indexname).SetUnique(unique),
	})
	return err
}

func (this *Mongo) Disconnect() {
	timeout, _ := context.WithTimeout(context.Background(), 10*time.Second)
	log.Println(this.client.Disconnect(timeout))
}

func getBsonFieldName(obj interface{}, fieldName string) (bsonName string, err error) {
	field, found := reflect.TypeOf(obj).FieldByName(fieldName)
	if !found {
		return "", errors.New("field '" + fieldName + "' not found")
	}
	tags, err := bsoncodec.DefaultStructTagParser.ParseStructTags(field)
	return tags.Name, err
}

func getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func getBsonFieldObject[T any]() T {
	v := new(T)
	err := fillObjectWithItsBsonFieldNames(v, nil, nil)
	if err != nil {
		panic(err)
	}
	return *v
}

func fillObjectWithItsBsonFieldNames(ptr interface{}, prefix []string, done []string) error {
	ptrval := reflect.ValueOf(ptr)
	objval := reflect.Indirect(ptrval)
	objecttype := objval.Type()
	objTypeStr := objecttype.Name()
	if slices.Contains(done, objTypeStr) {
		return nil
	}
	done = append(done, objTypeStr)
	for i := 0; i < objecttype.NumField(); i++ {
		field := objecttype.Field(i)
		if field.Type.Kind() == reflect.String {
			tags, err := bsoncodec.DefaultStructTagParser.ParseStructTags(field)
			if err != nil {
				return err
			}
			objval.Field(i).SetString(strings.Join(append(prefix, tags.Name), "."))
		}
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String {
			tags, err := bsoncodec.DefaultStructTagParser.ParseStructTags(field)
			if err != nil {
				return err
			}
			objval.Field(i).Set(reflect.ValueOf([]string{strings.Join(append(prefix, tags.Name), ".")}))
		}
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
			tags, err := bsoncodec.DefaultStructTagParser.ParseStructTags(field)
			if err != nil {
				return err
			}
			element := reflect.New(objval.Field(i).Type().Elem())
			if tags.Inline {
				err = fillObjectWithItsBsonFieldNames(element.Interface(), prefix, done)
			} else {
				err = fillObjectWithItsBsonFieldNames(element.Interface(), append(prefix, tags.Name), done)
			}
			if err != nil {
				return err
			}
			list := reflect.New(objval.Field(i).Type())
			list = reflect.Append(list.Elem(), element.Elem())
			objval.Field(i).Set(list)
		}
		if field.Type.Kind() == reflect.Struct {
			tags, err := bsoncodec.DefaultStructTagParser.ParseStructTags(field)
			if err != nil {
				return err
			}
			if tags.Inline {
				err = fillObjectWithItsBsonFieldNames(objval.Field(i).Addr().Interface(), prefix, done)
			} else {
				err = fillObjectWithItsBsonFieldNames(objval.Field(i).Addr().Interface(), append(prefix, tags.Name), done)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
