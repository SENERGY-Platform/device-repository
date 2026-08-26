# display_name is derived, not settable

## Applies when

Reading or setting a device's display name or its attributes through this
service — `PUT /devices/{id}`, `PUT /devices/{id}/attributes`,
`PUT /devices/{id}/display_name`.

**Not this if**: the question is about hubs or another resource type. Whether
those follow the same derivation and origin-filter rules is **not verified
here**.

## The derivation

`display_name` is a **computed field**. The mongo layer recomputes it on every
save from the `shared/nickname` attribute and falls back to the device `name`
when that attribute is missing **or empty**.

The `models.Device` payload of `PUT /devices/{id}` has no `display_name` field at
all — a value sent inside the device JSON is silently discarded and rebuilt from
the attributes. A client therefore cannot desync the two server-side; it can only
hold a stale copy locally, which is the bug this usually shows up as.

Three ways to change it, all equivalent on the server:

- `PUT /devices/{id}` with `shared/nickname` added, changed or removed in
  `attributes`
- `PUT /devices/{id}/attributes` with the full attribute list
- `PUT /devices/{id}/display_name` with a JSON string body — **`""` is the delete
  signal**: the endpoint sets the attribute to the empty string and the
  recomputation falls back to `name`

## update-only-same-origin-attributes

`PUT /devices/{id}` and `PUT /devices/{id}/attributes` accept
`?update-only-same-origin-attributes=<origins>`, comma-separated:

- attributes whose origin **is in the list** are replaced wholesale by the
  request — one missing from the request is **deleted**
- attributes of any **other** origin are carried over from the stored device,
  untouched

That is what lets a client rewrite its own and the shared attributes without
clobbering connector-owned ones.

## Where this came up

A "clearing the display name sometimes does not stick" report. Every backend path
turned out correct; a client kept a stale `display_name` after removing the
nickname attribute and re-submitted it through the display_name endpoint,
re-creating the nickname it had just deleted. Knowing that `""` is a valid delete
signal and that the payload field is ignored makes such a fix a two-liner.
