# The test suite needs -p 1

## Applies when

Running the whole suite locally with `go test ./...`. Observed once, in a run
where `lib/tests/repo_legacy` failed while the same package passed on its own
minutes earlier.

**Not this if**: the package reports a failing assertion. This one fails with no
`--- FAIL` line for any test at all — the package is marked FAIL and the only
clue is a log line. A real failure names the test.

## The symptom

```
{"level":"ERROR","msg":"fatal error while starting api","error":"listen tcp :8080: bind: address already in use"}
```

`go test` runs packages in parallel by default, each test env starts the api on
the configured port, and two packages that overlap in time collide. Which
package loses depends on timing, so the same run is green the next time and the
failure looks flaky rather than structural.

## Run it sequentially

```
go test -count=1 -p 1 ./...
```

`-p 1` serialises the packages and the collision cannot happen. `-count=1`
disables the test cache, which matters because a cached `ok` from a run before
your change is not a verification.

The full suite takes roughly seventeen minutes that way — worth starting in the
background rather than waiting on it.

## `-p 1` only covers one invocation

It serialises the packages *within* a run. Two `go test` processes started
independently still overlap, and then both fail the same way — which is easy to
walk into precisely because the run is long enough to forget it is still going.
Both results are worthless in that case, not just one: each run's containers take
ports the other wanted, so a green package means nothing either. Kill both and
start one, rather than trying to read the two logs.

Testcontainers' ryuk reaper cleans up the containers of a killed run on its own,
so there is nothing to tidy by hand.

## The real fix, not done

Each test env should take a free port instead of the one from `config.json`. Until
then a parallel run stays a coin flip, and every red suite costs a look at the
log to decide whether it was real.
