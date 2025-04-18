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

package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/device-repository/lib/api/util"
	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"github.com/SENERGY-Platform/service-commons/pkg/accesslog"
	"log"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
)

type EndpointMethod = func(config configuration.Config, router *http.ServeMux, ctrl Controller)

var endpoints = []interface{}{} //list of objects with EndpointMethod

func Start(ctx context.Context, config configuration.Config, control Controller) (err error) {
	log.Println("start api")
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
	}()
	router := GetRouter(config, control)
	server := &http.Server{Addr: ":" + config.ServerPort, Handler: router}
	go func() {
		log.Println("listening on ", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			debug.PrintStack()
			log.Fatal("FATAL:", err)
		}
	}()
	go func() {
		<-ctx.Done()
		log.Println("api shutdown", server.Shutdown(context.Background()))
	}()
	return
}

// GetRouter doc
// @title         Device-Repository API
// @version       0.1
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath  /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func GetRouter(config configuration.Config, control Controller) http.Handler {
	handler := GetRouterWithoutMiddleware(config, control)
	log.Println("add permissions endpoints")
	permForward := client.EmbedPermissionsClientIntoRouter(client.New(config.PermissionsV2Url), handler, "/permissions/", func(method string, path string) bool {
		if method == http.MethodDelete {
			return false
		}
		if strings.Contains(path, "admin") {
			return false
		}
		if strings.Contains(path, "import") {
			return false
		}
		if strings.Contains(path, "export") {
			return false
		}
		return true
	})
	log.Println("add cors")
	corsHandler := util.NewCors(permForward)
	log.Println("add logging")
	logger := accesslog.New(corsHandler)
	return logger
}

func GetRouterWithoutMiddleware(config configuration.Config, command Controller) http.Handler {
	router := http.NewServeMux()
	log.Println("add heart beat endpoint")
	router.HandleFunc("GET /{$}", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	for _, e := range endpoints {
		for name, call := range getEndpointMethods(e) {
			log.Println("add endpoint " + name)
			call(config, router, command)
		}
	}
	return router
}

func getEndpointMethods(e interface{}) map[string]func(config configuration.Config, router *http.ServeMux, ctrl Controller) {
	result := map[string]EndpointMethod{}
	objRef := reflect.ValueOf(e)
	methodCount := objRef.NumMethod()
	for i := 0; i < methodCount; i++ {
		m := objRef.Method(i)
		f, ok := m.Interface().(EndpointMethod)
		if ok {
			name := getTypeName(objRef.Type()) + "::" + objRef.Type().Method(i).Name
			result[name] = f
		}
	}
	return result
}

func getTypeName(t reflect.Type) (res string) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
