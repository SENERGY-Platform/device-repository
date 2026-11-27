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

package util

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
)

const cacheControlValue = "private, no-cache"

// cacheGroup says where the ETag of an endpoint comes from.
type cacheGroup int

const (
	// cacheGroupTimestamps: the response does not depend on the caller, the ETag is derived from the
	// last-update timestamps of a fixed set of collections and a 304 is answered without calling the handler.
	cacheGroupTimestamps cacheGroup = iota + 1
	// cacheGroupResponse: the response depends on the caller (ownership, rights, devices), the ETag is derived
	// from the response, the handler always runs.
	cacheGroupResponse
)

// cacheEndpoints is the positive list. An endpoint that is not listed gets no cache headers.
// A new endpoint whose result derives from ownership, rights or devices of the caller belongs into cacheGroupResponse,
// even if it returns metadata. Only GET is listed, HEAD is excluded by the middleware (ServeMux matches it too).
var cacheEndpoints = map[string]cacheGroup{
	// group timestamps
	"/v2/aspects":                                cacheGroupTimestamps,
	"/aspects":                                   cacheGroupTimestamps,
	"/aspects/{id}":                              cacheGroupTimestamps,
	"/aspects/{id}/measuring-functions":          cacheGroupTimestamps,
	"/aspects/{id}/controlling-functions":        cacheGroupTimestamps,
	"/aspect-classes":                            cacheGroupTimestamps,
	"/aspect-classes/{id}":                       cacheGroupTimestamps,
	"/v2/aspect-nodes":                           cacheGroupTimestamps,
	"/aspect-nodes":                              cacheGroupTimestamps,
	"/aspect-nodes/{id}":                         cacheGroupTimestamps,
	"/aspect-nodes/{id}/measuring-functions":     cacheGroupTimestamps,
	"/aspect-nodes/{id}/controlling-functions":   cacheGroupTimestamps,
	"/v2/concepts":                               cacheGroupTimestamps,
	"/v2/concepts-with-characteristics":          cacheGroupTimestamps,
	"/concepts/{id}":                             cacheGroupTimestamps,
	"/v2/characteristics":                        cacheGroupTimestamps,
	"/characteristics":                           cacheGroupTimestamps,
	"/characteristics/{id}":                      cacheGroupTimestamps,
	"/functions":                                 cacheGroupTimestamps,
	"/functions/{id}":                            cacheGroupTimestamps,
	"/measuring-functions":                       cacheGroupTimestamps,
	"/controlling-functions":                     cacheGroupTimestamps,
	"/v2/device-classes":                         cacheGroupTimestamps,
	"/device-classes":                            cacheGroupTimestamps,
	"/device-classes/{id}":                       cacheGroupTimestamps,
	"/device-classes/{id}/functions":             cacheGroupTimestamps,
	"/device-classes/{id}/controlling-functions": cacheGroupTimestamps,
	"/v3/device-types":                           cacheGroupTimestamps,
	"/device-types":                              cacheGroupTimestamps,
	"/device-types/{id}":                         cacheGroupTimestamps,
	"/protocols":                                 cacheGroupTimestamps,
	"/protocols/{id}":                            cacheGroupTimestamps,
	"/services/{id}":                             cacheGroupTimestamps,
	// group response
	"/devices":                     cacheGroupResponse,
	"/devices/{id}":                cacheGroupResponse,
	"/extended-devices":            cacheGroupResponse,
	"/extended-devices/{id}":       cacheGroupResponse,
	"/local-devices":               cacheGroupResponse,
	"/local-devices/{id}":          cacheGroupResponse,
	"/device-groups":               cacheGroupResponse,
	"/device-groups/{id}":          cacheGroupResponse,
	"/hubs":                        cacheGroupResponse,
	"/hubs/{id}":                   cacheGroupResponse,
	"/hubs/{id}/devices":           cacheGroupResponse,
	"/extended-hubs":               cacheGroupResponse,
	"/extended-hubs/{id}":          cacheGroupResponse,
	"/locations":                   cacheGroupResponse,
	"/locations/{id}":              cacheGroupResponse,
	"/extended-locations":          cacheGroupResponse,
	"/things":                      cacheGroupResponse,
	"/things/{id}":                 cacheGroupResponse,
	"/graphs":                      cacheGroupResponse,
	"/graphs/{id}":                 cacheGroupResponse,
	"/graphs/{id}/segments":        cacheGroupResponse,
	"/user-device-types":           cacheGroupResponse,
	"/invalid/device-types":        cacheGroupResponse,
	"/defaults/devices/attributes": cacheGroupResponse,
}

