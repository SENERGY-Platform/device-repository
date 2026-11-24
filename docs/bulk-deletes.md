# Bulk deletes check everything first and delete in one pass

## Applies when

Changing or calling `DELETE /<resource>` with a JSON array of ids, the
`Delete<Resources>` controller methods (`lib/controller/batchdelete.go` and the
resource files), or `removeMany` in `lib/database/mongo/sync.go`. Holds for the 14
resources that have it: aspect-classes, aspects, characteristics, concepts,
device-classes, device-groups, device-types, devices, functions, graphs, hubs,
locations, protocols, things. Written 2026-11-24 for v3.13.0 (SNRGY-5174).

**Not this if**: the request deletes one id through `DELETE /<resource>/{id}`.
The single deletes are unchanged and are what the bulk path has to match. Not
`/local-devices` or `/users` either, they have no bulk variant.

## The order inside one request

1. Ids are deduplicated; more than 10000 ids answer `400` before anything else
   (`MaxDeleteManyIds` in `lib/api/helper.go`). The body is read up to 8 MiB.
2. Rights, for all ids:
   - Resources only admins may delete answer like their single delete: `401`
     `token is not an admin`, for device-types `403`.
   - Resources with an Administrate right (devices, device-groups, graphs, hubs,
     locations, things) load the existing ids and ask permissions-v2 once with
     `ListComputedPermissionsContext`. Any id without Administrate answers `403`
     `access denied: <ids>` and nothing is deleted. Unknown ids are skipped.
   - Admin tokens skip that request. permissions-v2 answers an id without a
     rights entry with all permissions `false`, also for admins, while the
     single delete lets an admin through.
3. The preconditions of the single delete run for every id (the in-use checks,
   generated device groups). The first violation answers with its code and
   `unable to delete <id>: ...`; nothing is deleted.
4. `removeMany` marks the found documents with one `UpdateMany`
   (`sync_delete`), calls the batch sync handler once with all of them, then
   removes exactly the marked ids with one `DeleteMany` in a fresh context.

Up to the mark the request is all or nothing. After it, it behaves like the
single delete: a failing sync handler is logged, the request answers `200`, and
the documents stay marked until the `Retry*Sync` loop removes them one by one
with the single handler.

## What the batch sync handler bundles

- Rights: one `RemoveResourcesContext` per topic (permissions-v2 v1.2.0 or later,
  `DELETE /manage/{topic}`).
- Kafka: one message per id, all in one `WriteMessages` call; key and payload are
  the same as in the single delete.
- Graphs: `removeResourcesFromGraphs` lists the graphs of all ids with one query
  and writes each affected graph once. A graph write is a mongo update, a rights
  call and a kafka message; in v3.13.0 it happened once per id and graph, and
  deleting 354 things took 29 s. `TestDeleteThingsWritesEachGraphOnce` pins it.
- Per id, unchanged: ending thing links, resetting hubs and removing the
  generated device group of a device, aspect nodes of an aspect. A batch of many
  devices is therefore still a loop in the handler.

## Known limit

The in-use checks read the database as it is, not as it will be after the other
ids of the same request are gone. Deleting a concept and a characteristic it uses
in one request answers `400`; two requests in the right order work.

`DELETE /devices` existed before as a loop over single deletes that stopped at
the first error, leaving the ids before it deleted. It now uses this path: one id
without rights rejects the whole request.
