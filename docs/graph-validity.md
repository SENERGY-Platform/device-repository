# Graph validity

## Applies when

Reading or writing graphs whose elements change over time: a meter replaced on a
date, a device moved below another parent, a node taken out of the calculation
from a date. Also when a service that writes graphs (graph-provider) must not
lose that history. Contract DAS-270 V2, written 2026-11-24.

**Not this if**: the question is how a graph is validated in general (weights,
end node), see [writing-graphs.md](writing-graphs.md). Nothing here changes a
graph that has no `valid_from` or `valid_until` anywhere.

## The stored document and its projection

Nodes, edges and attributes of a graph, of its nodes and of its edges carry
`valid_from` and `valid_until` (RFC3339, both optional, omitted when unset). The
validity is half-open, `[valid_from, valid_until)`; a missing bound is unbounded;
`valid_from` is before `valid_until` where both are set. The fields are only
interpreted on graph, node and edge attributes; `models.Attribute` carries them
for every owner but a device or a thing neither writes nor reads them.

The stored document holds every version. The API hands out whole graphs:

| Request | Answer |
|---|---|
| `GET /graphs/{id}`, `GET /graphs` | the projection at now |
| `GET /graphs/{id}?at=<RFC3339>` | the projection at that instant |
| `GET /graphs/{id}/segments?start=&end=` | `[{from, until, graph}]`, see below |
| `GET /graphs/{id}?versions=all` | the stored document with all versions, for tooling and migration; not with `at` |

The projection keeps the nodes, edges and attributes whose validity contains the
instant and strips nothing else, so it is an ordinary graph and old readers see
what they saw before. The validity fields of the kept elements are still there.
An instant at which no node is valid is answered with 200, the graph with its
`id`, `owner`, the attributes valid then and empty `nodes` and `edges`, not 404.
`at` needs a zone; a `+` in an offset has to be url-encoded (`%2B`).

`/segments` cuts the range `[start, end)` at every validity bound inside it,
nodes, edges and attributes included, and answers each piece with its
projection. `start` and `end` are optional: without `start` the first piece has
`from: null`, without `end` the last has `until: null`, without both the answer is
the whole history. Pieces are not merged, so two neighbours may have the same
graph. A range with `end` not after `start` is a 400. This is the cut a consumer
that calculates over a period has to use.

The filters of `GET /graphs` (`device_ids`, `attributes`, `search`) match against
the stored document, so an ended device node still finds its graph. Each graph in
the answer is the projection at now.

## Identity

An element is the same element in two versions by:

- a node: its `id`. Node ids are unique over the whole document, so a node has one
  validity. `resource_type` and `resource_id` never change; a different resource is
  a different node, and a node that left and comes back is a new node with a new id.
- an edge: `from_node_id` and `to_node_id`. The versions of one edge are separate
  edges with their own `id`, which stays unique over the whole document; a new
  version gets a server-generated id. Versions of one edge do not overlap.
- an attribute: its owner (the graph, the node, the edge) and its `key`. Versions
  of one key on one owner do not overlap. An attribute is only valid while its
  owner is; the bounds of the owner cut it. Two attributes with the same key and no
  bounds at all stay legal as they were.

## Validation

`Valid()` holds for the projection of every segment between all the bounds of the
document, plus the identity rules above. The projection of each segment is
checked as a plain graph (weights 0 or 100 per node, exactly one end node that is
neither a device nor a thing, no loop, unique `resource_id`, no duplicate edge,
edges only between nodes that exist then). A 400 names the first invalid instant:

```
invalid at 2025-06-30T00:00:00Z: sum of outgoing edge weights for node m must be 100, got 70
```

Before the first bound the message says `invalid before <first bound>`. A segment
whose projection has no node at all is valid: the graph did not exist yet, or no
longer has content. A graph with no node anywhere is still invalid, as before.
Weights 90/10 until 30.06 and 70/30 from then on are two versions of each of two
edges, and valid; a node that has only the 90 from 30.06 on is a 400 naming that day.

## Writing: change and correction

`PUT /graphs/{id}?at=<RFC3339>&mode=change|correction` takes the **edited
projection at `at`** as its body. The server compares it with the stored
projection at `at` by identity and sets the validities. `at` defaults to now, `mode`
to `correction`. Instants are cut to milliseconds in UTC, as for things.

| | changed live version | added element | removed live version |
|---|---|---|---|
| `correction` | modified in place over its whole span | no bounds, except it ends where a later version of its identity starts and starts where an earlier one ended | deleted, it never was |
| `change` | split at `at`: the old ends, the new runs from `at` to the old one's end | valid from `at` until the next version of its identity, or open | ends at `at` |

A change made in the past therefore holds until the next change of the same
element. An edge is changed when its weight differs; a different attribute value
on an edge is a change of the attribute only and does not split the edge.

**Bounded change.** `until=<RFC3339>`, only with `mode=change`, after `at`, makes the
edit hold for exactly `[at, until)`. Whatever the touched identities have in the
span is replaced; a version that goes on beyond `until` is resumed there, as its own
version with its old content. An added element is valid for `[at, until)` only. A
month that was assigned to the wrong parent is one request: edge to A, body edge to
B, `at=01-01&until=04-01` gives A, B, A. A node has one validity, so a node that is
valid beyond `until` can not be removed for the span (400).

Further rules:

