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

package util

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/golang-jwt/jwt"
)

type MirrorPull interface {
	MirrorUpdate() error
}

func NewMirrorMiddleware(handler http.Handler, config configuration.Config, pull MirrorPull) *MirrorMiddleware {
	return &MirrorMiddleware{handler: handler, config: config, pull: pull}
}

type MirrorMiddleware struct {
	handler http.Handler
	config  configuration.Config
	token   string
	pull    MirrorPull
}

func (this *MirrorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.URL.String(), "query") && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete) {
		//forward request to source
		this.config.GetLogger().Info("forward update request to mirror source", "method", r.Method, "url", r.URL.String())
		endpoint := r.URL.Path
		if len(r.URL.Query()) > 0 {
			endpoint += "?" + r.URL.Query().Encode()
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req, err := http.NewRequest(r.Method, this.config.MgwMirrorSourceUrl+endpoint, strings.NewReader(string(body)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp.StatusCode < 300 {
			err = this.pull.MirrorUpdate()
			if err != nil {
				this.config.GetLogger().Error("unable to update mirror", "error", err)
			}
		} else {
			this.config.GetLogger().Error("forwarded request returned unexpected status-code", "status", resp.StatusCode)
		}

		defer resp.Body.Close()
		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			this.config.GetLogger().Error("unable to copy response body", "error", err)
		}
	} else {
		token, err := this.GetToken()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Header.Set("Authorization", token)
		this.handler.ServeHTTP(w, r)
	}
}

func (this *MirrorMiddleware) GetToken() (result string, err error) {
	if this.token != "" {
		return this.token, nil
	}
	userId, err := this.config.GetMgwMirrorUserId()
	if err != nil {
		return "", err
	}
	this.token, err = generateUserTokenById(userId)
	return this.token, err
}

func generateUserTokenById(userid string) (token string, err error) {
	claims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Issuer:    "device-repository-mirror",
		Subject:   userid,
	}
	jwtoken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedTokenString, err := jwtoken.SigningString()
	if err != nil {
		slog.Error("unable to generate user token", "error", err)
		return token, err
	}
	return fmt.Sprintf("Bearer %s.", signedTokenString), nil
}
