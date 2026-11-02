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
	"encoding/base64"
	"encoding/json"
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
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// tracingPermissions records the trace and the baggage username of every context that reaches the permissions client.
type tracingPermissions struct {
	permclient.Client
	mux       sync.Mutex
	traceIds  []trace.TraceID
	usernames []string
}

func (this *tracingPermissions) ListAccessibleResourceIdsContext(ctx context.Context, token string, topicId string, options permclient.ListOptions, permissions ...permclient.Permission) (ids []string, err error, code int) {
	this.mux.Lock()
	this.traceIds = append(this.traceIds, trace.SpanContextFromContext(ctx).TraceID())
	this.usernames = append(this.usernames, baggage.FromContext(ctx).Member("username").Value())
	this.mux.Unlock()
	return this.Client.ListAccessibleResourceIdsContext(ctx, token, topicId, options, permissions...)
}

func (this *tracingPermissions) calls() int {
	this.mux.Lock()
	defer this.mux.Unlock()
	return len(this.traceIds)
}

func (this *tracingPermissions) lastUsername() string {
	this.mux.Lock()
	defer this.mux.Unlock()
	if len(this.usernames) == 0 {
		return ""
	}
	return this.usernames[len(this.usernames)-1]
}

func (this *tracingPermissions) seen(traceId trace.TraceID) bool {
	this.mux.Lock()
	defer this.mux.Unlock()
	return slices.Contains(this.traceIds, traceId)
}

// newTracingTestServer starts the router with all middlewares on a test db and a recording permissions client
func newTracingTestServer(t *testing.T, ctx context.Context) (*httptest.Server, *tracingPermissions) {
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
	t.Cleanup(server.Close)
	return server, permissions
}

// recordSpans replaces the global OpenTelemetry setup, which GetRouter initialized with an OTLP exporter,
// by an in-memory recorder until the end of the test
func recordSpans(t *testing.T) (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	previousProvider, previousPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(previousProvider)
		otel.SetTextMapPropagator(previousPropagator)
	})
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return provider, recorder
}

// awaitServerSpan waits for an ended server span that matches, because the server ends its span
// after the response is written, so it may arrive after the client returned
func awaitServerSpan(recorder *tracetest.SpanRecorder, matches func(sdktrace.ReadOnlySpan) bool) (sdktrace.ReadOnlySpan, bool) {
	deadline := time.Now().Add(2 * time.Second)
	for {
		for _, s := range recorder.Ended() {
			if s.SpanKind() == trace.SpanKindServer && matches(s) {
				return s, true
			}
		}
		if time.Now().After(deadline) {
			return nil, false
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func spanAttribute(span sdktrace.ReadOnlySpan, key string) string {
	for _, a := range span.Attributes() {
		if string(a.Key) == key {
			return a.Value.AsString()
		}
	}
	return ""
}

// unsignedToken builds a jwt the service parses without validation
func unsignedToken(claims map[string]interface{}) string {
	encode := func(v interface{}) string {
		b, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return "Bearer " + encode(map[string]string{"alg": "HS256", "typ": "JWT"}) + "." + encode(claims) + ".sig"
}

func TestUsernameOutsideTheBaggageFormatIsServed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server, permissions := newTracingTestServer(t, ctx)
	_, recorder := recordSpans(t)

	for _, username := range []string{"first last", "jürgen", "a,b", "alice"} {
		token := unsignedToken(map[string]interface{}{
			"sub":                "testOwner",
			"preferred_username": username,
			"realm_access":       map[string][]string{"roles": {"user"}},
		})
		before := permissions.calls()
		_, err, code := client.NewClient(server.URL, client.Options{}).ListDevices(ctx, token, client.DeviceListOptions{Limit: 10})
		if err != nil {
			t.Errorf("username %q: %v (code %v)", username, err, code)
			continue
		}
		if permissions.calls() == before {
			t.Errorf("username %q: the request did not reach the controller", username)
			continue
		}
		if got := permissions.lastUsername(); got != username {
			t.Errorf("username %q: the baggage reaching the permissions client carries %q", username, got)
		}
		span, found := awaitServerSpan(recorder, func(s sdktrace.ReadOnlySpan) bool {
			return spanAttribute(s, "username") == username
		})
		if !found {
			t.Errorf("username %q: no server span with the username as attribute", username)
			continue
		}
		if got := spanAttribute(span, "user_id"); got != "testOwner" {
			t.Errorf("username %q: server span has user_id %q", username, got)
		}
	}
}

func TestTracePropagation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server, permissions := newTracingTestServer(t, ctx)
	provider, recorder := recordSpans(t)

	callerCtx, span := provider.Tracer("test").Start(ctx, "caller")
	_, err, _ := client.NewClient(server.URL, client.Options{}).ListDevices(callerCtx, userToken, client.DeviceListOptions{Limit: 10})
	span.End()
	if err != nil {
		t.Fatal(err)
	}
	traceId := span.SpanContext().TraceID()

	t.Run("client request continues the trace of the caller on the server", func(t *testing.T) {
		_, found := awaitServerSpan(recorder, func(s sdktrace.ReadOnlySpan) bool {
			return s.SpanContext().TraceID() == traceId && s.Parent().SpanID() == span.SpanContext().SpanID()
		})
		if !found {
			t.Fatal("no server span with the caller span as parent")
		}
	})

	t.Run("controller passes the trace on to the permissions client", func(t *testing.T) {
		if !permissions.seen(traceId) {
			t.Error("permissions client was called without the trace of the request")
		}
	})
}