// CacheTimestamps returns the last-update timestamps of the fixed set of collections that the endpoints of
// cacheGroupTimestamps read. It is called once per request of such an endpoint and must not cache.
type CacheTimestamps func(ctx context.Context) ([]model.LastUpdateTimestamp, error)

// NewCache adds ETag and Cache-Control to GET responses of the endpoints in cacheEndpoints and answers
// If-None-Match with 304.
// timestamps may be nil: the endpoints of cacheGroupTimestamps are then passed through without cache headers.
// The same happens for a single request if timestamps fails.
func NewCache(handler http.Handler, timestamps CacheTimestamps) http.Handler {
	mux := http.NewServeMux()
	for pattern := range cacheEndpoints {
		mux.Handle("GET "+pattern, http.NotFoundHandler()) //only the pattern is used
	}
	return &CacheMiddleware{handler: handler, timestamps: timestamps, mux: mux}
}

type CacheMiddleware struct {
	handler    http.Handler
	timestamps CacheTimestamps
	mux        *http.ServeMux
}

func (this *CacheMiddleware) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	group, ok := this.group(req)
	if !ok {
		this.handler.ServeHTTP(res, req)
		return
	}
	switch group {
	case cacheGroupTimestamps:
		this.serveTimestampGroup(res, req)
	case cacheGroupResponse:
		this.serveResponseGroup(res, req)
	default:
		this.handler.ServeHTTP(res, req)
	}
}

func (this *CacheMiddleware) group(req *http.Request) (group cacheGroup, ok bool) {
	if req.Method != http.MethodGet {
		return 0, false
	}
	_, pattern := this.mux.Handler(req)
	pattern = strings.TrimPrefix(pattern, "GET ")
	group, ok = cacheEndpoints[pattern]
	return group, ok
}

func (this *CacheMiddleware) serveTimestampGroup(res http.ResponseWriter, req *http.Request) {
	if this.timestamps == nil {
		this.handler.ServeHTTP(res, req)
		return
	}
	// read before the handler: the database layer writes the data first and the timestamp second, so a write that
	// runs in parallel can only cause one superfluous 200, never a wrong 304
	timestamps, err := this.timestamps(req.Context())
	if err != nil {
		this.handler.ServeHTTP(res, req)
		return
	}
	requestURI := req.RequestURI
	if requestURI == "" {
		requestURI = req.URL.RequestURI()
	}
	etag := timestampETag(timestamps, requestURI)
	match := newIfNoneMatch(req)
	if match.matchesTag(etag) {
		writeNotModified(res, etag)
		return
	}
	if match.any {
		// "*" matches only if the resource exists; without the handler that is unknown
		this.serveBuffered(res, req, func(*bufferedResponse) string { return etag }, match)
		return
	}
	this.handler.ServeHTTP(&etagWriter{ResponseWriter: res, etag: etag}, req)
}

func (this *CacheMiddleware) serveResponseGroup(res http.ResponseWriter, req *http.Request) {
	this.serveBuffered(res, req, responseETag, newIfNoneMatch(req))
}

func (this *CacheMiddleware) serveBuffered(res http.ResponseWriter, req *http.Request, etagOf func(*bufferedResponse) string, match ifNoneMatch) {
	buffered := newBufferedResponse()
	this.handler.ServeHTTP(buffered, req)
	if buffered.status != http.StatusOK {
		buffered.writeTo(res)
		return
	}
	etag := etagOf(buffered)
	if match.matches(etag) {
		writeNotModified(res, etag)
		return
	}
	buffered.header.Set("Cache-Control", cacheControlValue)
	buffered.header.Set("ETag", etag)
	buffered.writeTo(res)
}

