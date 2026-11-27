# Cache headers and when a 304 is correct

## Applies when

Adding or changing a GET endpoint, a `Set*`/`Remove*` method or a `Retry*Sync`
loop in `lib/database/mongo`, or a startup migration that writes one of the
collections listed below. Written 2026-11-27 for SNRGY-5178, read from
`lib/api/util/cache.go` and confirmed by `TestCacheHeaders` in
`lib/tests/cache_header_test.go`.

**Not this if**: a client gets stale data from its own in-memory cache.
`lib/client` keeps responses for its ttl; see
`the-client-cache-is-keyed-by-request-and-user.md`.

## What a response carries

Every GET of an endpoint in the positive list `cacheEndpoints`
(`lib/api/util/cache.go`) answers 200 with

```http
Cache-Control: private, no-cache
ETag: "<first 128 bit of a sha-256, hex>"
```

and a request whose `If-None-Match` matches gets `304` with an empty body,
`Cache-Control`, `Date` and `ETag`. An endpoint that is not in the list gets no
headers, so a new endpoint is uncached until someone adds it on purpose.
`HEAD` is never handled, errors stay errors.

No `Vary: Authorization`: `no-cache` forces a revalidation before every use, and
the ETag decides. **As soon as someone adds `max-age`, `Vary` becomes
mandatory**, because then the browser answers without asking.

`POST /clear-cache` answers `204` with `Clear-Site-Data: "cache"` and needs no
token. Frontends call it on logout; it clears the browser cache for the whole
API origin, not only this service. Any page can trigger it cross-origin, which
costs the victim a full reload of cached responses and nothing else. A mirror
answers it locally and does not forward it to its source.

## Two groups, two sources for the ETag

**Group timestamps** (metadata: aspects, aspect nodes, aspect classes,
concepts, characteristics, functions, device classes, device types, protocols,
services). The response does not depend on the caller. The ETag is a hash over
the last-update timestamps of a fixed set of collections plus the request URI.
A matching `If-None-Match` is answered with 304 **without calling the handler**.
The fixed set is the list in `GetCacheTimestamps` in `lib/controller/lastupdates.go`:
aspects, aspect nodes, aspect classes, concepts, characteristics, functions,
device classes, device types, device-type criteria, protocols. Every write to
any of them invalidates every ETag of this group.

**Group response** (devices, hubs, locations, device groups, things, graphs,
and the metadata endpoints that derive from the caller's devices:
`/user-device-types`, `/invalid/device-types`, `/defaults/devices/attributes`).
The ETag is a hash over the body and `X-Total-Count`. The handler always runs,
the saving is transfer, not database work. Timestamps cannot serve here: they
are written for the owner, permission changes in permissions-v2 leave none, and
the connection state moves none.

Rule for a new endpoint: if its result derives from ownership, rights or devices
of the caller, it belongs into group response, even if it returns metadata.

## What a 304 of group timestamps relies on

A 304 there is only correct if every change of a collection in the fixed set
moves its timestamp. The code holds that in three places:

- `Set*` and `Remove*` of these collections return an error wrapping
  `ErrLastUpdateTimestamp` if the timestamp cannot be written; they no longer
  swallow it.
- The `Retry*Sync` loops of these collections write the timestamp before they
  finish a job, and keep the job if that fails. A `Remove*` whose timestamp
  failed has already marked the document deleted, a second `Remove*` does not
  find it, so the retry loop is what moves the timestamp.
- `TestLastUpdateTimestampOfFixedSet` in `lib/database/mongo/lastupdates_test.go`
  calls every exported `Set*`/`Remove*` by reflection. A new method fails it
  until it is listed as writing the fixed set or not.

A startup migration writes past `Set*`. One that writes a collection of the
fixed set sets the timestamp itself (`writing-a-startup-migration.md`).

The middleware reads the timestamps before the handler runs: the database layer
writes data first and the timestamp second, so a concurrent write costs at most
one needless 200, never a wrong 304. There is no process cache for the
timestamps, because with two replicas a write on one and a read on the other
would answer 304 with the old state.

## Why the headers are what they are

