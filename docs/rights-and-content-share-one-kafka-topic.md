# Rights and content commands share one kafka topic

## Applies when

Consuming or producing this service's kafka topics, or adding a topic for a new
resource type. Read against `permissions-v2 v0.0.45`.

**Not this if**: the resource type keeps its permissions outside permissions-v2 —
then only the repository's own producer writes to the topic and the ordering
argument below is moot.

## Two producers, one topic

A resource type whose permissions live in permissions-v2 and whose data lives
here produces two kinds of message onto **the same** topic, from two different
processes:

- **permissions-v2** writes `{"command": "RIGHTS", "id": …, "rights": {…}}` with
  key `<id>/rights` — but **only if** the topic was registered with a non-empty
  `PublishToKafkaTopic`.
- **the repository** writes `{"command": "PUT"|"DELETE", "id": …, "<resource>": {…}}`
  with key `<id>`, from its producer in `lib/controller/publisher`.

**Per-resource ordering holds** because both sides use the same
`KeySeparationBalancer`: it splits the key at the first `/` and hashes only the
part in front, so `<id>` and `<id>/rights` land in the same partition. A consumer
therefore sees a resource's rights change and its content change in the order
they happened.

## Two traps when adding a topic for a new resource type

- **Registering it in permissions-v2 without `PublishToKafkaTopic`** gives
  permission checks and no events. Nothing fails — consumers are simply left
  polling the REST API, and it is easy to miss that this is why.
- **The publish runs inside the mongo sync handler.** A failed publish leaves the
  entry as `SyncTodo` and the log says "will be retried later" — which is only
  true if that resource type is wired into `Controller.Sync`. An implemented but
  never-called `Retry<Resource>Sync` is indistinguishable from a working one
  until kafka is actually down.

## Where this came up

Adding events for a resource type that had none. The topic was registered for
permission checks only, so nothing reached kafka and the planned consumer was
about to poll REST instead. Wiring it up surfaced the second trap: the retry
function existed, was covered by the database interface, and had never been
called.

The kafka contract itself is published as AsyncAPI in `docs/asyncapi.json`.
