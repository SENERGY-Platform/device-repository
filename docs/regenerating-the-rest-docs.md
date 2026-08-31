# Regenerating the REST docs rewrites a file you did not change

## Applies when

Adding or changing an endpoint and bringing `docs/devicerepository_swagger.*`
and `docs/devicerepository_docs.go` back in sync. Observed twice in one session.

**Not this if**: the question is `docs/asyncapi.json`. That file has a separate,
partly manual pipeline — see [asyncapi-is-half-generated.md](asyncapi-is-half-generated.md).
Nothing here applies to it, and a new kafka topic does not reach it by running
`go generate`.

## Run it from lib/api/doc_gen

The `go:generate` lines sit in `lib/api/doc_gen/gen.go`, so `go generate ./...`
from the repository root does reach them, but running it from that directory is
the smaller blast radius:

```
cd lib/api/doc_gen && go generate ./...
```

Two steps run: `go run gen.go`, which writes
`lib/api/generated_permissions.go`, and `swag init`, which writes the three
files in `docs/`.

## generated_permissions.go comes back reordered

The first step rewrites `lib/api/generated_permissions.go` with the same
annotation blocks in a different order — a diff of several hundred lines with no
change in content. Sampling the diff shows the same godoc blocks moved around,
and reverting the file afterwards leaves the regenerated swagger unchanged, so
the reordering carries nothing.

Revert it before committing, otherwise a one-line API addition ships as a
400-line diff nobody can review:

```
git checkout lib/api/generated_permissions.go
```

Do the revert **after** `swag init` has run, not before — the swagger step reads
that file.

## Swagger comments are the source

`swag init` reads the godoc blocks above the endpoint methods, not the routes. A
route changed in `router.HandleFunc` without changing the matching `@Router`
line produces a spec that documents a path the service does not serve, and
nothing fails. Change both, then regenerate, then grep the output for the old
path to confirm it is gone.

The committed spec is expected to match the code; CI regenerates and fails on a
diff. A spec that claims an outdated state is worse than none, because both a
human and an agent will trust it.
