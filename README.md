<a href="https://github.com/SENERGY-Platform/device-repository/actions/workflows/tests.yml" rel="nofollow">
    <img src="https://github.com/SENERGY-Platform/device-repository/actions/workflows/tests.yml/badge.svg?branch=master" alt="Tests" />
</a>

# device-repository

The platform's registry for devices and their meaning: devices, device types,
hubs, graphs and their attributes. It owns this data — every other service reads
it from here rather than keeping its own copy.

A **device type** is what carries the semantics of a measuring point:
`function_ids`, `aspect_ids` and a `characteristic_id` per content variable give
a value its meaning and its unit. A **device** is an instance of such a type.

Changes are published to kafka, so consumers can follow the registry without
polling. Permissions live in permissions-v2, which publishes onto the same
topics.

## OpenAPI
uses https://github.com/swaggo/swag

### generating
```
go generate ./...
```

### swagger ui
if the config variable UseSwaggerEndpoints is set to true, a swagger ui is accessible on /swagger/index.html (http://localhost:8080/swagger/index.html)

## Further documentation

`docs/` holds hand-written knowledge next to the generated specs — behaviour
that the API reference cannot express, all of it verified against a running
instance:

- [display_name is derived, not settable](docs/display-name-is-derived.md) —
  and how `update-only-same-origin-attributes` decides what survives a write
- [Attribute filters on a listing are ANDed](docs/attribute-filters-on-listings.md) —
  reads like an OR, is not, and what that does to a consumer looking for its own
  resources
- [Rights and content commands share one kafka topic](docs/rights-and-content-share-one-kafka-topic.md) —
  two producers, one ordering guarantee, and two traps when adding a topic
- [Writing graphs](docs/writing-graphs.md) — a new graph's id can only come
  from the server (a self-invented id fails as 403, not 404), and edge weights
  are validated: 0 is refused, outgoing weights sum to 0 or 100
- [asyncapi.json is only half generated](docs/asyncapi-is-half-generated.md) —
  the generator lives in its own module, emits an older spec version, and the
  committed file is a hand-made conversion
- [Not every device-type read passes the controller](docs/not-every-device-type-read-passes-the-controller.md) —
  two paths read or write device-types straight from the database, and one of
  them rewrites stored device-group criteria while doing so

Working in this repository:

- [Which tests break when a response gains a field](docs/tests-that-compare-whole-payloads.md) —
  the five packages that compare whole payloads, and the three shapes their
  expected values come in
- [Regenerating the REST docs rewrites a file you did not change](docs/regenerating-the-rest-docs.md) —
  `generated_permissions.go` comes back reordered and has to be reverted before
  committing
- [The test suite needs -p 1](docs/tests-collide-on-port-8080.md) — packages run
  in parallel and collide on port 8080, which looks like a flaky failure
