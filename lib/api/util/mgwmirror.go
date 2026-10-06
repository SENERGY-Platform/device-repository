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
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/golang-jwt/jwt"
)

type MirrorPull interface {
	MirrorUpdate() error
}

const defaultMissPullTimeout = 10 * time.Second
const defaultMissPullBackoff = 10 * time.Second
const defaultMissPullMaxBackoff = 5 * time.Minute

func NewMirrorMiddleware(handler http.Handler, config configuration.Config, pull MirrorPull) *MirrorMiddleware {
	//durations are validated by mgwmirror.StartSourcePullWorker()
	return &MirrorMiddleware{
		handler:         handler,
		config:          config,
		pull:            pull,
		missPullTimeout: durationOrDefault(config.MgwMirrorMissPullTimeout, defaultMissPullTimeout),
		missBackoff: newMissBackoff(
			durationOrDefault(config.MgwMirrorMissPullBackoff, defaultMissPullBackoff),
			durationOrDefault(config.MgwMirrorMissPullMaxBackoff, defaultMissPullMaxBackoff),
		),
	}
}

func durationOrDefault(value string, def time.Duration) time.Duration {
	if value == "" {
		return def
	}
	result, err := time.ParseDuration(value)
	if err != nil {
		return def
	}
	return result
}

type MirrorMiddleware struct {
	handler         http.Handler
	config          configuration.Config
	token           string
	pull            MirrorPull
	missPullTimeout time.Duration
	missBackoff     *missBackoff
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
		if r.Method == http.MethodGet {
			this.serveGetWithMissPull(w, r)
		} else {
			this.handler.ServeHTTP(w, r)
		}
	}
}

// serveGetWithMissPull answers a 404 only after the mirror asked the source for changes.
// The pull checks the last-update timestamps of the source, so a miss without changes in the source costs one request.
// If the pull takes longer than missPullTimeout (e.g. because the source is unreachable), the first 404 is returned
// and the pull continues in the background.
// A request that stays not found after its pull may pull again only after a backoff, see missBackoff.
func (this *MirrorMiddleware) serveGetWithMissPull(w http.ResponseWriter, r *http.Request) {
	key := r.URL.RequestURI()
	first := newBufferedResponse()
	this.handler.ServeHTTP(first, r)
	if first.status != http.StatusNotFound {
		this.missBackoff.found(key)
		first.writeTo(w)
		return
	}
	if !this.missBackoff.pullAllowed(key) {
		first.writeTo(w)
		return
	}
	done := make(chan error, 1)
	go func() {
		done <- this.pull.MirrorUpdate()
	}()
	select {
	case err := <-done:
		if err != nil {
			this.config.GetLogger().Warn("unable to update mirror after not found", "url", r.URL.String(), "error", err)
			this.missBackoff.missed(key)
			first.writeTo(w)
			return
		}
	case <-time.After(this.missPullTimeout):
		this.config.GetLogger().Warn("mirror update after not found exceeded timeout --> respond with not found", "url", r.URL.String(), "timeout", this.missPullTimeout.String())
		this.missBackoff.missed(key)
		first.writeTo(w)
		return
	}
	second := newBufferedResponse()
	this.handler.ServeHTTP(second, r)
	if second.status == http.StatusNotFound {
		this.missBackoff.missed(key)
	} else {
		this.missBackoff.found(key)
	}
	second.writeTo(w)
}

// missBackoff limits the pulls of not found reads, so that a client repeating a read of a missing entry
// does not send a request to the source each time.
// After a pull that did not find the entry (or failed, or exceeded the timeout), the same request may pull again
// only after a backoff, which doubles with every further miss up to max.
// A found response resets the request. So does a pause of more than max after the backoff ended,
// which also bounds the memory: such entries are swept.
type missBackoff struct {
	initial   time.Duration
	max       time.Duration
	now       func() time.Time
	mux       sync.Mutex
	entries   map[string]missBackoffEntry
	lastSweep time.Time
}

type missBackoffEntry struct {
	backoff time.Duration
	until   time.Time //no pull before this time
}

func newMissBackoff(initial time.Duration, max time.Duration) *missBackoff {
	return &missBackoff{initial: initial, max: max, now: time.Now, entries: map[string]missBackoffEntry{}}
}

func (this *missBackoff) pullAllowed(key string) bool {
	this.mux.Lock()
	defer this.mux.Unlock()
	entry, ok := this.entries[key]
	return !ok || !this.now().Before(entry.until)
}

func (this *missBackoff) missed(key string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	now := this.now()
	this.sweep(now)
	backoff := this.initial
	if entry, ok := this.entries[key]; ok && !this.expired(entry, now) {
		backoff = min(entry.backoff*2, this.max)
	}
	this.entries[key] = missBackoffEntry{backoff: backoff, until: now.Add(backoff)}
}

func (this *missBackoff) found(key string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	delete(this.entries, key)
}

func (this *missBackoff) expired(entry missBackoffEntry, now time.Time) bool {
	return now.After(entry.until.Add(this.max))
}

// sweep removes expired entries, at most once per max
func (this *missBackoff) sweep(now time.Time) {
	if now.Sub(this.lastSweep) < this.max {
		return
	}
	this.lastSweep = now
	for key, entry := range this.entries {
		if this.expired(entry, now) {
			delete(this.entries, key)
		}
	}
}

type bufferedResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func newBufferedResponse() *bufferedResponse {
	return &bufferedResponse{header: http.Header{}, status: http.StatusOK}
}

func (this *bufferedResponse) Header() http.Header {
	return this.header
}

func (this *bufferedResponse) WriteHeader(status int) {
	this.status = status
}

func (this *bufferedResponse) Write(b []byte) (int, error) {
	return this.body.Write(b)
}

func (this *bufferedResponse) writeTo(w http.ResponseWriter) {
	for k, vv := range this.header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(this.status)
	_, _ = w.Write(this.body.Bytes())
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
