# A request context carries the trace, not the cancellation

Every method of `api.Controller`, the controller and `lib/client` takes a
`context.Context` as first parameter, and the server passes the request's
context all the way to mongo and permissions-v2. What it passes is the trace:
the cancellation of the request is dropped on purpose.

## Applies when

Changing code on the request path of the server — a handler in `lib/api`, a
controller method, a sync handler — or wondering why a client that disconnects
does not abort the work it started. Since `v3.5.0`.

**Not this if**: the router was built with `GetRouterWithoutMiddleware`. It has
neither the tracing nor the middleware that drops the cancellation, so a handler
behind it sees the plain request context. Only test setups use it. Also not
this if the question is about the **client** side: there the caller's context
is used as is, so a caller's deadline or cancellation does end its request.

## Why the cancellation is dropped

Most writes here are not one write. A device is stored in mongo, its rights go
to permissions-v2, and the change is published to kafka, with no transaction
around the three. A cancellable context makes the gap between them abortable:
a client that hangs up after the mongo write leaves a document without its
permission entry or without the event its consumers wait for.

So the context keeps its values — the span, the baggage — and loses its
cancellation, in two places:

- `withoutCancel` in `lib/api/api.go` wraps the router inside the tracing
  handler, so every handler gets `context.WithoutCancel(request.Context())`.
- `getTimeoutContext(ctx)` in `lib/controller/controller.go` builds its 10 s
  timeout on `context.WithoutCancel(ctx)`. That covers callers without a
  request too, and it strips a deadline an earlier `getTimeoutContext` set, so
  every call gets its own 10 s as before.

## The token's user goes into the baggage, encoded

`GetRouter` wraps the router with `otelx.HTTPOpenTelemetry` from
gin-middleware. Its handler makes `sub` and `preferred_username` of the token
attributes of the server span and entries of the baggage, as `user_id` and
`username`. If the caller sent a non-empty baggage of its own, that one is what
reaches the service and everything further down, because `otelhttp` extracts it
after otelx has added its entries. So a service that calls with a service token
on behalf of a user keeps that user's `user_id` in the baggage.

Up to gin-middleware `v0.14.1` the handler answered
`500 Failed to add username baggage` when the baggage format could not carry the
value — a space, a comma, an umlaut — before the service saw the request. That
was the state of this service in `v3.5.0` and `v3.6.0`. gin-middleware now
percent-encodes such a value: the header carries `username=first%20last`, and
the span attribute and the value a downstream service parses are the original
string. Replacing the characters, a space by `_` for instance, was rejected: it
loses the original, and `first last` and `first_last` become the same user in
every trace.

`TestUsernameOutsideTheBaggageFormatIsServed` in `lib/tests/tracing_test.go`
fails against a gin-middleware without that fix, and if user id or username go
missing from the span or the baggage.

## How the context gets where it is needed

- **Sync handlers.** The database calls them with a fixed signature, e.g.
  `func(models.Hub) error`, so they cannot take a context parameter from it.
  `withContext` and `withContext2` in `lib/controller/controller.go` bind the
  caller's context into a closure. The retry loop in `Sync` binds the context
  of the sync loop the same way.
- **Paths without a request** pass `context.Background()`: the controller
  constructor, one startup migration, and the mirror pull in `lib/mgwmirror`.
  The pull is shared by every request that triggered it, so it belongs to none
  of their traces.
- **Outgoing requests.** `lib/client` builds every request with
  `http.NewRequestWithContext` and writes the trace headers through the global
  OpenTelemetry propagator before sending. The mirror middleware does the same
  when it forwards a write to its source.

## Rejected alternatives

- **A variadic `ctx ...context.Context` as last parameter.** Source compatible
  for callers, but it changes the type of every method, so every implementation
  of `client.Interface` breaks, and so does any narrow interface a consumer
  declares over the client's methods.
- **`WithContext(ctx)` on the client, returning a copy.** The test controller
  returned by `NewTestClient` implements the same interface and would have to
  return `client.Interface` as well, which `lib/controller` cannot import without
  an import cycle. A package function would avoid that, but the server side needs
  the context in the controller anyway.
- **A `*Context` variant next to every method**, as permissions-v2 has them.
  Keeps every caller compiling, at the price of a second method for each of the
  120. Backward compatibility was not required, so the existing methods were
  changed instead.
