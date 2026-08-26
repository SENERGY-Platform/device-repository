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
- [asyncapi.json is only half generated](docs/asyncapi-is-half-generated.md) —
  the generator lives in its own module, emits an older spec version, and the
  committed file is a hand-made conversion
