# Writing a startup migration

## Applies when

Adding a migration to `lib/database/mongo/migration.go`, or deciding whether an
existing one is safe to run again. Follows from the four that are there.

**Not this if**: the question is what a *particular* migration does. The aspect
ones are in
[the-aspect-filters-need-the-startup-migrations.md](the-aspect-filters-need-the-startup-migrations.md),
the concept-function one in
[a-concept-owns-a-function-pair.md](a-concept-owns-a-function-pair.md). This
document is the shape they share.

## Converting or creating decides everything else

There are two kinds, and the difference is not stylistic:

- A **converting** migration rewrites a stored shape into a newer one. It can
  recognise the old shape in the data — a field that only the old writer
  produced — so it detects its own work, skips when there is nothing to do, and
  survives any number of runs without a marker.
- A **creating** migration produces resources that did not exist. It cannot
  recognise its own result: a second run makes a second copy, and nothing in the
  data distinguishes what the migration made from what a user made afterwards.

Only the second kind needs a marker, and it is the deciding question because
everything below follows from it.

## A creating migration records itself, afterwards

`migrationHasRun` / `setMigrationHasRun` in `migrationstate.go` keep one document
per migration name in the collection named by
`mongo_migration_state_collection`. Two things about the order:

- The record is written **after** the migration succeeded, never before. A run
  that dies halfway therefore repeats on the next start — which is the safe
  direction, but it means such a migration has to survive **seeing its own
  partial result**. Write it so that a half-finished state converges rather than
  doubles.
- The marker is checked first and the migration returns early, so a marked
  migration costs one query.

A converting migration does not get a marker. Its detection query is the marker,
and it stays correct if someone restores an old backup.

## Publish only what a consumer has never seen

The default is **not** to publish. The converting migrations here replace whole
documents while keeping their sync info precisely so that nothing goes onto
kafka: they change a representation that no consumer reads yet, and an event for
that is noise the consumer cannot act on.

A creating migration is the opposite case. `runConceptFunctionsMigration` writes
functions nobody has seen, and a resource that reaches no consumer is invisible
to the whole semantic layer, so it publishes.

Ask which one applies before copying either.

## Reaching the controller

`lib/database/mongo` may not import `lib/controller`, and migrations regularly
need something that lives there — a normalisation, a derivation, a publisher.
The way through is the `MigrationMethods` interface, which the controller
satisfies and `lib.Start` hands in.

It has four methods now, three of which exist only because a package-level
helper in the controller is needed below it. That is a smell worth naming: the
alternative is to move those helpers into a small package both sides may import,
the way `lib/idmodifier` is cut. Adding a fifth method is the moment to do it
rather than the moment to add a fifth method.

## Startup cost is a real budget

Every migration runs on every start of every replica, marked or not — at minimum
its detection query. `runDeviceGroupMigration` scans all devices, and
`runGeneratedDeviceGroupCriteriaMigration` recomputes each generated group and
compares, because it has no marker and no detectable old shape. That is a
deliberate trade: exact idempotence without new infrastructure, paid for with one
device-type read per generated group at every start.

If a new migration cannot detect its old shape cheaply and is not a creating one,
that trade is the pattern to copy — but say so in its doc comment, because the
cost is invisible from the call site.

## Migrations are skipped entirely on a mirror

`lib.Start` guards the whole set with `conf.RunStartupMigrations &&
!conf.AsMgwMirror`. A mirror rewrites its local database from its source on the
next pull and normalises on the way in, so migrating it would convert data that
is about to be replaced. A migration that a mirror would need is a sign the
normalisation on the pull path is missing instead.
