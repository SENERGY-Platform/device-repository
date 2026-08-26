# Attribute filters on a listing are ANDed

## Applies when

Filtering a **listing** endpoint by attributes through the Go client in
`lib/client` — typically to find your own resources again by a provenance
attribute you wrote. Observed on `GET /graphs`; the filter is built the same way
for the other resource types, which was **not separately checked**.

**Not this if**: the filter names a single attribute key. The AND is only visible
from two keys onwards, and a one-key filter behaves exactly as documented.

## The plain `attributes` filter matches ALL listed keys

It is documented as matching an element that carries an attribute "that is in the
given list", which reads like an OR. The implementation builds one `$elemMatch`
per listed attribute and **ANDs** them, so a two-key filter returns only the
resources carrying **both**.

Why that bites: a resource written before the second attribute existed, or a
write interrupted between the two, is invisible to the filter. A consumer that
reads "not found" as "does not exist" then **creates a second resource beside the
one it cannot see**.

Confirmed against a running repository on 2026-08-26: a graph carrying only one
of two provenance keys did not appear in the two-key listing.

The way around it is one request per key and the union client-side — which is
also the cheaper request.

## `attributes_json` cannot be reached through the Go client

The filter that matches attribute *values* is unusable from the client: the
client percent-encodes the JSON and the query encoder escapes it a second time,
so the repository is handed something that is no longer JSON.

```
unexpected statuscode 400: unable to parse attributes_json:invalid character '%' looking for beginning of value
```

Setting an `Origin` on an attribute in the filter is what makes the client take
that path, so it is easy to trip over while only meaning to narrow the filter.

Until the client is fixed: filter on the **presence** of a key and compare values
yourself. That is usually affordable — the set behind one provenance key is small
by construction.
