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
bson.M{DeviceTypeCriteriaBson.AspectIds[0]: bson.M{"$in": append(node.DescendentIds, node.Id)}}
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

## `aspect_ids` inside one criterion is an AND over one content variable

`aspect_id` is deprecated and is an alias for `aspect_ids` with a single entry;
the controller folds it into the list on the way in, so nothing behind that reads
it. Several entries in one criterion are **ANDed**, and they are ANDed on the
**content variable**: `{"function_id": f, "aspect_ids": [a, b]}` asks for a
variable that carries both aspects, not for a device type that carries them in
two different places. Each entry still covers its own subtree, so the sentence
above applies per entry.

A path option answers with the aspects the query matched, in
`ServicePathOption.aspect_nodes`. `aspect_node` is deprecated and is the alias for a
single element list, so it holds the node with the alphabetically first id — the
only one an older client would ever have seen. There is **one option per path**,
not one per aspect: a content variable with two matched aspects is a single option
naming both. `Configurable` reads the same way, with the same two fields, and there
is one configurable per candidate rather than one per aspect.

That is a different AND from the one over the criteria list. A device type whose
one variable is `a` and whose other is `b` is returned by
`[{aspects: [a]}, {aspects: [b]}]` and is **not** returned by
`[{aspects: [a, b]}]`.

The **device-groups** listing spells the same request differently, because it
matches the stored `criteria` of a group rather than the criteria collection. The
reading is the same: an `$elemMatch` requires **one** stored criterion to carry all
the queried aspects, each of them again covering its own subtree. `criteria_short`
is not what answers this — it stays the rendering
`models.DeviceGroup.SetShortCriteria()` writes, for whoever reads it.

An **unset field of a criterion is not a filter**, on both sides. That is worth
saying because the device-groups listing used to disagree: it compared whole
`criteria_short` strings, so an empty `function_id`, `device_class_id` or aspect was
matched literally and found only stored criteria that were empty there too. A query
`{"function_id": f, "device_class_id": dc}` therefore used to return only groups
whose matching criterion carried **no** aspect, and one with an empty `function_id`
returned nothing at all, because no stored criterion has one.

## A generated device-group criterion is written twice over

`getDeviceGroupCriteriaOfDevice` emits, per content variable and interaction, both
the whole aspect list of that variable **and** one criterion per single aspect,
plus one per ancestor of those. The two say different things and both are load
bearing:

- the list records that **one** variable carries all of those aspects, which is
  exactly what a query over several aspects asks for
- the single ones carry the intersection. `GetDeviceGroupCriteria` intersects the
  criteria of the devices of a group by `Short()`, so a group of a device with
  `[a b]` and one with `[a]` keeps `a` only because `a` also stands alone. Drop the
  single ones and such a group loses the aspect entirely.

For a content variable with one aspect the two collapse into the same criterion, so
none of this is visible until a variable carries more than one.

Stored groups written before this carry only the single ones.
`runGeneratedDeviceGroupCriteriaMigration` rebuilds the criteria of the **auto
generated** groups on startup, and leaves the manually created ones alone: their
criteria may be hand written, and they converge anyway on the next write of one of
their device-types, because `setDeviceTypeSyncHandler` recomputes the criteria of
every group holding an affected device.

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

None of these is a compile error, and from outside they look alike: an answer
that is empty, or one far larger than expected. Two of them — the unexpanded
descendant and the unknown aspect id — produce exactly the symptom a caller will
read as "the platform has no such device", which sends the search for a cause to
the wrong side of the API.