// etagWriter sets the cache headers of a 200 response; every other status stays as the handler wrote it.
type etagWriter struct {
	http.ResponseWriter
	etag        string
	wroteHeader bool
}

func (this *etagWriter) WriteHeader(status int) {
	if this.wroteHeader {
		return
	}
	this.wroteHeader = true
	if status == http.StatusOK {
		this.Header().Set("Cache-Control", cacheControlValue)
		this.Header().Set("ETag", this.etag)
	}
	this.ResponseWriter.WriteHeader(status)
}

func (this *etagWriter) Write(b []byte) (int, error) {
	if !this.wroteHeader {
		this.WriteHeader(http.StatusOK)
	}
	return this.ResponseWriter.Write(b)
}

func (this *etagWriter) Unwrap() http.ResponseWriter {
	return this.ResponseWriter
}

func writeNotModified(res http.ResponseWriter, etag string) {
	res.Header().Set("Cache-Control", cacheControlValue)
	res.Header().Set("ETag", etag)
	res.Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
	res.WriteHeader(http.StatusNotModified)
}

// timestampETag binds the ETag to the state of the collections and to the URL: a matching validator can only come
// from an earlier 200 for the same URL at the same state, so a 304 never stands in for a 400 or 404.
func timestampETag(timestamps []model.LastUpdateTimestamp, requestURI string) string {
	sorted := slices.Clone(timestamps)
	slices.SortFunc(sorted, func(a, b model.LastUpdateTimestamp) int { return strings.Compare(a.Collection, b.Collection) })
	h := sha256.New()
	for _, t := range sorted {
		h.Write([]byte(strconv.Quote(t.Collection)))
		h.Write([]byte{0})
		h.Write([]byte(strconv.FormatInt(t.UnixTimestamp, 10)))
		h.Write([]byte{0})
	}
	h.Write([]byte(requestURI))
	return formatETag(h.Sum(nil))
}

// responseETag covers the body and the total count, which can change while the requested page stays the same.
func responseETag(res *bufferedResponse) string {
	h := sha256.New()
	h.Write([]byte(res.header.Get("X-Total-Count")))
	h.Write([]byte{0})
	h.Write(res.body.Bytes())
	return formatETag(h.Sum(nil))
}

// formatETag is a strong validator: the first 128 bit of the hash as hex in quotes
func formatETag(sum []byte) string {
	return `"` + hex.EncodeToString(sum[:16]) + `"`
}

// ifNoneMatch is the parsed If-None-Match header. The comparison is the weak one that RFC 9110 prescribes for it:
// the W/ prefix is ignored.
type ifNoneMatch struct {
	any  bool
	tags []string
}

func newIfNoneMatch(req *http.Request) (result ifNoneMatch) {
	for _, value := range req.Header.Values("If-None-Match") {
		parseIfNoneMatch(value, &result)
	}
	return result
}

func parseIfNoneMatch(value string, result *ifNoneMatch) {
	for {
		value = strings.TrimLeft(value, " \t,")
		if value == "" {
			return
		}
		if value[0] == '*' {
			result.any = true
			value = value[1:]
			continue
		}
		rest := strings.TrimPrefix(value, "W/")
		if len(rest) > 0 && rest[0] == '"' {
			end := strings.IndexByte(rest[1:], '"')
			if end < 0 {
				return
			}
			result.tags = append(result.tags, rest[:end+2])
			value = rest[end+2:]
			continue
		}
		// not an entity tag: skip to the next list element
		next := strings.IndexByte(value, ',')
		if next < 0 {
			return
		}
		value = value[next+1:]
	}
}

func (this ifNoneMatch) matchesTag(etag string) bool {
	return slices.Contains(this.tags, etag)
}

func (this ifNoneMatch) matches(etag string) bool {
	return this.any || this.matchesTag(etag)
}
