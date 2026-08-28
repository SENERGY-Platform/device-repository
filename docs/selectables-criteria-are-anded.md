# A selectables criteria list is ANDed, and an empty one matches everything

## Applies when

Querying `POST /v2/query/device-type-selectables` with more than one entry in the
criteria body, or with an empty one. Follows from the implementation in
`lib/controller/devicetype.go` and `lib/database/mongo/devicetype.go` rather than
from a single observation.

**Not this if**: the question is about the v1 endpoint
`POST /query/device-type-selectables`. Its criteria list is ANDed the same way,
but it carries a separate `interactionsFilter` instead of a per-criterion
`Interaction` and knows no `services_must_match_all_criteria` — so everything
below about that flag does not apply to it.

## The criteria list narrows, one criterion at a time

`GetDeviceTypeIdsByFilterCriteriaV2` calls the per-criterion filter in a loop and
feeds each result into the next as the `$in` set of device type ids. The list is
therefore an **AND** over device types.

So `[{function: power}, {function: energy}]` asks for a device type carrying
**both** — for two unrelated functions, none at all. A caller whose question
means "any matched function, in any matched aspect" has to send **one criterion
per request** and union the answers itself. That makes its request count the
product of the matched sets, which is worth capping.

## `services_must_match_all_criteria` does not turn that AND into an OR

Its swagger line reads *"toggle if filter criteria is 'and' or 'or' combination"*,
which promises more than the flag delivers. The device-type-level AND above is
unconditional — the flag never reaches the database query.

What it controls is narrower: whether a single **service** has to satisfy every
criterion before its paths are offered. It defaults to `false`, so by default a
device type is returned because it carries all criteria *somewhere*, and its
`ServicePathOptions` may come from services that each matched only one criterion.
A caller that needs one service to serve the whole request has to set the flag.

## An aspect criterion already covers the node's subtree

The filter expands it:

```go
filter[DeviceTypeCriteriaBson.AspectId] = bson.M{"$in": append(node.DescendentIds, node.Id)}
```

Passing descendants as additional criteria therefore ANDs a parent with its own
child and matches nothing.

An aspect id that is **not** found as an aspect node does not fail the request. It
logs

```
WARNING: filterDeviceTypeIdsByFilterCriteria() aspect id not found as aspect-node <id>
```

and falls back to an exact match on that id, which matches nothing. A stale or
mistyped aspect id is thus indistinguishable from an empty platform at the API.
Note the message names the v1 function even when the v2 path produced it, so
grepping logs by function name misleads about which endpoint was called.

## An empty criteria list is not an empty filter

```go
if len(query) == 0 {
	query = append(query, model.FilterCriteria{})
}
```

One empty criterion constrains nothing, so the `Distinct` runs unfiltered and
**every** device type comes back. A caller is better off refusing to send an
empty list than discovering that as a listing of everything.

## Why this is written down

None of the four is a compile error, and from outside they look alike: an answer
that is empty, or one far larger than expected. Two of them — the unexpanded
descendant and the unknown aspect id — produce exactly the symptom a caller will
read as "the platform has no such device", which sends the search for a cause to
the wrong side of the API.