- **A write never touches or drops a version that is not valid at `at`.** That is
  the invariant a graph author relies on. `PUT` without `at` is a correction at now,
  which is what every writer did before: GET, modify, PUT keeps the ended versions
  and the versions in the future that the body did not show. The one exception is a
  graph without any validity on either side, where the body is stored as it is.
- `valid_from` and `valid_until` of the body are ignored for elements that exist
  and a 400 for elements that are added: the edit sets the bounds. Edge ids the
  writer makes up are kept in a correction while the edge has a single version.
- **Added edges are cut to their nodes.** An added edge is valid only where both
  of its nodes are (and inside the added range), and an added node is valid only
  where one of its added outgoing edges is. A graph-provider that adds a device
  below a node that starts on a date `T` with `PUT` without `at` gets the edge
  and the new node from `T`. An added edge between nodes that are never valid
  together is a 400.
- **Attributes of added elements may not carry bounds** either, nor may added
  graph attributes: a 400, like the elements themselves.
- **A writer that changes weights on a graph with history needs `mode=change`.**
  A correction changes the live version in place over its whole span, so
  weights that were redistributed over versions with other weights in the past
  no longer add up there, and the result is a 400 naming the instant. That is the
  contract: a correction says "it was always so".
- A node removed in a correction is deleted from the document; if a version of an
  edge that is not valid at `at` still points to it, the result is invalid and the
  edit is a 400 naming the instant. Use `mode=change`, which ends the node.
- Unknown mode, `until` without `mode=change` and `until` not after `at` are a 400.
- The answer to every `PUT` is the stored graph's projection at **now**, like a
  plain `GET`, whatever `at` was.
- `POST /graphs` and a `PUT` to an id that does not exist store the graph as given.
- Owner and permission checks are unchanged.
- **Writes are compare-and-set.** Every stored graph has a `revision` (the time of
  its last write in nanoseconds, not part of the model, missing in documents from
  before it existed). A write is applied only if the revision is still the one that
  was read; otherwise the edit is applied again to what is stored now, which is
  correct because the body is a diff against the projection at `at`. After five
  lost races the answer is 409. Device and thing deletion use the same path.

## Guided actions: move and end a node

Two endpoints build the edit from the stored projection, so a client does not have to send a whole graph for a
single, common change. Both need write on the graph, answer the graph as it is now, name the first invalid instant in
a 400, and use the compare-and-set loop of `PUT` (409 after five lost races). Both are `mode=change`: history before
`at` is never touched. `at` is optional (now) and is cut to milliseconds in UTC.

`POST /graphs/{id}/nodes/{node_id}/move`, body `{"at", "to_node_id", "until"?}`: the outgoing edges of the node in the
projection at `at` are replaced by one edge to `to_node_id` with weight 100. Without `until` that holds until the next
version of the edge, with `until` (after `at`) for `[at, until)` only and the old edges are resumed there (A, B, A).
Children and everything above stay as they are. 400 if the node or the target is not valid at `at`, if the node
has no outgoing edge (the end node can not be moved), if it is already below the target with weight 100, or if the
result has a loop (a node below its own child).

`POST /graphs/{id}/nodes/{node_id}/end`, body `{"at"}` (an empty body is now): the node and its edges end at `at`, and
the edges that led to it are rerouted to where it led from `at` on, as `Graph.DeleteNode` does for a device delete.
400 if the node is not valid at `at` or the result is invalid (the end node with children).

**Limits.** An action is built from the projection at `at`, which does not show versions that start later. A node
whose outgoing edges (for end also the edges leading to it) have a version starting after `at` (for move: inside
`[at, until)`, or at all without `until`) is therefore refused up front with a 400 naming the instant, for example
`node M has changes from 2020-12-01T00:00:00Z; move it after that or bound the move with until, to end before it`. A
move bounded to end at or before that instant works; otherwise use `PUT` with `mode=change`, which sees all versions.
Rerouting on end divides weights as integers: a rerouted weight is `weight * destination / 100` rounded down, and what
is missing to 100 is added to the first outgoing edge, so a split such as 33/33/34 is not exact. The same holds when a
device is deleted. And the target of a move or the parent of an ended node may end inside the span, which is then a 400
at that instant from the general validation.

**Ending a node does not end the device link of a meter thing.** Whether the meter is taken out of service, replaced or
kept for another place is the caller's decision, made through `POST /things/{id}/device-links/{link_id}/end` or
`/swap`, see [things-and-device-links.md](things-and-device-links.md). A node and a link are separate facts and
the two calls are separate writes.

## Things and devices

`checkGraphThingRule` and the check at link creation compare validities, see
[things-and-device-links.md](things-and-device-links.md): a thing node and a device
that the thing links to may be in one graph only if the thing node, the device node
and the link have no instant in common. A device node is replaced by a meter thing
from a date by ending the device node and starting the thing node on that date.

**Deleting a device or a thing** removes its node and reroutes its edges
(`Graph.DeleteNode`) on the graph as it is now. A graph without validities is
written as that. A graph with validities is written as a `change` at the current
time: the node and its edges end now, the rerouted edges start now, the history
stays and no weight of an old version is touched.

## Kafka

The graph topic carries the stored document, with all versions, as
`{"command": "PUT", "id", "graph"}`, and `{"command": "DELETE", "id"}`. A consumer
that reads the body has to project it itself (`models.Graph.At`, `Segments`). The
consumer in the services next to this repository, graph-provider, only treats the
message as a trigger and does not read the body.
