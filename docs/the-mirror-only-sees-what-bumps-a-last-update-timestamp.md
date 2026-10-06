# The mirror only sees what bumps a last-update timestamp

## Applies when

Running with `AS_MGW_MIRROR=true`, or adding or changing a `Set*`/`Remove*`
method in `lib/database/mongo` that the mirror source serves. Read from the code
of `lib/mgwmirror/source.go` and confirmed by `TestMirror` in
`lib/tests/mirror_test.go`.

**Not this if**: the mirror keeps a deleted entry although the source's
timestamp did move. Then the pull ran and chose not to delete. That happens on
purpose in two cases, and the log tells them apart. An aborted listing or a
failed write logs `error while listing ...` or `error while setting ...` for the
collection, and nothing of that collection is removed in that pull. A
confirmation read that answered neither 404 nor 403 logs
`unable to confirm removal in mgw mirror source --> keep` with the status.

## How the mirror decides to pull a collection

The mirror asks the source for its last-update timestamps (`GetLastUpdateTimestamps`)
and pulls a collection only if the source's timestamp is newer than its own. The
first pull after start skips this check and pulls everything.

The source writes the timestamp in the database layer, inside every `Set*` and
`Remove*`. Resource types without owners (protocols, aspects, characteristics,
concepts, functions, device classes, device types, locations) write it for the
empty user. Devices and hubs write it for the **owner**, device-groups for the
**sync user**. The source answers the mirror user's own timestamps together with
the ownerless ones.

Besides the interval, two things start such a timestamp-checked pull: a write
the mirror forwards to the source, and a `GET` the mirror answers with 404. For
the latter the middleware in `lib/api/util/mgwmirror.go` holds the 404 back,
pulls, and answers the request again from the database, so an entry created in
the source since the last pull is found on the first read. If the pull takes
longer than `MGW_MIRROR_MISS_PULL_TIMEOUT` (default `10s`, for example because
the source is unreachable), the 404 is returned and the pull finishes in the
background. All pulls of a mirror run one after another through
`mgwmirror.Puller`. A caller waits for a pull that starts after its call, since
a running pull may have passed the collection already; all callers arriving
during a running pull share the one following pull, so a hanging source does
not queue up a pull per request. A found entry starts no pull, even if the source changed or removed it, and
neither does an empty listing.

A read that is still not found after its pull (or whose pull failed or timed
out) may pull again only after a backoff, so a client polling a missing entry
does not send a request to the source each time. The backoff is kept per
request URI including the query, starts at `MGW_MIRROR_MISS_PULL_BACKOFF`
(default `10s`) and doubles with every further miss up to
`MGW_MIRROR_MISS_PULL_MAX_BACKOFF` (default `5m`). During the backoff the 404 is
answered from the database alone. An entry created in the source in that time
becomes visible with the next pull of any kind: the interval, a forwarded write,
or the not found read of another request. A found response resets the backoff
of its request, as does a pause longer than the maximum after the backoff ended.

It follows that a change the mirror should see has to move one of those
timestamps, also for a not found read. Two things do not:

- **A method that forgets it.** `RemoveDeviceGroup` did not set a timestamp
  until the change that introduced this document. Against such a source, deleted
  device-groups (including the generated group of a deleted device) never reach
  the mirror, because nothing starts a device-group pull. A mirror restart
  catches up, since the first pull checks nothing. A new `Remove*` or `Set*`
  needs the timestamp as well, or the mirror misses it the same way.
- **A permission change.** Sharing or unsharing a resource in permissions-v2
  touches no timestamp here. The mirror handles an unshare correctly once the
  collection is pulled (see below), but the unshare alone does not start that
  pull.

## How the mirror removes entries

After a collection has been listed and written without error, the mirror lists
its own entries of that collection, all of them, before removing anything, and
takes those the source listing did not contain as candidates. Each candidate is
confirmed with a single read against the source:

| Read answers | Result |
|---|---|
| success | kept. The offset paging of the source listing skipped it, for example because the source changed during the pull |
| 404 | removed, the entry was deleted |
| 403 | removed, the mirror user may no longer see it |
| anything else | kept, logged as a warning |

Aspect-nodes have no single-node removal in the database. If a root has nodes
the source no longer knows, all nodes of that root are removed and the ones the
source listed are written again, which is what the controller does on an aspect
update.

## Where this came up

A mirror that had written every entry its source listed but never removed one.
Entries deleted in the source stayed in the mirror and were served to every
consumer on the gateway. Adding the removal showed the second cause: without the
timestamp in `RemoveDeviceGroup`, the new phase of `TestMirror` still found the
deleted device-groups in the mirror.
