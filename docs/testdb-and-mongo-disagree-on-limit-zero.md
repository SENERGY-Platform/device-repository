# The test double and mongo disagree on `Limit: 0`

## Applies when

Calling one of the paginated `List*` methods of `database.Database` from inside
the service — a controller, a migration, a sync handler — rather than from an
HTTP handler. Found on `ListFunctions`; the arithmetic is the same in the other
`lib/database/testdb` list methods.

**Not this if**: the call comes from an API handler. Those fill `Limit` from the
query string and default it to 100 before the options reach the database, so the
zero value never arrives. This is only about the paths that build a
`…ListOptions` in code.

## The same options mean opposite things in the two implementations

`Limit: 0` is not "no limit" in both:

- **mongo** passes it to `options.Find().SetLimit(0)`, and the driver reads zero
  as *unbounded*. Every match comes back.
- **`lib/database/testdb`** slices the result: `result[offset:min(len(result),
  int(offset+limit))]`. With `offset` and `limit` both zero that is
  `result[0:0]` — the **empty list**.

So an internal caller that leaves `Limit` unset gets everything against a real
database and nothing against the double. The tests that run on the double are
`lib/tests/semantic_legacy` through `NewPartialMockEnv`, which is where this
surfaces: a green mongo-backed test and an empty result in the mocked one, from
the same code.

Nothing errors in either direction. An empty list is a legitimate answer, so the
symptom is a feature that silently does nothing in one half of the suite.

## What to do

Pass an explicit cap and say in a comment why that number. Both internal callers
of `ListFunctions` do:

```go
existing, _, err := this.ListFunctions(ctx, model.FunctionListOptions{
	ConceptIds: []string{concept.Id},
	RdfType:    rdfType,
	Limit:      conceptFunctionsMigrationLimit,
	SortBy:     "id.asc",
})
```

A cap has its own cost — it truncates silently if the real count ever exceeds it
— so it belongs with a sentence about why the real count cannot. Where that
sentence is not writable, the honest shape is a batch loop over `util.IterBatch`,
as `ListDevices` is consumed in `setDeviceTypeSyncHandler`.

## Why not fix the double instead

Making `testdb` read zero as unbounded would match mongo and remove the trap. It
is also a change to the semantics every existing test on the double runs under,
for the benefit of callers that should be passing a limit anyway. Worth doing
deliberately, with the suite green before and after — not as a side effect of
the next feature that trips over it.