- **ETag, not `Last-Modified`.** HTTP dates have second resolution, and two
  replicas each write the timestamp from their own clock
  (`lib/database/mongo/lastupdates.go`). Two writes in the same second would be
  indistinguishable.
- **ETag strong, SHA-256 cut to 128 bit.** Group timestamps hashes the sorted
  pairs of collection name and timestamp (a missing one counts as 0) plus the
  `RequestURI`. The URI binds the ETag to the URL: a matching validator can only
  come from an earlier 200 of the same URL at the same state, so a 304 never
  replaces a 400 or 404. Group response hashes body and `X-Total-Count`,
  because the total of a list can change while the requested page stays the same.
- **`private`.** Only the user's browser may store, no shared proxy or gateway.
  Device lists then outlive a logout in the browser's disk cache, which is what
  `/clear-cache` is for.
- **`no-cache`.** Store, but revalidate before every use. `must-revalidate`
  only acts together with `max-age`.
- **No `no-store` for group response.** The browser would keep nothing, not even
  in memory, and device and hub lists would be transferred in full on every start.
- **No `max-age`, `stale-while-revalidate`, `stale-if-error`.** They hand out data
  without asking and trade correctness for speed.
- **No `Vary: Authorization`.** The cache key would contain the token value, and
  every refresh of the access token would discard the stored responses.
- **`If-None-Match: *` in group timestamps runs the handler.** Answering it before
  the handler would turn a 404 into a 304.
- **No process cache for the timestamps**, see above.

CORS: `If-None-Match` is in `Access-Control-Allow-Headers` and
`Access-Control-Expose-Headers` names `ETag, X-Total-Count`. The browser's own
revalidation works without either, it runs below `fetch`. A frontend that keeps
the validator itself needs both. `X-Total-Count` was not readable cross-origin
before, it stood only in `Allow-Headers`.

The middleware sits inside CORS (a 304 needs `Access-Control-Allow-Origin`), and
so inside access log and `MirrorMiddleware`. On a mirror, `MirrorMiddleware`
replaces every token by the mirror user's, so all users get the same response
and ETag, and a group response request is buffered twice.

Behind nginx with `gzip on`, a compressed 200 carries `W/"…"` instead of the
strong ETag; a 304 has no body and keeps the strong one. The middleware compares
`If-None-Match` weakly, as RFC 9110 requires, so `W/"x"` matches `"x"`.

## Not covered yet

- **Ingress and gateway.** Whether they rewrite or drop `Cache-Control`, `ETag`
  or `If-None-Match`, and whether they compress JSON at all, is unchecked.
- **Browsers.** That they revalidate `GET` on their own is to be measured, and it
  does not hold for the `POST /query/...` calls. Whether `Clear-Site-Data: "cache"`
  is supported in Chrome, Firefox and Safari, and whether a cross-origin `fetch`
  clears the cache partitioned by top-level site, is unchecked. If not, logout
  needs a navigation to the API origin or clearing through the gateway.
- **Logout without `/clear-cache`**: expired session, closed tab, single sign-out
  from another application. The cache stays; an entry is still served only after
  revalidation, and in group response the handler checks rights on that.
- **Memory.** Group response buffers the whole body; very large device and graph
  lists are not measured.
- **Cheaper group response.** A Kafka consumer on the `RIGHTS` messages that
  permissions-v2 writes to the same topics
  (`rights-and-content-share-one-kafka-topic.md`), together with a timestamp per
  user instead of per owner, would allow a 304 for devices without database
  work. It needs a timestamp on connection state changes
  (`SetDeviceConnectionState`, `SetHubConnectionState`, filtered on
  `connection_state: {$ne: state}`), where the owner is not known today, and
  every such timestamp makes the mgw mirror pull the whole collection.
- **Frontends** (app, web-ui, dashboard) still have to call `/clear-cache` in
  their logout flow before redirecting to Keycloak.

## The Go client revalidates

`lib/client` keeps an expired response with an ETag for another ten minutes
(`cacheRevalidationRetention`) and sends `If-None-Match` for it. A 304 renews
the entry with its body and `X-Total-Count`. A 304 that arrives after an
invalidation does not revive the entry; the client fetches again without the
header.
