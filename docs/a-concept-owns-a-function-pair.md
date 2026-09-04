# A concept owns a Get and a Set function

## Applies when

Creating or renaming a concept through `POST`/`PUT /concepts`, or deploying onto
a database whose concepts predate this. Follows from
`lib/controller/conceptfunctions.go` and `runConceptFunctionsMigration` in
`lib/database/mongo/migration.go`.

**Not this if**: the concept arrives through the import endpoint or through the
mgw mirror. Neither generates anything, on purpose — an export carries its own
functions and a mirror copies the state of its source, so generating there would
produce a second pair. Only `Controller.SetConcept` generates. That the two paths
write concepts straight past the controller is the same shape as
[not-every-device-type-read-passes-the-controller.md](not-every-device-type-read-passes-the-controller.md),
here as a reason to leave them alone rather than to follow them.

## The convention

A concept named `X` owns two functions:

| Function type | Name and display name |
|---|---|
| measuring | `Get-X` |
| controlling | `Set-X` |

Both `name` and `display_name` carry it. `model.ConceptFunctionName` is the only
place that spells it, and both the write path and the migration read it from
there.

## Created once, renamed while it is still ours

- **On create**, both functions are written with a fresh id.
- **On rename**, a function is renamed only while its name still is the one the
  convention gave it for the **old** concept name. A name someone else set is
  theirs and stays. The display name follows the name while it is empty or equal
  to it — an empty display name is not a choice, so it is filled rather than left
  behind.
- **On any other write**, nothing happens. In particular a concept whose function
  was deleted does **not** get it back on the next write: recreating it would undo
  the deletion rather than honour it.

So the pair can drift from the convention, and that is intended. What cannot
happen is the service overwriting a name a user typed.

## `DELETE /concepts/{id}` no longer works

`ValidateConceptDelete` refuses to delete a concept while any function references
it, and every concept now has two by construction. Deleting a concept therefore
means deleting both of its functions first — which itself only works while no
device-type uses them.

This is a deliberate decision, not an oversight: the alternative was for the
delete to remove the generated functions along with the concept, which makes a
delete destroy resources the caller did not name. Whoever wants the endpoint back
has to decide that trade, not repair a bug.

## The migration runs once, and says so in the database

`runConceptFunctionsMigration` establishes the pair for the concepts that existed
before. Unlike the migrations that convert a stored shape, it **creates**
resources and so cannot recognise its own result: a second run would produce a
second pair. It is guarded by a record in the `migration_state` collection
(`mongo_migration_state_collection`), written **after** it succeeds — a run that
dies halfway repeats on the next start, so it has to survive seeing its own
partial result.

Per concept and function type, what it does follows what it finds:

- **none** — create the function.
- **exactly one** — rename it if its name is not the convention's. This is the one
  place a name a user chose *is* overwritten, once: a concept can carry only one
  obvious function per type, and this is the run that decides which.
- **more than one** — create the function, and append `-deprecated` to the name
  and display name of every existing one. Picking one of several would be a
  guess, and deleting them would break the device-types that use them. **The ids
  are untouched**, so no device-type loses its reference; only the names say the
  functions are superseded.

It publishes what it writes. The other startup migrations deliberately do not —
they convert a representation no consumer reads yet — but this one produces
functions nobody has seen, and a function that reaches no consumer is invisible
to the whole semantic layer. The publisher reaches it through
`MigrationMethods.PublishFunction`, because `lib/database/mongo` may not import
`lib/controller`.

## Two concepts may share a name

`ValidateConcept` does not enforce unique names, so two concepts called `X` each
own a `Get-X`. Nothing breaks — functions are referenced by id — but a listing
shows the name twice with no way to tell them apart.
