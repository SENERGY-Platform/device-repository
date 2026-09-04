# The aspect filters need the startup migrations

## Applies when

Deploying a version from 2026-09-03 onwards onto a database written by an older
one, or deciding what `run_startup_migrations` may be set to. Follows from
`lib/database/mongo/migration.go` and the shapes it converts.

**Not this if**: the instance runs as an mgw mirror. `lib.Start` skips the
migrations entirely there (`conf.RunStartupMigrations && !conf.AsMgwMirror`),
which is correct, because the mirror rewrites its whole local database from the
source on the next pull and normalises what it writes on the way in. An empty
aspect filter on a mirror is a different problem.

## Two of the three are required, not convenience

`RunStartupMigrations` reads like a switch for optional housekeeping. For two of
the three migrations it is not: the code after this change reads a stored shape
that only the migration produces, and a database left unmigrated answers **empty
rather than failing**.

- **`runDeviceTypeCriteriaAspectIdsMigration`** rebuilds the derived
  `<device-type collection>_criteria` documents. They used to hold one row per
  aspect in a field `aspectid`; they now hold one row per content variable with
  `aspectids`. Every aspect filter in the service queries `aspectids`, so
  without this every device-type listing, every selectables request and every
  aspect `used-in` answer narrowed by an aspect comes back empty. It detects the
  old shape by the presence of `aspectid` and, having found it, drops the stale
  index and rebuilds the collection from the stored device-types.
- **`runDeviceGroupCriteriaAspectIdsMigration`** fills `aspect_ids` on the
  stored device-group criteria from the deprecated `aspect_id`. The device-group
  listing matches on `criteria.aspectids`, so an unmigrated group is invisible to
  every criteria query naming an aspect. It leaves `criteria_short` alone on
  purpose — that field is the model's own rendering and no longer the search key.
- **`runGeneratedDeviceGroupCriteriaMigration`** is the convergence one and the
  only one that can be skipped without breaking an existing caller. It rebuilds
  the criteria of the **auto generated** groups so they gain the criterion over a
  content variable's whole aspect list. Without it, only queries naming several
  aspects at once miss, which is a capability no caller had before.

The two required ones are cheap to recognise afterwards: a filter that used to
return devices returns none, and nothing is logged, because an aspect that
matches nothing is a legitimate answer.

## Why the first one needs the controller

The migration rebuilds the criteria from the stored device-types, and a
device-type written before `ContentVariable.AspectIds` carries only the
deprecated `AspectId`. Deriving the criteria from it without normalising first
would silently produce rows with an empty aspect list — the exact data loss
[not-every-device-type-read-passes-the-controller.md](not-every-device-type-read-passes-the-controller.md)
describes.

`lib/database/mongo` may not import `lib/controller`, so the normalisation is
handed in through the interface `MigrationMethods`, which the controller
satisfies. It has grown to four methods, three of which exist only because
package-level helpers in the controller are needed below it.

## They cost a scan on every start, by design

None of the three writes a marker. The first two skip cheaply — they detect the
old shape with a single query and return when it is absent. The third cannot: it
recomputes each generated group's criteria and compares, writing only what
differs. That is one device-type read per generated group on every start, in the
order of what `runDeviceGroupMigration` above it already spends, and it makes the
migration exactly idempotent without new infrastructure.

All three replace the whole document while keeping its sync info, so none of them
publishes to kafka. That is deliberate: they convert a representation, and no
consumer reads the new one yet.

## What they do not touch

Manually created device-groups keep their hand-written criteria; only the
generated ones are rebuilt. Note that this is a smaller guarantee than it sounds:
`setDeviceTypeSyncHandler` recomputes the criteria of **every** group holding a
device of a written device-type, generated or not, so hand-written criteria do
not survive the next write of one of their device-types either way. The migration
simply does not add a second occasion for that.
