/*
 * Copyright 2022 InfAI (CC SES)
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

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/api"
	permissions "github.com/SENERGY-Platform/permissions-v2/pkg/client"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type Interface interface {
	api.Controller
	GetPermissionsClient() permissions.Client
}

type Client struct {
	baseUrl string

	// optionalAuthTokenForApiGatewayRequest:
	//   - may be nil if used internally, without kong routing
	//   - is used for requests that internally don`t need an auth token but are forced to send one if the request is routed over the SENERGY-Platform api-gateway
	optionalAuthTokenForApiGatewayRequest func() (token string, err error)

	//shared by all copies from WithCache; read methods use it only if cacheTtl > 0
	cache    *cacheStore
	cacheTtl time.Duration
}

// NewClient creates a new client
// optionalAuthTokenForApiGatewayRequest:
//   - may be nil if used internally, without kong routing
//   - is used for requests that internally don`t need an auth token but are forced to send one if the request is routed over the SENERGY-Platform api-gateway
func NewClient(baseUrl string, optionalAuthTokenForApiGatewayRequest func() (token string, err error)) Interface {
	return &Client{baseUrl: baseUrl, optionalAuthTokenForApiGatewayRequest: optionalAuthTokenForApiGatewayRequest, cache: newCacheStore()}
}

func (c *Client) MirrorUpdate(ctx context.Context, uncheckedCollections ...string) error {
	return nil
}

// send adds the trace context of req to its headers, so the called service continues the trace of the caller.
// Without an OpenTelemetry setup in the calling service the global propagator is a no-op.
func send(req *http.Request) (*http.Response, error) {
	otel.GetTextMapPropagator().Inject(req.Context(), propagation.HeaderCarrier(req.Header))
	return http.DefaultClient.Do(req)
}

func do[T any](req *http.Request, optionalAuthTokenForApiGatewayRequest func() (token string, err error)) (result T, err error, code int) {
	if optionalAuthTokenForApiGatewayRequest != nil && req.Header.Get("Authorization") == "" {
		token, err := optionalAuthTokenForApiGatewayRequest()
		if err != nil {
			return result, err, http.StatusInternalServerError
		}
		req.Header.Set("Authorization", token)
	}
	resp, err := send(req)
	if err != nil {
		return result, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
		return result, fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp)), resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		_, _ = io.ReadAll(resp.Body) //ensure resp.Body is read to EOF
		return result, err, http.StatusInternalServerError
	}
	return
}

func doVoid(req *http.Request, optionalAuthTokenForApiGatewayRequest func() (token string, err error)) (err error, code int) {
	if optionalAuthTokenForApiGatewayRequest != nil && req.Header.Get("Authorization") == "" {
		token, err := optionalAuthTokenForApiGatewayRequest()
		if err != nil {
			return err, http.StatusInternalServerError
		}
		req.Header.Set("Authorization", token)
	}
	resp, err := send(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
		return fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp)), resp.StatusCode
	}
	return
}

func doWithTotalInResult[T any](req *http.Request, optionalAuthTokenForApiGatewayRequest func() (token string, err error)) (result T, total int64, err error, code int) {
	if optionalAuthTokenForApiGatewayRequest != nil && req.Header.Get("Authorization") == "" {
		token, err := optionalAuthTokenForApiGatewayRequest()
		if err != nil {
			return result, total, err, http.StatusInternalServerError
		}
		req.Header.Set("Authorization", token)
	}
	resp, err := send(req)
	if err != nil {
		return result, total, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body) //read error response end ensure that resp.Body is read to EOF
		return result, total, fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp)), resp.StatusCode
	}
	total, err = strconv.ParseInt(resp.Header.Get("X-Total-Count"), 10, 64)
	if err != nil {
		return result, total, fmt.Errorf("unable to read X-Total-Count header %w", err), http.StatusInternalServerError
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		_, _ = io.ReadAll(resp.Body) //ensure resp.Body is read to EOF
		return result, total, err, http.StatusInternalServerError
	}
	return
}

func (c *Client) GetPermissionsClient() permissions.Client {
	return permissions.New(c.baseUrl + "/permissions")
}

func (c *Client) validateWithToken(ctx context.Context, token string, path string, e interface{}) (err error, code int) {
	return c.validateWithTokenAndOptions(ctx, token, path, e, nil)
}

func (c *Client) validateWithTokenAndOptions(ctx context.Context, token string, path string, e interface{}, options url.Values) (err error, code int) {
	b, err := json.Marshal(e)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	optQuery := options.Encode()
	if optQuery != "" {
		optQuery = "&" + optQuery
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseUrl+path+"?dry-run=true", bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	resp, err := send(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if resp.StatusCode > 299 {
		temp, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("validation error: %v", string(temp)), resp.StatusCode
	}
	return nil, resp.StatusCode
}

func (c *Client) validate(ctx context.Context, path string, e interface{}) (err error, code int) {
	return c.validateWithOptions(ctx, path, e, nil)
}

func (c *Client) validateWithOptions(ctx context.Context, path string, e interface{}, options url.Values) (err error, code int) {
	if options == nil {
		//validate() passes nil; Set() on a nil url.Values panics
		options = url.Values{}
	}
	if c.optionalAuthTokenForApiGatewayRequest != nil {
		token, err := c.optionalAuthTokenForApiGatewayRequest()
		if err != nil {
			return err, http.StatusInternalServerError
		}
		return c.validateWithTokenAndOptions(ctx, token, path, e, options)
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	options.Set("dry-run", "true")
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseUrl+path+"?"+options.Encode(), bytes.NewBuffer(b))
	if err != nil {
		return err, http.StatusInternalServerError
	}
	resp, err := send(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp))
		return err, resp.StatusCode
	}
	return nil, resp.StatusCode
}

func (c *Client) validateDelete(ctx context.Context, path string) (err error, code int) {
	if c.optionalAuthTokenForApiGatewayRequest != nil {
		token, err := c.optionalAuthTokenForApiGatewayRequest()
		if err != nil {
			return err, http.StatusInternalServerError
		}
		return c.validateDeleteWithToken(ctx, token, path)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+path+"?dry-run=true", nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	resp, err := send(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp))
		return err, resp.StatusCode
	}
	return nil, resp.StatusCode
}

func (c *Client) validateDeleteWithToken(ctx context.Context, token string, path string) (err error, code int) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+path+"?dry-run=true", nil)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	req.Header.Set("Authorization", token)
	resp, err := send(req)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if resp.StatusCode >= 300 {
		temp, _ := io.ReadAll(resp.Body)
		err = fmt.Errorf("unexpected statuscode %v: %v", resp.StatusCode, string(temp))
		return err, resp.StatusCode
	}
	return nil, resp.StatusCode
}
