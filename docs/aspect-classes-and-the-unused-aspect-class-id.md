# aspect-classes exist; aspect_class_id does not work yet

## Applies when

Reading or writing aspect-classes, or reading `aspect_class_id` off an aspect or
an aspect-node. Describes the state as of 2026-08-31, when the resource was
added — the second half is a deliberate gap, not a bug report, so expect it to
change.

**Not this if**: the question is about aspects. An aspect is a tree with
sub-aspects and derived aspect-nodes; an aspect-class is flat and has none of
that. Nothing about aspect-node generation, leaf validation or the
`allow_none_leaf_aspect_nodes_in_device_types` option applies here.

## The resource

`models.AspectClass` is `{id, name}` and nothing else. Ids follow
`urn:infai:ses:aspect-class:`; `GenerateId()` fills one in when the request has
no id, so `POST` without an id works.

| | |
|---|---|
| `GET /aspect-classes` | paginated: `limit` (100), `offset`, `search`, `sort` (`name.asc`), `ids`; count in `X-Total-Count` |
| `GET /aspect-classes/{id}` | single, 404 when unknown |
| `POST /aspect-classes` | create, **admin only** |
| `PUT /aspect-classes/{id}` | replace, **admin only**, id in body must equal the path |
| `PUT /aspect-classes?dry-run=true` | validate without writing |
| `DELETE /aspect-classes/{id}` | **admin only**; `?dry-run=true` validates only |

Validation is minimal on purpose: id present, id carries the urn prefix, name
present. A non-admin token gets 401, not 403.

There is no `/v2/aspect-classes`. The `/v2` prefix on the sibling resources marks
the paginated list that replaced an older unpaginated one; a resource with no
history to keep does not need the split.

Writes publish onto a kafka topic of their own — `aspect_class_topic`, default
`aspect-classes` — and take part in `SyncRetry` like the other resources. The
collection is `mongo_aspect_class_collection`, default `aspectclass`.

## aspect_class_id is carried but never evaluated

The models release that brought `AspectClass` also put `aspect_class_id` on
`models.Aspect` **and** on `models.AspectNode`. Both fields are in the swagger
and in every response. This service does not interpret either of them:

- On an aspect the value is stored and returned as sent. It is never checked
  against an existing aspect-class, so a typo, or a reference to a class that was
  deleted meanwhile, survives a write.
- Deleting an aspect-class is not refused when aspects still reference it.
  `ValidateAspectClassDelete` in `lib/controller/aspectclasses.go` returns 200
  unconditionally and says so in a comment — that is where the usage check goes.
- On an aspect-node the field is **always empty**, whatever the aspect carries.
  `CreateAspectNodes` in `lib/controller/aspectnodes.go` builds the node field by
  field and does not copy it.

The third one is the one that misleads. Aspect-nodes are what most consumers
read, so a client that takes `aspect_class_id` from a node sees an empty string
and concludes the classification is unset everywhere — rather than
unimplemented. Read it from the aspect until this changes.

## When the check arrives

Three things belong together, and doing one without the others leaves a
half-state that is worse than the current honest gap:

1. copy `AspectClassId` in `CreateAspectNodes`
2. validate the reference in `ValidateAspect`
3. implement `ValidateAspectClassDelete` against the aspects that reference it

Step 1 also needs a decision for stored aspects: nodes are only rebuilt when
their aspect is written, so existing nodes keep the empty field until then. The
same shape of problem as in
[not-every-device-type-read-passes-the-controller.md](not-every-device-type-read-passes-the-controller.md).
