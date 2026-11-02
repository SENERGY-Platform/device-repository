/*
 * Copyright 2026 InfAI (CC SES)
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

package tests

import (
	"context"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/api"
	"github.com/SENERGY-Platform/device-repository/v3/lib/client"
	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/controller"
	"github.com/SENERGY-Platform/device-repository/v3/lib/controller/publisher"
	"github.com/SENERGY-Platform/device-repository/v3/lib/database/testdb"
	permclient "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// tracingPermissions records the trace of every context that reaches the permissions client.
type tracingPermissions struct {
	permclient.Client
	mux      sync.Mutex
	traceIds []trace.TraceID
}

func (this *tracingPermissions) ListAccessibleResourceIdsContext(ctx context.Context, token string, topicId string, options permclient.ListOptions, permissions ...permclient.Permission) (ids []string, err error, code int) {
	this.mux.Lock()
	this.traceIds = append(this.traceIds, trace.SpanContextFromContext(ctx).TraceID())
	this.mux.Unlock()
	return this.Client.ListAccessibleResourceIdsContext(ctx, token, topicId, options, permissions...)
}

func (this *tracingPermissions) seen(traceId trace.TraceID) bool {
	this.mux.Lock()
	defer this.mux.Unlock()
	return slices.Contains(this.traceIds, traceId)
}

func TestTracePropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("./../../config.json")
	if err != nil {
		t.Fatal(err)
	}
	permClient, err := permclient.NewTestClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	permissions := &tracingPermissions{Client: permClient}
	ctrl, err := controller.New(config, testdb.NewTestDB(config), publisher.Void{}, permissions)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(api.GetRouter(config, ctrl))
	defer server.Close()

	//GetRouter initializes the global OpenTelemetry setup with an OTLP exporter, the recorder replaces it afterwards
	previousProvider, previousPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	callerCtx, span := provider.Tracer("test").Start(ctx, "caller")
	_, err, _ = client.NewClient(server.URL, nil).ListDevices(callerCtx, userToken, client.DeviceListOptions{Limit: 10})
	span.End()
	if err != nil {
		t.Fatal(err)
	}
	traceId := span.SpanContext().TraceID()

	t.Run("client request continues the trace of the caller on the server", func(t *testing.T) {
		//the server ends its span after the response is written, so it may arrive after the client returned
		deadline := time.Now().Add(2 * time.Second)
		for {
			for _, s := range recorder.Ended() {
				if s.SpanKind() == trace.SpanKindServer && s.SpanContext().TraceID() == traceId && s.Parent().SpanID() == span.SpanContext().SpanID() {
					return
				}
			}
			if time.Now().After(deadline) {
				t.Fatal("no server span with the caller span as parent")
			}
			time.Sleep(10 * time.Millisecond)
		}
	})

	t.Run("controller passes the trace on to the permissions client", func(t *testing.T) {
		if !permissions.seen(traceId) {
			t.Error("permissions client was called without the trace of the request")
		}
	})
}
