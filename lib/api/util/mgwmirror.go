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
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/SENERGY-Platform/device-repository/lib/configuration"
	"github.com/golang-jwt/jwt"
)

func NewMirrorMiddleware(handler http.Handler, config configuration.Config) *MirrorMiddleware {
	return &MirrorMiddleware{handler: handler, config: config}
}

type MirrorMiddleware struct {
	handler http.Handler
	config  configuration.Config
	token   string
}

func (this *MirrorMiddleware) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if !strings.Contains(req.URL.String(), "query") && (req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodDelete) {
		http.Redirect(res, req, "mirror does not allow changes to the dataset", http.StatusBadRequest)
	} else {
		token, err := this.GetToken()
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", token)
		this.handler.ServeHTTP(res, req)
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
