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

package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SENERGY-Platform/device-repository/v3/lib/model"
)

// countingRepo answers every request with the number of requests so far as name,
// so a test can tell whether a result came from the server or from the cache
type countingRepo struct {
	mux      sync.Mutex
	requests int
	fail     bool
}

func (this *countingRepo) count() int {
	this.mux.Lock()
	defer this.mux.Unlock()
	return this.requests
}

func (this *countingRepo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.mux.Lock()
	this.requests++
	n := strconv.Itoa(this.requests)
	fail := this.fail
	this.mux.Unlock()
	if fail {
		http.Error(w, "fail", http.StatusInternalServerError)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/query/") {
		w.Header().Set("X-Total-Count", "2")
		json.NewEncoder(w).Encode([]map[string]string{{"id": "a", "name": n}, {"id": "b", "name": n}})
		return
	}
	segments := strings.Split(r.URL.Path, "/")
	json.NewEncoder(w).Encode(map[string]string{"id": segments[len(segments)-1], "name": n})
}

type testClock struct {
	now time.Time
}

func (this *testClock) Now() time.Time {
	return this.now
}

func newCacheTestClient(t *testing.T) (Interface, *countingRepo, *testClock) {
	repo := &countingRepo{}
	server := httptest.NewServer(repo)
	t.Cleanup(server.Close)
	c := NewClient(server.URL, nil)
	clock := &testClock{now: time.Now()}
	c.(*Client).cache.now = clock.Now
	return c, repo, clock
}

