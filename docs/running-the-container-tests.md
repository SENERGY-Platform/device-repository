# Running the container tests

## Applies when

Running or changing the tests that start containers: `lib/tests` and its
subpackages, and `lib/database/mongo`. Written 2026-11-12 after the suite was
moved from fixed sleeps to readiness signals (SNRGY-5135).

**Not this if**: a test fails with an assertion on data. This note is about the
environment the tests run in, not about what they check.

## The environment is ready when its constructor returns

`docker.NewEnv` and the other helpers in `lib/tests/docker` return only once
every container answers, so a test continues right after them without sleeping:

- **mongo** waits for its log line and port.
- **kafka** waits for `INFO Awaiting socket connections on` and then for a broker
  connection. `docker.InitTopics` creates the topics and returns once the leader
  of every partition answers an offset request. A metadata answer alone is not
  enough: the broker names a leader before its replica manager holds the
  partition, and the first produce then fails with `Unknown Topic Or Partition`.
- **permissions-v2** waits for `GET /health` with status 200.
  `wait.ForListeningPort` was not enough: it dials the host port, which the
  docker proxy accepts before anything listens in the container, and the first
  request to the container ip failed with `connection refused`.
- **the api** of the repository is bound synchronously by `api.Start`, so it
  accepts connections as soon as `lib.Start` returns.

Asynchronous effects inside a test (sync loop, mirror pull, kafka message) are
awaited with `docker.WaitFor(t, timeout, cond)`, which polls the condition the
test checks next. A few fixed sleeps remain on purpose, each commented: checks
that something does *not* happen within a window (mirror pull, criteria
recompute), and backoff or lock periods the test is about.

## Three things that look like flakes

**Sync jobs are selected by whole seconds.** `FetchSyncJobs` in
`lib/database/mongo/sync.go` takes a job only if its `sync_unix_timestamp` is
older than the cutoff in Unix seconds. A document written in the same second as
the sync run is skipped, and the next run comes after `sync_interval`, 10 minutes
by default. A test that writes sync jobs and then starts `lib.Start` calls
`docker.WaitForNextSecond()` in between (see `perm_sync_test.go`).

**kafka-go pools connections per broker address.** `kafka.DefaultTransport` keeps
connections and metadata for an address, and the next kafka container of the same
process often gets the same container ip. The repository's writers produce
without acks, so a produce into a stale pooled connection is lost without an
error. `docker.Kafka` resets the transport before it starts a broker and after it
terminates one. This would break tests that run kafka containers in parallel
within one process (`t.Parallel`); none do.

**Rootless docker sometimes cannot publish a port.** Under rootless docker the
daemon picks a host port without checking whether it is taken in its network
namespace, and the start fails with `driver failed programming external
connectivity ... address already in use`. testcontainers publishes every
`EXPOSE` port of an image even without `ExposedPorts`, so the publication cannot
be switched off. `startContainer` retries such a start up to five times. Rootful
docker in CI is not known to hit this.

## Packages run in parallel

Every test env takes a free port for the api (`GetFreePort`), and kafka publishes
no port the tests use, so packages no longer collide. CI runs
`go test -p 3 -short ./...` as the first job of the prod workflow and only tags
and builds when it passes. There are no automatic retries; rerun a failed job
with `gh run rerun <run-id> --failed`.

Containers are capped so three envs fit on a runner: kafka gets a 256 MiB heap,
mongo a 0.25 GB WiredTiger cache.

## Running the suite on a WSL machine with rootless docker

Container ips are routable only inside the network namespace of the rootless
daemon's child process. Kafka advertises its container ip, so every container is
reached by ip and the whole suite runs there. The `go` on such machines is often
a snap wrapper, which fails inside the namespace; call the real binary:

```
nsenter -U --preserve-credentials -t $(cat $XDG_RUNTIME_DIR/dockerd-rootless/child_pid) -n \
  /snap/go/current/bin/go test -p 2 -count=1 -short ./...
```

The whole suite took 7 to 8 minutes this way with `-p 2` on 2026-11-12. CI took
19 minutes for a green run before the change.
