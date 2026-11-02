# The client cache is keyed by request and user

`client.WithCache(c, ttl)` returns a copy of a client whose read methods answer
from a cache shared by all copies of the same client. A call on the client
itself, without `WithCache`, never reads the cache.

```go
c := client.NewClient(url, nil)
f, err, _ := client.WithCache(c, time.Minute).GetFunction(ctx, id) // asks the server, stores the response
f, err, _ = client.WithCache(c, time.Minute).GetFunction(ctx, id)  // answered from the cache
f, err, _ = c.GetFunction(ctx, id)                                 // asks the server
```

## Applies when

Reading through `lib/client` with `WithCache`, since `v3.6.0`, or adding a read
method to the client.

**Not this if**: the implementation of `client.Interface` in hand is not the one
from `NewClient` — the controller returned by `NewTestClient`, or a consumer's
own mock. `WithCache` returns those unchanged, so every call reaches them.

## What is cached

Read methods send their request through `doCached` or `doCachedWithTotal`
instead of `do` or `doWithTotalInResult`; nothing else is cached. That excludes
writes, validations and `GetLastUpdateTimestamps`, which exists to find out
whether something changed. **A new read method has to use `doCached`**, or it
runs uncached without any error. `TestWithCacheCoversReadMethods` calls every
`Read*`, `Get*` and `List*` method twice through `WithCache` and once without it,
and fails for a method that does not behave accordingly.

The cache holds the raw response body and the `X-Total-Count` header, so every
read decodes its own copy and a caller that changes a returned slice does not
change the cache. Errors are not cached. On success the code is `0`, as with
`do`, whether the answer came from the cache or not.

## The key

The HTTP method, the URL, the request body — many listings read through
`POST /query/...` — and the user id and roles of the `Authorization` header.
Results that depend on permissions are therefore never shared between users.

The token is parsed **without validation**. A refreshed token of the same user
reuses the cached responses, and an expired or revoked token can still be
answered from the cache until the ttl ends. A token that cannot be parsed
bypasses the cache.

## The ttl

An entry is used while it is younger than the ttl it was stored with **and**
younger than the ttl of the reading copy. A call site that needs fresher data
than another one stored passes a shorter ttl and gets a refresh; a longer ttl
does not extend an entry beyond the ttl it was stored with. Expired entries are
removed on a write, at most once a minute. There is no size limit.

## Rejected alternatives

- **A wrapper type implementing every read method.** 61 methods duplicated from
  `api.Controller`, keyed by the Go arguments. Caching at the request level needs
  one word changed per read method and keys by what actually goes over the wire.
- **`service-commons/pkg/cache`.** Its local layer hands out the stored object
  itself, so a caller writing into a returned slice changes what the next reader
  gets. A refresh for a shorter reading ttl would have to go through its
  validation callback, which logs every rejection as a warning.
- **The full token in the key.** Safe against revoked tokens, but every token
  refresh starts the user's cache from empty, so the hit rate is bounded by
  the token lifetime the identity provider is configured with.
