# The test suite needs -p 1

## Applies when

Running the whole suite locally with `go test ./...`. Observed once, in a run
where `lib/tests/repo_legacy` failed while the same package passed on its own
minutes earlier.

**Not this if**: the package reports a failing assertion. This one fails with no
`--- FAIL` line for any test at all — the package is marked FAIL and the only
clue is a log line. A real failure names the test.

The same log line also appears with `-p 1` when a process outside the test run
holds the port — see "Another process holds the port" below. The cause is the
same, the remedy is not: serialising the packages does nothing against it.

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

## Another process holds the port

A sequential run can still fail this way when something else on the machine
listens on 8080, for instance a container of another project published on
`127.0.0.1:8080`. Only the packages whose tests start the api on the configured
port fail; the rest of the suite stays green, which makes it look like a flaky
subset rather than a blocked port. Observed once, on 2026-11-02, with
`lib/tests` and `lib/tests/repo_legacy` failing and everything else passing.

Check before rerunning:

```
ss -ltnp | grep ':8080\b'
```

If the port is taken and the other process has to keep running, move the test
api instead. Every field of the configuration can be overridden by an
environment variable named after it, `ServerPort` becoming `SERVER_PORT`:

```
SERVER_PORT=18080 go test -count=1 -p 1 ./lib/tests/ ./lib/tests/repo_legacy/
```

## The real fix, not done

Each test env should take a free port instead of the one from `config.json`. Until
then a parallel run stays a coin flip, and every red suite costs a look at the
log to decide whether it was real.