// testToken builds an unsigned jwt; the cache parses tokens without validation
func testToken(sub string, jti string, roles ...string) string {
	encode := func(v interface{}) string {
		b, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return "Bearer " + encode(map[string]string{"alg": "HS256", "typ": "JWT"}) + "." +
		encode(map[string]interface{}{"sub": sub, "jti": jti, "realm_access": map[string][]string{"roles": roles}}) + ".sig"
}

func TestWithCacheSharesResultsBetweenViews(t *testing.T) {
	c, repo, _ := newCacheTestClient(t)
	ctx := context.Background()

	first, err, _ := WithCache(c, time.Minute).GetFunction(ctx, "f")
	if err != nil {
		t.Fatal(err)
	}
	second, err, code := WithCache(c, time.Minute).GetFunction(ctx, "f")
	if err != nil {
		t.Fatal(err)
	}
	if repo.count() != 1 || second.Name != first.Name || code != 0 {
		t.Errorf("expected one request and the cached result, got %v requests, %v and %v, code %v", repo.count(), first.Name, second.Name, code)
	}

	other, _, _ := WithCache(c, time.Minute).GetFunction(ctx, "g")
	if repo.count() != 2 || other.Id != "g" {
		t.Errorf("other arguments must not hit the cache, got %v requests and %v", repo.count(), other.Id)
	}
}

func TestWithoutCacheIgnoresCachedResults(t *testing.T) {
	c, repo, _ := newCacheTestClient(t)
	ctx := context.Background()

	set, err, _ := WithCache(c, time.Minute).GetFunction(ctx, "f")
	if err != nil {
		t.Fatal(err)
	}
	read, err, _ := WithCache(c, time.Minute).GetFunction(ctx, "f")
	if err != nil {
		t.Fatal(err)
	}
	if repo.count() != 1 || read.Name != set.Name {
		t.Fatalf("expected the second call to read the cache, got %v requests, %v and %v", repo.count(), set.Name, read.Name)
	}

	uncached, err, _ := c.GetFunction(ctx, "f")
	if err != nil {
		t.Fatal(err)
	}
	if repo.count() != 2 || uncached.Name == set.Name {
		t.Errorf("a call without WithCache must ask the server, got %v requests and %v from the cache", repo.count(), uncached.Name)
	}
}

func TestWithCacheKeysByRequestBody(t *testing.T) {
	c, repo, _ := newCacheTestClient(t)
	ctx := context.Background()

	WithCache(c, time.Minute).ListFunctions(ctx, model.FunctionListOptions{Limit: 10})
	WithCache(c, time.Minute).ListFunctions(ctx, model.FunctionListOptions{Limit: 20})
	if repo.count() != 2 {
		t.Errorf("queries with another body must not hit the cache, got %v requests", repo.count())
	}
	WithCache(c, time.Minute).ListFunctions(ctx, model.FunctionListOptions{Limit: 10})
	if repo.count() != 2 {
		t.Errorf("the same query must hit the cache, got %v requests", repo.count())
	}
}

func TestWithCacheTtl(t *testing.T) {
	c, repo, clock := newCacheTestClient(t)
	ctx := context.Background()

	WithCache(c, time.Minute).GetFunction(ctx, "f")

	clock.now = clock.now.Add(30 * time.Second)
	WithCache(c, time.Hour).GetFunction(ctx, "f")
	if repo.count() != 1 {
		t.Errorf("an entry younger than both ttls must be used, got %v requests", repo.count())
	}

	WithCache(c, 10*time.Second).GetFunction(ctx, "f")
	if repo.count() != 2 {
		t.Errorf("an entry older than the reading ttl must be refreshed, got %v requests", repo.count())
	}

	//the refresh above stored the entry with a ttl of 10s
	clock.now = clock.now.Add(11 * time.Second)
	WithCache(c, time.Hour).GetFunction(ctx, "f")
	if repo.count() != 3 {
		t.Errorf("an entry older than the ttl it was stored with must be refreshed, got %v requests", repo.count())
	}
}

func TestWithCacheKeysByUser(t *testing.T) {
	c, repo, _ := newCacheTestClient(t)
	ctx := context.Background()
	read := func(token string) {
		_, err, _ := WithCache(c, time.Minute).ReadDevice(ctx, "d", token, model.READ)
		if err != nil {
			t.Fatal(err)
		}
	}

	read(testToken("alice", "1", "user"))
	read(testToken("alice", "2", "user"))
	if repo.count() != 1 {
		t.Errorf("a refreshed token of the same user must hit the cache, got %v requests", repo.count())
	}

	read(testToken("bob", "1", "user"))
	if repo.count() != 2 {
		t.Errorf("another user must not hit the cache, got %v requests", repo.count())
	}

	read(testToken("alice", "3", "user", "admin"))
	if repo.count() != 3 {
		t.Errorf("the same user with other roles must not hit the cache, got %v requests", repo.count())
	}

	read("Bearer not-a-jwt")
	read("Bearer not-a-jwt")
	if repo.count() != 5 {
		t.Errorf("a token that can not be parsed must bypass the cache, got %v requests", repo.count())
	}
}

func TestWithCacheSkipsErrors(t *testing.T) {
	c, repo, _ := newCacheTestClient(t)
	ctx := context.Background()

	repo.fail = true
	_, err, _ := WithCache(c, time.Minute).GetFunction(ctx, "f")
	if err == nil {
		t.Fatal("expected error")
	}
	repo.fail = false
	_, err, _ = WithCache(c, time.Minute).GetFunction(ctx, "f")
	if err != nil || repo.count() != 2 {
		t.Errorf("an error must not be cached, got %v and %v requests", err, repo.count())
	}
}

func TestWithCacheReturnsCopies(t *testing.T) {
	c, repo, _ := newCacheTestClient(t)
	ctx := context.Background()

	first, total, err, _ := WithCache(c, time.Minute).ListFunctions(ctx, model.FunctionListOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	first[0].Name = "changed"
	second, secondTotal, err, _ := WithCache(c, time.Minute).ListFunctions(ctx, model.FunctionListOptions{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if repo.count() != 1 || second[0].Name != "1" || total != 2 || secondTotal != 2 {
		t.Errorf("expected an unchanged cached copy with total, got %v requests, %#v, totals %v and %v", repo.count(), second, total, secondTotal)
	}
}

func TestWithCacheReturnsOtherImplementationsUnchanged(t *testing.T) {
	type other struct {
		Interface
	}
	impl := other{}
	if WithCache(impl, time.Minute) != Interface(impl) {
		t.Error("expected the implementation itself")
	}
}

// nullRepo answers every request with a json null, which every result type decodes
type nullRepo struct {
	mux      sync.Mutex
	requests int
}

func (this *nullRepo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.mux.Lock()
	this.requests++
	this.mux.Unlock()
	w.Header().Set("X-Total-Count", "0")
	w.Write([]byte("null"))
}

func TestWithCacheCoversReadMethods(t *testing.T) {
	repo := &nullRepo{}
	server := httptest.NewServer(repo)
	defer server.Close()
	c := NewClient(server.URL, nil)
	cached := reflect.ValueOf(WithCache(c, time.Minute))
	uncached := reflect.ValueOf(c)

	clientType := reflect.TypeOf(c)
	for i := 0; i < clientType.NumMethod(); i++ {
		method := clientType.Method(i)
		isRead := strings.HasPrefix(method.Name, "Read") || strings.HasPrefix(method.Name, "Get") || strings.HasPrefix(method.Name, "List")
		if !isRead || method.Name == "GetLastUpdateTimestamps" || method.Name == "GetPermissionsClient" {
			continue
		}
		call := func(target reflect.Value) []reflect.Value {
			args := []reflect.Value{}
			for j := 1; j < method.Type.NumIn(); j++ {
				argType := method.Type.In(j)
				switch {
				case argType == reflect.TypeOf((*context.Context)(nil)).Elem():
					args = append(args, reflect.ValueOf(context.Background()))
				case method.Name == "GetFunctionsByType" && argType.Kind() == reflect.String:
					args = append(args, reflect.ValueOf(model.SES_ONTOLOGY_MEASURING_FUNCTION))
				case argType.Kind() == reflect.String:
					//every string may be a token, which has to be parseable to be cached; some methods put strings unescaped into the url, so no "Bearer " with its space
					args = append(args, reflect.ValueOf(strings.TrimPrefix(testToken("alice", "1", "user"), "Bearer ")).Convert(argType))
				default:
					args = append(args, reflect.Zero(argType))
				}
			}
			return target.MethodByName(method.Name).Call(args)
		}
		requests := func() int {
			repo.mux.Lock()
			defer repo.mux.Unlock()
			return repo.requests
		}
		results := call(cached)
		if err := results[len(results)-2]; !err.IsNil() {
			t.Errorf("%v: %v", method.Name, err.Interface())
			continue
		}
		before := requests()
		call(cached)
		if requests() != before {
			t.Errorf("%v is not cached", method.Name)
		}
		call(uncached)
		if requests() != before+1 {
			t.Errorf("%v without WithCache reads the cache", method.Name)
		}
	}
}
